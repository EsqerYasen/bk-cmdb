/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package remote

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"configcenter/src/common"
	"configcenter/src/storage/dal"
	"configcenter/src/storage/rpc"
	"configcenter/src/storage/types"
)

var _ dal.DB = (*Mongo)(nil)

// Mongo implement dal.DB interface
type Mongo struct {
	RequestID string // 请求ID,可选项
	TxnID     string // 事务ID,uuid
	rpc       *rpc.Client
	getServer types.GetServerFunc
	parent    *Mongo
}

// NewWithDiscover returns new DB
func NewWithDiscover(getServer types.GetServerFunc) (dal.DB, error) {
	servers, err := getServer()
	if err != nil {
		return nil, err
	}

	rpccli, err := rpc.DialHTTPPath("tcp", servers[0], "/txn/v3/rpc")
	if err != nil {
		return nil, err
	}
	return &Mongo{
		rpc:       rpccli,
		getServer: getServer,
	}, nil
}

// New returns new DB
func New(uri string) (dal.DB, error) {
	rpccli, err := rpc.DialHTTPPath("tcp", uri, "/txn/v3/rpc")
	if err != nil {
		return nil, err
	}
	return &Mongo{
		rpc: rpccli,
	}, nil
}

// Close replica client
func (c *Mongo) Close() error {
	return c.rpc.Close()
}

// Ping replica client
func (c *Mongo) Ping() error {
	return c.rpc.Ping()
}

// Clone create a new DB instance
func (c *Mongo) Clone() dal.DB {
	nc := Mongo{
		TxnID:     c.TxnID,
		RequestID: c.RequestID,
		rpc:       c.rpc,
		parent:    c,
	}
	return &nc
}

// IsDuplicatedError check the error
func (c *Mongo) IsDuplicatedError(error) bool {
	return false
}

// IsNotFoundError check the error
func (c *Mongo) IsNotFoundError(error) bool {
	return false
}

// Table collection operation
func (c *Mongo) Table(collection string) dal.Table {

	col := Collection{
		RequestID: c.RequestID,
		TxnID:     c.TxnID,
	}
	col.collection = collection
	col.rpc = c.rpc

	return &col
}

// NextSequence 获取新序列号(非事务)
func (c *Mongo) NextSequence(ctx context.Context, sequenceName string) (uint64, error) {
	// build msg
	msg := types.OPFindAndModifyOperation{}
	msg.OPCode = types.OPFindAndModifyCode
	msg.Collection = common.BKTableNameIDgenerator
	if err := msg.DOC.Encode(types.Document{
		"$inc": types.Document{"SequenceID": 1},
	}); err != nil {
		return 0, err
	}
	if err := msg.Selector.Encode(types.Document{
		"_id": sequenceName,
	}); err != nil {
		return 0, err
	}
	msg.Upsert = true
	msg.ReturnNew = true

	// set txn
	opt, ok := ctx.Value(common.CCContextKeyJoinOption).(dal.JoinOption)
	if ok {
		msg.RequestID = opt.RequestID
		// msg.TxnID = opt.TxnID // because NextSequence was not supported for transaction in mongo
	}

	// call
	reply := types.OPReply{}
	err := c.rpc.Call(types.CommandRDBOperation, &msg, &reply)
	if err != nil {
		return 0, err
	}
	if !reply.Success {
		return 0, errors.New(reply.Message)
	}

	if len(reply.Docs) <= 0 {
		return 0, dal.ErrDocumentNotFound
	}

	return strconv.ParseUint(fmt.Sprint(reply.Docs[0]["SequenceID"]), 10, 64)
}

// HasTable 判断是否存在集合
func (c *Mongo) HasTable(tablename string) (bool, error) {
	return false, dal.ErrNotImplemented
}

// DropTable 移除集合
func (c *Mongo) DropTable(tablename string) error {
	return dal.ErrNotImplemented
}

// CreateTable 创建集合
func (c *Mongo) CreateTable(tablename string) error {
	return dal.ErrNotImplemented
}

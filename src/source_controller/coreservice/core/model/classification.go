/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.,
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the ",License",); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an ",AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

import (
	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/errors"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/common/universalsql/mongo"
	"configcenter/src/source_controller/coreservice/core"
	"configcenter/src/storage/dal"
)

type modelClassification struct {
	model   *modelManager
	dbProxy dal.RDB
}

func (m *modelClassification) CreateOneModelClassification(ctx core.ContextParams, inputParam metadata.CreateOneModelClassification) (*metadata.CreateOneDataResult, error) {

	if 0 == len(inputParam.Data.ClassificationID) {
		blog.Errorf("request(%s): it is failed to create the model classification, because of the classificationID (%v) is not set", ctx.ReqID, inputParam.Data)
		return &metadata.CreateOneDataResult{}, ctx.Error.Errorf(common.CCErrCommParamsNeedSet, metadata.ClassFieldClassificationID)
	}

	_, exists, err := m.isExists(ctx, inputParam.Data.ClassificationID)
	if nil != err {
		blog.Errorf("request(%s): it is failed to check if the classification ID (%s)is exists, error info is %s", ctx.ReqID, inputParam.Data.ClassificationID, err.Error())
		return nil, err
	}
	if exists {
		blog.Errorf("classification (%v)is duplicated", inputParam.Data)
		return nil, ctx.Error.Error(common.CCErrCommDuplicateItem)
	}

	id, err := m.save(ctx, inputParam.Data)
	if nil != err {
		blog.Errorf("request(%s): it is failed to save the classification(%v), error info is %s", ctx.ReqID, inputParam.Data, err.Error())
		return &metadata.CreateOneDataResult{}, err
	}
	return &metadata.CreateOneDataResult{Created: metadata.CreatedDataResult{ID: id}}, err
}

func (m *modelClassification) CreateManyModelClassification(ctx core.ContextParams, inputParam metadata.CreateManyModelClassifiaction) (*metadata.CreateManyDataResult, error) {

	dataResult := &metadata.CreateManyDataResult{}

	addExceptionFunc := func(idx int64, err errors.CCErrorCoder, classification *metadata.Classification) {
		dataResult.CreateManyInfoResult.Exceptions = append(dataResult.CreateManyInfoResult.Exceptions, metadata.ExceptionResult{
			OriginIndex: idx,
			Message:     err.Error(),
			Code:        int64(err.GetCode()),
			Data:        classification,
		})
	}

	for itemIdx, item := range inputParam.Data {

		if 0 == len(item.ClassificationID) {
			blog.Errorf("request(%s): it is failed to create the model classification, because of the classificationID (%v) is not set", ctx.ReqID, item.ClassificationID)
			addExceptionFunc(int64(itemIdx), ctx.Error.Errorf(common.CCErrCommParamsNeedSet, metadata.ClassFieldClassificationID).(errors.CCErrorCoder), &item)
			continue
		}

		_, exists, err := m.isExists(ctx, item.ClassificationID)
		if nil != err {
			blog.Errorf("request(%s): it is failed to check the classification ID (%s) is exists, error info is %s", ctx.ReqID, item.ClassificationID, err.Error())
			addExceptionFunc(int64(itemIdx), err.(errors.CCErrorCoder), &item)
			continue
		}

		if exists {
			dataResult.Repeated = append(dataResult.Repeated, metadata.RepeatedDataResult{OriginIndex: int64(itemIdx), Data: mapstr.NewFromStruct(item, "field")})
			continue
		}

		id, err := m.save(ctx, item)
		if nil != err {
			blog.Errorf("request(%s): it is failed to save the clasisfication(%v), error info is %s", ctx.ReqID, item, err.Error())
			addExceptionFunc(int64(itemIdx), err.(errors.CCErrorCoder), &item)
			continue
		}

		dataResult.Created = append(dataResult.Created, metadata.CreatedDataResult{
			ID: id,
		})

	}

	return dataResult, nil
}
func (m *modelClassification) SetManyModelClassification(ctx core.ContextParams, inputParam metadata.SetManyModelClassification) (*metadata.SetDataResult, error) {

	dataResult := &metadata.SetDataResult{}

	addExceptionFunc := func(idx int64, err errors.CCErrorCoder, classification *metadata.Classification) {
		dataResult.Exceptions = append(dataResult.Exceptions, metadata.ExceptionResult{
			OriginIndex: idx,
			Message:     err.Error(),
			Code:        int64(err.GetCode()),
			Data:        classification,
		})
	}

	for itemIdx, item := range inputParam.Data {

		if 0 == len(item.ClassificationID) {
			blog.Errorf("request(%s): it is failed to create the model classification, because of the classificationID (%v) is not set", ctx.ReqID, item.ClassificationID)
			addExceptionFunc(int64(itemIdx), ctx.Error.Errorf(common.CCErrCommParamsNeedSet, metadata.ClassFieldClassificationID).(errors.CCErrorCoder), &item)
			continue
		}

		origin, exists, err := m.isExists(ctx, item.ClassificationID)
		if nil != err {
			blog.Errorf("request(%s): it is failed to check the classification ID (%s) is exists, error info is %s", ctx.ReqID, item.ClassificationID, err.Error())
			addExceptionFunc(int64(itemIdx), err.(errors.CCErrorCoder), &item)
			continue
		}

		if exists {

			cond := mongo.NewCondition()
			cond.Element(&mongo.Eq{Key: metadata.ClassificationFieldID, Val: origin.ID})
			if _, err := m.update(ctx, mapstr.NewFromStruct(item, "field"), cond); nil != err {
				blog.Errorf("request(%s): it is failed to update some fields(%v) of the classification by the condition(%v), error info is %s", ctx.ReqID, item, cond.ToMapStr(), err.Error())
				addExceptionFunc(int64(itemIdx), err.(errors.CCErrorCoder), &item)
				continue
			}

			dataResult.UpdatedCount.Count++
			dataResult.Updated = append(dataResult.Updated, metadata.UpdatedDataResult{
				OriginIndex: int64(itemIdx),
				ID:          uint64(origin.ID),
			})
			continue
		}

		id, err := m.save(ctx, item)
		if nil != err {
			blog.Errorf("request(%s): it is failed to save the classification(%v), error info is %s", ctx.ReqID, item, err.Error())
			addExceptionFunc(int64(itemIdx), err.(errors.CCErrorCoder), &item)
			continue
		}

		dataResult.CreatedCount.Count++
		dataResult.Created = append(dataResult.Created, metadata.CreatedDataResult{
			ID: id,
		})

	}

	return dataResult, nil
}

func (m *modelClassification) SetOneModelClassification(ctx core.ContextParams, inputParam metadata.SetOneModelClassification) (*metadata.SetDataResult, error) {

	if 0 == len(inputParam.Data.ClassificationID) {
		blog.Errorf("request(%s): it is failed to set the model classification, because of the classificationID (%v) is not set", ctx.ReqID, inputParam.Data)
		return &metadata.SetDataResult{}, ctx.Error.Errorf(common.CCErrCommParamsNeedSet, metadata.ClassFieldClassificationID)
	}

	origin, exists, err := m.isExists(ctx, inputParam.Data.ClassificationID)
	if nil != err {
		blog.Errorf("request(%s): it is failed to check the classification ID (%s) is exists, error info is %s", ctx.ReqID, inputParam.Data.ClassificationID, err.Error())
		return &metadata.SetDataResult{}, err
	}

	dataResult := &metadata.SetDataResult{}

	addExceptionFunc := func(idx int64, err errors.CCErrorCoder, classification *metadata.Classification) {
		dataResult.Exceptions = append(dataResult.Exceptions, metadata.ExceptionResult{
			OriginIndex: idx,
			Message:     err.Error(),
			Code:        int64(err.GetCode()),
			Data:        classification,
		})
	}
	if exists {

		cond := mongo.NewCondition()
		cond.Element(&mongo.Eq{Key: metadata.ClassificationFieldID, Val: origin.ID})
		if _, err := m.update(ctx, mapstr.NewFromStruct(inputParam.Data, "field"), cond); nil != err {
			blog.Errorf("request(%s): it is failed to update some fields(%v) for a classification by the condition(%v), error info is %s", ctx.ReqID, inputParam.Data, cond.ToMapStr(), err.Error())
			addExceptionFunc(0, err.(errors.CCErrorCoder), &inputParam.Data)
			return dataResult, nil
		}
		dataResult.UpdatedCount.Count++
		dataResult.Updated = append(dataResult.Updated, metadata.UpdatedDataResult{ID: uint64(origin.ID)})
		return dataResult, err
	}

	id, err := m.save(ctx, inputParam.Data)
	if nil != err {
		blog.Errorf("request(%s): it is failed to save the classification(%v), error info is %s", ctx.ReqID, inputParam.Data, err.Error())
		addExceptionFunc(0, err.(errors.CCErrorCoder), origin)
		return dataResult, err
	}
	dataResult.CreatedCount.Count++
	dataResult.Created = append(dataResult.Created, metadata.CreatedDataResult{ID: id})
	return dataResult, err
}

func (m *modelClassification) UpdateModelClassification(ctx core.ContextParams, inputParam metadata.UpdateOption) (*metadata.UpdatedCount, error) {

	cond, err := mongo.NewConditionFromMapStr(inputParam.Condition)
	if nil != err {
		blog.Errorf("request(%s): it is failed to convert the condition(%v) from mapstr into condition object, error info is %s", ctx.ReqID, inputParam.Condition, err.Error())
		return &metadata.UpdatedCount{}, err
	}

	cond.Element(&mongo.Eq{Key: metadata.AttributeFieldSupplierAccount, Val: ctx.SupplierAccount})
	cnt, err := m.update(ctx, inputParam.Data, cond)
	if nil != err {
		blog.Errorf("request(%s): it is failed to update some fields(%v) for some classifications by the condition(%v), error info is %s", ctx.ReqID, inputParam.Data, inputParam.Condition, err.Error())
		return &metadata.UpdatedCount{}, err
	}
	return &metadata.UpdatedCount{Count: cnt}, nil
}

func (m *modelClassification) DeleteModelClassificaiton(ctx core.ContextParams, inputParam metadata.DeleteOption) (*metadata.DeletedCount, error) {

	deleteCond, err := mongo.NewConditionFromMapStr(inputParam.Condition)
	if nil != err {
		blog.Errorf("request(%s): it is failed to convert the condition (%v) from mapstr into condition object, error info is %s", ctx.ReqID, inputParam.Condition, err.Error())
		return &metadata.DeletedCount{}, ctx.Error.New(common.CCErrCommHTTPInputInvalid, err.Error())
	}

	deleteCond.Element(&mongo.Eq{Key: metadata.ClassFieldClassificationSupplierAccount, Val: ctx.SupplierAccount})
	cnt, exists, err := m.hasModel(ctx, deleteCond)
	if nil != err {
		blog.Errorf("request(%s): it is failed to check whether the classifications which are marked by the condition (%v) have some models, error info is %s", ctx.ReqID, deleteCond.ToMapStr(), err.Error())
		return &metadata.DeletedCount{}, err
	}
	if exists {
		return &metadata.DeletedCount{}, ctx.Error.Error(common.CCErrTopoObjectClassificationHasObject)
	}

	cnt, err = m.delete(ctx, deleteCond)
	if nil != err {
		blog.Errorf("request(%s): it is failed to delete the classification whci are marked by the condition(%v), error info is %s", ctx.ReqID, deleteCond.ToMapStr(), err.Error())
		return &metadata.DeletedCount{}, err
	}

	return &metadata.DeletedCount{Count: cnt}, nil
}

func (m *modelClassification) CascadeDeleteModeClassification(ctx core.ContextParams, inputParam metadata.DeleteOption) (*metadata.DeletedCount, error) {

	deleteCond, err := mongo.NewConditionFromMapStr(inputParam.Condition)
	if nil != err {
		blog.Errorf("request(%s): it is failed to convert the condition (%v) from mapstr into condition object, error info is %s", ctx.ReqID, inputParam.Condition, err.Error())
		return &metadata.DeletedCount{}, err
	}
	deleteCond.Element(&mongo.Eq{Key: metadata.ClassFieldClassificationSupplierAccount, Val: ctx.SupplierAccount})

	classificationItems, err := m.search(ctx, deleteCond)
	if nil != err {
		blog.Errorf("request(%s): it is failed to search some classifications by the condition (%v) , error info is %s", ctx.ReqID, inputParam.Condition, err.Error())
		return &metadata.DeletedCount{}, err
	}

	classificationIDS := []string{}
	for _, item := range classificationItems {
		classificationIDS = append(classificationIDS, item.ClassificationID)
	}

	if _, err := m.cascadeDeleteModel(ctx, classificationIDS); nil != err {
		blog.Error("request(%s): it is failed to cascade delete some models by the classificationIDS (%v), error info is %s", ctx.ReqID, classificationIDS, err.Error())
		return &metadata.DeletedCount{}, err
	}

	if _, err := m.delete(ctx, deleteCond); nil != err {
		blog.Errorf("request(%s): it is failed to delete some classifications by the condition (%v), error info is %s", ctx.ReqID, deleteCond.ToMapStr(), err.Error())
		return &metadata.DeletedCount{}, err
	}

	return &metadata.DeletedCount{Count: uint64(len(classificationItems))}, nil
}

func (m *modelClassification) SearchModelClassification(ctx core.ContextParams, inputParam metadata.QueryCondition) (*metadata.QueryModelClassificationDataResult, error) {

	searchCond, err := mongo.NewConditionFromMapStr(inputParam.Condition)
	if nil != err {
		blog.Errorf("request(%s): it is failed to convert the condition (%v) from mapstr into condition object, error info is %s", ctx.ReqID, inputParam.Condition, err.Error())
		return &metadata.QueryModelClassificationDataResult{}, err
	}

	totalCount, err := m.count(ctx, searchCond)
	if nil != err {
		blog.Errorf("request(%s): it is failed to get the count by the condition (%v), error info is %s", ctx.ReqID, searchCond.ToMapStr(), err.Error())
		return &metadata.QueryModelClassificationDataResult{}, err
	}

	searchCond.Element(&mongo.Eq{Key: metadata.ClassFieldClassificationSupplierAccount, Val: ctx.SupplierAccount})
	classificationItems, err := m.search(ctx, searchCond)
	if nil != err {
		blog.Errorf("request(%s): it is failed to search some classifications by the condition (%v), error info is %s", ctx.ReqID, searchCond.ToMapStr(), err.Error())
		return &metadata.QueryModelClassificationDataResult{}, err
	}

	return &metadata.QueryModelClassificationDataResult{Count: int64(totalCount), Info: classificationItems}, nil
}
package service

import (
	"hr-dynamic-form-engine/src/cbase"
	"hr-dynamic-form-engine/src/dao"
	"hr-dynamic-form-engine/src/dto"
	"hr-dynamic-form-engine/src/model"
)

type AuditServer struct{}

func (AuditServer) QueryService(targetType, targetId, formId string, page, pageSize int) *dto.PageResult[model.AuditLog] {
	res, err := dao.AuditDbDefault().Query(targetType, targetId, formId, page, pageSize)
	if err != nil {
		cbase.PanicMsg(err.Error())
	}
	return res
}

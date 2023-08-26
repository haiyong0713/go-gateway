package report

import (
	"go-common/library/log"

	"github.com/pkg/errors"
)

//go:generate easyjson -all report.go

//easyjson:json
type DislikeReportData struct {
	UniqueID   string `json:"unique_id,omitempty"`
	MaterialID int64  `json:"material_id,omitempty"`
}

func BuildDislikeReportData(creativeId int64, posRecUniqueID string) string {
	if creativeId == 0 && posRecUniqueID == "" {
		return ""
	}
	dislikeReportData := &DislikeReportData{
		UniqueID:   posRecUniqueID,
		MaterialID: creativeId,
	}
	drdByte, err := dislikeReportData.MarshalJSON()
	if err != nil {
		log.Error("Failed to marshal DislikeReportData: %+v", errors.WithStack(err))
		return ""
	}
	return string(drdByte)
}

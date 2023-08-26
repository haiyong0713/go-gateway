package fawkes

import (
	apmmdl "go-gateway/app/app-svr/fawkes/service/model/apm"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	"testing"
)

func TestParse(t *testing.T) {
	filter := &apmmdl.Filter{
		AndType:   "1",
		Column:    "app_key",
		EqualType: "1",
		Values:    "iphone",
		ValueType: "STRING",
	}
	filter2 := &apmmdl.Filter{
		AndType:   "1",
		Column:    "app_key",
		EqualType: "1",
		Values:    "android,android64",
		ValueType: "STRING",
	}
	filter3 := &apmmdl.Filter{
		AndType:   "1",
		Column:    "error_stack",
		EqualType: "7",
		Values:    "bilibili,tv",
		ValueType: "STRING",
	}
	filter4 := &apmmdl.Filter{
		AndType:   "1",
		Column:    "error_stack",
		EqualType: "9",
		Values:    "",
		ValueType: "",
	}
	var filters []*apmmdl.Filter
	filters = append(filters, filter, filter2, filter3, filter4)
	match := &apmmdl.MatchOption{}
	match.Filters = filters
	condition, args, err := parseMatchOptionV2(match)
	log.Info("condition:%v args:%+v err:%v", condition, args, err)
}

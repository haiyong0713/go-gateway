package generator

import (
	"fmt"
)

func ProcessMappingRule(mappingRule *MappingRuleDetail) (err error) {
	dest := mappingRule.Dest
	src := mappingRule.Src
	if src == "$mid" {
		src = "__getMid(ctx)"
	}
	if mappingRule.MapFunc == "stringToint64" {
		src = fmt.Sprintf("__stringToint64(%s)", src)
	} else if mappingRule.MapFunc == "int64Tostring" {
		src = fmt.Sprintf("__int64Tostring(%s)", src)
	}

	mappingRule.From = src
	mappingRule.To = dest
	return
}

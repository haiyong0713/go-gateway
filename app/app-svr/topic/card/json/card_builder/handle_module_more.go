package cardbuilder

import (
	jsonwebcard "go-gateway/app/app-svr/topic/card/json"
)

func handleModuleMore() *jsonwebcard.ModuleMore {
	return &jsonwebcard.ModuleMore{
		ThreePointItems: []*jsonwebcard.ThreePointItems{
			{
				Type:  jsonwebcard.ThreePointItemReport,
				Label: "举报",
			},
			{
				Type:  jsonwebcard.ThreePointItemTopicIrrelevant,
				Label: "与话题无关",
			},
		},
	}
}

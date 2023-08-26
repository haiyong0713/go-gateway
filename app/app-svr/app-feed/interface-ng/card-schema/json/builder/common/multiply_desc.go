package jsoncommon

import (
	"fmt"
	"time"

	appcardmodel "go-gateway/app/app-svr/app-card/interface/model"
	jsoncard "go-gateway/app/app-svr/app-feed/interface-ng/card-schema/json"
	arcgrpc "go-gateway/app/app-svr/archive/service/api"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
)

func ConstructMultiplyDesc(in *arcgrpc.Arc, author *accountgrpc.Card, requestAt time.Time) *jsoncard.MultiplyDesc {
	descType, extra := resolveMultiplyDescTypeAndExtra(in, requestAt)
	return &jsoncard.MultiplyDesc{
		AuthorName: resolveAuthorName(in, author),
		Extra:      extra,
		Type:       descType,
	}
}

func resolveAuthorName(in *arcgrpc.Arc, author *accountgrpc.Card) string {
	if author != nil {
		return author.Name
	}
	return in.Author.Name
}

func resolveMultiplyDescTypeAndExtra(in *arcgrpc.Arc, requestAt time.Time) (int8, string) {
	const (
		_unionDescType = 1
	)
	pubTime := appcardmodel.PubDataByRequestAt(in.PubDate.Time(), requestAt)
	if in.Rights.IsCooperation == 1 {
		return _unionDescType, fmt.Sprintf("等%d人联合创作 · %s", len(in.StaffInfo)+1, pubTime)
	}
	return 0, fmt.Sprintf(" · %s", pubTime)
}

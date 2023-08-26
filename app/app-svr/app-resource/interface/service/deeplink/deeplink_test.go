package deeplink

import (
	"testing"

	"go-gateway/app/app-svr/app-resource/interface/model/deeplink"

	"github.com/stretchr/testify/assert"
)

var (
	testOgvOriginLink = "bilibili://bangumi/season/39311?h5awaken=b3Blbl9hcHBfZnJvbV90eXBlPWRlZXBsaW5rX3R0bGgtYW5kLXJkLTMxeWF6LTE2Nzc5NTY3Mzc5MjgxOTkmb3Blbl9hcHBfdXJsPXNzMzkzMTE=&from_spmid=out_open_deeplink_ttlh-and-rd-31yaz-1677956737928199&"
	testUgcOriginLink = "bilibili://video/BV13p4y1k7Wr?h5awaken=b3Blbl9hcHBfZnJvbV90eXBlPWRlZXBsaW5rX3R0bGgtYW5kLXJkLTMxeWF6LTE2Nzc5NTY3Mzc5MjgxOTkmb3Blbl9hcHBfdXJsPUJWMTNwNHkxazdXcg==&from_spmid=out_open_deeplink_ttlh-and-rd-31yaz-1677956737928199&"
	testSpecialLink   = "bilibili://video/BV1pK4y1T7sF?h5awaken=b3Blbl9hcHBfZnJvbV90eXBlPWRlZXBsaW5rX2h3cHBzbGgtYW5kLTMwZGF5cy15YXotMzU3NjgxMzg1MzY1ODMxODA4Jm9wZW5fYXBwX3VybD1CVjFwSzR5MVQ3c0Yt56S+56eR5Lq65paHLTIxMDgwNC00OA==&from_spmid=out_open_deeplink_hwppslh-and-30days-yaz-357681385365831808"
)

func TestResolveInnerMetaInOriginLink(t *testing.T) {
	meta1, err := resolveInnerMetaInOriginLink(testOgvOriginLink)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, &deeplink.AiDeeplinkMaterial{InnerId: "39311", InnerType: 2, SourceName: "deeplink_ttlh", AccountId: "1677956737928199"}, meta1)
	meta2, err := resolveInnerMetaInOriginLink(testUgcOriginLink)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, &deeplink.AiDeeplinkMaterial{InnerId: "BV13p4y1k7Wr", InnerType: 1, SourceName: "deeplink_ttlh", AccountId: "1677956737928199"}, meta2)
	meta3, err := resolveInnerMetaInOriginLink(testSpecialLink)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, &deeplink.AiDeeplinkMaterial{InnerId: "BV1pK4y1T7sF", InnerType: 1, SourceName: "deeplink_hwppslh", AccountId: "357681385365831808"}, meta3)
}

var (
	testSeasonInnerLink = "bilibili://bangumi/season/38398?from_spmid=out_open_deeplink_hwppslh-and-lsyazwjh"
)

func TestParseInnerIdAndInnerType(t *testing.T) {
	innerId, innerType := parseInnerIdAndInnerType(testSeasonInnerLink)
	assert.Equal(t, "38398", innerId)
	assert.Equal(t, int64(2), innerType)
}

package tag

import (
	"context"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestTagInfos(t *testing.T) {
	Convey(t.Name(), t, func() {
		var (
			tags []int64
			mid  int64
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.mInfo).Reply(200).JSON(`{
			"code": 0,
			"data": [
				{
					"tag_id": 9222,
					"tag_name": "英雄联盟",
					"cover": "/sp/87/8732428c57ecc95acc4757ce4d241cf1{IMG}.jpg",
					"content": "《英雄联盟》是由美国Riot Games开发的3D大型竞技场战网游戏，其主创团队是由实力强劲的Dota-Allstars的核心人物，以及暴雪等著名游戏公司的美术、程序、策划人员组成，将DOTA的玩法从对战平台延伸到网络游戏世界。除了DotA",
					"type": 0,
					"state": 0,
					"ctime": 1433151310,
					"count": {
						"view": 0,
						"use": 45372,
						"atten": 4570
					},
					"is_atten": 1,
					"likes": 12,
					"hates": 0,
					"attribute": 0,
					"liked": 0,
					"hated": 1
				},
				{
					"tag_id": 9222,
					"tag_name": "英雄联盟",
					"cover": "/sp/87/8732428c57ecc95acc4757ce4d241cf1{IMG}.jpg",
					"content": "《英雄联盟》是由美国Riot Games开发的3D大型竞技场战网游戏，其主创团队是由实力强劲的Dota-Allstars的核心人物，以及暴雪等著名游戏公司的美术、程序、策划人员组成，将DOTA的玩法从对战平台延伸到网络游戏世界。除了DotA",
					"type": 0,
					"state": 0,
					"ctime": 1433151310,
					"count": {
						"view": 0,
						"use": 45372,
						"atten": 4570
					},
					"is_atten": 1,
					"likes": 12,
					"hates": 0,
					"attribute": 0,
					"liked": 0,
					"hated": 1
				}
			],
			"message": "ok"
		}`)
		res, err := d.TagInfos(context.Background(), tags, mid)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	})
}

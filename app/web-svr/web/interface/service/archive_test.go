package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	tagmdl "go-gateway/app/web-svr/web/interface/model"

	chmdl "go-gateway/app/web-svr/web/interface/model/channel"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestRemoveDuplicateTags(t *testing.T) {
	testVideoTags := []*chmdl.VideoTag{
		{
			Tag: tagmdl.Tag{
				ID:   143724,
				Name: "测试一下",
			},
			JumpUrl: "",
			TagType: "old_channel",
		},
		{
			Tag: tagmdl.Tag{
				ID:   232,
				Name: "测试一下2",
			},
			JumpUrl: "",
			TagType: "new_topic",
		},
		{
			Tag: tagmdl.Tag{
				ID:   1793,
				Name: "测试一下",
			},
			TagType: "topic",
		},
	}

	testVideoTagsRes := []*chmdl.VideoTag{
		{
			Tag: tagmdl.Tag{
				ID:   143724,
				Name: "测试一下",
			},
			JumpUrl: "",
			TagType: "old_channel",
		},
		{
			Tag: tagmdl.Tag{
				ID:   232,
				Name: "测试一下2",
			},
			JumpUrl: "",
			TagType: "new_topic",
		},
	}

	assert.Equal(t, nil, nil)
	assert.Equal(t, testVideoTagsRes, removeDuplicateTags(testVideoTags))
}

func TestService_View(t *testing.T) {
	Convey("test archive view", t, WithService(func(s *Service) {
		var (
			mid int64 = 27515256
			aid int64 = 10110688
			cid int64 = 1
		)
		res, _, err := s.View(context.Background(), aid, cid, mid, "", "", "")
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
		str, _ := json.Marshal(res)
		Printf("%s", str)
	}))
}

func TestService_ArchiveStat(t *testing.T) {
	Convey("test archive archiveStat", t, WithService(func(s *Service) {
		var aid int64 = 5464686
		res, err := s.ArchiveStat(context.Background(), aid)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_AddShare(t *testing.T) {
	Convey("test archive AddShare", t, WithService(func(s *Service) {
		var (
			mid int64 = 27515256
			aid int64 = 5464686
		)
		res, _, err := s.AddShare(context.Background(), aid, mid, 0, 0, 0, 0, "", "", "", "")
		So(err, ShouldBeNil)
		So(res, ShouldBeGreaterThan, 0)
	}))
}

func TestService_Description(t *testing.T) {
	Convey("test archive Description", t, WithService(func(s *Service) {
		var (
			aid  int64 = 5464686
			page int64 = 1
		)
		res, err := s.Description(context.Background(), aid, page)
		So(err, ShouldBeNil)
		So(res, ShouldNotBeEmpty)
	}))
}

func TestService_ArcReport(t *testing.T) {
	Convey("test archive ArcReport", t, WithService(func(s *Service) {
		var (
			mid    int64 = 27515256
			aid    int64 = 5464686
			tp     int64
			reason string
			pics   string
		)
		err := s.ArcReport(context.Background(), mid, aid, tp, reason, pics)
		So(err, ShouldBeNil)
	}))
}

func TestService_AppealTags(t *testing.T) {
	Convey("test archive AppealTags", t, WithService(func(s *Service) {
		res, err := s.AppealTags(context.Background())
		So(err, ShouldBeNil)
		So(len(res), ShouldBeGreaterThan, 0)
	}))
}

func TestService_AuthorRecommend(t *testing.T) {
	Convey("test archive AuthorRecommend", t, WithService(func(s *Service) {
		var aid int64 = 5464686
		res, err := s.AuthorRecommend(context.Background(), aid)
		So(err, ShouldBeNil)
		So(len(res), ShouldBeGreaterThan, 0)
	}))
}

func TestService_RelatedArcs(t *testing.T) {
	Convey("test archive RelatedArcs", t, WithService(func(s *Service) {
		var aid int64 = 5464686
		res, _, err := s.RelatedArcs(context.Background(), aid, 0, "", false, false, nil)
		So(err, ShouldBeNil)
		So(len(res), ShouldBeGreaterThan, 0)
	}))
}

func TestService_Detail(t *testing.T) {
	Convey("test archive Detail", t, WithService(func(s *Service) {
		var aid int64 = 10113300
		type key string
		var caller key = "caller"
		c := context.Background()
		ctx := context.WithValue(c, caller, "main.web-svr.web-interface")
		res, err := s.Detail(ctx, aid, 0, "", "", "", true, true, true, true)
		data, _ := json.MarshalIndent(res, "", "\t")
		fmt.Printf("%+v\n", string(data))
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

func TestService_DetailTag(t *testing.T) {
	Convey("test archive DetailTag", t, WithService(func(s *Service) {
		var aid int64 = 10113300
		c := context.Background()
		type key string
		var caller key = "caller"
		ctx := context.WithValue(c, caller, "main.web-svr.web-interface")
		res, err := s.DetailTag(ctx, aid, 0, nil)
		data, _ := json.MarshalIndent(res, "", "\t")
		fmt.Printf("%+v\n", string(data))
		So(err, ShouldBeNil)
		So(res, ShouldNotBeNil)
	}))
}

package service

import (
	"encoding/json"
	"testing"
)

var (
	tmpService = new(Service)
)

// go test -v auto_subscribe.go favorite.go grpc.go grpc_test.go guess.go live.go match.go match_active.go pointdata.go s10.go s10_score_analysis.go s10_tab.go s9.go search.go service.go
func TestGrpcServiceBiz(t *testing.T) {
	teamListStr := `{"1255":{"id":1255,"title":"title1","sub_title":"sub_title","e_title":"","create_time":0,"area":"","logo":"logo","uid":0,"members":"","dic":"","is_deleted":0,"video_url":"video_url","profile":"profile","leida_tid":0,"reply_id":0,"team_type":0,"region_id":0},"1256":{"id":1256,"title":"title1","sub_title":"sub_title","e_title":"","create_time":0,"area":"","logo":"logo","uid":0,"members":"","dic":"","is_deleted":0,"video_url":"video_url","profile":"profile","leida_tid":0,"reply_id":0,"team_type":0,"region_id":0}}`
	contestListStr := `{"7609":{"id":7609,"game_stage":"game_stage","stime":0,"etime":0,"home_id":0,"away_id":0,"home_score":0,"away_score":0,"live_room":0,"aid":0,"collection":0,"collection_bvid":"","game_state":0,"dic":"dic","ctime":"2020-10-09T07:15:47+08:00","mtime":"2020-10-09T07:15:47+08:00","status":0,"sid":0,"mid":0,"season":null,"home_team":null,"away_team":null,"special":0,"success_team":0,"success_teaminfo":null,"special_name":"special_n","special_tips":"spec","special_image":"special_image","playback":"playback","collection_url":"collection_url","live_url":"live_url","data_type":0,"match_id":0,"guess_type":0,"guess_show":0,"bvid":"","game_stage1":"game_stage1","game_stage2":"game_stage2","live_status":0,"live_popular":0,"live_cover":"","push_switch":0,"live_title":""},"7610":{"id":7610,"game_stage":"game_stage","stime":0,"etime":0,"home_id":0,"away_id":0,"home_score":0,"away_score":0,"live_room":0,"aid":0,"collection":0,"collection_bvid":"","game_state":0,"dic":"dic","ctime":"2020-10-09T07:15:47+08:00","mtime":"2020-10-09T07:15:47+08:00","status":0,"sid":0,"mid":0,"season":null,"home_team":null,"away_team":null,"special":0,"success_team":0,"success_teaminfo":null,"special_name":"special_n","special_tips":"spec","special_image":"special_image","playback":"playback","collection_url":"collection_url","live_url":"live_url","data_type":0,"match_id":0,"guess_type":0,"guess_show":0,"bvid":"","game_stage1":"game_stage1","game_stage2":"game_stage2","live_status":0,"live_popular":0,"live_cover":"","push_switch":0,"live_title":""},"7611":{"id":7611,"game_stage":"game_stage","stime":0,"etime":0,"home_id":0,"away_id":0,"home_score":0,"away_score":0,"live_room":0,"aid":0,"collection":0,"collection_bvid":"","game_state":0,"dic":"dic","ctime":"2020-10-09T07:15:48+08:00","mtime":"2020-10-09T07:15:48+08:00","status":0,"sid":0,"mid":0,"season":null,"home_team":null,"away_team":null,"special":0,"success_team":0,"success_teaminfo":null,"special_name":"special_n","special_tips":"spec","special_image":"special_image","playback":"playback","collection_url":"collection_url","live_url":"live_url","data_type":0,"match_id":0,"guess_type":0,"guess_show":0,"bvid":"","game_stage1":"game_stage1","game_stage2":"game_stage2","live_status":0,"live_popular":0,"live_cover":"","push_switch":0,"live_title":""}}`
	seasonListStr := `{"179":{"id":179,"mid":0,"title":"title1","sub_title":"sub_title1","stime":0,"etime":0,"sponsor":"sponsor","logo":"logo","dic":"dic","status":0,"ctime":1602198427,"mtime":1602198427,"rank":0,"is_app":0,"url":"url","data_focus":"data_focus","focus_url":"focus_url","leida_sid":0,"game_type":0,"search_image":"search_image","sync_platform":0},"180":{"id":180,"mid":0,"title":"title1","sub_title":"sub_title1","stime":0,"etime":0,"sponsor":"sponsor","logo":"logo","dic":"dic","status":0,"ctime":1602198430,"mtime":1602198430,"rank":0,"is_app":0,"url":"url","data_focus":"data_focus","focus_url":"focus_url","leida_sid":0,"game_type":0,"search_image":"search_image","sync_platform":0}}`

	if err := json.Unmarshal([]byte(teamListStr), &specifiedTeamMap); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal([]byte(contestListStr), &specifiedContestMap); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal([]byte(seasonListStr), &specifiedSeasonMap); err != nil {
		t.Error(err)
	}

	t.Run("teamBiz", teamBiz)
	t.Run("contestBiz", contestBiz)
	t.Run("seasonBiz", seasonBiz)
}

func teamBiz(t *testing.T) {
	m, missList := fetchTeamListFromMemoryCache([]int64{1255})
	if len(missList) > 0 {
		t.Error("team should be in memory cache")
	}

	for _, v := range m {
		tmp := v.DeepCopy()
		tmp.ID = 111
		if v.ID == tmp.ID {
			t.Error("deepCopy biz is wrong")
		}
	}
}

func seasonBiz(t *testing.T) {
	m, missList := fetchSeasonListFromMemoryCache([]int64{179})
	if len(missList) > 0 {
		t.Error("contest should be in memory cache")
	}

	for _, v := range m {
		tmp := v.DeepCopy()
		tmp.ID = 111
		if v.ID == tmp.ID {
			t.Error("deepCopy biz is wrong")
		}
	}
}

func contestBiz(t *testing.T) {
	m, missList := fetchContestListFromMemoryCache([]int64{7609})
	if len(missList) > 0 {
		t.Error("contest should be in memory cache")
	}

	for _, v := range m {
		tmp := v.DeepCopy()
		tmp.ID = 111
		if v.ID == tmp.ID {
			t.Error("deepCopy biz is wrong")
		}
	}
}

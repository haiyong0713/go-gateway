package recommend

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-card/interface/model"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

func httpMock(method, url string) *gock.Request {
	r := gock.New(url)
	r.Method = strings.ToUpper(method)
	return r
}

func TestRecommend(t *testing.T) {
	Convey("Recommend", t, func() {
		var (
			plat              int8
			buvid             string
			mid               int64
			build, loginEvent int
			zoneID            int64
			group             int
			interest, network string
			style             int
			column            model.ColumnStatus
			flush             int
			autoplay          string
			now               time.Time
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", fmt.Sprintf(d.rcmd, group)).Reply(200).JSON(`{
			"code": 0,
			"data": [
				{
					"av_feature": "{\"ctr\":0.0367,\"fctr\":0.0178,\"wdlks\":0.1285,\"dlr\":0.0,\"fls\":0.0,\"rankscore\":0.2243,\"fo\":0,\"reasontype\":3,\"fms\":0.1206,\"av_play\":327203,\"rid\":3,\"d\":\" |d 4\",\"v_cl\":\" |v_cl 9\",\"v_bl\":\" |v_bl 9\",\"v_fl\":\" |v_fl 9\",\"real_matchtype\":\" |real_matchtype 4$2\",\"source_len\":\" |source_len 2\",\"matchtype\":\" |matchtype 16$9\",\"nonclick_show_region_num\":\" |nonclick_show_region_num 6\",\"nonclick_show_tag_num\":\" |nonclick_show_tag_num 4\",\"m_k_word\":\" |m_k_word 自制 校园 学习 全能打卡挑战\",\"m_k_w\":\" |m_k_w 4\",\"ysession_state\":\" |ysession_state no_click_x\",\"dr_class_match\":\" |dr_class_match 2_1\",\"play_show_region_num\":\" |play_show_region_num 6\",\"play_show_tag_num\":\"|play_show_tag_num 6\",\"p_xsession_state\":\" |p_xsession_state play_x\",\"play_region_num\":\" |play_region_num 1\",\"play_tag_num\":\" |play_tag_num 1\",\"play_tag\":\" |play_tag 792\",\"pr_class_match\":\" |pr_class_match 2_1_3\",\"r_m_6\":\" |r_m_6 0.166667\",\"r_m_32\":\" |r_m_32 0.375\"}",
					"goto": "av",
					"id": 75882187,
					"rcmd_reason": {
						"content": "4万点赞",
						"corner_mark": 2,
						"jumpgoto": "",
						"jumpid": 0,
						"style": 2
					},
					"source": "online_tag$online_av2av",
					"tid": 11657551,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				},
				{
					"av_feature": "{\"ctr\":0.0214,\"fctr\":0.0,\"wdlks\":0.0,\"dlr\":0.0,\"fls\":0.0,\"rankscore\":0.0,\"fo\":0,\"reasontype\":0,\"fms\":0.0,\"av_play\":-1,\"lup_area\":\" |lup_area 5 192\"}",
					"goto": "live",
					"id": 3981708,
					"rcmd_reason": null,
					"source": "live",
					"tid": 0,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				},
				{
					"av_feature": "{\"ctr\":0.0097,\"fctr\":0.0109,\"wdlks\":0.1255,\"dlr\":0.0,\"fls\":0.0,\"rankscore\":0.1744,\"fo\":0,\"reasontype\":9,\"fms\":0.1113,\"av_play\":591289,\"rid\":9,\"d\":\" |d 4\",\"v_cl\":\" |v_cl 9\",\"v_bl\":\" |v_bl 9\",\"v_fl\":\" |v_fl 9\",\"real_matchtype\":\" |real_matchtype 2\",\"source_len\":\" |source_len 1\",\"matchtype\":\" |matchtype 25\",\"nonclick_show_region_num\":\" |nonclick_show_region_num 6\",\"nonclick_show_tag_num\":\" |nonclick_show_tag_num 4\",\"m_k_w\":\" |m_k_w 0\",\"ysession_state\":\" |ysession_state no_click_x\",\"dr_class_match\":\" |dr_class_match 2_1\",\"play_show_region_num\":\" |play_show_region_num 6\",\"play_show_tag_num\":\" |play_show_tag_num 6\",\"p_xsession_state\":\" |p_xsession_state play_x\",\"play_region_num\":\" |play_region_num 1\",\"play_tag_num\":\" |play_tag_num 1\",\"play_tag\":\" |play_tag 1742\",\"pr_class_match\":\" |pr_class_match 2_1_3\",\"r_m_6\":\" |r_m_6 0.166667\",\"r_m_32\":\" |r_m_32 0.375\"}",
					"goto": "av",
					"id": 75871289,
					"rcmd_reason": {
						"content": "互动视频 9.4分",
						"corner_mark": 2,
						"jumpgoto": "",
						"jumpid": 0,
						"style": 2
					},
					"source": "interactive_av",
					"tid": 12080024,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				},
				{
					"av_feature": "{\"ctr\":0.0145,\"fctr\":0.0139,\"wdlks\":0.0735,\"dlr\":0.0003,\"fls\":0.0,\"rankscore\":0.1066,\"fo\":0,\"reasontype\":0,\"fms\":0.0728,\"av_play\":66305,\"d\":\" |d 2\",\"v_cl\":\" |v_cl 9\",\"v_bl\":\" |v_bl 8\",\"v_fl\":\" |v_fl 9\",\"real_matchtype\":\" |real_matchtype 2$4$1\",\"source_len\":\" |source_len 3\",\"matchtype\":\" |matchtype 10$9$15$2\",\"nonclick_show_region_num\":\" |nonclick_show_region_num 3\",\"nonclick_show_tag_num\":\" |nonclick_show_tag_num 3\",\"m_k_w\":\" |m_k_w 0\",\"ysession_state\":\" |ysession_state no_click_x\",\"dr_class_match\":\" |dr_class_match 1_2\",\"play_show_region_num\":\" |play_show_region_num 4\",\"play_show_tag_num\":\" |play_show_tag_num 4\",\"p_xsession_state\":\" |p_xsession_state play_x\",\"play_region_num\":\" |play_region_num 0\",\"play_tag_num\":\" |play_tag_num 0\",\"pr_class_match\":\" |pr_class_match 1_2_0\",\"r_m_6\":\" |r_m_6 0\",\"r_m_32\":\" |r_m_32 0\"}",
					"goto": "av",
					"id": 77766948,
					"rcmd_reason": null,
					"source": "app_end$offline_tag$online_tag$region_dynamic",
					"tid": 12407472,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				},
				{
					"av_feature": "{\"ctr\":0.0168,\"fctr\":0.0078,\"wdlks\":0.1178,\"dlr\":0.0,\"fls\":0.0,\"rankscore\":0.0911,\"fo\":0,\"reasontype\":0,\"fms\":0.1021,\"av_play\":13986,\"d\":\" |d 7\",\"v_cl\":\" |v_cl 8\",\"v_bl\":\" |v_bl 5\",\"v_fl\":\" |v_fl 9\",\"real_matchtype\":\" |real_matchtype 4\",\"source_len\":\" |source_len 1\",\"matchtype\":\" |matchtype 16\",\"nonclick_show_region_num\":\" |nonclick_show_region_num 3\",\"nonclick_show_tag_num\":\" |nonclick_show_tag_num 1\",\"m_k_w\":\" |m_k_w 0\",\"ysession_state\":\" |ysession_state no_click_x\",\"dr_class_match\":\" |dr_class_match 0_1\",\"play_show_region_num\":\" |play_show_region_num 5\",\"play_show_tag_num\":\" |play_show_tag_num 2\",\"p_xsession_state\":\" |p_xsession_state play_x\",\"play_region_num\":\" |play_region_num 0\",\"play_tag_num\":\" |play_tag_num 0\",\"pr_class_match\":\" |pr_class_match 0_2_0\",\"r_m_6\":\" |r_m_6 0\",\"r_m_32\":\" |r_m_32 0\"}",
					"goto": "av",
					"id": 55930208,
					"rcmd_reason": null,
					"source": "online_av2av",
					"tid": 306801,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				},
				{
					"av_feature": "{\"ctr\":0.0277,\"fctr\":0.0107,\"wdlks\":0.08,\"dlr\":0.0,\"fls\":0.0,\"rankscore\":0.0885,\"fo\":0,\"reasontype\":0,\"fms\":0.0748,\"av_play\":527493,\"d\":\" |d 5\",\"v_cl\":\" |v_cl 9\",\"v_bl\":\" |v_bl 9\",\"v_fl\":\" |v_fl 9\",\"real_matchtype\":\" |real_matchtype 4\",\"source_len\":\" |source_len 1\",\"matchtype\":\" |matchtype 16\",\"nonclick_show_region_num\":\" |nonclick_show_region_num 0\",\"nonclick_show_tag_num\":\" |nonclick_show_tag_num 0\",\"m_k_word\":\" |m_k_word 经验分享\",\"m_k_w\":\" |m_k_w 1\",\"ysession_state\":\" |ysession_state no_click_x\",\"dr_class_match\":\" |dr_class_match 0_0\",\"play_show_region_num\":\" |play_show_region_num 0\",\"play_show_tag_num\":\" |play_show_tag_num 0\",\"p_xsession_state\":\" |p_xsession_state play_x\",\"play_region_num\":\" |play_region_num 0\",\"play_tag_num\":\" |play_tag_num 0\",\"pr_class_match\":\" |pr_class_match 0_0_0\",\"r_m_6\":\" |r_m_6 0\",\"r_m_32\":\" |r_m_32 0\"}",
					"goto": "av",
					"id": 71287487,
					"rcmd_reason": null,
					"source": "online_av2av",
					"tid": 3162912,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				},
				{
					"av_feature": "{\"ctr\":0.0136,\"fctr\":0.0,\"wdlks\":0.0,\"dlr\":0.0,\"fls\":0.0,\"rankscore\":0.0,\"fo\":0,\"reasontype\":6,\"fms\":0.0,\"av_play\":-1,\"rid\":6,\"up_mid\":\" |up_mid 3295\",\"followed_mid\":\" |followed_mid 27593118\"}",
					"goto": "picture",
					"id": 329606239956136911,
					"rcmd_reason": {
						"content": "",
						"corner_mark": 2,
						"followed_mid": 27593118,
						"jumpgoto": "",
						"jumpid": 0,
						"style": 4
					},
					"source": "dynamic",
					"tid": 0,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				},
				{
					"av_feature": "{\"ctr\":0.0,\"fctr\":0.0,\"wdlks\":0.0,\"dlr\":0.0,\"fls\":0.0,\"rankscore\":0.0,\"fo\":0,\"reasontype\":0,\"fms\":0.0,\"av_play\":-1}",
					"goto": "av",
					"id": 77051467,
					"rcmd_reason": null,
					"source": "dalao",
					"tid": 199924,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				},
				{
					"av_feature": "{\"ctr\":0.0163,\"fctr\":0.0077,\"wdlks\":0.1076,\"dlr\":0.0002,\"fls\":0.0,\"rankscore\":0.0829,\"fo\":0,\"reasontype\":0,\"fms\":0.0939,\"av_play\":50647,\"d\":\" |d 0\",\"v_cl\":\" |v_cl 9\",\"v_bl\":\" |v_bl 8\",\"v_fl\":\" |v_fl 8\",\"real_matchtype\":\" |real_matchtype 4$2$1\",\"source_len\":\" |source_len 3\",\"matchtype\":\" |matchtype 16$15$15$2\",\"nonclick_show_region_num\":\" |nonclick_show_region_num 6\",\"nonclick_show_tag_num\":\" |nonclick_show_tag_num 5\",\"m_k_word\":\" |m_k_word 美食 美食圈\",\"m_k_w\":\" |m_k_w 2\",\"ysession_state\":\" |ysession_state no_click_x\",\"dr_class_match\":\" |dr_class_match 1_1\",\"play_show_region_num\":\" |play_show_region_num 6\",\"play_show_tag_num\":\" |play_show_tag_num 6\",\"p_xsession_state\":\" |p_xsession_state play_x\",\"play_region_num\":\" |play_region_num 0\",\"play_tag_num\":\" |play_tag_num 0\",\"pr_class_match\":\" |pr_class_match 1_1_0\",\"r_m_6\":\" |r_m_6 0\",\"r_m_32\":\" |r_m_32 0\"}",
					"goto": "av",
					"id": 78143682,
					"rcmd_reason": null,
					"source": "offline_tag$region_dynamic$online_av2av$long_term_tag",
					"tid": 12551626,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				},
				{
					"av_feature": "{\"ctr\":0.0259,\"fctr\":0.0144,\"wdlks\":0.0323,\"dlr\":0.0001,\"fls\":0.0,\"rankscore\":0.0567,\"fo\":0,\"reasontype\":5,\"fms\":0.0404,\"av_play\":49393,\"rid\":5,\"d\":\" |d 1\",\"v_cl\":\" |v_cl 9\",\"v_bl\":\" |v_bl 8\",\"v_fl\":\" |v_fl 9\",\"real_matchtype\":\" |real_matchtype 2$5\",\"source_len\":\" |source_len 3\",\"matchtype\":\" |matchtype 9$8$15$5\",\"nonclick_show_region_num\":\" |nonclick_show_region_num 0\",\"nonclick_show_tag_num\":\" |nonclick_show_tag_num 2\",\"m_k_word\":\" |m_k_word 明星 vlog\",\"m_k_w\":\" |m_k_w 2\",\"ysession_state\":\" |ysession_state no_click_x\",\"dr_class_match\":\" |dr_class_match 0_0\",\"play_show_region_num\":\" |play_show_region_num 0\",\"play_show_tag_num\":\" |play_show_tag_num 4\",\"p_xsession_state\":\" |p_xsession_state play_x\",\"play_region_num\":\" |play_region_num 0\",\"play_tag_num\":\" |play_tag_num 1\",\"play_tag\":\" |play_tag 2848\",\"pr_class_match\":\" |pr_class_match 0_0_0\",\"r_m_6\":\" |r_m_6 0\",\"r_m_32\":\" |r_m_32 0\"}",
					"goto": "av",
					"id": 77928553,
					"rcmd_reason": {
						"content": "数码·点赞飙升",
						"corner_mark": 2,
						"jumpgoto": "",
						"jumpid": 0,
						"style": 2
					},
					"source": "offline_tag$online_tag$new_dynamic$av_boost",
					"tid": 10297101,
					"trackid": "all_17.shylf-ai-recsys-87.1575623748962.517"
				}
			],
			"dislike_exp": 1,
			"user_feature": "{\"click_dislike_count\":\"0\",\"cand_size\":0,\"config_string\":\"{\\\"fo_like_alpha\\\":0.0,\\\"fo_like_beta\\\":0.8,\\\"follow_alpha\\\":0.0,\\\"follow_beta\\\":100.0,\\\"follow_rank_mode\\\":0,\\\"foup_strategy\\\":\\\"global_fps_thr\\\",\\\"like_alpha\\\":0.5,\\\"like_beta\\\":0.0,\\\"low_fctr_thr\\\":0.003,\\\"low_fps_thr\\\":0.007,\\\"max_fo_rank\\\":50,\\\"part_rank_count\\\":500,\\\"rank_mode\\\":4,\\\"rm_fls_count\\\":250,\\\"rm_lr_count\\\":500,\\\"wd_ctr_param\\\":1.0,\\\"wd_like_alpha\\\":0.0,\\\"wd_like_beta\\\":90.0}\",\"real\":\" |real 78145881 75767910 73564307\",\"fresh_idx\":10,\"network\":\"mobile\",\"autoplay_card\":\"2|2\",\"rank_type\":\"wide_deep\",\"action_rank_type\":\"wide_deep_action\",\"recsys_mode\":0,\"fomode_adjustshow\":0,\"foup_cnt\":\"0\",\"low_score\":0.0451,\"low_fo_score\":0.0015,\"max_fo_score\":0.0141,\"filter_low_fps\":\"19\",\"has_foup_detail\":\"22|\",\"x_fresh\":\"10\",\"t_fresh\":\"14\",\"is_fallback\":\"0\",\"explore_fair\":\"0\",\"explore_exp\":\"0\",\"last_detail\":{\"click_region\":[],\"click_tag\":[],\"display_region\":[21,76,124,39,122],\"display_tag\":[20215,1207642,11657551,6942,1742,13160,1833,4149,239855,1283883]},\"pd_tag\":\" |pd_tag 6942 7729 1207642 11657551 1833 20215 5417 536\",\"p_tag\":\" |p_tag 253801 2511282 55564 70561 1742 530003 8816 8224300 6020954 13160 37366 198984 792 1057109 3390 2848\",\"pd_r\":\" |pd_r 124 39 76 138\",\"f_idx\":\" |f_idx 10\",\"p_real\":\" |p_real 28830939 39736806 38023415\",\"p_r\":\" |p_r 182 21\",\"sp_tag\":\" |sp_tag 530003 1742 13160 1217 1833 3390 188295 1057109\",\"sp_tag_1h\":\" |sp_tag_1h 8816 8224300 13160 37366 2848 1057109 3390 1742\",\"last_play_cnt\":\" |last_play_cnt 0.484848\",\"last_play_time\":\" |last_play_time 0.431548\"}"
		}`)
		res, _, _, _, err := d.Recommend(context.Background(), plat, buvid, mid, build, loginEvent, zoneID, group, interest, network, style, column, flush, autoplay, now)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestHots(t *testing.T) {
	Convey("Hots", t, func() {
		d.clientAsyn.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.hot).Reply(200).JSON(`{
			"note": false,
			"source_date": "2019-12-19",
			"code": 0,
			"num": 500,
			"list": [
				{
					"aid": 79753087,
					"mid": 9824766,
					"score": 1023030,
					"desc": ""
				},
				{
					"aid": 79656005,
					"mid": 14110780,
					"score": 697778,
					"desc": ""
				}
			]
		}`)
		res, err := d.Hots(context.Background())
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestTagTop(t *testing.T) {
	Convey("TagTop", t, func() {
		var (
			mid, tid int64
			rn       int
		)
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.top).Reply(200).JSON(`{
			"code": 0,
			"data": [
				79115187
			],
			"ctop": 79115187,
			"cbottom": 79115187
		}`)
		res, err := d.TagTop(context.Background(), mid, tid, rn)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestGroup(t *testing.T) {
	Convey("Group", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.group).Reply(200).JSON(`{
			"6219077": 12,
			"773939": 6,
			"40866": 9,
			"48946674": 19,
			"3336336": 14
		}`)
		res, err := d.Group(context.Background())
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestRecommendHot(t *testing.T) {
	Convey("RecommendHot", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.rcmdHot).Reply(200).JSON(`{
			"note": false,
			"source_date": "2019-12-19",
			"code": 0,
			"num": 250,
			"list": [
				{
					"aid": 79753087,
					"mid": 9824766,
					"score": 1039006,
					"desc": ""
				},
				{
					"aid": 79656005,
					"mid": 14110780,
					"score": 703814,
					"desc": ""
				},
				{
					"aid": 79644071,
					"mid": 7552204,
					"score": 646948,
					"desc": ""
				}
			]
		}`)
		res, err := d.RecommendHot(context.Background())
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

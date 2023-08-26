package recommend

import (
	"fmt"
	"testing"
	"time"

	"go-gateway/app/app-svr/app-card/interface/model"

	. "github.com/smartystreets/goconvey/convey"
	gock "gopkg.in/h2non/gock.v1"
)

func TestRecommend(t *testing.T) {
	Convey("Recommend", t, func() {
		var (
			plat                                                       int8
			buvid                                                      string
			mid                                                        int64
			build, loginEvent, parentMode, recsysMode, teenagersMode   int
			zoneID                                                     int64
			group                                                      int
			interest, network, applist                                 string
			style                                                      int
			column                                                     model.ColumnStatus
			flush, count, deviceType                                   int
			avAdResource                                               int64
			autoplay, deviceName, openEvent, bannerHash, userInterests string
			resourceID, bannerExp                                      int
			now                                                        time.Time
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
		res, err := d.Recommend(ctx(), plat, buvid, mid, build, loginEvent, parentMode, recsysMode, teenagersMode, zoneID, group, interest, network,
			style, column, flush, count, deviceType, avAdResource, autoplay, deviceName, openEvent, bannerHash, applist, userInterests, resourceID, bannerExp, now)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestTagTop(t *testing.T) {
	Convey("TagTop", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.top).Reply(200).JSON(`{
			"note": false,
			"source_date": "2019-12-06",
			"code": 0,
			"num": 500,
			"list": [
				{
					"aid": 78014605,
					"mid": 546195,
					"score": 1514314,
					"desc": ""
				},
				{
					"aid": 78128190,
					"mid": 6574487,
					"score": 1213318,
					"desc": ""
				},
				{
					"aid": 78192525,
					"mid": 129240403,
					"score": 1184036,
					"desc": ""
				},
				{
					"aid": 78172723,
					"mid": 279991456,
					"score": 714372,
					"desc": ""
				},
				{
					"aid": 78160466,
					"mid": 437316738,
					"score": 704573,
					"desc": ""
				},
				{
					"aid": 78094006,
					"mid": 26139491,
					"score": 673792,
					"desc": ""
				},
				{
					"aid": 78207123,
					"mid": 384298638,
					"score": 625409,
					"desc": ""
				},
				{
					"aid": 78143483,
					"mid": 196356191,
					"score": 581722,
					"desc": ""
				},
				{
					"aid": 78113248,
					"mid": 390461123,
					"score": 574403,
					"desc": ""
				},
				{
					"aid": 78037395,
					"mid": 168598,
					"score": 573980,
					"desc": ""
				},
				{
					"aid": 77849398,
					"mid": 485159045,
					"score": 559618,
					"desc": ""
				},
				{
					"aid": 78133226,
					"mid": 10451557,
					"score": 550123,
					"desc": ""
				},
				{
					"aid": 78095203,
					"mid": 108572682,
					"score": 549555,
					"desc": ""
				},
				{
					"aid": 78078278,
					"mid": 2378908,
					"score": 544899,
					"desc": ""
				},
				{
					"aid": 78106598,
					"mid": 337521240,
					"score": 540755,
					"desc": ""
				},
				{
					"aid": 78004733,
					"mid": 37439823,
					"score": 525948,
					"desc": ""
				},
				{
					"aid": 78135967,
					"mid": 37090048,
					"score": 517987,
					"desc": ""
				},
				{
					"aid": 77962879,
					"mid": 179512321,
					"score": 498368,
					"desc": ""
				},
				{
					"aid": 78146738,
					"mid": 10462362,
					"score": 480074,
					"desc": ""
				},
				{
					"aid": 78106861,
					"mid": 482917999,
					"score": 476896,
					"desc": ""
				},
				{
					"aid": 78109086,
					"mid": 11403305,
					"score": 447338,
					"desc": ""
				},
				{
					"aid": 78152802,
					"mid": 54992199,
					"score": 434729,
					"desc": ""
				},
				{
					"aid": 77373228,
					"mid": 632887,
					"score": 434609,
					"desc": ""
				},
				{
					"aid": 78031734,
					"mid": 222103174,
					"score": 391988,
					"desc": ""
				},
				{
					"aid": 78048985,
					"mid": 471902481,
					"score": 384409,
					"desc": ""
				},
				{
					"aid": 78005909,
					"mid": 193262326,
					"score": 384382,
					"desc": ""
				},
				{
					"aid": 78079036,
					"mid": 7560829,
					"score": 379782,
					"desc": ""
				},
				{
					"aid": 78211736,
					"mid": 50329118,
					"score": 373959,
					"desc": ""
				},
				{
					"aid": 78100473,
					"mid": 268810504,
					"score": 362910,
					"desc": ""
				},
				{
					"aid": 78122450,
					"mid": 290526283,
					"score": 360340,
					"desc": ""
				},
				{
					"aid": 78200247,
					"mid": 25503580,
					"score": 356701,
					"desc": ""
				},
				{
					"aid": 78157203,
					"mid": 96793524,
					"score": 355231,
					"desc": ""
				},
				{
					"aid": 78142831,
					"mid": 94281836,
					"score": 344522,
					"desc": ""
				},
				{
					"aid": 78118341,
					"mid": 79061224,
					"score": 339154,
					"desc": ""
				},
				{
					"aid": 77939471,
					"mid": 2072832,
					"score": 334802,
					"desc": ""
				},
				{
					"aid": 78143018,
					"mid": 382193067,
					"score": 332155,
					"desc": ""
				},
				{
					"aid": 78109738,
					"mid": 439021394,
					"score": 332008,
					"desc": ""
				},
				{
					"aid": 78155850,
					"mid": 50329118,
					"score": 331672,
					"desc": ""
				},
				{
					"aid": 78083909,
					"mid": 384298638,
					"score": 328051,
					"desc": ""
				},
				{
					"aid": 78088286,
					"mid": 37260118,
					"score": 327461,
					"desc": ""
				},
				{
					"aid": 78207759,
					"mid": 808171,
					"score": 325837,
					"desc": ""
				},
				{
					"aid": 78198239,
					"mid": 3301199,
					"score": 324581,
					"desc": ""
				},
				{
					"aid": 78147146,
					"mid": 29296192,
					"score": 321068,
					"desc": ""
				},
				{
					"aid": 78156597,
					"mid": 7953030,
					"score": 315891,
					"desc": ""
				},
				{
					"aid": 78161963,
					"mid": 1935882,
					"score": 311219,
					"desc": ""
				},
				{
					"aid": 78169726,
					"mid": 456664753,
					"score": 310521,
					"desc": ""
				},
				{
					"aid": 78140550,
					"mid": 3331521,
					"score": 309058,
					"desc": ""
				},
				{
					"aid": 78152718,
					"mid": 374716732,
					"score": 308931,
					"desc": ""
				},
				{
					"aid": 78003966,
					"mid": 7560829,
					"score": 306986,
					"desc": ""
				},
				{
					"aid": 77887230,
					"mid": 79577853,
					"score": 306571,
					"desc": ""
				},
				{
					"aid": 78130724,
					"mid": 415479453,
					"score": 306378,
					"desc": ""
				},
				{
					"aid": 78020360,
					"mid": 10851726,
					"score": 301251,
					"desc": ""
				},
				{
					"aid": 77905674,
					"mid": 1958342,
					"score": 299356,
					"desc": ""
				},
				{
					"aid": 77963556,
					"mid": 14110780,
					"score": 295106,
					"desc": ""
				},
				{
					"aid": 78125911,
					"mid": 107486042,
					"score": 292974,
					"desc": ""
				},
				{
					"aid": 78168867,
					"mid": 893053,
					"score": 291997,
					"desc": ""
				},
				{
					"aid": 77895236,
					"mid": 12394995,
					"score": 289577,
					"desc": ""
				},
				{
					"aid": 78120411,
					"mid": 32365949,
					"score": 280801,
					"desc": ""
				},
				{
					"aid": 78120563,
					"mid": 94114029,
					"score": 280039,
					"desc": ""
				},
				{
					"aid": 78047539,
					"mid": 455876411,
					"score": 268811,
					"desc": ""
				},
				{
					"aid": 78157014,
					"mid": 176756724,
					"score": 266029,
					"desc": ""
				},
				{
					"aid": 78126101,
					"mid": 324753357,
					"score": 265927,
					"desc": ""
				},
				{
					"aid": 78009071,
					"mid": 61196169,
					"score": 260951,
					"desc": ""
				},
				{
					"aid": 78004647,
					"mid": 250648682,
					"score": 257070,
					"desc": ""
				},
				{
					"aid": 78133760,
					"mid": 8960728,
					"score": 256512,
					"desc": ""
				},
				{
					"aid": 78036563,
					"mid": 431047137,
					"score": 243225,
					"desc": ""
				},
				{
					"aid": 77921590,
					"mid": 443182852,
					"score": 239616,
					"desc": ""
				},
				{
					"aid": 78137495,
					"mid": 416128940,
					"score": 238828,
					"desc": ""
				},
				{
					"aid": 77927319,
					"mid": 2986310,
					"score": 237034,
					"desc": ""
				},
				{
					"aid": 78158523,
					"mid": 5957440,
					"score": 236816,
					"desc": ""
				},
				{
					"aid": 77779131,
					"mid": 401372201,
					"score": 235043,
					"desc": ""
				},
				{
					"aid": 77920056,
					"mid": 324753357,
					"score": 234739,
					"desc": ""
				},
				{
					"aid": 77810653,
					"mid": 12394995,
					"score": 227238,
					"desc": ""
				},
				{
					"aid": 77715541,
					"mid": 17411953,
					"score": 223371,
					"desc": ""
				},
				{
					"aid": 78149044,
					"mid": 43111066,
					"score": 219988,
					"desc": ""
				},
				{
					"aid": 78151394,
					"mid": 3221649,
					"score": 216967,
					"desc": ""
				},
				{
					"aid": 78128571,
					"mid": 480366389,
					"score": 213997,
					"desc": ""
				},
				{
					"aid": 78188176,
					"mid": 11336264,
					"score": 213920,
					"desc": ""
				},
				{
					"aid": 78047946,
					"mid": 19308863,
					"score": 210406,
					"desc": ""
				},
				{
					"aid": 77996215,
					"mid": 412633068,
					"score": 207797,
					"desc": ""
				},
				{
					"aid": 77930202,
					"mid": 294152720,
					"score": 207603,
					"desc": ""
				},
				{
					"aid": 78057915,
					"mid": 253350665,
					"score": 207591,
					"desc": ""
				},
				{
					"aid": 77985582,
					"mid": 26027608,
					"score": 206402,
					"desc": ""
				},
				{
					"aid": 78108476,
					"mid": 20165629,
					"score": 204366,
					"desc": ""
				},
				{
					"aid": 77794536,
					"mid": 281039,
					"score": 204310,
					"desc": ""
				},
				{
					"aid": 77927663,
					"mid": 131452749,
					"score": 203785,
					"desc": ""
				},
				{
					"aid": 78105772,
					"mid": 40966108,
					"score": 201866,
					"desc": ""
				},
				{
					"aid": 78086924,
					"mid": 18491201,
					"score": 199686,
					"desc": ""
				},
				{
					"aid": 78030991,
					"mid": 414641554,
					"score": 199572,
					"desc": ""
				},
				{
					"aid": 78107624,
					"mid": 15183062,
					"score": 196731,
					"desc": ""
				},
				{
					"aid": 78055327,
					"mid": 2206456,
					"score": 194432,
					"desc": ""
				},
				{
					"aid": 77935473,
					"mid": 65538469,
					"score": 192955,
					"desc": ""
				},
				{
					"aid": 78154162,
					"mid": 29440965,
					"score": 190867,
					"desc": ""
				},
				{
					"aid": 78095496,
					"mid": 16307541,
					"score": 190559,
					"desc": ""
				},
				{
					"aid": 78161258,
					"mid": 10330740,
					"score": 189751,
					"desc": ""
				},
				{
					"aid": 78091516,
					"mid": 10139490,
					"score": 185212,
					"desc": ""
				},
				{
					"aid": 78034506,
					"mid": 450979444,
					"score": 183443,
					"desc": ""
				},
				{
					"aid": 77940931,
					"mid": 19515012,
					"score": 180032,
					"desc": ""
				},
				{
					"aid": 78158195,
					"mid": 415479453,
					"score": 178852,
					"desc": ""
				},
				{
					"aid": 77938993,
					"mid": 203337614,
					"score": 178523,
					"desc": ""
				},
				{
					"aid": 77877157,
					"mid": 258457966,
					"score": 176295,
					"desc": ""
				},
				{
					"aid": 78170734,
					"mid": 427841873,
					"score": 168260,
					"desc": ""
				},
				{
					"aid": 78108022,
					"mid": 20165629,
					"score": 168177,
					"desc": ""
				},
				{
					"aid": 78090377,
					"mid": 434334701,
					"score": 167042,
					"desc": ""
				},
				{
					"aid": 78081348,
					"mid": 125526,
					"score": 166827,
					"desc": ""
				},
				{
					"aid": 78059508,
					"mid": 16720403,
					"score": 166474,
					"desc": ""
				},
				{
					"aid": 77891986,
					"mid": 373867,
					"score": 166282,
					"desc": ""
				},
				{
					"aid": 78148363,
					"mid": 357440158,
					"score": 166231,
					"desc": ""
				},
				{
					"aid": 78046987,
					"mid": 396848107,
					"score": 166057,
					"desc": ""
				},
				{
					"aid": 78155845,
					"mid": 231048690,
					"score": 164226,
					"desc": ""
				},
				{
					"aid": 78149498,
					"mid": 392876519,
					"score": 161903,
					"desc": ""
				},
				{
					"aid": 77754245,
					"mid": 124633700,
					"score": 160527,
					"desc": ""
				},
				{
					"aid": 78112387,
					"mid": 390371228,
					"score": 160045,
					"desc": ""
				},
				{
					"aid": 77912295,
					"mid": 8076065,
					"score": 159927,
					"desc": ""
				},
				{
					"aid": 78120805,
					"mid": 7815300,
					"score": 159631,
					"desc": ""
				},
				{
					"aid": 78004916,
					"mid": 927587,
					"score": 154831,
					"desc": ""
				},
				{
					"aid": 78081434,
					"mid": 320491072,
					"score": 154567,
					"desc": ""
				},
				{
					"aid": 78166446,
					"mid": 1950209,
					"score": 154410,
					"desc": ""
				},
				{
					"aid": 78031102,
					"mid": 386043247,
					"score": 152320,
					"desc": ""
				},
				{
					"aid": 77714616,
					"mid": 45882815,
					"score": 152237,
					"desc": ""
				},
				{
					"aid": 77875670,
					"mid": 382189062,
					"score": 151600,
					"desc": ""
				},
				{
					"aid": 78095908,
					"mid": 258457966,
					"score": 151504,
					"desc": ""
				},
				{
					"aid": 78170273,
					"mid": 5646546,
					"score": 151062,
					"desc": ""
				},
				{
					"aid": 77874289,
					"mid": 29329085,
					"score": 150478,
					"desc": ""
				},
				{
					"aid": 78131498,
					"mid": 33683045,
					"score": 149591,
					"desc": ""
				},
				{
					"aid": 77893589,
					"mid": 15324420,
					"score": 149266,
					"desc": ""
				},
				{
					"aid": 78149793,
					"mid": 303469491,
					"score": 148408,
					"desc": ""
				},
				{
					"aid": 78158507,
					"mid": 35359510,
					"score": 147011,
					"desc": ""
				},
				{
					"aid": 78157851,
					"mid": 8784855,
					"score": 146759,
					"desc": ""
				},
				{
					"aid": 77813314,
					"mid": 379963545,
					"score": 146118,
					"desc": ""
				},
				{
					"aid": 77959619,
					"mid": 54992199,
					"score": 146032,
					"desc": ""
				},
				{
					"aid": 77918627,
					"mid": 145149047,
					"score": 142488,
					"desc": ""
				},
				{
					"aid": 77730023,
					"mid": 157056,
					"score": 142360,
					"desc": ""
				},
				{
					"aid": 78041651,
					"mid": 279583114,
					"score": 142291,
					"desc": ""
				},
				{
					"aid": 77956746,
					"mid": 330383888,
					"score": 141130,
					"desc": ""
				},
				{
					"aid": 78033382,
					"mid": 37781521,
					"score": 140711,
					"desc": ""
				},
				{
					"aid": 78157630,
					"mid": 387086717,
					"score": 140167,
					"desc": ""
				},
				{
					"aid": 78096219,
					"mid": 22042016,
					"score": 138969,
					"desc": ""
				},
				{
					"aid": 78139605,
					"mid": 8112659,
					"score": 136147,
					"desc": ""
				},
				{
					"aid": 78016903,
					"mid": 427494870,
					"score": 135893,
					"desc": ""
				},
				{
					"aid": 78175829,
					"mid": 383326579,
					"score": 133591,
					"desc": ""
				},
				{
					"aid": 78046182,
					"mid": 11061327,
					"score": 133400,
					"desc": ""
				},
				{
					"aid": 77970251,
					"mid": 4271117,
					"score": 132587,
					"desc": ""
				},
				{
					"aid": 77989109,
					"mid": 13354765,
					"score": 130542,
					"desc": ""
				},
				{
					"aid": 78160602,
					"mid": 280421456,
					"score": 129758,
					"desc": ""
				},
				{
					"aid": 78212320,
					"mid": 482917999,
					"score": 128787,
					"desc": ""
				},
				{
					"aid": 78220009,
					"mid": 10500463,
					"score": 128360,
					"desc": ""
				},
				{
					"aid": 77988803,
					"mid": 176037767,
					"score": 127588,
					"desc": ""
				},
				{
					"aid": 77680413,
					"mid": 5465995,
					"score": 127175,
					"desc": ""
				},
				{
					"aid": 78017420,
					"mid": 7792521,
					"score": 126896,
					"desc": ""
				},
				{
					"aid": 77835417,
					"mid": 253584248,
					"score": 126871,
					"desc": ""
				},
				{
					"aid": 77687380,
					"mid": 19071708,
					"score": 126869,
					"desc": ""
				},
				{
					"aid": 77967886,
					"mid": 109204686,
					"score": 126231,
					"desc": ""
				},
				{
					"aid": 77392684,
					"mid": 4162287,
					"score": 124835,
					"desc": ""
				},
				{
					"aid": 77860443,
					"mid": 432048447,
					"score": 124231,
					"desc": ""
				},
				{
					"aid": 78044062,
					"mid": 6739643,
					"score": 123836,
					"desc": ""
				},
				{
					"aid": 78109900,
					"mid": 2379178,
					"score": 122943,
					"desc": ""
				},
				{
					"aid": 78115060,
					"mid": 10330740,
					"score": 122931,
					"desc": ""
				},
				{
					"aid": 78152241,
					"mid": 33253839,
					"score": 122175,
					"desc": ""
				},
				{
					"aid": 78128363,
					"mid": 402663819,
					"score": 122160,
					"desc": ""
				},
				{
					"aid": 78152671,
					"mid": 37839963,
					"score": 122081,
					"desc": ""
				},
				{
					"aid": 77810607,
					"mid": 207413896,
					"score": 120836,
					"desc": ""
				},
				{
					"aid": 77853663,
					"mid": 386614469,
					"score": 119737,
					"desc": ""
				},
				{
					"aid": 78176663,
					"mid": 475961,
					"score": 118376,
					"desc": ""
				},
				{
					"aid": 78136291,
					"mid": 328531988,
					"score": 118309,
					"desc": ""
				},
				{
					"aid": 77894117,
					"mid": 79061224,
					"score": 118124,
					"desc": ""
				},
				{
					"aid": 78035986,
					"mid": 7487399,
					"score": 116811,
					"desc": ""
				},
				{
					"aid": 78095883,
					"mid": 258457966,
					"score": 116285,
					"desc": ""
				},
				{
					"aid": 78137012,
					"mid": 20601151,
					"score": 115553,
					"desc": ""
				},
				{
					"aid": 78172247,
					"mid": 2539073,
					"score": 115242,
					"desc": ""
				},
				{
					"aid": 77625547,
					"mid": 382986457,
					"score": 114753,
					"desc": ""
				},
				{
					"aid": 77903095,
					"mid": 274595297,
					"score": 114708,
					"desc": ""
				},
				{
					"aid": 78108338,
					"mid": 16720403,
					"score": 113642,
					"desc": ""
				},
				{
					"aid": 78162340,
					"mid": 19071708,
					"score": 112025,
					"desc": ""
				},
				{
					"aid": 77401671,
					"mid": 12394995,
					"score": 111854,
					"desc": ""
				},
				{
					"aid": 78146680,
					"mid": 95845925,
					"score": 111810,
					"desc": ""
				},
				{
					"aid": 77789052,
					"mid": 8150009,
					"score": 111688,
					"desc": ""
				},
				{
					"aid": 78010707,
					"mid": 18248995,
					"score": 111414,
					"desc": ""
				},
				{
					"aid": 78217225,
					"mid": 26023642,
					"score": 111318,
					"desc": ""
				},
				{
					"aid": 77810815,
					"mid": 390371228,
					"score": 109714,
					"desc": ""
				},
				{
					"aid": 77458533,
					"mid": 7817472,
					"score": 108987,
					"desc": ""
				},
				{
					"aid": 78050712,
					"mid": 479277242,
					"score": 107172,
					"desc": ""
				},
				{
					"aid": 77934176,
					"mid": 22553659,
					"score": 106744,
					"desc": ""
				},
				{
					"aid": 77738647,
					"mid": 427368140,
					"score": 106656,
					"desc": ""
				},
				{
					"aid": 78144674,
					"mid": 27374685,
					"score": 105732,
					"desc": ""
				},
				{
					"aid": 78007851,
					"mid": 1858682,
					"score": 103852,
					"desc": ""
				},
				{
					"aid": 78012677,
					"mid": 13043933,
					"score": 103595,
					"desc": ""
				},
				{
					"aid": 77574152,
					"mid": 2342610,
					"score": 103572,
					"desc": ""
				},
				{
					"aid": 78120578,
					"mid": 946974,
					"score": 103388,
					"desc": ""
				},
				{
					"aid": 78029112,
					"mid": 2680131,
					"score": 102864,
					"desc": ""
				},
				{
					"aid": 77921384,
					"mid": 74450120,
					"score": 102850,
					"desc": ""
				},
				{
					"aid": 78060096,
					"mid": 287795639,
					"score": 102493,
					"desc": ""
				},
				{
					"aid": 31273868,
					"mid": 97177641,
					"score": 102260,
					"desc": ""
				},
				{
					"aid": 78162752,
					"mid": 388784772,
					"score": 102044,
					"desc": ""
				},
				{
					"aid": 78112102,
					"mid": 336731767,
					"score": 101742,
					"desc": ""
				},
				{
					"aid": 78130208,
					"mid": 10330740,
					"score": 101673,
					"desc": ""
				},
				{
					"aid": 77876675,
					"mid": 258457966,
					"score": 101657,
					"desc": ""
				},
				{
					"aid": 78194984,
					"mid": 107436435,
					"score": 101455,
					"desc": ""
				},
				{
					"aid": 78007744,
					"mid": 85898420,
					"score": 101103,
					"desc": ""
				},
				{
					"aid": 77944847,
					"mid": 17819768,
					"score": 100853,
					"desc": ""
				},
				{
					"aid": 77954817,
					"mid": 427841873,
					"score": 99495,
					"desc": ""
				},
				{
					"aid": 78084571,
					"mid": 4370617,
					"score": 99476,
					"desc": ""
				},
				{
					"aid": 78127615,
					"mid": 1893045,
					"score": 98921,
					"desc": ""
				},
				{
					"aid": 78190779,
					"mid": 481391705,
					"score": 98302,
					"desc": ""
				},
				{
					"aid": 78203392,
					"mid": 75,
					"score": 97652,
					"desc": ""
				},
				{
					"aid": 78074457,
					"mid": 47291,
					"score": 97510,
					"desc": ""
				},
				{
					"aid": 78183476,
					"mid": 405981431,
					"score": 97263,
					"desc": ""
				},
				{
					"aid": 77912036,
					"mid": 25911961,
					"score": 97127,
					"desc": ""
				},
				{
					"aid": 78148575,
					"mid": 50660116,
					"score": 96995,
					"desc": ""
				},
				{
					"aid": 77998490,
					"mid": 14890867,
					"score": 96656,
					"desc": ""
				},
				{
					"aid": 78150770,
					"mid": 322010320,
					"score": 96651,
					"desc": ""
				},
				{
					"aid": 77830714,
					"mid": 7552204,
					"score": 96415,
					"desc": ""
				},
				{
					"aid": 77985384,
					"mid": 607588,
					"score": 96220,
					"desc": ""
				},
				{
					"aid": 78190915,
					"mid": 102984190,
					"score": 95940,
					"desc": ""
				},
				{
					"aid": 78118987,
					"mid": 259333,
					"score": 95870,
					"desc": ""
				},
				{
					"aid": 78046470,
					"mid": 1935882,
					"score": 95466,
					"desc": ""
				},
				{
					"aid": 77693891,
					"mid": 39422678,
					"score": 95183,
					"desc": ""
				},
				{
					"aid": 78028550,
					"mid": 299732210,
					"score": 94851,
					"desc": ""
				},
				{
					"aid": 77803093,
					"mid": 25623387,
					"score": 94349,
					"desc": ""
				},
				{
					"aid": 78200092,
					"mid": 286700005,
					"score": 94141,
					"desc": ""
				},
				{
					"aid": 78204150,
					"mid": 43222001,
					"score": 93859,
					"desc": ""
				},
				{
					"aid": 78152481,
					"mid": 42870908,
					"score": 93739,
					"desc": ""
				},
				{
					"aid": 78148642,
					"mid": 6534506,
					"score": 93734,
					"desc": ""
				},
				{
					"aid": 78067912,
					"mid": 123064133,
					"score": 93091,
					"desc": ""
				},
				{
					"aid": 78160030,
					"mid": 30625977,
					"score": 92925,
					"desc": ""
				},
				{
					"aid": 77912034,
					"mid": 456664753,
					"score": 92749,
					"desc": ""
				},
				{
					"aid": 77966775,
					"mid": 392597589,
					"score": 91945,
					"desc": ""
				},
				{
					"aid": 77826157,
					"mid": 26931963,
					"score": 91313,
					"desc": ""
				},
				{
					"aid": 77898959,
					"mid": 9285234,
					"score": 91164,
					"desc": ""
				},
				{
					"aid": 78161176,
					"mid": 325272035,
					"score": 90699,
					"desc": ""
				},
				{
					"aid": 78143767,
					"mid": 10330740,
					"score": 90078,
					"desc": ""
				},
				{
					"aid": 78011672,
					"mid": 39304265,
					"score": 88522,
					"desc": ""
				},
				{
					"aid": 78234712,
					"mid": 388063772,
					"score": 87616,
					"desc": ""
				},
				{
					"aid": 77925899,
					"mid": 562197,
					"score": 87504,
					"desc": ""
				},
				{
					"aid": 77829019,
					"mid": 2920960,
					"score": 87361,
					"desc": ""
				},
				{
					"aid": 78177884,
					"mid": 427652006,
					"score": 87318,
					"desc": ""
				},
				{
					"aid": 78045747,
					"mid": 36416153,
					"score": 87201,
					"desc": ""
				},
				{
					"aid": 77929401,
					"mid": 20165629,
					"score": 86902,
					"desc": ""
				},
				{
					"aid": 78036347,
					"mid": 432182566,
					"score": 86783,
					"desc": ""
				},
				{
					"aid": 78154961,
					"mid": 471902481,
					"score": 86623,
					"desc": ""
				},
				{
					"aid": 78021337,
					"mid": 472807480,
					"score": 86535,
					"desc": ""
				},
				{
					"aid": 77991103,
					"mid": 154021609,
					"score": 86153,
					"desc": ""
				},
				{
					"aid": 78143234,
					"mid": 429132321,
					"score": 85246,
					"desc": ""
				},
				{
					"aid": 78003913,
					"mid": 3682229,
					"score": 84496,
					"desc": ""
				},
				{
					"aid": 77608305,
					"mid": 39627524,
					"score": 84384,
					"desc": ""
				},
				{
					"aid": 77637120,
					"mid": 355262034,
					"score": 84302,
					"desc": ""
				},
				{
					"aid": 78205505,
					"mid": 50111839,
					"score": 84139,
					"desc": ""
				},
				{
					"aid": 77873494,
					"mid": 254233663,
					"score": 84101,
					"desc": ""
				},
				{
					"aid": 78173334,
					"mid": 303469491,
					"score": 83671,
					"desc": ""
				},
				{
					"aid": 77598202,
					"mid": 384298638,
					"score": 83666,
					"desc": ""
				},
				{
					"aid": 78120024,
					"mid": 50063223,
					"score": 83497,
					"desc": ""
				},
				{
					"aid": 78045688,
					"mid": 54992199,
					"score": 83318,
					"desc": ""
				},
				{
					"aid": 77934749,
					"mid": 437316738,
					"score": 82903,
					"desc": ""
				},
				{
					"aid": 77941687,
					"mid": 26408970,
					"score": 82632,
					"desc": ""
				},
				{
					"aid": 78173376,
					"mid": 1718674,
					"score": 82418,
					"desc": ""
				},
				{
					"aid": 78149716,
					"mid": 5581898,
					"score": 81525,
					"desc": ""
				},
				{
					"aid": 77938799,
					"mid": 10119428,
					"score": 80838,
					"desc": ""
				},
				{
					"aid": 77906550,
					"mid": 23212442,
					"score": 80020,
					"desc": ""
				},
				{
					"aid": 78111067,
					"mid": 256283682,
					"score": 79928,
					"desc": ""
				},
				{
					"aid": 77918141,
					"mid": 330383888,
					"score": 79705,
					"desc": ""
				},
				{
					"aid": 77878703,
					"mid": 99336697,
					"score": 79675,
					"desc": ""
				},
				{
					"aid": 77930610,
					"mid": 50063223,
					"score": 79512,
					"desc": ""
				},
				{
					"aid": 78102110,
					"mid": 194061276,
					"score": 79404,
					"desc": ""
				},
				{
					"aid": 77910067,
					"mid": 99617722,
					"score": 79401,
					"desc": ""
				},
				{
					"aid": 78037576,
					"mid": 430641591,
					"score": 79343,
					"desc": ""
				},
				{
					"aid": 78175831,
					"mid": 398582016,
					"score": 79258,
					"desc": ""
				},
				{
					"aid": 77862151,
					"mid": 19577966,
					"score": 78924,
					"desc": ""
				},
				{
					"aid": 77884148,
					"mid": 15503317,
					"score": 78917,
					"desc": ""
				},
				{
					"aid": 77919070,
					"mid": 481018493,
					"score": 78588,
					"desc": ""
				},
				{
					"aid": 78073306,
					"mid": 279991456,
					"score": 77557,
					"desc": ""
				},
				{
					"aid": 77836327,
					"mid": 437316738,
					"score": 77422,
					"desc": ""
				},
				{
					"aid": 77975207,
					"mid": 28810067,
					"score": 77223,
					"desc": ""
				},
				{
					"aid": 78029187,
					"mid": 79061224,
					"score": 76904,
					"desc": ""
				},
				{
					"aid": 78051053,
					"mid": 95494347,
					"score": 76702,
					"desc": ""
				},
				{
					"aid": 78169403,
					"mid": 381881579,
					"score": 76629,
					"desc": ""
				},
				{
					"aid": 78055015,
					"mid": 113362335,
					"score": 76575,
					"desc": ""
				},
				{
					"aid": 78200653,
					"mid": 152389373,
					"score": 76340,
					"desc": ""
				},
				{
					"aid": 78094948,
					"mid": 35961388,
					"score": 75772,
					"desc": ""
				},
				{
					"aid": 77523825,
					"mid": 219608244,
					"score": 75493,
					"desc": ""
				},
				{
					"aid": 78158795,
					"mid": 33098691,
					"score": 75293,
					"desc": ""
				},
				{
					"aid": 77657287,
					"mid": 7071284,
					"score": 74667,
					"desc": ""
				},
				{
					"aid": 78148844,
					"mid": 454461988,
					"score": 74638,
					"desc": ""
				},
				{
					"aid": 78204298,
					"mid": 13046,
					"score": 74307,
					"desc": ""
				},
				{
					"aid": 78035474,
					"mid": 403207842,
					"score": 74263,
					"desc": ""
				},
				{
					"aid": 77871897,
					"mid": 7112166,
					"score": 74094,
					"desc": ""
				},
				{
					"aid": 77887412,
					"mid": 390461123,
					"score": 73923,
					"desc": ""
				},
				{
					"aid": 77938547,
					"mid": 37084080,
					"score": 73601,
					"desc": ""
				},
				{
					"aid": 78128977,
					"mid": 13620441,
					"score": 73571,
					"desc": ""
				},
				{
					"aid": 78097482,
					"mid": 1690404,
					"score": 73183,
					"desc": ""
				},
				{
					"aid": 78239116,
					"mid": 1935882,
					"score": 72390,
					"desc": ""
				},
				{
					"aid": 78020787,
					"mid": 39922227,
					"score": 72025,
					"desc": ""
				},
				{
					"aid": 78131735,
					"mid": 483927833,
					"score": 71903,
					"desc": ""
				},
				{
					"aid": 78138657,
					"mid": 95064321,
					"score": 71618,
					"desc": ""
				},
				{
					"aid": 78050158,
					"mid": 52001259,
					"score": 70830,
					"desc": ""
				},
				{
					"aid": 77525478,
					"mid": 249608727,
					"score": 70768,
					"desc": ""
				},
				{
					"aid": 77991367,
					"mid": 57676402,
					"score": 70232,
					"desc": ""
				},
				{
					"aid": 78074560,
					"mid": 6014992,
					"score": 70201,
					"desc": ""
				},
				{
					"aid": 77920349,
					"mid": 88671169,
					"score": 70133,
					"desc": ""
				},
				{
					"aid": 77369706,
					"mid": 395408274,
					"score": 69794,
					"desc": ""
				},
				{
					"aid": 78177150,
					"mid": 291780438,
					"score": 69762,
					"desc": ""
				},
				{
					"aid": 77854834,
					"mid": 480913392,
					"score": 69140,
					"desc": ""
				},
				{
					"aid": 78045539,
					"mid": 7057999,
					"score": 69065,
					"desc": ""
				},
				{
					"aid": 78050406,
					"mid": 382651856,
					"score": 67902,
					"desc": ""
				},
				{
					"aid": 78088453,
					"mid": 837470,
					"score": 67902,
					"desc": ""
				},
				{
					"aid": 77733003,
					"mid": 393063099,
					"score": 67803,
					"desc": ""
				},
				{
					"aid": 77984340,
					"mid": 142656172,
					"score": 67751,
					"desc": ""
				},
				{
					"aid": 78001749,
					"mid": 481838232,
					"score": 67727,
					"desc": ""
				},
				{
					"aid": 77041478,
					"mid": 165511,
					"score": 67712,
					"desc": ""
				},
				{
					"aid": 78052722,
					"mid": 21778636,
					"score": 67256,
					"desc": ""
				},
				{
					"aid": 77972972,
					"mid": 18690024,
					"score": 67251,
					"desc": ""
				},
				{
					"aid": 77468878,
					"mid": 59428434,
					"score": 67174,
					"desc": ""
				},
				{
					"aid": 78095261,
					"mid": 2859372,
					"score": 66826,
					"desc": ""
				},
				{
					"aid": 78213299,
					"mid": 34149649,
					"score": 66763,
					"desc": ""
				},
				{
					"aid": 78167600,
					"mid": 240260506,
					"score": 66420,
					"desc": ""
				},
				{
					"aid": 78119711,
					"mid": 438345816,
					"score": 66026,
					"desc": ""
				},
				{
					"aid": 78102369,
					"mid": 367927,
					"score": 65378,
					"desc": ""
				},
				{
					"aid": 77998334,
					"mid": 477631979,
					"score": 65326,
					"desc": ""
				},
				{
					"aid": 78021363,
					"mid": 155692515,
					"score": 65120,
					"desc": ""
				},
				{
					"aid": 78177131,
					"mid": 3670216,
					"score": 64894,
					"desc": ""
				},
				{
					"aid": 77933933,
					"mid": 1565155,
					"score": 64778,
					"desc": ""
				},
				{
					"aid": 78041397,
					"mid": 17546432,
					"score": 64519,
					"desc": ""
				},
				{
					"aid": 77431368,
					"mid": 27501480,
					"score": 64474,
					"desc": ""
				},
				{
					"aid": 78108516,
					"mid": 321765262,
					"score": 63968,
					"desc": ""
				},
				{
					"aid": 78010943,
					"mid": 15355639,
					"score": 63855,
					"desc": ""
				},
				{
					"aid": 77808046,
					"mid": 7309250,
					"score": 63572,
					"desc": ""
				},
				{
					"aid": 78130742,
					"mid": 5211734,
					"score": 63544,
					"desc": ""
				},
				{
					"aid": 77934020,
					"mid": 7487399,
					"score": 63363,
					"desc": ""
				},
				{
					"aid": 78153259,
					"mid": 64925986,
					"score": 63236,
					"desc": ""
				},
				{
					"aid": 77865408,
					"mid": 99015667,
					"score": 62965,
					"desc": ""
				},
				{
					"aid": 78142090,
					"mid": 60719450,
					"score": 62676,
					"desc": ""
				},
				{
					"aid": 77949654,
					"mid": 86936310,
					"score": 62654,
					"desc": ""
				},
				{
					"aid": 78133076,
					"mid": 34977510,
					"score": 62443,
					"desc": ""
				},
				{
					"aid": 78150031,
					"mid": 13904634,
					"score": 62420,
					"desc": ""
				},
				{
					"aid": 78166207,
					"mid": 8065474,
					"score": 61977,
					"desc": ""
				},
				{
					"aid": 78090555,
					"mid": 7458285,
					"score": 61802,
					"desc": ""
				},
				{
					"aid": 77438892,
					"mid": 31832612,
					"score": 60839,
					"desc": ""
				},
				{
					"aid": 77863593,
					"mid": 29302925,
					"score": 60789,
					"desc": ""
				},
				{
					"aid": 78025117,
					"mid": 168064909,
					"score": 60714,
					"desc": ""
				},
				{
					"aid": 77504043,
					"mid": 411442023,
					"score": 60189,
					"desc": ""
				},
				{
					"aid": 78046433,
					"mid": 296616625,
					"score": 59893,
					"desc": ""
				},
				{
					"aid": 78146527,
					"mid": 10094840,
					"score": 59887,
					"desc": ""
				},
				{
					"aid": 78103906,
					"mid": 349991143,
					"score": 59251,
					"desc": ""
				},
				{
					"aid": 78118207,
					"mid": 29243277,
					"score": 58718,
					"desc": ""
				},
				{
					"aid": 77970933,
					"mid": 8188433,
					"score": 58713,
					"desc": ""
				},
				{
					"aid": 77481835,
					"mid": 60068509,
					"score": 58700,
					"desc": ""
				},
				{
					"aid": 78113079,
					"mid": 129240403,
					"score": 58654,
					"desc": ""
				},
				{
					"aid": 78167041,
					"mid": 8893909,
					"score": 58596,
					"desc": ""
				},
				{
					"aid": 77687026,
					"mid": 4321133,
					"score": 58473,
					"desc": ""
				},
				{
					"aid": 78071736,
					"mid": 384298638,
					"score": 58218,
					"desc": ""
				},
				{
					"aid": 78106952,
					"mid": 408672114,
					"score": 57811,
					"desc": ""
				},
				{
					"aid": 77712427,
					"mid": 50329118,
					"score": 57635,
					"desc": ""
				},
				{
					"aid": 77772812,
					"mid": 3240555,
					"score": 57618,
					"desc": ""
				},
				{
					"aid": 77511792,
					"mid": 297242063,
					"score": 57527,
					"desc": ""
				},
				{
					"aid": 78052365,
					"mid": 430662909,
					"score": 57409,
					"desc": ""
				},
				{
					"aid": 78039859,
					"mid": 485099197,
					"score": 57330,
					"desc": ""
				},
				{
					"aid": 78049386,
					"mid": 439004370,
					"score": 57313,
					"desc": ""
				},
				{
					"aid": 78075698,
					"mid": 456664753,
					"score": 57222,
					"desc": ""
				},
				{
					"aid": 77933212,
					"mid": 32302761,
					"score": 57173,
					"desc": ""
				},
				{
					"aid": 78119177,
					"mid": 477429068,
					"score": 57089,
					"desc": ""
				},
				{
					"aid": 77840882,
					"mid": 8054411,
					"score": 57056,
					"desc": ""
				},
				{
					"aid": 77786038,
					"mid": 20679868,
					"score": 57043,
					"desc": ""
				},
				{
					"aid": 78151717,
					"mid": 6185598,
					"score": 56991,
					"desc": ""
				},
				{
					"aid": 78179606,
					"mid": 2103351,
					"score": 56851,
					"desc": ""
				},
				{
					"aid": 78187313,
					"mid": 437862593,
					"score": 56818,
					"desc": ""
				},
				{
					"aid": 77442854,
					"mid": 437316738,
					"score": 56680,
					"desc": ""
				},
				{
					"aid": 77940438,
					"mid": 14333871,
					"score": 56430,
					"desc": ""
				},
				{
					"aid": 78127715,
					"mid": 406472610,
					"score": 56218,
					"desc": ""
				},
				{
					"aid": 78161091,
					"mid": 108569350,
					"score": 56192,
					"desc": ""
				},
				{
					"aid": 78134075,
					"mid": 1581626,
					"score": 56182,
					"desc": ""
				},
				{
					"aid": 77474370,
					"mid": 346170914,
					"score": 56094,
					"desc": ""
				},
				{
					"aid": 77360276,
					"mid": 250858633,
					"score": 55621,
					"desc": ""
				},
				{
					"aid": 77760474,
					"mid": 17390165,
					"score": 55433,
					"desc": ""
				},
				{
					"aid": 77696975,
					"mid": 12394995,
					"score": 55301,
					"desc": ""
				},
				{
					"aid": 78161660,
					"mid": 55431669,
					"score": 54915,
					"desc": ""
				},
				{
					"aid": 78138294,
					"mid": 213826349,
					"score": 54882,
					"desc": ""
				},
				{
					"aid": 78029162,
					"mid": 355454313,
					"score": 53944,
					"desc": ""
				},
				{
					"aid": 78171887,
					"mid": 44473221,
					"score": 53875,
					"desc": ""
				},
				{
					"aid": 77721080,
					"mid": 2612610,
					"score": 53863,
					"desc": ""
				},
				{
					"aid": 77983809,
					"mid": 168598,
					"score": 53643,
					"desc": ""
				},
				{
					"aid": 78052923,
					"mid": 16720403,
					"score": 53416,
					"desc": ""
				},
				{
					"aid": 77685492,
					"mid": 60509716,
					"score": 52894,
					"desc": ""
				},
				{
					"aid": 78161847,
					"mid": 389858754,
					"score": 52825,
					"desc": ""
				},
				{
					"aid": 77434255,
					"mid": 13826185,
					"score": 52713,
					"desc": ""
				},
				{
					"aid": 78029657,
					"mid": 1420982,
					"score": 52497,
					"desc": ""
				},
				{
					"aid": 78020708,
					"mid": 339800866,
					"score": 52463,
					"desc": ""
				},
				{
					"aid": 78155979,
					"mid": 18739124,
					"score": 52281,
					"desc": ""
				},
				{
					"aid": 77941158,
					"mid": 6739643,
					"score": 52203,
					"desc": ""
				},
				{
					"aid": 77898292,
					"mid": 434716461,
					"score": 52166,
					"desc": ""
				},
				{
					"aid": 78001652,
					"mid": 2142762,
					"score": 52140,
					"desc": ""
				},
				{
					"aid": 77695297,
					"mid": 208557615,
					"score": 51831,
					"desc": ""
				},
				{
					"aid": 78057495,
					"mid": 5957440,
					"score": 51694,
					"desc": ""
				},
				{
					"aid": 77661110,
					"mid": 562197,
					"score": 51549,
					"desc": ""
				},
				{
					"aid": 78154775,
					"mid": 334491384,
					"score": 51480,
					"desc": ""
				},
				{
					"aid": 78105054,
					"mid": 20165629,
					"score": 51355,
					"desc": ""
				},
				{
					"aid": 78114232,
					"mid": 5260378,
					"score": 51105,
					"desc": ""
				},
				{
					"aid": 77724126,
					"mid": 477570713,
					"score": 51041,
					"desc": ""
				},
				{
					"aid": 77982065,
					"mid": 13144496,
					"score": 50696,
					"desc": ""
				},
				{
					"aid": 77916779,
					"mid": 33818200,
					"score": 50524,
					"desc": ""
				},
				{
					"aid": 77814961,
					"mid": 16853896,
					"score": 50503,
					"desc": ""
				},
				{
					"aid": 77915302,
					"mid": 58852835,
					"score": 50312,
					"desc": ""
				},
				{
					"aid": 78012637,
					"mid": 452309333,
					"score": 50057,
					"desc": ""
				},
				{
					"aid": 78140969,
					"mid": 5838917,
					"score": 49980,
					"desc": ""
				},
				{
					"aid": 77985290,
					"mid": 418062450,
					"score": 49975,
					"desc": ""
				},
				{
					"aid": 78101500,
					"mid": 24140344,
					"score": 49932,
					"desc": ""
				},
				{
					"aid": 78128198,
					"mid": 27043295,
					"score": 49811,
					"desc": ""
				},
				{
					"aid": 77654604,
					"mid": 35130689,
					"score": 49732,
					"desc": ""
				},
				{
					"aid": 77947528,
					"mid": 437842498,
					"score": 49724,
					"desc": ""
				},
				{
					"aid": 78141705,
					"mid": 22752620,
					"score": 49416,
					"desc": ""
				},
				{
					"aid": 77540232,
					"mid": 176269428,
					"score": 49387,
					"desc": ""
				},
				{
					"aid": 77883148,
					"mid": 334408083,
					"score": 49379,
					"desc": ""
				},
				{
					"aid": 77353708,
					"mid": 1712395,
					"score": 49245,
					"desc": ""
				},
				{
					"aid": 78052981,
					"mid": 319438468,
					"score": 49215,
					"desc": ""
				},
				{
					"aid": 77990434,
					"mid": 384298638,
					"score": 49124,
					"desc": ""
				},
				{
					"aid": 77932655,
					"mid": 146574258,
					"score": 49031,
					"desc": ""
				},
				{
					"aid": 78080826,
					"mid": 383724059,
					"score": 48591,
					"desc": ""
				},
				{
					"aid": 78094529,
					"mid": 64137054,
					"score": 48499,
					"desc": ""
				},
				{
					"aid": 77720885,
					"mid": 515993,
					"score": 48427,
					"desc": ""
				},
				{
					"aid": 77711928,
					"mid": 14341074,
					"score": 48369,
					"desc": ""
				},
				{
					"aid": 78146655,
					"mid": 372274816,
					"score": 48338,
					"desc": ""
				},
				{
					"aid": 77841372,
					"mid": 8111051,
					"score": 48279,
					"desc": ""
				},
				{
					"aid": 77698746,
					"mid": 57676402,
					"score": 48037,
					"desc": ""
				},
				{
					"aid": 78056962,
					"mid": 322010320,
					"score": 47953,
					"desc": ""
				},
				{
					"aid": 77476633,
					"mid": 18491201,
					"score": 47587,
					"desc": ""
				},
				{
					"aid": 77946992,
					"mid": 105502937,
					"score": 47509,
					"desc": ""
				},
				{
					"aid": 77516824,
					"mid": 424144,
					"score": 47435,
					"desc": ""
				},
				{
					"aid": 78141160,
					"mid": 443305053,
					"score": 47377,
					"desc": ""
				},
				{
					"aid": 78158656,
					"mid": 4469745,
					"score": 47323,
					"desc": ""
				},
				{
					"aid": 77849116,
					"mid": 86735192,
					"score": 47260,
					"desc": ""
				},
				{
					"aid": 78032668,
					"mid": 94281836,
					"score": 47211,
					"desc": ""
				},
				{
					"aid": 78165949,
					"mid": 322568183,
					"score": 47141,
					"desc": ""
				},
				{
					"aid": 78008105,
					"mid": 485118594,
					"score": 46779,
					"desc": ""
				},
				{
					"aid": 77923051,
					"mid": 602129,
					"score": 46742,
					"desc": ""
				},
				{
					"aid": 78218029,
					"mid": 12018,
					"score": 46520,
					"desc": ""
				},
				{
					"aid": 77645137,
					"mid": 3205593,
					"score": 46338,
					"desc": ""
				},
				{
					"aid": 77948365,
					"mid": 52914908,
					"score": 46332,
					"desc": ""
				},
				{
					"aid": 77663412,
					"mid": 269869945,
					"score": 45993,
					"desc": ""
				},
				{
					"aid": 77443306,
					"mid": 352442341,
					"score": 45922,
					"desc": ""
				},
				{
					"aid": 77800629,
					"mid": 69205685,
					"score": 45780,
					"desc": ""
				},
				{
					"aid": 77752028,
					"mid": 25503580,
					"score": 45712,
					"desc": ""
				},
				{
					"aid": 78018384,
					"mid": 72956117,
					"score": 45504,
					"desc": ""
				},
				{
					"aid": 78159548,
					"mid": 16364623,
					"score": 45439,
					"desc": ""
				},
				{
					"aid": 78153310,
					"mid": 909376,
					"score": 45239,
					"desc": ""
				},
				{
					"aid": 77839188,
					"mid": 37090048,
					"score": 45119,
					"desc": ""
				},
				{
					"aid": 78051884,
					"mid": 101229184,
					"score": 44878,
					"desc": ""
				},
				{
					"aid": 78038761,
					"mid": 6101834,
					"score": 44796,
					"desc": ""
				},
				{
					"aid": 77907394,
					"mid": 91399769,
					"score": 44740,
					"desc": ""
				},
				{
					"aid": 77947269,
					"mid": 361511359,
					"score": 44694,
					"desc": ""
				},
				{
					"aid": 78136954,
					"mid": 25631197,
					"score": 44651,
					"desc": ""
				},
				{
					"aid": 77538216,
					"mid": 4474705,
					"score": 44543,
					"desc": ""
				},
				{
					"aid": 78092074,
					"mid": 93415,
					"score": 44468,
					"desc": ""
				},
				{
					"aid": 77553803,
					"mid": 130220798,
					"score": 44161,
					"desc": ""
				},
				{
					"aid": 77926554,
					"mid": 3957971,
					"score": 44138,
					"desc": ""
				},
				{
					"aid": 77909233,
					"mid": 346826376,
					"score": 43953,
					"desc": ""
				},
				{
					"aid": 77975618,
					"mid": 471507721,
					"score": 43953,
					"desc": ""
				},
				{
					"aid": 78098954,
					"mid": 375504219,
					"score": 43947,
					"desc": ""
				},
				{
					"aid": 77980934,
					"mid": 314216,
					"score": 43426,
					"desc": ""
				},
				{
					"aid": 78011377,
					"mid": 46708782,
					"score": 43407,
					"desc": ""
				},
				{
					"aid": 77828723,
					"mid": 16419172,
					"score": 43388,
					"desc": ""
				},
				{
					"aid": 77901245,
					"mid": 393293810,
					"score": 43155,
					"desc": ""
				},
				{
					"aid": 77512431,
					"mid": 37291782,
					"score": 42876,
					"desc": ""
				},
				{
					"aid": 78153473,
					"mid": 37128669,
					"score": 42870,
					"desc": ""
				},
				{
					"aid": 77937395,
					"mid": 415479453,
					"score": 42859,
					"desc": ""
				},
				{
					"aid": 77565506,
					"mid": 31578954,
					"score": 42744,
					"desc": ""
				},
				{
					"aid": 77496384,
					"mid": 30222835,
					"score": 42367,
					"desc": ""
				},
				{
					"aid": 77461638,
					"mid": 281135499,
					"score": 42114,
					"desc": ""
				},
				{
					"aid": 77587303,
					"mid": 270494247,
					"score": 42033,
					"desc": ""
				},
				{
					"aid": 77756638,
					"mid": 431039037,
					"score": 41978,
					"desc": ""
				},
				{
					"aid": 77988314,
					"mid": 395936853,
					"score": 41977,
					"desc": ""
				},
				{
					"aid": 78033571,
					"mid": 138885741,
					"score": 41946,
					"desc": ""
				},
				{
					"aid": 78098117,
					"mid": 39893537,
					"score": 41915,
					"desc": ""
				},
				{
					"aid": 77758850,
					"mid": 385965342,
					"score": 41695,
					"desc": ""
				},
				{
					"aid": 77483006,
					"mid": 16726189,
					"score": 41674,
					"desc": ""
				},
				{
					"aid": 78011204,
					"mid": 1858682,
					"score": 41634,
					"desc": ""
				},
				{
					"aid": 77698716,
					"mid": 65538469,
					"score": 41623,
					"desc": ""
				},
				{
					"aid": 78163359,
					"mid": 471792754,
					"score": 41623,
					"desc": ""
				},
				{
					"aid": 77630487,
					"mid": 3916908,
					"score": 41481,
					"desc": ""
				},
				{
					"aid": 77437014,
					"mid": 38802656,
					"score": 41446,
					"desc": ""
				},
				{
					"aid": 77486315,
					"mid": 1581626,
					"score": 41156,
					"desc": ""
				},
				{
					"aid": 77785738,
					"mid": 10278125,
					"score": 40987,
					"desc": ""
				},
				{
					"aid": 78094008,
					"mid": 441381282,
					"score": 40951,
					"desc": ""
				},
				{
					"aid": 78046317,
					"mid": 37254001,
					"score": 40891,
					"desc": ""
				},
				{
					"aid": 78150221,
					"mid": 327750329,
					"score": 40798,
					"desc": ""
				},
				{
					"aid": 77811339,
					"mid": 386565895,
					"score": 40740,
					"desc": ""
				},
				{
					"aid": 77402472,
					"mid": 11831050,
					"score": 40664,
					"desc": ""
				},
				{
					"aid": 77686601,
					"mid": 19642758,
					"score": 40592,
					"desc": ""
				},
				{
					"aid": 77810146,
					"mid": 215190289,
					"score": 40304,
					"desc": ""
				},
				{
					"aid": 78090054,
					"mid": 360739161,
					"score": 40300,
					"desc": ""
				},
				{
					"aid": 77882339,
					"mid": 9131680,
					"score": 40177,
					"desc": ""
				},
				{
					"aid": 77883847,
					"mid": 1295350,
					"score": 40074,
					"desc": ""
				},
				{
					"aid": 78097306,
					"mid": 3974880,
					"score": 39884,
					"desc": ""
				},
				{
					"aid": 78153342,
					"mid": 11357018,
					"score": 39652,
					"desc": ""
				},
				{
					"aid": 78031899,
					"mid": 357451181,
					"score": 39409,
					"desc": ""
				},
				{
					"aid": 77490981,
					"mid": 808171,
					"score": 39354,
					"desc": ""
				},
				{
					"aid": 78049509,
					"mid": 1950209,
					"score": 39321,
					"desc": ""
				},
				{
					"aid": 78045659,
					"mid": 15816282,
					"score": 39281,
					"desc": ""
				},
				{
					"aid": 78048842,
					"mid": 35359510,
					"score": 39080,
					"desc": ""
				},
				{
					"aid": 77943064,
					"mid": 16720403,
					"score": 38825,
					"desc": ""
				},
				{
					"aid": 77542288,
					"mid": 337019942,
					"score": 38585,
					"desc": ""
				},
				{
					"aid": 77482692,
					"mid": 24360247,
					"score": 38373,
					"desc": ""
				},
				{
					"aid": 77474704,
					"mid": 38351330,
					"score": 38354,
					"desc": ""
				}
			]
		}`)
		var (
			mid, tid int64
			rn       int
		)
		res, err := d.TagTop(ctx(), mid, tid, rn)
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestGroup(t *testing.T) {
	Convey("Group", t, func() {
		d.clientAsyn.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.group).Reply(200).JSON(`{
			"6219077": 12,
			"773939": 6,
			"40866": 9,
			"48946674": 19,
			"9004197": 17,
			"24841473": 12,
			"5694148": 12,
			"96239263": 6,
			"691608": 14,
			"417851": 12,
			"37220829": 5,
			"72630464": 5,
			"1481151": 5,
			"1758140": 5,
			"11164742": 5,
			"1345743": 19,
			"369373780": 5,
			"90138218": 13,
			"28009181": 13,
			"28008860": 13,
			"105232234": 12,
			"390643": 12,
			"1411291": 7,
			"12754559": 5,
			"16817687": 5,
			"529313": 5,
			"44255105": 5,
			"398644909": 5,
			"1671898": 5,
			"11270809": 5,
			"9328218": 5,
			"160257447": 4,
			"333": 7,
			"96945971": 0,
			"16685661": 13,
			"198349491": 13,
			"899512": 13,
			"692": 13,
			"496065": 13,
			"14135892": 7,
			"59709220": 17,
			"58108412": 1,
			"431797187": 12,
			"298712220": 18,
			"8218715": 7,
			"266665413": 1,
			"27637930": 17,
			"435354096": 17,
			"21673742": 0,
			"33426": 16,
			"4126121": 8,
			"482": 12,
			"615292": 1,
			"6703497": 0,
			"108669": 0,
			"388745658": 7,
			"2835423": 14,
			"626800": 10,
			"3250475": 11,
			"9599990": 13,
			"3479095": 17,
			"284226": 12,
			"48155": 12,
			"317800647": 19,
			"35053212": 12,
			"208259": 1,
			"260982266": 14,
			"132551": 1,
			"57870006": 13,
			"20980174": 0,
			"336799590": 17,
			"29562880": 11,
			"388010110": 14,
			"276486432": 18,
			"477676144": 8,
			"758462": 16,
			"7651628": 11,
			"1396885": 0,
			"7438477": 0,
			"387907443": 0,
			"3769679": 0,
			"25108815": 0,
			"3336336": 10
		}`)
		res, err := d.Group(ctx())
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestFollowModeList(t *testing.T) {
	Convey("FollowModeList", t, func() {
		d.clientAsyn.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.followModeList).Reply(200).JSON(`{
			"code": 0,
			"data": [
				10908,
				15229,
				18482,
				24301,
				28883,
				29824,
				43907,
				45265,
				48155,
				58152
			]
		}`)
		res, err := d.FollowModeList(ctx())
		So(res, ShouldNotBeNil)
		So(err, ShouldBeNil)
	})
}

func TestConvergeList(t *testing.T) {
	Convey("ConvergeList", t, func() {
		d.client.SetTransport(gock.DefaultTransport)
		httpMock("GET", d.recommand).Reply(200).JSON(`{
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
	})
	var (
		plat             int8
		buvid            string
		mid, convergeID  int64
		build, displayID int
		convergeParam    string
		convergeType     int
		now              time.Time
	)
	res, _, err := d.ConvergeList(ctx(), plat, buvid, mid, convergeID, build, displayID, convergeParam, convergeType, now)
	So(res, ShouldNotBeNil)
	So(err, ShouldBeNil)
}

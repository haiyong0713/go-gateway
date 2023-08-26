package archive

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"go-gateway/app/app-svr/app-view/interface/model/view"

	advo "git.bilibili.co/bapis/bapis-go/bcg/sunspot/ad/vo"

	"gopkg.in/h2non/gock.v1"

	"github.com/gogo/protobuf/types"
	"github.com/smartystreets/goconvey/convey"
)

func TestRelateAids(t *testing.T) {
	var (
		c   = context.TODO()
		aid = int64(1)
	)
	convey.Convey("Ping", t, func(ctx convey.C) {
		_, err := d.RelateAids(c, aid)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestNewRelateAids(t *testing.T) {
	var (
		c = context.Background()
	)
	convey.Convey("NewRelateAids", t, func(ctx convey.C) {
		_, _, err := d.NewRelateAids(c, 11, 1, 0, 1, 1, 1, 1, "", "", "", "", "", 1)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

func TestNewRelateAidsV2(t *testing.T) {
	convey.Convey("NewRelateAidsV2", t, func(ctx convey.C) {
		d.client.SetTransport(gock.DefaultTransport)
		mock := httpMock("GET", d.relateRecURL).Reply(200).JSON(`
    {
    "biz_data":{
        "code":0,
        "data":{
            "ads_control":"CjB0eXBlLmdvb2dsZWFwaXMuY29tL2JpbGliaWxpLmFkLnYxLkFkc0NvbnRyb2xEdG8=",
            "ads_info":{
                "2029":{
                    "2030":{
                        "av_id":375189419,
                        "card_index":1,
                        "card_type":6,
                        "is_ad":true,
                        "jump_url":"",
                        "promotion_target_id":"",
                        "promotion_target_type":"",
                        "source_contents":"CjN0eXBlLmdvb2dsZWFwaXMuY29tL2JpbGliaWxpLmFkLnYxLlNvdXJjZUNvbnRlbnREdG8SOQogMTYyMjEwNjczMjQ4MXExNzJhMjVhMjIyYTY2cTM0NDAQoBIYnxIgASoLCP///////////wFAAQ=="
                    }
                }
            }
        }
    },
    "biz_pk_code":1,
    "code":0,
    "dalao_exp":1,
    "data":[
        {
            "av_feature":"{\"ctr\":0.0163,\"fctr\":0.0077,\"wdlks\":0.1076,\"dlr\":0.0002,\"fls\":0.0,\"rankscore\":0.0829,\"fo\":0,\"reasontype\":0,\"fms\":0.0939,\"av_play\":50647,\"d\":\" |d 0\",\"v_cl\":\" |v_cl 9\",\"v_bl\":\" |v_bl 8\",\"v_fl\":\" |v_fl 8\",\"real_matchtype\":\" |real_matchtype 4$2$1\",\"source_len\":\" |source_len 3\",\"matchtype\":\" |matchtype 16$15$15$2\",\"nonclick_show_region_num\":\" |nonclick_show_region_num 6\",\"nonclick_show_tag_num\":\" |nonclick_show_tag_num 5\",\"m_k_word\":\" |m_k_word 美食 美食圈\",\"m_k_w\":\" |m_k_w 2\",\"ysession_state\":\" |ysession_state no_click_x\",\"dr_class_match\":\" |dr_class_match 1_1\",\"play_show_region_num\":\" |play_show_region_num 6\",\"play_show_tag_num\":\" |play_show_tag_num 6\",\"p_xsession_state\":\" |p_xsession_state play_x\",\"play_region_num\":\" |play_region_num 0\",\"play_tag_num\":\" |play_tag_num 0\",\"pr_class_match\":\" |pr_class_match 1_1_0\",\"r_m_6\":\" |r_m_6 0\",\"r_m_32\":\" |r_m_32 0\"}",
            "goto":"av",
            "id":78143682,
            "rcmd_reason":null,
            "source":"offline_tag$region_dynamic$online_av2av$long_term_tag",
            "tid":12551626,
            "trackid":"all_17.shylf-ai-recsys-87.1575623748962.517"
        },
        {
            "av_feature":"{\"ctr\":0.0259,\"fctr\":0.0144,\"wdlks\":0.0323,\"dlr\":0.0001,\"fls\":0.0,\"rankscore\":0.0567,\"fo\":0,\"reasontype\":5,\"fms\":0.0404,\"av_play\":49393,\"rid\":5,\"d\":\" |d 1\",\"v_cl\":\" |v_cl 9\",\"v_bl\":\" |v_bl 8\",\"v_fl\":\" |v_fl 9\",\"real_matchtype\":\" |real_matchtype 2$5\",\"source_len\":\" |source_len 3\",\"matchtype\":\" |matchtype 9$8$15$5\",\"nonclick_show_region_num\":\" |nonclick_show_region_num 0\",\"nonclick_show_tag_num\":\" |nonclick_show_tag_num 2\",\"m_k_word\":\" |m_k_word 明星 vlog\",\"m_k_w\":\" |m_k_w 2\",\"ysession_state\":\" |ysession_state no_click_x\",\"dr_class_match\":\" |dr_class_match 0_0\",\"play_show_region_num\":\" |play_show_region_num 0\",\"play_show_tag_num\":\" |play_show_tag_num 4\",\"p_xsession_state\":\" |p_xsession_state play_x\",\"play_region_num\":\" |play_region_num 0\",\"play_tag_num\":\" |play_tag_num 1\",\"play_tag\":\" |play_tag 2848\",\"pr_class_match\":\" |pr_class_match 0_0_0\",\"r_m_6\":\" |r_m_6 0\",\"r_m_32\":\" |r_m_32 0\"}",
            "goto":"av",
            "id":77928553,
            "rcmd_reason":{
                "content":"数码·点赞飙升",
                "corner_mark":2,
                "jumpgoto":"",
                "jumpid":0,
                "style":2
            },
            "source":"offline_tag$online_tag$new_dynamic$av_boost",
            "tid":10297101,
            "trackid":"all_17.shylf-ai-recsys-87.1575623748962.517"
        }
    ],
    "gamecard_style_exp":1,
    "play_param":0,
    "dislike_exp":1,
    "pv_feature":"{\"new\":0,\"rid\":36,\"srid\":201,\"uptime\":0,\"no1\":0,\"tfea\":0,\"exp\":0,\"rrblock\":0,\"nsrc\":0,\"rn\":0,\"fba\":1,\"ap\":0,\"sp\":\"0\",\"record\":\"|latest_avid 416819748_0 416819748_0 416819748_0 416819748_0 416819748_0 416819748_0 416819748_0 |stuff_avid 416819748_0 416819748_0 416819748_0\",\"setype\":1,\"acti\":0,\"r_pl\":\" -_- \",\"r_dm\":\" -_- \",\"r_du\":\" -_- \",\"r_lk\":\" -_- \",\"r_ag\":\" -_- \",\"nr_pl\":\" 17_17 17_17 17_17 \",\"nr_dm\":\" 9_9 9_9 9_9 \",\"nr_du\":\" 6_2 6_2 6_2 \",\"nr_lk\":\" 9_9 9_9 9_9 \",\"nr_ag\":\" 7_7 7_7 7_7 \",\"nr_pl_f_0\":\" 1 \",\"nr_dm_f_0\":\" 0.9 \",\"nr_du_f_0\":\" 0.6 \",\"nr_lk_f_0\":\" 0.9 \",\"nr_ag_f_0\":\" 0.7 \",\"nr_pl_t_0\":\" 1 \",\"nr_dm_t_0\":\" 0.9 \",\"nr_du_t_0\":\" 0.2 \",\"nr_lk_t_0\":\" 0.9 \",\"nr_ag_t_0\":\" 0.7 \",\"nr_pl_f_1\":\" 1 \",\"nr_dm_f_1\":\" 0.9 \",\"nr_du_f_1\":\" 0.6 \",\"nr_lk_f_1\":\" 0.9 \",\"nr_ag_f_1\":\" 0.7 \",\"nr_pl_t_1\":\" 1 \",\"nr_dm_t_1\":\" 0.9 \",\"nr_du_t_1\":\" 0.2 \",\"nr_lk_t_1\":\" 0.9 \",\"nr_ag_t_1\":\" 0.7 \",\"nr_pl_f_2\":\" 1 \",\"nr_dm_f_2\":\" 0.9 \",\"nr_du_f_2\":\" 0.6 \",\"nr_lk_f_2\":\" 0.9 \",\"nr_ag_f_2\":\" 0.7 \",\"nr_pl_t_2\":\" 1 \",\"nr_dm_t_2\":\" 0.9 \",\"nr_du_t_2\":\" 0.2 \",\"nr_lk_t_2\":\" 0.9 \",\"nr_ag_t_2\":\" 0.7 \",\"r_pl_f_0\":\" -1 \",\"r_dm_f_0\":\" -1 \",\"r_du_f_0\":\" -1 \",\"r_lk_f_0\":\" -1 \",\"r_ag_f_0\":\" -1 \",\"r_pl_t_0\":\" -1 \",\"r_dm_t_0\":\" -1 \",\"r_du_t_0\":\" -1 \",\"r_lk_t_0\":\" -1 \",\"r_ag_t_0\":\" -1  \",\"outd\":\" 5  \",\"f_pv\":\" 8 \",\"f_cpp\":\" 12 \",\"f_ump\":\" 8 \",\"f_umc\":\" 8 \",\"f_tmp\":\" 3 \",\"f_tmc\":\" 7 \",\"f_smp\":\" 5 \",\"f_smc\":\" 9 \",\"f_t0p\":\" 7 \",\"f_t0c\":\" 3 \",\"f_t1p\":\" 2 \",\"f_t1c\":\" 8 \",\"f_t3p\":\" 8 \",\"f_t3c\":\" 9 \",\"f_t5p\":\" 5 \",\"f_t5c\":\" 5 \",\"f_umr\":\" 9 \",\"f_tmr\":\" 8 \",\"f_smr\":\" 9 \",\"f_t0r\":\" 3 \",\"f_t1r\":\" 8 \",\"f_t3r\":\" 9 \",\"f_t5r\":\" 9\"}",
    "user_feature":"new=0,rid=36,srid=201,uptime=0,no1=0,tfea=0,exp=0,rrblock=0,nsrc=0,rn=40,fba=1,acti=0,"
}
			`)
		res := &view.RelateResV2{}
		_ = json.Unmarshal(mock.BodyBuffer, res)
		decodeBytes, err := base64.StdEncoding.DecodeString(res.BizData.Data.AdsControl)
		fmt.Println(string(decodeBytes))
		a := &advo.SunspotAdReplyForView{}
		a.AdsControl = &types.Any{}
		//a.AdsControl.Unmarshal()
		m := a.AdsControl.Unmarshal(decodeBytes)

		//c := json.Unmarshal(decodeBytes, &a.AdsControl)
		fmt.Println(1, err, m)
		recommendReq := &view.RecommendReq{
			//以下参数调试不能变
			Aid:   416819748,
			Mid:   349646365,
			AdExp: 1,
			//
			Cmd:        "related",
			Build:      61802100,
			Network:    "wifi",
			ZoneId:     4308992,
			Plat:       1,
			AdResource: "2029,2335",
		}
		res, code, err := d.NewRelateAidsV2(context.Background(), recommendReq)
		fmt.Println(code, err, res)
		ctx.Convey("Then err should not be nil.", func(ctx convey.C) {
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
}

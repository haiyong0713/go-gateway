package exporttask

import (
	"fmt"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/component"
	"go-gateway/app/web-svr/activity/admin/model/lottery"
	"strconv"
)

type ExportConfArg struct {
	Validate string
}

type ExportOutputField struct {
	Name   string
	Title  string
	Format func(string) string
}

type ExportConf struct {
	Params  map[string]*ExportConfArg
	Execute ExportTask
}

var exportConf = map[uint8]*ExportConf{
	1: { //预约数据源
		Params: map[string]*ExportConfArg{},
		Execute: &taskExportReserve{
			reserveSQL: &taskExportSQL{
				SQL: "SELECT id,mid,num,state,mtime,ctime FROM act_reserve_%02d WHERE sid = ? AND mtime BETWEEN ? AND ? AND id > ? ORDER BY id ASC limit 100000",
				Tablet: func(m map[string]string) int {
					sid, _ := strconv.ParseInt(m["sid"], 10, 64)
					return int(sid % 100)
				},
				Args: []string{
					"sid",
					"start_time",
					"end_time",
					"primary_key",
				},
				PrimaryKey:     "id",
				PrimaryDefault: 0,
			},
			userStatSQL: &taskExportSQL{
				SQL: "SELECT id,mid,task_id,cnt,round_count,ctime FROM task_user_state_%02d WHERE foreign_id = ? AND mid IN (%s) AND id > ? ORDER BY id ASC limit 100000",
				Builder: func(s string, m map[string]string) string {
					sid, _ := strconv.ParseInt(m["sid"], 10, 64)
					return fmt.Sprintf(s, sid%100, m["mid_list"])
				},
				Args: []string{
					"sid",
					"primary_key",
				},
				PrimaryKey:     "id",
				PrimaryDefault: 0,
			},
		},
	},
	2: { // 活动数据源列表
		Params: map[string]*ExportConfArg{
			"state": {
				Validate: "required,min=-1",
			},
		},
		Execute: &taskExportSQL{
			SQL: "SELECT id,name,type,stime,etime,ctime,mtime,author FROM act_subject WHERE state=? AND id > ? ORDER BY id ASC limit 1000",
			Args: []string{
				"state",
				"primary_key",
			},
			PrimaryKey:     "id",
			PrimaryDefault: 0,
			formatter: &simpleFormatter{
				Output: []*ExportOutputField{
					{
						Name: "id",
					},
					{
						Name:  "name",
						Title: "活动名称",
					},
					{
						Name: "type",
					},
					{
						Name:   "stime",
						Format: formatTimeString,
					},
					{
						Name:   "etime",
						Format: formatTimeString,
					},
					{
						Name:   "ctime",
						Format: formatTimeString,
					},
					{
						Name:   "mtime",
						Format: formatTimeString,
					},
					{
						Name: "author",
					},
				},
			},
		},
	},
	3: { // 视频稿件
		Params: map[string]*ExportConfArg{},
		Execute: &taskExportSQL{
			SQL: "SELECT likes.id,sid,mid,wid,state,likes.mtime,like_extend.like FROM likes force index (ix_like_0) LEFT JOIN like_extend ON likes.id=like_extend.lid WHERE sid = ? AND likes.mtime BETWEEN ? AND ? AND likes.id > ? ORDER BY likes.id ASC LIMIT 1000",
			Args: []string{
				"sid",
				"start_time",
				"end_time",
				"primary_key",
			},
			PrimaryKey:     "id",
			PrimaryDefault: 0,
			formatter: &simpleFormatter{
				Output: []*ExportOutputField{
					{
						Name:  "sid",
						Title: "数据源ID",
					},
					{
						Name: "mid",
					},
					{
						Name:  "nickname",
						Title: "昵称",
					},
					{
						Name:  "wid",
						Title: "avid",
					},
					{
						Name: "bvid",
					},
					{
						Name:  "type_name",
						Title: "分区名字",
					},
					{
						Name:  "view",
						Title: "播放数",
					},
					{
						Name:  "arc_like",
						Title: "视频点赞数",
					},
					{
						Name:  "fav",
						Title: "收藏",
					},
					{
						Name:  "coin",
						Title: "硬币",
					},
					{
						Name:  "share",
						Title: "分享",
					},
					{
						Name:  "title",
						Title: "视频标题",
					},
					{
						Name:  "like",
						Title: "投票数",
					},
					{
						Name:  "state",
						Title: "稿件状态(1:通过,-1:未通过,0:待审核)",
					},
				},
			},
			Append: []appended{
				&VideoAppend{
					Field: "wid",
				},
				&AccountAppend{
					Field: "mid",
				},
				&BvidAppend{
					Field: "wid",
				},
			},
		},
	},
	4: { // 问卷
		Params: map[string]*ExportConfArg{},
		Execute: &taskExportQuestion{
			questionSQL: &taskExportSQL{
				Builder: func(s string, m map[string]string) string {
					return fmt.Sprintf(s, m["like_content_table"])
				},
				SQL: "SELECT likes.id,sid,mid,wid,likes.mtime,message FROM likes force index (ix_like_0) INNER JOIN %s b ON likes.id = b.id WHERE likes.sid = ? AND likes.mtime BETWEEN ? AND ? AND likes.id > ? ORDER BY likes.id ASC LIMIT 1000",
				Args: []string{
					"sid",
					"start_time",
					"end_time",
					"primary_key",
				},
				PrimaryKey:     "id",
				PrimaryDefault: 0,
			},
		},
	},
	5: { // 图片
		Params: map[string]*ExportConfArg{},
		Execute: &taskExportSQL{
			SQL: "SELECT likes.id,sid,mid,wid,likes.mtime,image,message,like_extend.like,likes.state FROM likes force index (ix_like_0) INNER JOIN like_content ON likes.id = like_content.id LEFT JOIN like_extend ON likes.id=like_extend.lid WHERE likes.sid = ? AND likes.mtime BETWEEN ? AND ? AND likes.id > ? ORDER BY likes.id ASC LIMIT 1000",
			Args: []string{
				"sid",
				"start_time",
				"end_time",
				"primary_key",
			},
			PrimaryKey:     "id",
			PrimaryDefault: 0,
			formatter: &simpleFormatter{
				Output: []*ExportOutputField{
					{
						Name:  "sid",
						Title: "数据源ID",
					},
					{
						Name: "mid",
					},
					{
						Name:   "mtime",
						Title:  "日期",
						Format: formatTimeString,
					},
					{
						Name:  "image",
						Title: "图片地址",
					},
					{
						Name:  "message",
						Title: "图片信息",
					},
					{
						Name:  "like",
						Title: "点赞数",
					},
					{
						Name:  "state",
						Title: "图片状态(1:通过,-1:未通过,0:待审核)",
					},
				},
			},
		},
	},
	6: { // 抽奖未中奖导出
		Params: map[string]*ExportConfArg{},
		Execute: &taskExportSQL{
			SQL: "SELECT id, mid, cid, state, type, ctime FROM act_lottery_action_%d WHERE gift_id = 0 and ctime BETWEEN ? AND ? AND id > ? ORDER BY id ASC LIMIT 1000",
			Args: []string{
				"start_time",
				"end_time",
				"primary_key",
			},
			Tablet: func(m map[string]string) int {
				lof := new(lottery.RuleInfo)
				if err := component.GlobalOrm.Where("sid=?", m["sid"]).Find(&lof).Error; err != nil {
					log.Error("act_lottery_action_ Tablet db.Where(sid:%d).Find error(%v)", m["sid"], err)
					return 0
				}
				return int(lof.ID)
			},
			PrimaryKey:     "id",
			PrimaryDefault: 0,
			formatter: &simpleFormatter{
				Output: []*ExportOutputField{
					{
						Name: "mid",
					},
					{
						Name:  "cid",
						Title: "抽奖次数配置id",
					},
					{
						Name:  "state",
						Title: "状态(0:正常,1:不正常)",
					},
					{
						Name:  "type",
						Title: "抽奖次数类型",
					},
					{
						Name:   "ctime",
						Title:  "日期",
						Format: formatTimeString,
					},
				},
			},
		},
	},
	7: { // 点赞明细
		Params: map[string]*ExportConfArg{},
		Execute: &taskExportSQL{
			SQL: "SELECT like_action.id, action, like_action.mtime, like_action.mid, lid, like_action.sid, likes.wid FROM like_action INNER JOIN likes ON like_action.lid=likes.id WHERE like_action.sid = ? AND like_action.mtime BETWEEN ? AND ? AND like_action.id > ? ORDER BY like_action.id ASC LIMIT 1000",
			Args: []string{
				"sid",
				"start_time",
				"end_time",
				"primary_key",
			},
			PrimaryKey:     "id",
			PrimaryDefault: 0,
			formatter: &simpleFormatter{
				Output: []*ExportOutputField{
					{
						Name:  "sid",
						Title: "数据源ID",
					},
					{
						Name:  "lid",
						Title: "对象id",
					},
					{
						Name:  "wid",
						Title: "稿件id",
					},
					{
						Name: "mid",
					},
					{
						Name:  "nickname",
						Title: "昵称",
					},
					{
						Name:   "mtime",
						Title:  "时间",
						Format: formatTimeString,
					},
				},
			},
			Append: []appended{
				&AccountAppend{
					Field: "mid",
				},
			},
		},
	},
}

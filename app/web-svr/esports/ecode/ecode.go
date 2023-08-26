package ecode

import xecode "go-common/library/ecode"

var (
	EsportsContestNotExist                  = xecode.New(83001) // 你所订阅的赛程不存在
	EsportsContestMaxCount                  = xecode.New(83002) // 你订阅赛程数已达上限
	EsportsContestFavDel                    = xecode.New(83003) // 该赛程未订阅，不能取消哦~
	EsportsContestFavExist                  = xecode.New(83004) // 该赛程已订阅，不能重复订阅哦~
	EsportsContestNotDay                    = xecode.New(83005) // 仅可订阅15天内的赛事哦~
	EsportsContestStart                     = xecode.New(83006) // 你订阅的赛程已经开始啦~快来直播间观看吧~
	EsportsContestEnd                       = xecode.New(83007) // 你订阅的赛程已经结束啦~可以点击回放和集锦进行观看哦~
	EsportsContestFavNot                    = xecode.New(83008) // 该赛程不可订阅哦~
	EsportsActNotExist                      = xecode.New(83009) // 赛事活动不存在
	EsportsActVideoNotExist                 = xecode.New(83010) // 赛事活动视频不存在
	EsportsActPointNotExist                 = xecode.New(83011) // 比赛数据不存在
	EsportsActKnockNotExist                 = xecode.New(83012) // 淘汰赛数据不存在
	EsportsModNameErr                       = xecode.New(83050) // 模块名称重复~
	EsportsActModNot                        = xecode.New(83051) // 模块不属于该赛事活动~
	EsportsActModErr                        = xecode.New(83052) // 模块信息不正确~
	EsportsModArcErr                        = xecode.New(83053) // 模块稿件不正确~
	EsportsTreeNodeErr                      = xecode.New(83054) // 节点不属于该赛事详情~
	EsportsTreeDetailErr                    = xecode.New(83055) // 赛事活动详情不存在~
	EsportsTreeEmptyErr                     = xecode.New(83056) // 当前没有任何记录，请编辑后提交~
	EsportsMultiEdit                        = xecode.New(83057) // 节点不能多人同时保存~
	EsportsArcServerErr                     = xecode.New(83058) // 稿件服务出错~
	EsportsContestDataErr                   = xecode.New(83059) // 比赛数据不正确~
	EsportsGuessEndErr                      = xecode.New(83060) // 竞猜活动已结束~
	EsportsGuessNOTFound                    = xecode.New(83061) // 竞猜活动不存在~
	EsportsMatchLiveInvalid                 = xecode.New(83062) // 赛程直播地址不存在
	EsportsLiveNoList                       = xecode.New(83063) // 没有直播列表数据
	EsportsLiveNoInfo                       = xecode.New(83064) // 没有直播详情数据
	EsportsAutoSubed                        = xecode.New(83065) // 您已完成一键订阅哦～
	EsportsDrawPost                         = xecode.New(63065) // 生成海报失败
	EsportsComponentErr                     = xecode.New(63066) // 组件接口出错
	EsportsContestSeriesNotFound            = xecode.New(83066) // 该阶段不存在~
	EsportsContestSeriesExtraConfigFound    = xecode.New(83067) // 该阶段已配置~
	EsportsContestSeriesExtraConfigNotFound = xecode.New(83068) // 该阶段未配置~
	EsportsContestSeriesExtraConfigErr      = xecode.New(83069) // 该阶段配置错误~
)

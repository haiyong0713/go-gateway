package like

import (
	"context"
	"fmt"

	likemdl "go-gateway/app/web-svr/activity/interface/model/like"
)

// likeKey likes table line cache
func likeKey(id int64) string {
	return fmt.Sprintf("go_l_id_%d", id)
}

// actSubjectKey act_subject table line cache .
func actSubjectKey(id int64) string {
	return fmt.Sprintf("go_s_id_%d", id)
}

// actSubjectMaxIDKey act_subject table max id cache
func actSubjectMaxIDKey() string {
	return "go_sub_id_max"
}

// likeMaxIDKey likes table max id cache
func likeMaxIDKey() string {
	return "go_like_id_max"
}

// likeMissionBuffKey .
func likeMissionBuffKey(sid, mid int64) string {
	return fmt.Sprintf("go_l_m_a_%d_%d", sid, mid)
}

// likeMissionGroupIDkey .
func likeMissionGroupIDkey(lid int64) string {
	return fmt.Sprintf("go_l_m_g_id_%d", lid)
}

// likeActMissionKey flag has buff or not.
func likeActMissionKey(sid, lid, mid int64) string {
	return fmt.Sprintf("go:b-a:m:l:%d:%d:%d", sid, lid, mid)
}

// actAchieveKey .
func actAchieveKey(sid int64) string {
	return fmt.Sprintf("go:a:achs:%d", sid)
}

// actMissionFriendsKey .
func actMissionFriendsKey(sid, lid int64) string {
	return fmt.Sprintf("go:a:m:frd:%d:%d", sid, lid)
}

// actUserAchieveKey .
func actUserAchieveKey(id int64) string {
	return fmt.Sprintf("go:a:u:m:%d", id)
}

// actUserAchieveAwardKey .
func actUserAchieveAwardKey(id int64) string {
	return fmt.Sprintf("go:a:u:a:%d", id)
}

func subjectStatKey(sid int64) string {
	return fmt.Sprintf("ob_s_%d", sid)
}

func viewRankKey(sid int64, typ string) string {
	if typ != "" {
		return fmt.Sprintf("v_r_%d_%s", sid, typ)
	}
	return fmt.Sprintf("v_r_%d", sid)
}

func likeContentKey(lid int64) string {
	return fmt.Sprintf("go_l_ct_%d", lid)
}

func sourceItemKey(sid int64) string {
	return fmt.Sprintf("so_i_%d", sid)
}

func subjectProtocolKey(sid int64) string {
	return fmt.Sprintf("go_s_pt_%d", sid)
}

func textOnlyOneKey(sid, mid int64) string {
	return fmt.Sprintf("go_s_m_oly_%d_%d", sid, mid)
}

func reserveOnlyKey(sid, mid int64) string {
	return fmt.Sprintf("go_resv_oly_%d_%d", mid, sid)
}

func likeActLikesKey(sid, mid int64) string {
	return fmt.Sprintf("like_act_likes_all_%d_%d", mid, sid)
}

func likeActHisKey(sid int64) string {
	return fmt.Sprintf("likeact_his_%d", sid)
}

func actSubjectWithStateKey(id int64) string {
	return fmt.Sprintf("go_s_id_v2_%d", id)
}

func GetUpActReserveRelationInfoBySid(sid int64) string {
	return fmt.Sprintf("up_act_reserve_relation_info_%d", sid)
}

func getReserveCounterGroupIDBySidKey(sid int64) string {
	return fmt.Sprintf("rsv_counter_group_gid_%d", sid)
}

func getReserveCounterGroupInfoByGidKey(gid int64) string {
	return fmt.Sprintf("rsv_counter_group_info_%d", gid)
}

func getReserveCounterNodeByGidKey(gid int64) string {
	return fmt.Sprintf("rsv_counter_node_info_gid_%d", gid)
}
func GetUpActReserveRelationInfo4SpaceCardIDs(mid int64) string {
	return fmt.Sprintf("up_act_reserve_relation_info_4_space_card_ids_%d", mid)
}

func GetUpActReserveRelationInfo4Live(upmid int64) string {
	return fmt.Sprintf("up_act_reserve_relation_info_4_live_%d", upmid)
}

func GetWebViewDataByVidKey(vid int64) string {
	return fmt.Sprintf("w_v_d_v_%d", vid)
}

func GetOnlineWebViewDataByVidKey(vid int64) string {
	return fmt.Sprintf("o_w_v_d_v_%d", vid)
}

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -sync=true -struct_name=Dao
	Like(c context.Context, id int64) (*likemdl.Item, error)
	// bts: -sync=true -struct_name=Dao
	Likes(c context.Context, ids []int64) (map[int64]*likemdl.Item, error)
	// bts: -sync=true -struct_name=Dao
	ActSubject(c context.Context, id int64) (*likemdl.SubjectItem, error)
	// bts: -struct_name=Dao
	ActSubjects(c context.Context, ids []int64) (map[int64]*likemdl.SubjectItem, error)
	// bts: -sync=true -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	LikeMissionBuff(ctx context.Context, sid int64, mid int64) (res int64, err error)
	// bts: -sync=true -struct_name=Dao
	MissionGroupItems(ctx context.Context, lids []int64) (map[int64]*likemdl.MissionGroup, error)
	// bts: -sync=true -nullcache=-1 -check_null_code=$!=nil&&$==-1 -struct_name=Dao
	ActMission(ctx context.Context, sid int64, lid int64, mid int64) (res int64, err error)
	// bts:-sync=true -struct_name=Dao
	ActLikeAchieves(ctx context.Context, sid int64) (res *likemdl.Achievements, err error)
	// bts:-sync=true -struct_name=Dao
	ActMissionFriends(ctx context.Context, sid int64, lid int64) (res *likemdl.ActMissionGroups, err error)
	// bts:-sync=true -struct_name=Dao
	ActUserAchieve(ctx context.Context, id int64) (res *likemdl.ActLikeUserAchievement, err error)
	// bts:-struct_name=Dao
	MatchSubjects(c context.Context, ids []int64) (map[int64]*likemdl.Object, error)
	// bts:-sync=true -struct_name=Dao
	LikeContent(c context.Context, ids []int64) (map[int64]*likemdl.LikeContent, error)
	// bts:-struct_name=Dao
	SourceItemData(c context.Context, sid int64) ([]int64, error)
	// bts:-sync=true -struct_name=Dao
	ActSubjectProtocol(c context.Context, sid int64) (res *likemdl.ActSubjectProtocol, err error)
	// bts: -struct_name=Dao
	ActSubjectProtocols(c context.Context, sid []int64) (map[int64]*likemdl.ActSubjectProtocol, error)
	// bts: -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	LikeTotal(c context.Context, sid int64) (int64, error)
	// bts: -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	LikeMidTotal(c context.Context, mid int64, sids []int64) (int64, error)
	// bts: -nullcache=&likemdl.Item{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	LikeCheck(c context.Context, mid int64, sid int64) (*likemdl.Item, error)
	// bts: -nullcache=[]*likemdl.LidItem{{Lid:-1}} -check_null_code=len($)==1&&$[0].Lid==-1 -struct_name=Dao
	LikeActLids(c context.Context, sid int64, mid int64) ([]*likemdl.LidItem, error)
	// bts: -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	TextOnly(c context.Context, sid int64, mid int64) (int, error)
	// bts: -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	ActUserAward(c context.Context, id int64) (int64, error)
	// bts: -nullcache=&likemdl.HasReserve{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao -sync=true
	ReserveOnly(c context.Context, sid int64, mid int64) (*likemdl.HasReserve, error)
	// bts: -struct_name=Dao
	ReservesTotal(c context.Context, sid []int64) (map[int64]int64, error)
	// bts: -singleflight=true -nullcache=[]int64{-1} -check_null_code=len($)==1&&$[0]==-1 -struct_name=Dao
	ActStochastic(c context.Context, sid int64, ltype int64) (res []int64, err error)
	// bts: -singleflight=true -paging=true -ignores=||start,end -nullcache=&likemdl.EsLikesReply{Lids:[]int64{-1}} -check_null_code=len($.Lids)==1&&$.Lids[0]==-1 -struct_name=Dao
	ActEsLikesIDs(c context.Context, sid int64, ltype int64, start int64, end int64) (res *likemdl.EsLikesReply, err error)
	// bts: -singleflight=true -paging=true -ignores=||start,end -nullcache=[]int64{-1} -check_null_code=len($)==1&&$[0]==-1 -struct_name=Dao
	ActRandom(c context.Context, sid int64, ltype int64, start int64, end int64) (res []int64, err error)
	// bts:-sync=true -struct_name=Dao
	LikeExtendToken(ctx context.Context, sid int64, mid int64) (res *likemdl.ExtendTokenDetail, err error)
	// bts:-sync=true -struct_name=Dao
	LikeExtendInfo(ctx context.Context, sid int64, token string) (res *likemdl.ExtendTokenDetail, err error)
	// bts:-sync=true -struct_name=Dao -nullcache=&likemdl.AwardSubject{ID:-1} -check_null_code=$!=nil&&$.ID==-1
	AwardSubject(ctx context.Context, sid int64) (res *likemdl.AwardSubject, err error)
	// bts:-sync=true -struct_name=Dao -nullcache=&likemdl.AwardSubject{ID:-1} -check_null_code=$!=nil&&$.ID==-1
	AwardSubjectByID(ctx context.Context, id int64) (res *likemdl.AwardSubject, err error)
	// bts: -nullcache=&likemdl.ActUp{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	ActUp(c context.Context, mid int64) (res *likemdl.ActUp, err error)
	// bts: -nullcache=&likemdl.ActUp{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	ActUpBySid(c context.Context, sid int64) (res *likemdl.ActUp, err error)
	// bts: -nullcache=&likemdl.ActUp{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	ActUpByAid(c context.Context, aid int64) (res *likemdl.ActUp, err error)
	// bts: -nullcache=[]*likemdl.SubjectRule{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1 -struct_name=Dao
	SubjectRulesBySid(c context.Context, sid int64) ([]*likemdl.SubjectRule, error)
	// bts: -nullcache=[]*likemdl.SubjectRule{{ID:-1}} -check_null_code=len($)==1&&$[0]!=nil&&$[0].ID==-1 -struct_name=Dao
	SubjectRulesBySids(c context.Context, sids []int64) (map[int64][]*likemdl.SubjectRule, error)
	// bts: -nullcache=[]*likemdl.Item{{ID:-1}} -check_null_code=len($)==1&&$[0].ID==-1 -struct_name=Dao
	ActivityArchives(ctx context.Context, sid int64, mid int64) ([]*likemdl.Item, error)
	// bts:-sync=true -struct_name=Dao -nullcache=&likemdl.ActRelationInfo{ID:-1} -check_null_code=$!=nil&&$.ID==-1
	GetActRelationInfo(ctx context.Context, id int64) (res *likemdl.ActRelationInfo, err error)
	// bts: -sync=true -struct_name=Dao
	ActSubjectWithState(c context.Context, id int64) (*likemdl.SubjectItem, error)
	// bts: -sync=true -struct_name=Dao
	ActSubjectsWithState(c context.Context, ids []int64) (map[int64]*likemdl.SubjectItem, error)
	// bts:-sync=true -struct_name=Dao
	GetUpActReserveRelationInfoBySid(ctx context.Context, sids []int64) (res map[int64]*likemdl.UpActReserveRelationInfo, err error)
	// bts:-sync=true -struct_name=Dao
	GetReserveCounterGroupIDBySid(ctx context.Context, sid int64) (res []int64, err error)
	// bts:-sync=true -struct_name=Dao
	GetReserveCounterGroupInfoByGid(ctx context.Context, gid []int64) (res map[int64]*likemdl.ReserveCounterGroupItem, err error)
	// bts:-sync=true -struct_name=Dao
	GetReserveCounterNodeByGid(ctx context.Context, gid []int64) (res map[int64][]*likemdl.ReserveCounterNodeItem, err error)
	// bts:-sync=true -struct_name=Dao
	GetUpActReserveRelationInfo4SpaceCardIDs(ctx context.Context, mid int64) (res []int64, err error)
	// bts:-sync=true -struct_name=Dao -nullcache=-1 -check_null_code=$==-1
	GetUpActReserveRelationInfo4Live(ctx context.Context, upMid int64) (res int64, err error)
	// bts:-sync=true -struct_name=Dao -nullcache=[]*likemdl.WebDataItem{{ID:-1}} -check_null_code=len($)==1&&$[0].ID==-1
	GetWebViewDataByVid(ctx context.Context, vid int64) ([]*likemdl.WebDataItem, error)
	// bts:-sync=true -struct_name=Dao -nullcache=[]*likemdl.WebDataItem{{ID:-1}} -check_null_code=len($)==1&&$[0].ID==-1
	GetOnlineWebViewDataByVid(ctx context.Context, vid int64) ([]*likemdl.WebDataItem, error)
}

//go:generate kratos tool mcgen
type _mc interface {
	// mc: -key=likeKey -struct_name=Dao
	CacheLike(c context.Context, id int64) (*likemdl.Item, error)
	// mc: -key=likeKey -struct_name=Dao
	CacheLikes(c context.Context, id []int64) (map[int64]*likemdl.Item, error)
	// mc: -key=likeKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheLikes(c context.Context, items map[int64]*likemdl.Item) error
	// mc: -key=likeKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheLike(c context.Context, key int64, value *likemdl.Item) error
	// mc: -key=actSubjectKey -struct_name=Dao
	CacheActSubject(c context.Context, id int64) (*likemdl.SubjectItem, error)
	// mc: -key=actSubjectKey -expire=d.mcRegularExpire -encode=pb -struct_name=Dao
	AddCacheActSubject(c context.Context, key int64, value *likemdl.SubjectItem) error
	// mc: -key=actSubjectKey -struct_name=Dao
	CacheActSubjects(c context.Context, ids []int64) (map[int64]*likemdl.SubjectItem, error)
	// mc: -key=actSubjectKey -expire=d.mcRegularExpire -encode=pb -struct_name=Dao
	AddCacheActSubjects(c context.Context, data map[int64]*likemdl.SubjectItem) error
	// mc: -key=actSubjectMaxIDKey -struct_name=Dao
	CacheActSubjectMaxID(c context.Context) (res int64, err error)
	// mc: -key=actSubjectMaxIDKey -expire=d.mcPerpetualExpire -encode=raw -struct_name=Dao
	AddCacheActSubjectMaxID(c context.Context, sid int64) error
	// mc: -key=likeMaxIDKey -struct_name=Dao
	CacheLikeMaxID(c context.Context) (res int64, err error)
	// mc: -key=likeMaxIDKey -expire=d.mcPerpetualExpire -encode=raw -struct_name=Dao
	AddCacheLikeMaxID(c context.Context, lid int64) error
	//mc: -key=likeMissionBuffKey -struct_name=Dao
	CacheLikeMissionBuff(c context.Context, sid int64, mid int64) (res int64, err error)
	//mc: -key=likeMissionBuffKey -expire=d.mcActMissionExpire -struct_name=Dao
	AddCacheLikeMissionBuff(c context.Context, sid int64, val int64, mid int64) error
	//mc: -key=likeMissionGroupIDkey -struct_name=Dao
	CacheMissionGroupItems(ctx context.Context, lids []int64) (map[int64]*likemdl.MissionGroup, error)
	//mc: -key=likeMissionGroupIDkey -expire=d.mcItemExpire -encode=pb -struct_name=Dao
	AddCacheMissionGroupItems(ctx context.Context, val map[int64]*likemdl.MissionGroup) error
	//mc: -key=likeActMissionKey -struct_name=Dao
	CacheActMission(c context.Context, sid int64, lid int64, mid int64) (res int64, err error)
	//mc: -key=likeActMissionKey -expire=d.mcActMissionExpire -encode=raw -struct_name=Dao
	AddCacheActMission(c context.Context, sid int64, val int64, lid int64, mid int64) error
	//mc: -key=actAchieveKey -struct_name=Dao
	CacheActLikeAchieves(c context.Context, sid int64) (res *likemdl.Achievements, err error)
	//mc: -key=actAchieveKey -expire=d.mcItemExpire -encode=pb -struct_name=Dao
	AddCacheActLikeAchieves(c context.Context, sid int64, res *likemdl.Achievements) error
	//mc: -key=actMissionFriendsKey -struct_name=Dao
	CacheActMissionFriends(c context.Context, sid int64, lid int64) (res *likemdl.ActMissionGroups, err error)
	//mc: -key=actMissionFriendsKey -struct_name=Dao
	DelCacheActMissionFriends(c context.Context, sid int64, lid int64) error
	//mc: -key=actMissionFriendsKey -expire=d.mcItemExpire -encode=pb -struct_name=Dao
	AddCacheActMissionFriends(c context.Context, sid int64, res *likemdl.ActMissionGroups, lid int64) error
	//mc: -key=actUserAchieveKey -struct_name=Dao
	CacheActUserAchieve(c context.Context, id int64) (res *likemdl.ActLikeUserAchievement, err error)
	//mc: -key=actUserAchieveKey -expire=d.mcItemExpire -encode=pb -struct_name=Dao
	AddCacheActUserAchieve(c context.Context, id int64, val *likemdl.ActLikeUserAchievement) error
	//mc: -key=actUserAchieveAwardKey -struct_name=Dao
	CacheActUserAward(c context.Context, id int64) (res int64, err error)
	//mc: -key=actUserAchieveAwardKey -expire=d.mcActMissionExpire -encode=raw -struct_name=Dao
	AddCacheActUserAward(c context.Context, id int64, val int64) error
	// mc: -key=subjectStatKey -struct_name=Dao
	CacheSubjectStat(c context.Context, sid int64) (*likemdl.SubjectStat, error)
	// mc: -key=subjectStatKey -expire=d.mcSubStatExpire -encode=json -struct_name=Dao
	AddCacheSubjectStat(c context.Context, sid int64, value *likemdl.SubjectStat) error
	// mc: -key=viewRankKey -struct_name=Dao
	CacheViewRank(c context.Context, sid int64, typ string) (string, error)
	// mc: -key=viewRankKey -expire=d.mcViewRankExpire -encode=raw -struct_name=Dao
	AddCacheViewRank(c context.Context, sid int64, value string, typ string) error
	// mc: -key=likeContentKey -struct_name=Dao
	CacheLikeContent(c context.Context, lids []int64) (res map[int64]*likemdl.LikeContent, err error)
	// mc: -key=likeContentKey -expire=d.mcRegularExpire -encode=pb -struct_name=Dao
	AddCacheLikeContent(c context.Context, val map[int64]*likemdl.LikeContent) error
	// mc: -key=sourceItemKey -struct_name=Dao
	CacheSourceItemData(c context.Context, sid int64) ([]int64, error)
	// mc: -key=sourceItemKey -expire=d.mcSourceItemExpire -encode=json -struct_name=Dao
	AddCacheSourceItemData(c context.Context, sid int64, lids []int64) error
	// mc: -key=subjectProtocolKey -struct_name=Dao
	CacheActSubjectProtocol(c context.Context, sid int64) (res *likemdl.ActSubjectProtocol, err error)
	// mc: -key=subjectProtocolKey -expire=d.mcRegularExpire -encode=pb -struct_name=Dao
	AddCacheActSubjectProtocol(c context.Context, sid int64, value *likemdl.ActSubjectProtocol) error
	// mc: -key=subjectProtocolKey -struct_name=Dao
	DelCacheActSubjectProtocol(c context.Context, sid int64) error
	// mc: -key=subjectProtocolKey -struct_name=Dao
	CacheActSubjectProtocols(c context.Context, sids []int64) (map[int64]*likemdl.ActSubjectProtocol, error)
	// mc: -key=subjectProtocolKey -expire=d.mcProtocolExpire -encode=pb -struct_name=Dao
	AddCacheActSubjectProtocols(c context.Context, data map[int64]*likemdl.ActSubjectProtocol) error
	// mc: -key=textOnlyOneKey -struct_name=Dao
	CacheTextOnly(c context.Context, sid int64, mid int64) (res int, err error)
	// mc: -key=textOnlyOneKey -expire=d.mcUserCheckExpire -encode=raw -struct_name=Dao
	AddCacheTextOnly(c context.Context, sid int64, val int, mid int64) error
	// mc: -key=ipRequestKey -struct_name=Dao
	CacheIPRequestCheck(c context.Context, ip string) (res int, err error)
	// mc: -key=ipRequestKey -expire=d.mcLikeIPExpire -encode=raw -struct_name=Dao
	AddCacheIPRequestCheck(c context.Context, ip string, val int) error
	// mc: -key=reserveOnlyKey -struct_name=Dao
	CacheReserveOnly(c context.Context, sid int64, mid int64) (res *likemdl.HasReserve, err error)
	// mc: -key=reserveOnlyKey -expire=d.mcReserveOnlyExpire -encode=pb -struct_name=Dao
	AddCacheReserveOnly(c context.Context, sid int64, val *likemdl.HasReserve, mid int64) error
	// mc: -key=reserveOnlyKey -struct_name=Dao
	DelCacheReserveOnly(c context.Context, sid int64, mid int64) error
	// mc: -key=actSubjectWithStateKey -struct_name=Dao
	CacheActSubjectWithState(c context.Context, id int64) (*likemdl.SubjectItem, error)
	// mc: -key=actSubjectWithStateKey -expire=d.mcRegularExpire -encode=pb -struct_name=Dao
	AddCacheActSubjectWithState(c context.Context, key int64, value *likemdl.SubjectItem) error
	// mc: -key=actSubjectWithStateKey -struct_name=Dao
	CacheActSubjectsWithState(c context.Context, ids []int64) (map[int64]*likemdl.SubjectItem, error)
	// mc: -key=actSubjectWithStateKey -expire=d.mcRegularExpire -encode=pb -struct_name=Dao
	AddCacheActSubjectsWithState(c context.Context, data map[int64]*likemdl.SubjectItem) error
	//  mc: -key=GetUpActReserveRelationInfoBySid -struct_name=Dao
	CacheGetUpActReserveRelationInfoBySid(c context.Context, sids []int64) (map[int64]*likemdl.UpActReserveRelationInfo, error)
	//  mc: -key=GetUpActReserveRelationInfoBySid -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheGetUpActReserveRelationInfoBySid(c context.Context, data map[int64]*likemdl.UpActReserveRelationInfo) (map[int64]*likemdl.UpActReserveRelationInfo, error)
	//  mc: -key=getReserveCounterGroupIDBySidKey -struct_name=Dao
	CacheGetReserveCounterGroupIDBySid(ctx context.Context, sid int64) (res []int64, err error)
	//  mc: -key=getReserveCounterGroupIDBySidKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheGetReserveCounterGroupIDBySid(ctx context.Context, sid int64, data []int64) error
	//  mc: -key=getReserveCounterGroupIDBySidKey -struct_name=Dao
	DelCacheGetReserveCounterGroupIDBySid(ctx context.Context, sid int64) error
	//  mc: -key=getReserveCounterGroupInfoByGidKey -struct_name=Dao
	CacheGetReserveCounterGroupInfoByGid(ctx context.Context, gid []int64) (res map[int64]*likemdl.ReserveCounterGroupItem, err error)
	//  mc: -key=getReserveCounterGroupInfoByGidKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheGetReserveCounterGroupInfoByGid(ctx context.Context, data map[int64]*likemdl.ReserveCounterGroupItem) error
	//  mc: -key=getReserveCounterGroupInfoByGidKey -struct_name=Dao
	DelCacheGetReserveCounterGroupInfoByGid(ctx context.Context, gid int64) error
	//  mc: -key=getReserveCounterNodeByGidKey -struct_name=Dao
	CacheGetReserveCounterNodeByGid(ctx context.Context, gid []int64) (res map[int64][]*likemdl.ReserveCounterNodeItem, err error)
	//  mc: -key=getReserveCounterNodeByGidKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheGetReserveCounterNodeByGid(ctx context.Context, data map[int64][]*likemdl.ReserveCounterNodeItem) error
	//  mc: -key=getReserveCounterNodeByGidKey -struct_name=Dao
	DelCacheGetReserveCounterNodeByGid(ctx context.Context, gid int64) error
	//  mc: -key=GetUpActReserveRelationInfo4SpaceCardIDs -struct_name=Dao
	CacheGetUpActReserveRelationInfo4SpaceCardIDs(c context.Context, mid int64) ([]int64, error)
	//  mc: -key=GetUpActReserveRelationInfo4SpaceCardIDs -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheGetUpActReserveRelationInfo4SpaceCardIDs(c context.Context, mid int64, ids []int64) ([]int64, error)
	//  mc: -key=GetUpActReserveRelationInfo4Live -struct_name=Dao
	CacheGetUpActReserveRelationInfo4Live(c context.Context, upmid int64) (int64, error)
	//  mc: -key=GetUpActReserveRelationInfo4Live -expire=d.mcRegularExpire -struct_name=Dao
	AddCacheGetUpActReserveRelationInfo4Live(c context.Context, upmid int64, sids int64) ([]int64, error)
	//  mc: -key=GetOnlineWebViewDataByVidKey -expire=d.mcSourceItemExpire -struct_name=Dao
	AddCacheGetOnlineWebViewDataByVid(c context.Context, vid int64, list []*likemdl.WebDataItem) error
	//  mc: -key=GetOnlineWebViewDataByVidKey -expire=d.mcSourceItemExpire -struct_name=Dao
	CacheGetOnlineWebViewDataByVid(c context.Context, vid int64) ([]*likemdl.WebDataItem, error)
	//  mc: -key=GetOnlineWebViewDataByVidKey -expire=d.mcSourceItemExpire -struct_name=Dao
	DelCacheGetOnlineWebViewDataByVid(c context.Context, vid int64) error
	//  mc: -key=GetWebViewDataByVidKey -expire=d.mcSourceItemExpire -struct_name=Dao
	AddCacheGetWebViewDataByVid(c context.Context, vid int64, list []*likemdl.WebDataItem) error
	//  mc: -key=GetWebViewDataByVidKey -expire=d.mcSourceItemExpire -struct_name=Dao
	CacheGetWebViewDataByVid(c context.Context, vid int64) ([]*likemdl.WebDataItem, error)
	//  mc: -key=GetWebViewDataByVidKey -expire=d.mcSourceItemExpire -struct_name=Dao
	DelCacheGetWebViewDataByVid(c context.Context, vid int64) error
}

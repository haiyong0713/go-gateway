package native

import (
	"context"
	"fmt"

	v1 "go-gateway/app/web-svr/native-page/interface/api"
	dynmdl "go-gateway/app/web-svr/native-page/interface/model/dynamic"
	"go-gateway/app/web-svr/native-page/interface/model/white_list"
)

func nativePageKey(id int64) string {
	return fmt.Sprintf("nat_pg_%d", id)
}

func nativeForeignKey(id int64, pageType int64) string {
	return fmt.Sprintf("nat_pt_fr_%d_%d", id, pageType)
}

func nativeTabBindKey(id int64, category int32) string {
	return fmt.Sprintf("nat_tab_ct_%d_%d", id, category)
}

func nativeModuleKey(id int64) string {
	return fmt.Sprintf("nat_modu_%d", id)
}

func nativeClickKey(id int64) string {
	return fmt.Sprintf("nat_m_cli_%d", id)
}

func nativeDynamicKey(id int64) string {
	return fmt.Sprintf("nat_m_dyna_%d", id)
}

func nativeParticipationKey(id int64) string {
	return fmt.Sprintf("nat_m_pt_%d", id)
}

func nativeVideoKey(id int64) string {
	return fmt.Sprintf("nat_m_vid_%d", id)
}

func nativeUkeyKey(pid int64, ukey string) string {
	return fmt.Sprintf("nat_p_uk_%d_%s", pid, ukey)
}

func nativeMixtureKey(id int64) string {
	return fmt.Sprintf("nat_m_mixture_%d", id)
}

func nativeTabKey(id int64) string {
	return fmt.Sprintf("nat_tab_%d", id)
}

func nativeTabModuleKey(id int64) string {
	return fmt.Sprintf("nat_tmodu_%d", id)
}

func nativeTabSortKey(id int64) string {
	return fmt.Sprintf("nat_m_st_%d", id)
}

func (d *Dao) cacheSFNativeTabSort(id int64) string {
	return fmt.Sprintf("nat_sf_st_%d", id)
}

func ntTsPageKey(pid int64) string {
	return fmt.Sprintf("nat_ts_page_%d", pid)
}

func ntTsModuleExtKey(id int64) string {
	return fmt.Sprintf("nat_ts_md_ext_%d", id)
}

func (d *Dao) cacheSFNatTagIDExist(id int64) string {
	return fmt.Sprintf("nat_sf_td_%d", id)
}

func natTagIDExistKey(id int64) string {
	return fmt.Sprintf("nat_ts_ex_%d", id)
}

func natIDsByActTypeKey(actType int64) string {
	return fmt.Sprintf("nat_ids_act_type_%d", actType)
}

func nativeClickIDsKey(id int64) string {
	return fmt.Sprintf("nat_m_cl_s_%d", id)
}

func (d *Dao) cacheSFNativeClickIDs(id int64) string {
	return fmt.Sprintf("nat_sf_cli_%d", id)
}

func nativeActIDsKey(id int64) string {
	return fmt.Sprintf("nat_m_at_s_%d", id)
}

func (d *Dao) cacheSFNativeActIDs(id int64) string {
	return fmt.Sprintf("nat_sf_at_%d", id)
}

func nativeVideoIDsKey(id int64) string {
	return fmt.Sprintf("nat_m_vio_s_%d", id)
}

func (d *Dao) cacheSFNativeVideoIDs(id int64) string {
	return fmt.Sprintf("nat_sf_vio_%d", id)
}

func nativeDynamicIDsKey(id int64) string {
	return fmt.Sprintf("nat_m_dyc_s_%d", id)
}

func (d *Dao) cacheSFNativeDynamicIDs(id int64) string {
	return fmt.Sprintf("nat_sf_dyc_%d", id)
}

func userSpaceByMidKey(mid int64) string {
	return fmt.Sprintf("nat_user_space_%d", mid)
}

func sponsoredUpKey(mid int64) string {
	return fmt.Sprintf("sponsored_up_%d", mid)
}

func natPageExtendKey(pid int64) string {
	return fmt.Sprintf("nat_page_ext_%d", pid)
}

func (d *Dao) cacheSFNativeExtend(pid int64) string {
	return fmt.Sprintf("nat_sf_page_ext_%d", pid)
}

//go:generate kratos tool btsgen
type _bts interface {
	// bts: -struct_name=Dao
	NativePages(c context.Context, ids []int64) (map[int64]*v1.NativePage, error)
	// bts: -nullcache=&v1.NativePageDyn{Id:-1} -check_null_code=$!=nil&&$.Id==-1 -struct_name=Dao
	NativePagesExt(c context.Context, ids []int64) (map[int64]*v1.NativePageDyn, error)
	// bts: -struct_name=Dao
	NtTsPages(c context.Context, ids []int64) (map[int64]*v1.NativeTsPage, error)
	// bts: -singleflight=true -paging=true -ignores=||start,end -nullcache=&dynmdl.ModuleRankReply{IDs:[]*dynmdl.RankInfo{{ID:-1,Score:3786883200}}} -check_null_code=$!=nil&&len($.IDs)==1&&$.IDs[0]!=nil&&$.IDs[0].ID==-1 -struct_name=Dao
	NtTsUIDs(c context.Context, uid int64, start int64, end int64) (res *dynmdl.ModuleRankReply, err error)
	// bts: -singleflight=true -paging=true -ignores=||start,end -nullcache=&dynmdl.ModuleRankReply{IDs:[]*dynmdl.RankInfo{{ID:-1,Score:3786883200}}} -check_null_code=$!=nil&&len($.IDs)==1&&$.IDs[0]!=nil&&$.IDs[0].ID==-1 -struct_name=Dao
	NtTsOnlineIDs(c context.Context, uid int64, start int64, end int64) (res *dynmdl.ModuleRankReply, err error)
	// bts: -singleflight=true -paging=true -ignores=||start,end -nullcache=&dynmdl.ModuleIDsReply{IDs:[]int64{-1}} -check_null_code=$!=nil&&len($.IDs)==1&&$.IDs[0]==-1 -struct_name=Dao
	NtTsModuleIDs(c context.Context, tsID int64, start int64, end int64) (res *dynmdl.ModuleIDsReply, err error)
	// bts: -struct_name=Dao
	NtTsModulesExt(c context.Context, ids []int64) (map[int64]*dynmdl.NativeTsModuleExt, error)
	// bts: -singleflight=true -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	NtPidToTsID(c context.Context, pid int64) (int64, error)
	// bts: -struct_name=Dao -nullcache=-1 -check_null_code=$==-1
	NtPidToTsIDs(c context.Context, pids []int64) (map[int64]int64, error)
	// bts: -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	NativeForeigns(c context.Context, ids []int64, pageType int64) (map[int64]int64, error)
	// bts: -singleflight=true -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	NatTagIDExist(c context.Context, id int64) (int64, error)
	// bts: -struct_name=Dao
	NativeModules(c context.Context, ids []int64) (map[int64]*v1.NativeModule, error)
	// bts: -struct_name=Dao
	NativeClicks(c context.Context, ids []int64) (map[int64]*v1.NativeClick, error)
	// bts: -struct_name=Dao
	NativeDynamics(c context.Context, ids []int64) (map[int64]*v1.NativeDynamicExt, error)
	// bts: -struct_name=Dao
	NativeVideos(c context.Context, ids []int64) (map[int64]*v1.NativeVideoExt, error)
	// bts: -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	NativeUkey(c context.Context, pid int64, ukey string) (int64, error)
	// bts: -struct_name=Dao
	NativeMixtures(c context.Context, ids []int64) (map[int64]*v1.NativeMixtureExt, error)
	// bts: -nullcache=&v1.NativeActTab{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	NativeTabs(c context.Context, ids []int64) (res map[int64]*v1.NativeActTab, err error)
	// bts: -nullcache=&v1.NativeTabModule{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	NativeTabModules(c context.Context, ids []int64) (res map[int64]*v1.NativeTabModule, err error)
	// bts: -singleflight=true -nullcache=[]int64{-1} -check_null_code=len($)==1&&$[0]==-1 -struct_name=Dao
	NativeTabSort(c context.Context, id int64) (res []int64, err error)
	// bts: -singleflight=true -nullcache=[]int64{-1} -check_null_code=len($)==1&&$[0]==-1 -struct_name=Dao
	NativeClickIDs(c context.Context, id int64) (res []int64, err error)
	// bts: -singleflight=true -nullcache=[]int64{-1} -check_null_code=len($)==1&&$[0]==-1 -struct_name=Dao
	NativeDynamicIDs(c context.Context, id int64) (res []int64, err error)
	// bts: -singleflight=true -nullcache=[]int64{-1} -check_null_code=len($)==1&&$[0]==-1 -struct_name=Dao
	NativeActIDs(c context.Context, id int64) (res []int64, err error)
	// bts: -singleflight=true -nullcache=[]int64{-1} -check_null_code=len($)==1&&$[0]==-1 -struct_name=Dao
	NativeVideoIDs(c context.Context, id int64) (res []int64, err error)
	// bts: -nullcache=&v1.NativeParticipationExt{ID:-1} -check_null_code=$!=nil&&$.ID==-1 -struct_name=Dao
	NativePart(c context.Context, ids []int64) (res map[int64]*v1.NativeParticipationExt, err error)
	// bts: -nullcache=-1 -check_null_code=$==-1 -struct_name=Dao
	NativeTabBind(c context.Context, ids []int64, category int32) (map[int64]int64, error)
	// bts: -singleflight=true -paging=true -ignores=||start,end -nullcache=&dynmdl.ModuleIDsReply{IDs:[]int64{-1}} -check_null_code=$!=nil&&len($.IDs)==1&&$.IDs[0]==-1 -struct_name=Dao
	ModuleIDs(c context.Context, nid int64, pType int32, start int64, end int64) (res *dynmdl.ModuleIDsReply, err error)
	// bts: -singleflight=true -paging=true -ignores=||start,end -nullcache=&dynmdl.ModuleIDsReply{IDs:[]int64{-1}} -check_null_code=$!=nil&&len($.IDs)==1&&$.IDs[0]==-1 -struct_name=Dao
	NatMixIDs(c context.Context, moduleID int64, pType int32, start int64, end int64) (res *dynmdl.ModuleIDsReply, err error)
	// bts: -singleflight=true -paging=true -ignores=||start,end -nullcache=&dynmdl.ModuleIDsReply{IDs:[]int64{-1}} -check_null_code=$!=nil&&len($.IDs)==1&&$.IDs[0]==-1 -struct_name=Dao
	NatAllMixIDs(c context.Context, moduleID int64, start int64, end int64) (res *dynmdl.ModuleIDsReply, err error)
	// bts: -singleflight=true -paging=true -ignores=||start,end -nullcache=&dynmdl.ModuleIDsReply{IDs:[]int64{-1}} -check_null_code=$!=nil&&len($.IDs)==1&&$.IDs[0]==-1 -struct_name=Dao
	PartPids(c context.Context, pid int64, start int64, end int64) (res *dynmdl.ModuleIDsReply, err error)
	// bts: -struct_name=Dao
	NatIDsByActType(c context.Context, actType int64) ([]int64, error)
	// bts: -struct_name=Dao -singleflight=true -check_null_code=$!=nil&&$.ID==-1 -nullcache=&white_list.WhiteList{ID:-1}
	WhiteListByMid(c context.Context, mid int64) (*white_list.WhiteList, error)
	// bts: -struct_name=Dao -nullcache=&v1.NativeUserSpace{Id:-1} -check_null_code=$!=nil&&$.Id==-1
	UserSpaceByMid(c context.Context, mid int64) (*v1.NativeUserSpace, error)
	// bts: -singleflight=true -nullcache=&v1.NativePageExtend{Id:-1} -check_null_code=$!=nil&&$.Id==-1 -struct_name=Dao
	NativeExtend(c context.Context, pid int64) (res *v1.NativePageExtend, err error)
}

//go:generate kratos tool redisgen
type _redis interface {
	// redis: -key=natPageExtendKey -struct_name=Dao
	CacheNativeExtend(c context.Context, pid int64) (*v1.NativePageExtend, error)
	// redis: -key=natPageExtendKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheNativeExtend(c context.Context, pid int64, val *v1.NativePageExtend) error
	// redis: -key=natPageExtendKey -struct_name=Dao
	DelCacheNativeExtend(c context.Context, id int64) error
	// redis: -key=natTagIDExistKey -struct_name=Dao
	CacheNatTagIDExist(c context.Context, id int64) (int64, error)
	// redis: -key=natTagIDExistKey -expire=d.mcRegularExpire -encode=raw -struct_name=Dao
	AddCacheNatTagIDExist(c context.Context, id int64, val int64) error
	// redis: -key=natTagIDExistKey -struct_name=Dao
	DelCacheNatTagIDExist(c context.Context, id int64) error
	// redis: -key=nativeTabBindKey -struct_name=Dao
	DelCacheNativeTabBind(c context.Context, id int64, category int32) error
	// redis: -key=nativeParticipationKey -struct_name=Dao
	DelCacheNativePart(c context.Context, id int64) error
	// redis: -key=nativeTabSortKey -struct_name=Dao
	CacheNativeTabSort(c context.Context, id int64) ([]int64, error)
	// redis: -key=nativeTabSortKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
	AddCacheNativeTabSort(c context.Context, id int64, val []int64) error
	// redis: -key=nativeTabSortKey -struct_name=Dao
	DelNativeTabSort(c context.Context, id int64) error
	// redis: -key=nativeClickIDsKey -struct_name=Dao
	CacheNativeClickIDs(c context.Context, id int64) ([]int64, error)
	// redis: -key=nativeClickIDsKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
	AddCacheNativeClickIDs(c context.Context, id int64, val []int64) error
	// redis: -key=nativeClickIDsKey -struct_name=Dao
	DelNativeClickIDs(c context.Context, id int64) error
	// redis: -key=nativeDynamicIDsKey -struct_name=Dao
	CacheNativeDynamicIDs(c context.Context, id int64) ([]int64, error)
	// redis: -key=nativeDynamicIDsKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
	AddCacheNativeDynamicIDs(c context.Context, id int64, val []int64) error
	// redis: -key=nativeDynamicIDsKey -struct_name=Dao
	DelNativeDynamicIDs(c context.Context, id int64) error
	// redis: -key=nativeActIDsKey -struct_name=Dao
	CacheNativeActIDs(c context.Context, id int64) ([]int64, error)
	// redis: -key=nativeActIDsKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
	AddCacheNativeActIDs(c context.Context, id int64, val []int64) error
	// redis: -key=nativeActIDsKey -struct_name=Dao
	DelNativeActIDs(c context.Context, id int64) error
	// redis: -key=nativeVideoIDsKey -struct_name=Dao
	CacheNativeVideoIDs(c context.Context, id int64) ([]int64, error)
	// redis: -key=nativeVideoIDsKey -struct_name=Dao -expire=d.mcRegularExpire -encode=json
	AddCacheNativeVideoIDs(c context.Context, id int64, val []int64) error
	// redis: -key=nativeVideoIDsKey -struct_name=Dao
	DelNativeVideoIDs(c context.Context, id int64) error
	// redis: -key=nativeTabKey -struct_name=Dao
	DelCacheNativeTab(c context.Context, id int64) error
	// redis: -key=nativeTabModuleKey -struct_name=Dao
	DelCacheNativeTabModule(c context.Context, ids int64) error
	// redis: -key=nativePageKey -struct_name=Dao
	DelCacheNativePages(c context.Context, ids []int64) error
	// redis: -key=ntTsPageKey -struct_name=Dao
	DelCacheNtTsPages(c context.Context, ids []int64) error
	// redis: -key=ntTsModuleExtKey -struct_name=Dao
	DelCacheNtTsModulesExt(c context.Context, ids []int64) error
	// redis: -key=nativeForeignKey -struct_name=Dao
	CacheNativeForeign(c context.Context, id int64, pageType int64) (int64, error)
	// redis: -key=nativeForeignKey -struct_name=Dao
	DelCacheNativeForeign(c context.Context, id int64, pageType int64) error
	// redis: -key=nativeModuleKey -struct_name=Dao
	DelCacheNativeModules(c context.Context, ids []int64) error
	// redis: -key=nativeClickKey -struct_name=Dao
	DelCacheNativeClicks(c context.Context, ids []int64) error
	// redis: -key=nativeDynamicKey -struct_name=Dao
	DelCacheNativeDynamics(c context.Context, ids []int64) error
	// redis: -key=nativeVideoKey -struct_name=Dao
	DelCacheNativeVideos(c context.Context, ids []int64) error
	// redis: -key=nativeUkeyKey -struct_name=Dao
	CacheNativeUkey(c context.Context, pid int64, ukey string) (int64, error)
	// redis: -key=nativeUkeyKey -struct_name=Dao
	DelCacheNativeUkey(c context.Context, pid int64, ukey string) error
	// redis: -key=nativeMixtureKey -struct_name=Dao
	DelCacheNativeMixtures(c context.Context, ids []int64) error
	// redis: -key=natIDsByActTypeKey -struct_name=Dao
	CacheNatIDsByActType(c context.Context, actType int64) ([]int64, error)
	// redis: -key=natIDsByActTypeKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheNatIDsByActType(c context.Context, actType int64, val []int64) error
	// redis: -key=natIDsByActTypeKey -struct_name=Dao
	DelCacheNatIDsByActType(c context.Context, actType int64) error
	// redis: -key=userSpaceByMidKey -struct_name=Dao
	CacheUserSpaceByMid(c context.Context, mid int64) (*v1.NativeUserSpace, error)
	// redis: -key=userSpaceByMidKey -expire=d.mcRegularExpire -encode=json -struct_name=Dao
	AddCacheUserSpaceByMid(c context.Context, mid int64, val *v1.NativeUserSpace) error
	// redis: -key=userSpaceByMidKey -struct_name=Dao
	DelCacheUserSpaceByMid(c context.Context, mid int64) error
	// redis: -key=sponsoredUpKey -struct_name=Dao
	CacheSponsoredUp(c context.Context, mid int64) (bool, error)
}

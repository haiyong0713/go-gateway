package dao

import (
	"context"
	"fmt"
	"time"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"

	favMdl "git.bilibili.co/bapis/bapis-go/community/model/favorite"
	favSvc "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	ugcSeasonSvc "git.bilibili.co/bapis/bapis-go/ugc-season/service"
	"go-common/library/sync/errgroup.v2"
)

type favFolderInfo struct {
	Meta      model.FavFolderMeta
	Folder    *favMdl.Folder
	UgcSeason *ugcSeasonSvc.Season
}

type FavFoldersInfoOpt struct {
	Dev   *device.Device
	IP    string
	Metas []model.FavFolderMeta
	Mid   int64
}

func (d *dao) favFoldersInfo(ctx context.Context, req *favSvc.FoldersReq, resCh chan favFolderInfo) error {
	eg := errgroup.WithContext(ctx)
	const folderBatch = 20
	for i := 0; i < len(req.GetIds()); i += folderBatch {
		var part []*favSvc.FolderID
		if i+folderBatch > len(req.GetIds()) {
			part = req.Ids[i:]
		} else {
			part = req.Ids[i : i+folderBatch]
		}
		eg.Go(func(c context.Context) error {
			partReq := &favSvc.FoldersReq{Typ: req.Typ, Ids: part, Mid: req.Mid}
			fsInfo, err := d.favGRPC.Folders(c, partReq)
			if err != nil {
				return wrapDaoError(err, "favGRPC.Folders", partReq)
			}
			for i := range fsInfo.GetRes() {
				thisFolder := fsInfo.GetRes()[i]
				resCh <- favFolderInfo{Meta: model.FavFolderMeta{Typ: req.Typ, Mid: thisFolder.Mid, Fid: thisFolder.ID}, Folder: thisFolder}
			}
			return nil
		})
	}
	return eg.Wait()
}

//nolint:biligowordcheck
func (d *dao) FavFoldersInfo(ctx context.Context, opt FavFoldersInfoOpt) (ret map[string]model.FavFolder, err error) {
	fids := make(map[int32][]*favSvc.FolderID)
	for _, meta := range opt.Metas {
		fids[meta.Typ] = append(fids[meta.Typ], &favSvc.FolderID{Fid: meta.Fid, Mid: meta.Mid})
	}
	ret = make(map[string]model.FavFolder)
	eg := errgroup.WithContext(ctx)
	resCh := make(chan favFolderInfo, 3)
	for k := range fids {
		kp := k
		eg.Go(func(c context.Context) error {
			// ugc 合集类型获取
			if kp == model.FavTypeUgcSeason {
				ss := make([]int64, 0, len(fids[kp]))
				for _, f := range fids[kp] {
					ss = append(ss, f.Fid)
				}
				ugcSS, err := d.UgcSeasonsInfo(c, ss)
				if err != nil {
					return err
				}
				for i := range ugcSS {
					thisFolder := ugcSS[i]
					resCh <- favFolderInfo{Meta: model.FavFolderMeta{Typ: kp, Mid: thisFolder.Mid, Fid: thisFolder.ID}, UgcSeason: thisFolder}
				}
			} else {
				// 正常读取收藏夹
				req := &favSvc.FoldersReq{Ids: fids[kp], Typ: kp, Mid: opt.Mid}
				return d.favFoldersInfo(c, req, resCh)
			}
			return nil
		})
	}
	go func() {
		err = eg.Wait()
		close(resCh)
	}()
	var mids []int64
	midDedup := make(map[int64]struct{})
	for {
		res, ok := <-resCh
		if !ok {
			break
		}
		ret[res.Meta.Hash()] = model.FavFolder{CurrentMid: opt.Mid, Folder: res.Folder, UgcSeason: res.UgcSeason}
		// 收集收藏夹创建人信息
		var owner int64
		if res.Folder != nil {
			owner = res.Folder.Mid
		} else if res.UgcSeason != nil {
			owner = res.UgcSeason.Mid
		} else {
			panic(fmt.Sprintf("programmer error: illegal favFolderInfo(%+v)", res))
		}
		if _, ok := midDedup[owner]; !ok {
			if owner > 0 {
				mids = append(mids, owner)
			} else {
				log.Warnc(ctx, "unexpected folder owner<=0. Discarded. Meta(%+v), FavFolder(%+v), UgcSS(%+v)", res.Meta, res.Folder, res.UgcSeason)
				continue
			}
		}
		midDedup[owner] = struct{}{}
	}
	// 前一步有错误的话先返回
	if err != nil {
		return nil, err
	}

	eg2 := errgroup.WithContext(ctx)
	eg2.Go(func(c context.Context) error {
		d.fillFavFolderCover(c, FillFavFolderCoverOpt{Mid: opt.Mid, Dev: opt.Dev, RemoteIp: opt.IP}, ret)
		return nil
	})
	if len(mids) > 0 {
		eg2.Go(func(c context.Context) error {
			resp, err := d.UpInfoByMids(c, mids, opt.IP)
			if err != nil {
				return err
			}
			var ok bool
			for i := range ret {
				f := ret[i]
				f.OwnerInfo, ok = resp[f.GetOwnerMid()]
				if ok {
					ret[i] = f
				} else {
					log.Warnc(ctx, "FavFolder owner info not found for mid(%d) folder(%+v)", f.GetOwnerMid(), f)
				}
			}
			return nil
		})
	}
	err = eg2.Wait()
	return
}

type recentFavItem struct {
	Ftyp   int32
	Fid    int64
	Typ    int32
	Oid    int64
	Folder *favMdl.Folder
}

func (rfi recentFavItem) ItemHash() string {
	return fmt.Sprintf("%d-%d", rfi.Typ, rfi.Oid)
}

func (rfi recentFavItem) SetAvIfMatch(avInfos map[int64]model.ArchiveInfo, ep2Aids map[int32]int64) bool {
	switch rfi.Typ {
	case model.FavTypeVideo:
		if av, ok := avInfos[rfi.Oid]; ok && av.Arc != nil {
			rfi.Folder.Cover = av.Arc.Pic
			return true
		}
	case model.FavTypeOgv:
		if ep2Aids == nil {
			return false
		}
		if aid, ok := ep2Aids[int32(rfi.Oid)]; ok && aid != 0 {
			if av, ok := avInfos[aid]; ok && av.Arc != nil {
				rfi.Folder.Cover = av.Arc.Pic
				return true
			}
		}
	}
	return false
}

func (rfi recentFavItem) SetAuIfMatch(auInfos map[int64]model.SongItem) bool {
	if rfi.Typ == model.FavTypeAudio {
		if au, ok := auInfos[rfi.Oid]; ok {
			rfi.Folder.Cover = au.Cover
			return true
		}
	}
	return false
}

type FillFavFolderCoverOpt struct {
	Mid      int64
	Dev      *device.Device
	RemoteIp string
}

// 填充收藏夹封面信息
func (d *dao) fillFavFolderCover(ctx context.Context, opt FillFavFolderCoverOpt, folders interface{}) {
	if folders == nil {
		return
	}
	recent := make(map[string]recentFavItem)
	// 最多查询最近1个收藏稿件
	// 请勿随意更改，否则会race
	maxRecent := 1
	switch fs := folders.(type) {
	case []model.FavFolder:
		for _, f := range fs {
			if len(f.GetCover()) > 0 {
				continue
			}
			for i, ri := range f.GetRecentRes() {
				if i >= maxRecent {
					break
				}
				item := recentFavItem{Ftyp: f.Type, Fid: f.ID, Typ: ri.Typ, Oid: ri.Oid, Folder: f.Folder}
				recent[item.ItemHash()] = item
			}
		}
	case map[string]model.FavFolder:
		for _, f := range fs {
			if len(f.GetCover()) > 0 {
				continue
			}
			for i, ri := range f.GetRecentRes() {
				if i >= maxRecent {
					break
				}
				item := recentFavItem{Ftyp: f.Type, Fid: f.ID, Typ: ri.Typ, Oid: ri.Oid, Folder: f.Folder}
				recent[item.ItemHash()] = item
			}
		}
	default:
		panic(fmt.Sprintf("fillFavFolderCover: programmer error unknown type %T", folders))
	}
	if len(recent) == 0 {
		return
	}
	var epids []int32
	var aids, auids []int64
	for _, r := range recent {
		switch r.Typ {
		case model.FavTypeVideo:
			aids = append(aids, r.Oid)
		case model.FavTypeOgv:
			epids = append(epids, int32(r.Oid))
		case model.FavTypeAudio:
			auids = append(auids, r.Oid)
			continue
		case model.FavTypeMediaList, model.FavTypeUgcSeason:
			// 不用获取这两种类型资源的封面
			continue
		default:
			log.Warnc(ctx, "fillFavFolderCover: unknown recent item, Discarded (%+v)", r)
			continue
		}
	}
	var ep2Aids map[int32]int64
	var err error
	if len(epids) > 0 {
		ep2Aids, err = d.Epids2Aids(ctx, epids)
		if err != nil {
			log.Warnc(ctx, "fillFavFolderCover.Epids2Aids failed: %v Discarded", err)
		} else {
			for _, aid := range ep2Aids {
				aids = append(aids, aid)
			}
		}
	}

	eg := errgroup.WithContext(ctx)
	if len(aids) > 0 {
		eg.Go(func(c context.Context) error {
			arcInfos, err := d.ArchiveInfos(c, ArchiveInfoOpt{Aids: aids, Mid: opt.Mid, Device: opt.Dev})
			if err != nil {
				log.Warnc(c, "fillFavFolderCover.ArchiveInfos failed: %v Discarded", err)
				return nil
			}
			for _, r := range recent {
				if len(r.Folder.Cover) > 0 {
					continue
				}
				r.SetAvIfMatch(arcInfos, ep2Aids)
			}
			return nil
		})
	}
	if len(auids) > 0 {
		eg.Go(func(c context.Context) error {
			auInfos, err := d.SongInfosV1(c, SongInfosOpt{SongIds: auids, RemoteIP: opt.RemoteIp})
			if err != nil {
				log.Warnc(c, "fillFavFolderCover.SongInfos failed: %v Discarded", err)
				return nil
			}
			for _, r := range recent {
				if len(r.Folder.Cover) > 0 {
					continue
				}
				r.SetAuIfMatch(auInfos)
			}
			return nil
		})
	}

	_ = eg.Wait()
}

type FavFolderListOpt struct {
	Mid      int64
	FavTypes []int32
	Dev      *device.Device
	IP       string
}

//nolint:biligowordcheck
func (d *dao) FavFolderList(ctx context.Context, opt FavFolderListOpt) (ret []model.FavFolder, err error) {
	eg := errgroup.WithContext(ctx)
	resCh := make(chan *favMdl.Folder, 10)
	for _, tp := range opt.FavTypes {
		tpCopy := tp
		eg.Go(func(c context.Context) error {
			req := &favSvc.UserFoldersReq{
				Typ: tpCopy, Mid: opt.Mid, Vmid: opt.Mid,
			}
			folders, err := d.favGRPC.UserFolders(c, req)
			if err != nil {
				return wrapDaoError(err, "favGRPC.UserFolders", req)
			}
			for i := range folders.GetRes() {
				resCh <- folders.Res[i]
			}
			return nil
		})
	}
	go func() {
		err = eg.Wait()
		close(resCh)
	}()
	var mids []int64
	midDedup := make(map[int64]struct{})
	for {
		res, ok := <-resCh
		if !ok {
			break
		}
		ret = append(ret, model.FavFolder{CurrentMid: opt.Mid, Folder: res})
		// 收集收藏夹创建人信息
		if _, ok := midDedup[res.Mid]; !ok {
			mids = append(mids, res.Mid)
		}
		midDedup[res.Mid] = struct{}{}
	}
	eg2 := errgroup.WithContext(ctx)
	eg2.Go(func(c context.Context) error {
		d.fillFavFolderCover(c, FillFavFolderCoverOpt{Mid: opt.Mid, Dev: opt.Dev, RemoteIp: opt.IP}, ret)
		return nil
	})
	if len(mids) > 0 {
		eg2.Go(func(c context.Context) error {
			resp, err := d.UpInfoByMids(c, mids, opt.IP)
			if err != nil {
				return err
			}
			for i := range ret {
				f := ret[i]
				f.OwnerInfo = resp[f.GetOwnerMid()]
				ret[i] = f
			}
			return nil
		})
	}

	err = eg2.Wait()
	return
}

type FavFolderDetailPagedOpt struct {
	Mid      int64
	Folder   model.FavFolderMeta
	PageSize int32
	LastTp   int32
	LastAvid int64
}

type FavFolderDetailPagedResp struct {
	List  []model.FavItemDetail
	Total int32
}

// 只针对 收藏夹服务分页
func (d *dao) FavFolderDetailPaged(ctx context.Context, opt FavFolderDetailPagedOpt) (ret *FavFolderDetailPagedResp, err error) {
	// 可能是ugc合集收藏夹
	if opt.Folder.Typ == model.FavTypeUgcSeason {
		sDetail, err := d.UgcSeasonDetail(ctx, opt.Folder.Fid)
		if err != nil {
			return nil, err
		}
		list := sDetail.ToModelFavItemDetails(opt.Folder)
		return &FavFolderDetailPagedResp{
			List: list, Total: int32(len(list)),
		}, nil
	}
	ret = new(FavFolderDetailPagedResp)
	// 可能需要获取完整的收藏夹创建人信息
	if opt.Folder.Mid < _favPartitionBase {
		req := &favSvc.FoldersReq{
			Typ: opt.Folder.Typ, Mid: opt.Mid, Ids: []*favSvc.FolderID{{Fid: opt.Folder.Fid, Mid: opt.Folder.Mid}},
		}
		fInfos, err := d.favGRPC.Folders(ctx, req)
		if err != nil {
			return nil, wrapDaoError(err, "favGRPC.Folders", req)
		}
		if len(fInfos.GetRes()) != 1 {
			return nil, fmt.Errorf("unexpected fav folder num %d, expect 1", len(fInfos.GetRes()))
		}
		// 获取完整的收藏夹mid信息
		opt.Folder.Mid = fInfos.GetRes()[0].Mid
	}

	firstPage := false
	if opt.LastAvid == 0 {
		firstPage = true
	}
	// 收藏夹原生服务需要pagesize大一个才能保证返回size大小的页面
	if !firstPage {
		opt.PageSize += 1
	}
	req := &favSvc.FavoritesListReq{
		Tp: opt.LastTp, CurrentMid: opt.Mid, FavMid: opt.Folder.Mid,
		Fid: opt.Folder.Fid, FirstPage: firstPage, Desc: true, Ps: opt.PageSize, Avid: opt.LastAvid,
	}
	res, err := d.favGRPC.FavoritesList(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "favGRPC.FavoritesList", req)
		return
	}
	ret.Total = res.GetRes().GetPage().GetTotalCount()
	ret.List = append(ret.List, toFavItemModels(res.GetRes().GetList())...)
	return
}

type FavFolderDetailOpt struct {
	Mid     int64
	Folder  model.FavFolderMeta
	Anchor  model.FavItemMeta
	MaxSize int32
}

// 默认最大获取数量
const defaultMaxSize = 1000

// 收藏夹服务有分表
const _favPartitionBase = 100

// 获取单个收藏夹内的收藏元素信息 有最大数量限制
// 可以获取锚点附近的MaxSize个稿件
func (d *dao) FavFolderDetail(ctx context.Context, opt FavFolderDetailOpt) (ret []model.FavItemDetail, err error) {
	if opt.MaxSize <= 0 {
		opt.MaxSize = defaultMaxSize
	}
	if opt.Folder.Typ == model.FavTypeUgcSeason {
		ret, err = d.allFavFolderItems(ctx, opt.Folder, opt.Mid)
		if err != nil {
			return nil, err
		}
	} else {
		if opt.Folder.Mid < _favPartitionBase {
			req := &favSvc.FoldersReq{
				Typ: opt.Folder.Typ, Mid: opt.Mid, Ids: []*favSvc.FolderID{{Fid: opt.Folder.Fid, Mid: opt.Folder.Mid}},
			}
			fInfos, err := d.favGRPC.Folders(ctx, req)
			if err != nil {
				return nil, wrapDaoError(err, "favGRPC.Folders", req)
			}
			if len(fInfos.GetRes()) != 1 {
				return nil, fmt.Errorf("unexpected fav folder num %d, expect 1", len(fInfos.GetRes()))
			}
			// 获取完整的收藏夹mid信息
			opt.Folder.Mid = fInfos.GetRes()[0].Mid
		}
		// 正常从收藏夹获取
		req := &favSvc.FavoritesListReq{
			Tp: opt.Anchor.Otype, Avid: opt.Anchor.Oid,
			CurrentMid: opt.Mid, FavMid: opt.Folder.Mid, Fid: opt.Folder.Fid,
			Desc: true,
		}
		ret, err = d.favRes(ctx, req, opt.MaxSize)
		if err != nil {
			return nil, err
		}
	}
	if int32(len(ret)) <= opt.MaxSize {
		return ret, nil
	}
	// 超长 需要截断
	if opt.Anchor.Oid == 0 || opt.Anchor.Otype == 0 {
		return ret[0:opt.MaxSize], nil
	}
	var idx int32
	found := false
	for _, item := range ret {
		if item.Item.Type == opt.Anchor.Otype && item.Item.Oid == opt.Anchor.Oid {
			idx = item.Item.Index
			found = true
			break
		}
	}
	if !found {
		return nil, model.ErrAnchorNotFound
	}
	head, tail := model.CalculateIdx(idx, int32(len(ret)), opt.MaxSize)
	return ret[head:tail], nil
}

// 获取anchor两端size大小的内容
func (d *dao) favRes(ctx context.Context, req *favSvc.FavoritesListReq, size int32) (ret []model.FavItemDetail, err error) {
	eg := errgroup.WithContext(ctx)
	if req.Tp == 0 || req.Avid == 0 {
		// 无翻页信息 返回前size个
		return d.favResDo(ctx, req, size)
	}

	var befor, after []model.FavItemDetail
	eg.Go(func(c context.Context) (err error) {
		befor, err = d.favResDo(c, &favSvc.FavoritesListReq{
			Tp: req.Tp, Avid: req.Avid,
			CurrentMid: req.CurrentMid, FavMid: req.FavMid, Fid: req.Fid,
			Desc: false,
		}, size)
		return
	})
	eg.Go(func(c context.Context) (err error) {
		after, err = d.favResDo(c, &favSvc.FavoritesListReq{
			Tp: req.Tp, Avid: req.Avid,
			CurrentMid: req.CurrentMid, FavMid: req.FavMid, Fid: req.Fid,
			Desc: true,
		}, size)
		return
	})
	err = eg.Wait()
	if err != nil {
		return
	}
	ret = make([]model.FavItemDetail, 0, size*2)
	// 前面的是正序获取的，需要倒序回来
	for i := len(befor) - 1; i > 0; i-- {
		ret = append(ret, befor[i])
	}
	ret = append(ret, after...)

	return
}

func (d *dao) favResDo(ctx context.Context, req *favSvc.FavoritesListReq, size int32) (ret []model.FavItemDetail, err error) {
	req.Ps = 1000
	if req.Tp == 0 || req.Avid == 0 {
		req.FirstPage = true
	}
	ret = make([]model.FavItemDetail, 0, size)
	firstLoop := true
	for {
		res, err := d.favGRPC.FavoritesList(ctx, req)
		if err != nil {
			return nil, wrapDaoError(err, "favGRPC.FavoritesList", req)
		}
		if len(res.GetRes().GetList()) <= 0 || (len(ret) > 0 && len(res.GetRes().GetList()) <= 1) {
			return ret, nil
		}
		if firstLoop {
			ret = append(ret, toFavItemModels(res.GetRes().GetList())...)
		} else {
			ret = append(ret, toFavItemModels(res.GetRes().GetList()[1:])...)
		}
		if int32(len(ret)) >= size || res.GetRes().GetPage().GetTotalCount() <= ret[len(ret)-1].Item.Index+1 {
			break
		}
		firstLoop = false
		req.FirstPage = false
		req.Tp = ret[len(ret)-1].Item.Type
		req.Avid = ret[len(ret)-1].Item.Oid
	}
	return ret, nil
}

type FavFolderDetailsOpt struct {
	Mid     int64
	Folders []model.FavFolderMeta
}

type favFolderDts struct {
	Details []model.FavItemDetail
	Meta    model.FavFolderMeta
}

// 获取多个收藏夹内全部内容
//
//nolint:biligowordcheck
func (d *dao) FavFoldersDetail(ctx context.Context, opt FavFolderDetailsOpt) (ret map[string][]model.FavItemDetail, err error) {
	ret = make(map[string][]model.FavItemDetail)
	if len(opt.Folders) > 1 {
		eg := errgroup.WithContext(ctx)
		resCh := make(chan favFolderDts, 3)
		for i := range opt.Folders {
			idx := i
			eg.Go(func(c context.Context) error {
				meta := opt.Folders[idx]
				fdts, err := d.allFavFolderItems(c, meta, opt.Mid)
				if err != nil {
					return err
				}
				resCh <- favFolderDts{
					Details: fdts,
					Meta:    meta,
				}
				return nil
			})
		}
		go func() {
			err = eg.Wait()
			close(resCh)
		}()
		for {
			res, ok := <-resCh
			if !ok {
				break
			}
			ret[res.Meta.Hash()] = res.Details
		}
	} else {
		meta := opt.Folders[0]
		dts, err := d.allFavFolderItems(ctx, meta, opt.Mid)
		if err != nil {
			return nil, err
		}
		ret[meta.Hash()] = dts
	}
	return
}

// 获取单个收藏夹内所有内容
func (d *dao) allFavFolderItems(ctx context.Context, meta model.FavFolderMeta, mid int64) (ret []model.FavItemDetail, err error) {
	// 可能是ugc合集收藏夹
	if meta.Typ == model.FavTypeUgcSeason {
		sDetail, err := d.UgcSeasonDetail(ctx, meta.Fid)
		if err != nil {
			return nil, err
		}
		return sDetail.ToModelFavItemDetails(meta), nil
	}
	// 正常从收藏夹服务获取
	// 如果mid只有两位 那么先尝试获取完整的收藏夹信息
	if meta.Mid < _favPartitionBase {
		req := &favSvc.FoldersReq{
			Typ: meta.Typ, Mid: mid, Ids: []*favSvc.FolderID{{Fid: meta.Fid, Mid: meta.Mid}},
		}
		fInfos, err := d.favGRPC.Folders(ctx, req)
		if err != nil {
			return nil, wrapDaoError(err, "favGRPC.Folders", req)
		}
		if len(fInfos.GetRes()) != 1 {
			return nil, fmt.Errorf("unexpected fav folder num %d, expect 1", len(fInfos.GetRes()))
		}
		// 获取完整的收藏夹mid信息
		meta.Mid = fInfos.GetRes()[0].Mid
	}

	// 第一页
	req := &favSvc.FavoritesListReq{
		CurrentMid: mid,
		FavMid:     meta.Mid,
		Fid:        meta.Fid,
		FirstPage:  true,
		Ps:         1000,
		Desc:       true,
	}
	res, err := d.favGRPC.FavoritesList(ctx, req)
	if err != nil {
		return nil, wrapDaoError(err, "favGRPC.FavoritesList", req)
	}
	ret = append(ret, toFavItemModels(res.GetRes().GetList())...)
	// 对于视频收藏夹 继续读取后续页
	if meta.Typ == model.FavTypeVideo && len(ret) > 0 && res.GetRes().GetPage().GetTotalCount() > ret[len(ret)-1].Item.Index+1 {
		req.FirstPage = false
		for {
			req.Avid = ret[len(ret)-1].Item.Oid
			req.Tp = ret[len(ret)-1].Item.Type
			more, err := d.favGRPC.FavoritesList(ctx, req)
			if err != nil {
				return nil, wrapDaoError(err, "favGRPC.FavoritesList", req)
			}
			// 翻页的最上面一个是上一页的最后一个，跳过
			if len(more.GetRes().GetList()) > 1 {
				ret = append(ret, toFavItemModels(more.GetRes().GetList()[1:])...)
			}
			if more.GetRes().GetPage().GetTotalCount() <= ret[len(ret)-1].Item.Index+1 {
				break
			}
		}
	}
	return
}

type FavFolderCreateOpt struct {
	FolderType int32
	Name, Desc string
	Public     int32
	Mid        int64
}

func (d *dao) FavFolderCreate(ctx context.Context, opt FavFolderCreateOpt) (meta model.FavFolderMeta, err error) {
	req := &favSvc.AddFolderReq{
		Typ:         opt.FolderType,
		Mid:         opt.Mid,
		Name:        opt.Name,
		Description: opt.Desc,
		Public:      opt.Public,
	}
	reply, err := d.favGRPC.AddFolder(ctx, req)
	if err != nil {
		err = wrapDaoError(err, "favGRPC.AddFolder", req)
		return
	}
	return model.FavFolderMeta{Typ: opt.FolderType, Fid: reply.GetFid(), Mid: opt.Mid}, nil
}

type FavFolderDeleteOpt struct {
	Meta model.FavFolderMeta
	Dev  *device.Device
}

func (d *dao) FavFolderDelete(ctx context.Context, opt FavFolderDeleteOpt) (err error) {
	req := &favSvc.DelFolderReq{
		Typ:      opt.Meta.Typ,
		Mid:      opt.Meta.Mid,
		Fid:      opt.Meta.Fid,
		MobiApp:  opt.Dev.RawMobiApp,
		Platform: opt.Dev.RawPlatform,
		Device:   opt.Dev.Device,
	}
	_, err = d.favGRPC.DelFolder(ctx, req)

	return wrapDaoError(err, "favGRPC.DelFolder", req)
}

type FavItemAddOpt struct {
	Meta     model.FavItemAddAndDelMeta
	Device   *device.Device
	Network  *network.Network
	NoSilver bool
	NoReport bool
}

func (d *dao) FavItemAdd(ctx context.Context, opt FavItemAddOpt) (err error) {
	req := &favSvc.AddFavReq{
		Tp:       opt.Meta.Tp,
		Mid:      opt.Meta.Mid,
		Fid:      opt.Meta.Fid,
		Oid:      opt.Meta.Oid,
		Otype:    opt.Meta.Otype,
		MobiApp:  opt.Device.RawMobiApp,
		Platform: opt.Device.RawPlatform,
		Device:   opt.Device.Device,
		From:     favMdl.SourceFromListenerSingle,
	}
	var aid int64
	switch opt.Meta.Otype {
	case model.FavTypeVideo, model.FavTypeAudio:
		aid = opt.Meta.Oid
	case model.FavTypeOgv:
		if !opt.NoSilver {
			ep2aid, err := d.Epids2Aids(ctx, []int32{int32(opt.Meta.Oid)})
			if err != nil {
				return err
			}
			aid = ep2aid[int32(opt.Meta.Oid)]
		}
	}
	if !opt.NoSilver && aid > 0 && (opt.Meta.Otype == model.FavTypeOgv || opt.Meta.Otype == model.FavTypeVideo) {
		arcs, err := d.ArchiveInfos(ctx, ArchiveInfoOpt{Aids: []int64{aid}, Mid: opt.Meta.Mid, Device: opt.Device})
		if err != nil {
			return err
		}
		arcInfo := arcs[aid]
		if arcInfo.Arc != nil {
			silverOpt := interactSilverOpt{
				Action:   _silverActionVideoFav,
				Oid:      opt.Meta.Oid,
				ItemType: playItem2GaiaItemType[model.Fav2Play[opt.Meta.Otype]],
				Title:    arcInfo.Arc.GetTitle(),
				UpMid:    arcInfo.Arc.GetAuthor().Mid,
				PubTime:  arcInfo.Arc.GetPubDate().Time().Format(time.RFC3339),
				PlayNum:  arcInfo.Arc.GetStat().View,
			}
			resp, err := d.favSilver(ctx, silverOpt)
			if err != nil {
				log.Warnc(ctx, "SilverBullet failed to check fav event(%+v)", opt)
			} else {
				if resp.IsRejected() {
					return ErrSilverBulletHit
				}
			}
		}
	}
	_, err = d.favGRPC.AddFav(ctx, req)
	if err == nil {
		if !opt.NoReport {
			d.favActReport(ctx, opt.Device, opt.Network, opt.Meta.Mid, d.resolveFavItem(ctx, opt.Meta.Otype, opt.Meta.Oid), model.Fav2Play[opt.Meta.Otype], _actDo)
		}
	} else {
		err = wrapDaoError(err, "favGRPC.AddFav", req)
	}

	return err
}

type FavItemDeleteOpt struct {
	Meta     model.FavItemAddAndDelMeta
	Device   *device.Device
	Network  *network.Network
	NoReport bool
}

func (d *dao) FavItemDelete(ctx context.Context, opt FavItemDeleteOpt) (err error) {
	req := &favSvc.DelFavReq{
		Tp:       opt.Meta.Tp,
		Mid:      opt.Meta.Mid,
		Fid:      opt.Meta.Fid,
		Oid:      opt.Meta.Oid,
		Otype:    opt.Meta.Otype,
		MobiApp:  opt.Device.RawMobiApp,
		Platform: opt.Device.RawPlatform,
		Device:   opt.Device.Device,
	}
	_, err = d.favGRPC.DelFav(ctx, req)
	if err == nil {
		if !opt.NoReport {
			d.favActReport(ctx, opt.Device, opt.Network, opt.Meta.Mid, d.resolveFavItem(ctx, opt.Meta.Otype, opt.Meta.Oid), model.Fav2Play[opt.Meta.Otype], _actCancel)
		}
	} else {
		err = wrapDaoError(err, "favGRPC.DelFav", req)
	}

	return err
}

func toFavItemModels(favs []*favSvc.ModelFavorite) (ret []model.FavItemDetail) {
	ret = make([]model.FavItemDetail, 0, len(favs))
	for i := range favs {
		ret = append(ret, model.FavItemDetail{Item: favs[i]})
	}
	return
}

type FavoredInFoldersOpt struct {
	FolderTypes []int32
	OType       int32
	Oid, Mid    int64
}

//nolint:biligowordcheck
func (d *dao) FavoredInFolders(ctx context.Context, opt FavoredInFoldersOpt) (ret []model.FavFolderMeta, err error) {
	eg := errgroup.WithContext(ctx)
	resCh := make(chan model.FavFolderMeta, 3)
	for _, typ := range opt.FolderTypes {
		typCopy := typ
		eg.Go(func(c context.Context) error {
			req := &favSvc.FavoriteFolderIdsReq{
				Type: typCopy, Oid: opt.Oid, Otype: opt.OType, Mid: opt.Mid,
			}
			folders, err := d.favGRPC.FavoriteFolderIds(c, req)
			if err != nil {
				return wrapDaoError(err, "favGRPC.FavoriteFolderIds", req)
			}
			for _, f := range folders.GetFids() {
				resCh <- model.FavFolderMeta{
					Typ: typCopy,
					Fid: f,
					Mid: opt.Mid,
				}
			}
			return nil
		})
	}
	go func() {
		err = eg.Wait()
		close(resCh)
	}()
	for {
		res, ok := <-resCh
		if !ok {
			break
		}
		ret = append(ret, res)
	}
	return
}

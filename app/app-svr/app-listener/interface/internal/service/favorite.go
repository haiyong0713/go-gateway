package service

import (
	"context"
	"fmt"

	"go-common/component/metadata/device"
	"go-common/component/metadata/network"
	"go-common/library/ecode"
	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/internal/dao"
	"go-gateway/app/app-svr/app-listener/interface/internal/model"
	mainErr "go-gateway/ecode"

	"github.com/pkg/errors"
	"go-common/library/sync/errgroup.v2"
)

var (
	// 限制本接口可以查询的fav类型
	validFavTypes = map[int32]struct{}{
		model.FavTypeVideo:     {},
		model.FavTypeMediaList: {},
	}
	// 默认查询 视频收藏 和 订阅
	defaultFavTypes = []int32{model.FavTypeVideo, model.FavTypeMediaList}

	// 有效的收藏夹类型
	validFavFolderTypes = map[int32]struct{}{
		model.FavTypeVideo:     {},
		model.FavTypeMediaList: {},
		model.FavTypeUgcSeason: {},
	}
)

func validateFavTypes(_ context.Context, in []int32) error {
	for _, tp := range in {
		if _, ok := validFavTypes[tp]; !ok {
			return errors.WithMessagef(ecode.RequestErr, "illegal fav types %v", in)
		}
	}
	return nil
}

func validateFavFolder(_ context.Context, fid int64, ftype int32, fdTypes ...int32) error {
	if len(fdTypes) > 0 {
		found := false
		for _, fdt := range fdTypes {
			if ftype == fdt {
				found = true
				break
			}
		}
		if !found {
			return errors.WithMessagef(ecode.RequestErr, "illegal folder type %d", ftype)
		}
	} else if _, ok := validFavFolderTypes[ftype]; !ok {
		return errors.WithMessagef(ecode.RequestErr, "illegal folder type %d", ftype)
	}
	if fid <= 0 {
		return errors.WithMessagef(ecode.RequestErr, "illegal folder id %d", fid)
	}
	return nil
}

// convert  isFavItemAddReq_Item/isFavItemDelReq_Item to general favItem
func toFavItemMeta(ctx context.Context, obj interface{}) (ret model.FavItemAddAndDelMeta, err error) {
	var playItem *v1.PlayItem
	var favItem *v1.FavItem
	switch item := obj.(type) {
	case *v1.FavItemAddReq_Play:
		playItem = item.Play
	case *v1.FavItemAddReq_Fav:
		favItem = item.Fav
	case *v1.FavItemDelReq_Play:
		playItem = item.Play
	case *v1.FavItemDelReq_Fav:
		favItem = item.Fav
	case *v1.FavItemBatchReq_Play:
		playItem = item.Play
	case *v1.FavItemBatchReq_Fav:
		favItem = item.Fav
	default:
		err = fmt.Errorf("programmer error: unknown type %T", obj)
		return
	}

	if playItem != nil {
		err = validatePlayItem(ctx, playItem, 1)
		if err != nil {
			return
		}
		ret.Otype = model.Play2Fav[playItem.ItemType]
		switch playItem.ItemType {
		case model.PlayItemUGC, model.PlayItemOGV, model.PlayItemAudio:
			ret.Oid = playItem.Oid
		default:
			err = fmt.Errorf("unknown playItem %+v", *playItem)
			return
		}
	} else if favItem != nil {
		// 端上的favtype都是映射后的
		ret.Oid, ret.Otype = favItem.Oid, model.Play2Fav[favItem.ItemType]
		ret.Fid, ret.Mid = extractFidAndMid(favItem.Fid)
	} else {
		err = fmt.Errorf("unexpected empty playItem and favItem")
	}
	return
}

func (s *Service) FavItemBatch(ctx context.Context, req *v1.FavItemBatchReq) (resp *v1.FavItemBatchResp, err error) {
	if len(req.Actions) == 0 || req.Item.Size() == 0 {
		return nil, errors.WithMessage(ecode.RequestErr, "empty req.Action or req.Item")
	}
	dev, net, auth := DevNetAuthFromCtx(ctx)
	var delFrom, addTo []model.FavItemAddAndDelMeta
	for _, f := range req.Actions {
		if err = validateFavFolder(ctx, f.Fid, f.FolderType, model.FavTypeVideo); err != nil {
			return nil, err
		}
		meta, err := toFavItemMeta(ctx, req.Item)
		if err != nil {
			return nil, errors.WithMessagef(ecode.RequestErr, "can not convert FavItemBatch.Item to meta: %v", err)
		}
		meta.Tp = f.FolderType
		meta.Fid, meta.Mid = extractFidAndMid(f.Fid)
		meta.Mid = auth.Mid

		switch f.Action {
		case v1.FavFolderAction_ADD:
			addTo = append(addTo, meta)
		case v1.FavFolderAction_DEL:
			delFrom = append(delFrom, meta)
		default:
			log.Warnc(ctx, "unknown FavFolderAction (%v). Discarded", f.Action)
		}
	}

	eg := errgroup.WithContext(ctx)
	for i := range delFrom {
		iCopy := i
		eg.Go(func(c context.Context) error {
			err := s.dao.FavItemDelete(c, dao.FavItemDeleteOpt{Meta: delFrom[iCopy], Device: dev, Network: net, NoReport: len(addTo) > 0 || iCopy != len(delFrom)-1})
			if err != nil {
				return err
			}
			return nil
		})
	}
	for i := range addTo {
		iCopy := i
		eg.Go(func(c context.Context) error {
			err := s.dao.FavItemAdd(c, dao.FavItemAddOpt{Meta: addTo[iCopy], Device: dev, Network: net, NoReport: len(delFrom) > len(addTo) || iCopy != len(addTo)-1})
			if err != nil {
				return err
			}
			return nil
		})
	}
	err = eg.Wait()
	if err == nil {
		resp = new(v1.FavItemBatchResp)
		switch {
		case len(addTo) > 0:
			resp.Message = s.C.Res.Text.FavBatchAdd
		case len(delFrom) > 0:
			resp.Message = s.C.Res.Text.FavBatchDel
		default:
			resp.Message = s.C.Res.Text.FavBatch
		}
	} else if err == dao.ErrSilverBulletHit {
		return nil, errors.WithMessagef(mainErr.SliverBulletFavReject, "silverBullet rejected batch favAdd: mid(%d) req(%+v)", auth.Mid, req)
	}
	return
}

func (s *Service) FavItemAdd(ctx context.Context, req *v1.FavItemAddReq) (resp *v1.FavItemAddResp, err error) {
	if err = validateFavFolder(ctx, req.Fid, req.FolderType, model.FavTypeVideo); err != nil {
		return
	}
	dev, net, auth := DevNetAuthFromCtx(ctx)
	meta, err := toFavItemMeta(ctx, req.Item)
	if err != nil {
		return nil, errors.WithMessagef(ecode.RequestErr, "can not convert FavItemAddReq.Item to meta: %v", err)
	}
	meta.Tp = req.FolderType
	// 对于添加收藏的行为，理论上只会有 item 为 PlayItem 的请求
	meta.Fid, meta.Mid = extractFidAndMid(req.Fid)
	// 使用当前用户的mid覆盖请求的mid 避免恶意请求
	meta.Mid = auth.Mid

	err = s.dao.FavItemAdd(ctx, dao.FavItemAddOpt{Meta: meta, Device: dev, Network: net})
	if err != nil {
		if err == dao.ErrSilverBulletHit {
			return nil, errors.WithMessagef(mainErr.SliverBulletFavReject, "silverBullet rejected favAdd: mid(%d) req(%+v)", auth.Mid, req)
		}
		return
	}
	return &v1.FavItemAddResp{
		Message: s.C.Res.Text.AddFav,
	}, nil
}

func (s *Service) FavItemDel(ctx context.Context, req *v1.FavItemDelReq) (resp *v1.FavItemDelResp, err error) {
	if err = validateFavFolder(ctx, req.Fid, req.FolderType, model.FavTypeVideo); err != nil {
		return
	}
	dev, net, auth := DevNetAuthFromCtx(ctx)
	meta, err := toFavItemMeta(ctx, req.Item)
	if err != nil {
		return nil, errors.WithMessagef(ecode.RequestErr, "can not convert FavItemDelReq.Item to meta: %v", err)
	}
	meta.Tp = req.FolderType
	meta.Fid, meta.Mid = extractFidAndMid(req.Fid)
	meta.Mid = auth.Mid

	err = s.dao.FavItemDelete(ctx, dao.FavItemDeleteOpt{Meta: meta, Device: dev, Network: net})
	if err != nil {
		return
	}
	return &v1.FavItemDelResp{
		Message: s.C.Res.Text.DelFav,
	}, nil
}

func (s *Service) FavFolderList(ctx context.Context, req *v1.FavFolderListReq) (resp *v1.FavFolderListResp, err error) {
	if req.FolderTypes != nil && len(req.FolderTypes) > 0 {
		if err = validateFavTypes(ctx, req.FolderTypes); err != nil {
			return
		}
	} else {
		req.FolderTypes = defaultFavTypes
	}
	dev, net, auth := DevNetAuthFromCtx(ctx)
	allFolders, err := s.dao.FavFolderList(ctx, dao.FavFolderListOpt{Mid: auth.Mid, FavTypes: req.FolderTypes, Dev: dev, IP: net.RemoteIP})
	if err != nil {
		return
	}
	resp = new(v1.FavFolderListResp)

	// 过滤出收藏与订阅
	var collectedFolders []model.FavFolderMeta
	for _, f := range allFolders {
		switch f.Type {
		case model.FavTypeVideo:
			resp.List = append(resp.List, f.ToV1FavFolder())
		case model.FavTypeMediaList:
			collectedFolders = append(collectedFolders, model.FavFolderMeta{Typ: model.FavTypeMediaList, Mid: f.Mid, Fid: f.ID})
		default:
			log.Warnc(ctx, "unknown fav folder type %d. Discarded", f.Type)
		}
	}
	// 获取订阅收藏夹的信息
	if len(collectedFolders) > 0 {
		folders, err := s.dao.FavFoldersDetail(ctx, dao.FavFolderDetailsOpt{Mid: auth.Mid, Folders: collectedFolders})
		if err != nil {
			return nil, err
		}
		var toFill []model.FavItemDetail
		for _, fs := range folders {
			for i := range fs {
				toFill = append(toFill, fs[i])
			}
		}
		err = s.fillSubscription(ctx, toFill, auth.Mid, dev, net)
		if err != nil {
			return nil, err
		}
		for _, f := range toFill {
			resp.List = append(resp.List, f.ToV1FavFolder())
		}
	}
	if req.Item != nil {
		// 获取稿件所在收藏夹的信息
		if err := validatePlayItem(ctx, req.Item, 1); err != nil {
			return nil, err
		}
		folders, err := s.dao.FavoredInFolders(ctx, dao.FavoredInFoldersOpt{
			FolderTypes: req.FolderTypes, OType: model.Play2Fav[req.Item.ItemType],
			Oid: req.Item.Oid, Mid: auth.Mid,
		})
		if err != nil {
			return nil, err
		}
		if len(folders) > 0 {
			lkMap := make(map[string]struct{})
			for _, f := range folders {
				lkMap[f.Hash()] = struct{}{}
			}
			for i, f := range resp.List {
				if _, ok := lkMap[model.HashV1FavFolder(f)]; ok {
					resp.List[i].FavState = 1
				}
			}
		}
	}
	return
}

func (s *Service) fillSubscription(ctx context.Context, folders []model.FavItemDetail, mid int64, dev *device.Device, net *network.Network) (err error) {
	if len(folders) <= 0 {
		return nil
	}
	var favs []model.FavFolderMeta
	favLookup := make(map[string]int)
	for i, f := range folders {
		switch f.Item.Type {
		case model.FavTypeMediaList:
			// type = 11  Oid为收藏夹id
			meta := model.FavFolderMeta{Typ: f.Item.Type, Mid: f.Item.Oid % 100, Fid: f.Item.Oid / 100}
			favs = append(favs, meta)
			favLookup[meta.Hash()] = i
		case model.FavTypeUgcSeason:
			// type = 21 Oid为ugc合集id
			meta := model.FavFolderMeta{Typ: f.Item.Type, Fid: f.Item.Oid}
			favs = append(favs, meta)
			favLookup[meta.Hash()] = i
		}
	}
	fdInfos, err := s.dao.FavFoldersInfo(ctx, dao.FavFoldersInfoOpt{Mid: mid, Metas: favs, Dev: dev, IP: net.RemoteIP})
	if err != nil {
		return err
	}
	for metaHash := range fdInfos {
		if index, ok := favLookup[metaHash]; ok {
			fCopy := fdInfos[metaHash]
			folders[index].FavFolder = &fCopy
		}
	}
	return
}

//nolint:gocognit
func (s *Service) FavFolderDetail(ctx context.Context, req *v1.FavFolderDetailReq) (resp *v1.FavFolderDetailResp, err error) {
	// 1. 首先对request进行检查与预处理
	if err = validateFavFolder(ctx, req.Fid, req.FolderType, model.FavTypeVideo, model.FavTypeUgcSeason); err != nil {
		return nil, err
	}
	if req.PageSize <= 0 || req.PageSize > 30 {
		req.PageSize = 20
	}

	dev, net, auth := DevNetAuthFromCtx(ctx)

	var itemList []model.FavItemDetail
	folderMeta := model.FavFolderMeta{
		Typ: req.FolderType,
		Mid: req.FavMid,
		Fid: req.Fid,
	}
	if req.FolderType != model.FavTypeUgcSeason {
		var fmid int64
		folderMeta.Fid, fmid = extractFidAndMid(req.Fid)
		// 端上目前对于自己的收藏夹可能会不填mid
		// 如果自己的mid编码符合要求，则填入自己的mid
		if folderMeta.Mid == 0 && fmid == auth.Mid%100 {
			folderMeta.Mid = auth.Mid
		} else if folderMeta.Mid == 0 {
			// 否则，填入编码的id
			folderMeta.Mid = fmid
		}
	}

	var lAid int64
	var lTp int32
	if req.LastItem != nil {
		lAid = req.LastItem.Oid
		// 把端上的item type转换回收藏夹的类型
		req.LastItem.ItemType = model.Play2Fav[req.LastItem.ItemType]
		lTp = req.LastItem.ItemType
	}
	reply, err := s.dao.FavFolderDetailPaged(ctx, dao.FavFolderDetailPagedOpt{
		Mid: auth.Mid, PageSize: int32(req.PageSize), LastAvid: lAid, LastTp: lTp,
		Folder: folderMeta,
	})
	if err != nil {
		return nil, err
	}
	itemList = reply.List

	resp = new(v1.FavFolderDetailResp)
	resp.Total = uint32(reply.Total)

	// 2. 处理翻页逻辑
	if req.LastItem != nil {
		anchorFound := false
		for i, item := range itemList {
			// 假如找到last item，舍弃该item及其之前的所有item
			if item.Item.Oid == req.LastItem.Oid &&
				item.Item.Type == req.LastItem.ItemType {
				if i < len(itemList)-1 {
					// 数组越界
					itemList = itemList[i+1:]
				} else {
					itemList = []model.FavItemDetail{}
				}
				anchorFound = true
				break
			}
		}
		// 整个itemList中未找到lastItem
		if !anchorFound {
			return nil, errors.WithMessagef(ecode.RequestErr, "cannot find last_item in the favorite folder: req(%+v)", req)
		}
	}

	if req.FolderType == model.FavTypeUgcSeason {
		if uint32(len(itemList)) <= req.PageSize {
			resp.ReachEnd = true
		} else {
			resp.ReachEnd = false
			itemList = itemList[:req.PageSize]
		}
	} else if len(itemList) > 0 {
		// 收藏夹则通过idx值判断
		resp.ReachEnd = uint32(itemList[len(itemList)-1].Item.Index+1) >= resp.Total
	}
	if len(itemList) == 0 {
		resp.ReachEnd = true
		return
	}

	// 3. 遍历item列表，根据item类型，找出不同类型的id。
	var videoIDs, audioIDs []int64
	var epIDs []int32

	for _, item := range itemList {
		switch item.Item.Type {
		// 视频
		case model.FavTypeVideo:
			videoIDs = append(videoIDs, item.Item.Oid)
		// OGV
		case model.FavTypeOgv:
			epIDs = append(epIDs, int32(item.Item.Oid))
		// 音频
		case model.FavTypeAudio:
			audioIDs = append(audioIDs, item.Item.Oid)
		default:
			log.Warnc(ctx, "unknown fav item %+v, Discarded", item.Item)
		}
	}

	var epCards map[int32]model.EpCard
	if len(epIDs) > 0 {
		epCards, err = s.dao.OGVEpCards(ctx, dao.OGVEpCardOpt{Eps: epIDs})
		if err != nil {
			return nil, err
		}
		for _, ep := range epCards {
			videoIDs = append(videoIDs, ep.Ec.GetAid())
		}
	}

	// 3. 请求相应下游服务，获得item资源详细信息
	eg := errgroup.WithContext(ctx)

	var avInfos map[int64]model.ArchiveInfo
	var audioInfos map[int64]model.SongItem
	var filterAvs map[int64]string

	if len(videoIDs) > 0 {
		eg.Go(func(c context.Context) (err error) {
			// 获取稿件信息
			avInfos, err = s.dao.ArchiveInfos(c, dao.ArchiveInfoOpt{
				Aids:   videoIDs,
				Mid:    auth.Mid,
				Device: dev,
			})
			return err
		})
		eg.Go(func(c context.Context) (err error) {
			// 获取服务端负向过滤信息
			filterAvs, err = s.dao.FilterArchives(c, videoIDs)
			return err
		})
	}
	if len(audioIDs) > 0 {
		eg.Go(func(c context.Context) error {
			audioInfos, err = s.dao.SongInfosV1(c, dao.SongInfosOpt{SongIds: audioIDs, RemoteIP: net.RemoteIP})
			return err
		})
	}

	if req.NeedFolderInfo {
		// 如果客户端需要，一并返回收藏夹信息
		eg.Go(func(c context.Context) error {
			retMap, err := s.dao.FavFoldersInfo(c, dao.FavFoldersInfoOpt{Mid: auth.Mid, Metas: []model.FavFolderMeta{folderMeta}, Dev: dev, IP: net.RemoteIP})
			if err != nil {
				return err
			}
			if f, ok := retMap[folderMeta.Hash()]; ok {
				resp.FolderInfo = f.ToV1FavFolder()
			}
			return nil
		})
	}

	err = eg.Wait()
	if err != nil {
		return nil, err
	}

	// 4. 最终拼出结果response并返回
	fillOpt := model.FillV1FavItemOpt{
		ArchiveInfos: avInfos,
		AudioInfos:   audioInfos,
		FilterAvs:    filterAvs,
		EpCards:      epCards,
	}
	for _, item := range itemList {
		itemDetail := item.FillV1FavItemDetail(folderMeta, fillOpt)
		if itemDetail == nil {
			log.Errorc(ctx, "failed to get item detail for %+v in folder %+v. Discarded", item, folderMeta)
			continue
		}
		resp.List = append(resp.List, itemDetail)
	}
	return
}

func (s *Service) FavFolderCreate(ctx context.Context, req *v1.FavFolderCreateReq) (resp *v1.FavFolderCreateResp, err error) {
	if len(req.Name) == 0 || (req.Public>>1 != 0) {
		return nil, errors.WithMessagef(ecode.RequestErr, "empty req.Name or illegal req.Public val")
	}
	_, _, auth := DevNetAuthFromCtx(ctx)
	fd, err := s.dao.FavFolderCreate(ctx, dao.FavFolderCreateOpt{
		FolderType: model.FavTypeVideo, Name: req.Name, Desc: req.Desc,
		Public: req.Public, Mid: auth.Mid,
	})
	if err != nil {
		return nil, err
	}
	// 返回的fid并未编码
	return &v1.FavFolderCreateResp{
		FolderType: model.FavTypeVideo,
		Fid:        fd.EncodeFolderID(),
		Message:    s.C.Res.Text.CreateFavFolder,
	}, nil
}

func (s *Service) FavFolderDelete(ctx context.Context, req *v1.FavFolderDeleteReq) (resp *v1.FavFolderDeleteResp, err error) {
	if err = validateFavFolder(ctx, req.Fid, model.FavTypeVideo, model.FavTypeVideo); err != nil {
		return
	}
	dev, _, auth := DevNetAuthFromCtx(ctx)
	fid, _ := extractFidAndMid(req.Fid)
	err = s.dao.FavFolderDelete(ctx, dao.FavFolderDeleteOpt{
		Meta: model.FavFolderMeta{Typ: model.FavTypeVideo, Fid: fid, Mid: auth.Mid},
		Dev:  dev,
	})
	if err != nil {
		return nil, err
	}
	return &v1.FavFolderDeleteResp{
		Message: s.C.Res.Text.DeleteFavFolder,
	}, err
}

// 前端的fid为编码后的，与收藏服务交互需要分表的folder id（单表自增主键）和 partition number（mid后两位）
//
//nolint:gomnd
func extractFidAndMid(encoded int64) (fid, mid int64) {
	return encoded / 100, encoded % 100
}

func (s *Service) FavoredInAnyFolders(ctx context.Context, req *v1.FavoredInAnyFoldersReq) (resp *v1.FavoredInAnyFoldersResp, err error) {
	if err = validatePlayItem(ctx, req.Item, 1); err != nil {
		return nil, err
	}
	if len(req.FolderTypes) == 0 {
		req.FolderTypes = []int32{model.FavTypeVideo}
	}
	_, _, auth := DevNetAuthFromCtx(ctx)
	oid := req.Item.Oid
	folder, err := s.dao.FavoredInFolders(ctx, dao.FavoredInFoldersOpt{
		FolderTypes: req.FolderTypes, OType: model.Play2Fav[req.Item.ItemType],
		Oid: oid, Mid: auth.Mid,
	})
	if err != nil {
		return
	}
	resp = &v1.FavoredInAnyFoldersResp{Item: req.Item}
	for _, f := range folder {
		resp.Folders = append(resp.Folders, &v1.FavFolderMeta{
			FolderType: f.Typ, Fid: f.EncodeFolderID(),
		})
	}
	return
}

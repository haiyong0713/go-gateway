package service

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	sysMsgGRPC "git.bilibili.co/bapis/bapis-go/system-msg/interface"

	midrpc "git.bilibili.co/bapis/bapis-go/account/service"
	"github.com/jinzhu/gorm"

	"go-gateway/app/web-svr/space/ecode"

	xecode "go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/web-svr/space/admin/model"

	"go-common/library/sync/errgroup.v2"
)

func (s *Service) TopPhotoArcs(ctx context.Context, mids []int64) (map[int64]*model.TopPhotoArc, error) {
	if len(mids) == 0 {
		return nil, xecode.RequestErr
	}
	midSplitMap := make(map[int64][]int64)
	for _, v := range mids {
		midSplitMap[v%10] = append(midSplitMap[v%10], v)
	}
	eg := errgroup.WithContext(ctx)
	var mutex sync.Mutex
	res := make(map[int64]*model.TopPhotoArc, len(mids))
	for k, v := range midSplitMap {
		part := k
		partMids := v
		if len(partMids) == 0 {
			continue
		}
		eg.Go(func(ctx context.Context) error {
			var partData []*model.TopPhotoArc
			photoArc := &model.TopPhotoArc{Mid: part}
			if err := s.dao.DB.Table(photoArc.TableName()).Where("mid IN (?)", partMids).Find(&partData).Error; err != nil {
				log.Errorc(ctx, "TopPhotoArcs mids:%v error:%+v", partMids, err)
				return nil
			}
			mutex.Lock()
			for _, item := range partData {
				if item != nil {
					res[item.Mid] = item
				}
			}
			mutex.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "TopPhotoArcs mids:%v eg.Wait error:%v", mids, err)
		return nil, err
	}
	return res, nil
}

// GetTopPhotoList  获取头图审核列表
func (s *Service) GetTopPhotoList(ctx context.Context, params *model.MemberUploadTopPhotoSearchParams, pn int, ps int) (list *model.TopPhotoRes, err error) {
	var (
		topPhotos     []*model.MemberUploadTopPhoto
		topPhotoShows = make([]*model.MemberUploadTopPhotoShow, 0)
		backTimes     map[int64]int
		fans          map[int64]int64
		infos         map[int64]*midrpc.Info
		cards         map[int64]*midrpc.Card
		mids          = make([]int64, 0)
		total         int
	)

	// 头图审核信息
	if topPhotos, total, err = s.dao.GetMemberUploadTopPhotoByPage(params, pn, ps); err != nil {
		log.Error("TopPhoto Service: GetTopPhotoList GetMemberUploadTopPhotoByPage Error %+v", err)
		return nil, err
	}

	for _, topPhoto := range topPhotos {
		mids = append(mids, topPhoto.MID)
	}

	eg := errgroup.WithContext(ctx)
	// 驳回次数
	eg.Go(func(ctx context.Context) (err error) {
		if backTimes, err = s.dao.GetBackTimes(mids); err != nil {
			log.Errorc(ctx, "TopPhoto Service: GetTopPhotoList GetBackTimes Error %+v", err)
		}
		return err
	})
	//fans
	eg.Go(func(ctx context.Context) (err error) {
		if fans, err = s.dao.GetFans(mids); err != nil {
			log.Errorc(ctx, "TopPhoto Service: GetTopPhotoList GetFans Error %+v", err)
		}
		return err
	})

	// account info 账户信息
	eg.Go(func(ctx context.Context) (err error) {
		if infos, err = s.dao.GetAccountInfosByMIDs(mids); err != nil {
			log.Errorc(ctx, "TopPhoto Service: GetTopPhotoList GetAccountInfosByMIDs Error +%v", err)
		}
		return err
	})
	// account certification 账户认证信息
	eg.Go(func(ctx context.Context) (err error) {
		if cards, err = s.dao.GetAccountCertification(mids); err != nil {
			log.Errorc(ctx, "TopPhoto Service: GetTopPhotoList GetAccountCertification Error +%v", err)
		}
		return err
	})

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "Service GetTopPhotoList eg.Wait Error %+v", err)
		return nil, err
	}

	for _, photo := range topPhotos {
		topPhotoShow := &model.MemberUploadTopPhotoShow{
			MemberUploadTopPhoto: *photo,
			BackTimes:            backTimes[photo.MID],
			Fans:                 fans[photo.MID],
			Nickname:             infos[photo.MID].Name,
			Certification:        cards[photo.MID].Official.Type,
		}
		topPhotoShows = append(topPhotoShows, topPhotoShow)
	}

	pager := &model.Pager{
		CurrentPage: pn,
		PageSize:    ps,
		TotalItems:  total,
	}

	list = &model.TopPhotoRes{
		Items: topPhotoShows,
		Pager: pager,
	}

	return

}

// PassPhoto 通过
func (s *Service) PassPhoto(ctx context.Context, ids []int64, uname string, uid int64) (err error) {
	var (
		photos []*model.MemberUploadTopPhoto
	)

	if photos, err = s.dao.GetPhotoByIDs(ids); err != nil {
		log.Error("TopPhoto Service: PassPhoto GetPhotoByIDs Error %+v", err)
		return err
	}

	eg := errgroup.WithContext(ctx)

	for _, photo := range photos {
		_photo := photo
		eg.Go(func(ctx context.Context) error {
			return s.realPass(ctx, _photo, uname, uid)
		})
	}

	if err = eg.Wait(); err != nil {
		log.Errorc(ctx, "TopPhoto Service: PassPhoto realPass Error %+v", err)
		return err
	}

	return

}

// realPass 详细通过流程
func (s *Service) realPass(ctx context.Context, photo *model.MemberUploadTopPhoto, uname string, uid int64) (err error) {
	if photo == nil {
		return xecode.Error(xecode.RequestErr, "realPass参数为空")
	}

	if photo.Deleted == 1 || photo.Status != 0 {
		log.Warn("TopPhoto Service: realPass (%d) Invalid Item", photo.ID)
		return nil
	}

	var (
		vipInfo    *midrpc.VipInfo
		tx         = s.dao.DB.Begin()
		egAddition = errgroup.WithContext(context.Background()) //用于删除缓存等操作
	)
	defer func() {
		if err != nil {
			err = tx.Rollback().Error
		} else {
			err = tx.Commit().Error
		}
	}()

	// 审核状态变更，通过
	passUpdateMap := map[string]interface{}{
		"status": model.TOP_PHOTO_PASSED,
	}
	if err = s.dao.EditPhotoByID(photo.ID, passUpdateMap, tx); err != nil {
		log.Errorc(ctx, "TopPhoto Service: realPass EditPhotoByID (%d, %+v) Error %+v", photo.ID, passUpdateMap, err)
		return
	}

	// 审核记录更新
	if err = s.addVipAuditLog(photo, uname, model.VIP_AUDIT_LOG_REASON_PASS, "审核通过", tx); err != nil {
		log.Error("TopPhoto Service: realPass addVipAuditLog (%d) addVipAuditLog Error %+v", photo.ID, err)
		return
	}

	// 获取VIP信息
	if vipInfo, err = s.dao.GetVipInfo(photo.MID); err != nil {
		log.Error("TopPhoto Service: realPass (%d) GetVipInfo (%d) Error %+v", photo.ID, photo.MID, err)
		return
	}

	// 用户头图更新
	var (
		isActivated int
		platfrom    int
	)
	switch photo.PlatFrom {
	case model.MEMBER_UPLOAD_TOPPHOTO_FROM_IOS, model.MEMBER_UPLOAD_TOPPHOTO_FROM_ANDROID: // ios android
		isActivated = 0
		platfrom = model.TOPPHOTO_PLATFORM_MOBILE
	case model.MEMBER_UPLOAD_TOPPHOTO_FROM_IPAD, model.MEMBER_UPLOAD_TOPPHOTO_FROM_WEB: // ipad web
		isActivated = 1
		platfrom = model.TOPPHOTO_PLATFORM_CLIENT
	default:
		isActivated = 1
		platfrom = 0
	}

	// 删除
	if err = s.dao.DeleteByParams(photo.MID, platfrom, tx); err != nil {
		log.Error("TopPhoto Service: realPass (%d) DeleteByParams (%d) Error  %+v", photo.ID, photo.MID, err)
		return
	}
	//如果是web及ipad端先清除
	if isActivated == 1 {
		if err = s.dao.EditByMidToNotActivated(photo.MID, tx); err != nil {
			log.Error("TopPhoto Service: realPass (%d) EditByMidToNotActivated (%d) Error %+v", photo.ID, photo.MID, err)
			return
		}
	}
	//添加
	toAdd := &model.MemberTopPhoto{
		MID:         photo.MID,
		SID:         photo.ID,
		Expire:      vipInfo.DueDate / 1000,
		IsActivated: isActivated,
		PlatFrom:    platfrom,
		ModifyTime:  time.Now().Format("2006-01-02 15:04:05"),
	}
	if err = s.dao.AddTopPhoto(toAdd, tx); err != nil {
		log.Error("TopPhoto Service: PassPhoto (%d) AddTopPhoto Error %+v", photo.ID, err)
		return
	}

	egAddition.Go(func(ctx context.Context) (err error) {
		// 删除缓存
		var res *model.ResRaw
		if res, err = s.dao.PurgeCache(photo.MID); err != nil || res.Code != 0 {
			log.Error("TopPhoto Service: PassPhoto (%d) PurgeCache (%d) Error %+v", photo.ID, photo.MID, err)
		}

		// 行为日志
		if err = s.addPassActionLog(uname, uid, photo); err != nil {
			log.Error("TopPhoto Service: PassPhoto (%d) addPassActionLog Error %+v", photo.ID, err)
		}

		return nil
	})
	egAddition.Go(func(ctx context.Context) (err error) {
		// 删除头图缓存
		if err = s.dao.ClearCacheTopPhoto(ctx, photo.MID); err != nil {
			log.Error("TopPhoto Service: PassPhoto ClearCacheTopPhoto id:(%d), err:(%v)", photo.ID, err)
		}
		return nil
	})
	if err := egAddition.Wait(); err != nil {
		log.Error("TopPhoto Service: PassPhoto id:(%d), err:(%v)", photo.ID, err)
	}

	return
}

// RePass 驳回再通过
func (s *Service) RePass(id int64, uname string, uid int64) (err error) {
	var (
		ids      = []int64{id}
		list     []*model.MemberUploadTopPhoto
		imgPath  string
		nowRes   *model.MemberUploadTopPhoto
		validNow *model.MemberUploadTopPhoto
		bfsRes   *model.BfsRes
		tx       = s.dao.DB.Begin()
		eg       = errgroup.WithContext(context.Background())
	)
	defer func() {
		if err != nil {
			err = tx.Rollback().Error
		} else {
			err = tx.Commit().Error
		}
	}()

	if list, err = s.dao.GetPhotoByIDs(ids); err != nil {
		log.Error("TopPhoto Service: RePass (%d) GetPhotoByIDs Error %+v", id, err)
		return
	}
	if len(list) == 0 {
		return ecode.TopPhotoNotFound
	}

	toRePass := list[0]
	if toRePass == nil || toRePass.Status != 2 {
		log.Error("TopPhoto Service: RePass (%d) Invalid ID", id)
		return ecode.TopPhotoRequestError
	}

	if nowRes, err = s.dao.GetNowValidPhoto(toRePass.MID, toRePass.PlatFrom, toRePass.ID); err != nil {
		log.Error("TopPhoto Service: RePass GetNowValidPhoto (%+v) Error %+v", toRePass, err)
		return
	}

	if nowRes != nil && nowRes.ID != 0 {
		validNow = nowRes
		//如果现在已生效的后上传 以 后上传的 为准
		if validNow.UploadDate > toRePass.UploadDate {
			return
		}

		//否则以 repass的为准 删除已生效的
		updateMap := map[string]interface{}{
			"deleted": model.MEMBER_UPLOAD_TOP_PHOTO_DELETED,
		}
		if err = s.dao.EditPhotoByID(validNow.ID, updateMap, tx); err != nil {
			log.Error("TopPhoto Service: RePass (%d) EditPhotoDeleteById Error %+v", id, err)
			return
		}
	}

	//变更审核状态 并 更新到数据库
	newPath := strings.Replace(toRePass.ImgPath, "private", "space", 1)
	if err = s.dao.RePassPhotoEditByID(toRePass.ID, newPath, tx); err != nil {
		log.Error("TopPhoto Service: RePass (%d) RePassPhotoEditByID Error %+v", toRePass.ID, err)
		return
	}

	// VipAuditLog 审核记录更新
	if err = s.addVipAuditLog(toRePass, uname, model.VIP_AUDIT_LOG_REASON_PASS, "驳回再通过", tx); err != nil {
		log.Error("TopPhoto Service: RePass (%d) addVipAuditLog Error %+v", id, err)
		return
	}

	var (
		isActivated int
		platfrom    int
		vipInfo     *midrpc.VipInfo
	)
	switch toRePass.PlatFrom {
	case model.MEMBER_UPLOAD_TOPPHOTO_FROM_IOS, model.MEMBER_UPLOAD_TOPPHOTO_FROM_ANDROID: // ios android
		isActivated = 0
		platfrom = model.TOPPHOTO_PLATFORM_MOBILE

	case model.MEMBER_UPLOAD_TOPPHOTO_FROM_IPAD, model.MEMBER_UPLOAD_TOPPHOTO_FROM_WEB: // ipad web
		isActivated = 1
		platfrom = model.TOPPHOTO_PLATFORM_CLIENT

	default:
		isActivated = 1
		platfrom = 0

	}

	if vipInfo, err = s.dao.GetVipInfo(toRePass.MID); err != nil {
		log.Error("TopPhoto Service: RePass (%d) GetVipInfo (%d) Error %+v", id, toRePass.MID, err)
		return
	}

	if err = s.dao.DeleteByParams(toRePass.MID, platfrom, tx); err != nil {
		log.Error("TopPhoto Service: RePass (%d) DeleteByParams Error %+v", id, err)
		return
	}

	//如果是web及ipad端先清除
	if isActivated == 1 {
		if err = s.dao.EditByMidToNotActivated(toRePass.MID, tx); err != nil {
			log.Error("TopPhoto Service: RePass (%d) EditByMidToNotActivated (%d) Error %+v", id, toRePass.MID, err)
			return
		}
	}

	toAdd := &model.MemberTopPhoto{
		MID:         toRePass.MID,
		SID:         toRePass.ID,
		Expire:      vipInfo.DueDate / 1000,
		IsActivated: isActivated,
		PlatFrom:    platfrom,
		ModifyTime:  time.Now().Format("2006-01-02 15:04:05"),
	}

	if err = s.dao.AddTopPhoto(toAdd, tx); err != nil {
		log.Error("TopPhoto Service: RePass (%d) AddTopPhoto Error %+v", id, err)
		return
	}

	eg.Go(func(ctx context.Context) (err error) {
		// 将图片移动到 space 中
		if !strings.Contains(toRePass.ImgPath, "http") {
			imgPath = "/" + toRePass.ImgPath
		} else {
			imgPath = toRePass.ImgPath
		}
		if bfsRes, err = s.dao.BFSMove(imgPath, uname, "space"); err != nil || bfsRes.Code != 0 {
			log.Error("TopPhoto Service: RePass (%d) BFSMove Error %+v", id, err)
		}

		// 删除缓存
		var purgeRes *model.ResRaw
		if purgeRes, err = s.dao.PurgeCache(toRePass.MID); err != nil || purgeRes == nil || purgeRes.Code != 0 {
			log.Error("TopPhoto Service: RePass (%d) PurgeCache res %+v Error %+v", toRePass.MID, purgeRes, err)
		}

		// 行为日志
		if err = s.addRePassActionLog(uname, uid, toRePass); err != nil {
			log.Error("TopPhoto Service: RePass (%d) addRePassActionLog Error %+v", id, err)
		}

		return nil
	})

	return
}

// BackPhoto 驳回
func (s *Service) BackPhoto(toBack *model.BackPhotoParam, toBlock *model.AccountBlockParam, uname string, uid int64) (err error) {
	var (
		backeds       []*model.MemberUploadTopPhoto
		ids           = []int64{toBack.ID}
		reasonDefault string
		reason        int
		imgPath       string
		bfsRes        *model.BfsRes
		tx            = s.dao.DB.Begin()
		eg            = errgroup.WithContext(context.Background())
	)
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		tx.Commit()
	}()

	if backeds, err = s.dao.GetPhotoByIDs(ids); err != nil {
		log.Error("TopPhoto Service: BackPhoto (%d) GetPhotoByIDs Error %+v", toBack.ID, err)
		return
	}
	if len(backeds) == 0 {
		return ecode.TopPhotoNotFound
	}

	backed := backeds[0]
	if backed == nil || backed.Deleted == 1 {
		log.Error("TopPhoto Service: BackPhoto (%d) Invalid ID", toBack.ID)
		return ecode.TopPhotoRequestError
	}

	reasonDefault, reason = s.getReasonDefault(toBack.Reason, toBack.ReasonDefault)

	//审核状态及路径 数据库更新
	newPath := strings.Replace(backed.ImgPath, "space", "private", 1)
	up := map[string]interface{}{
		"status":   model.TOP_PHOTO_NOTPASSED,
		"deleted":  model.MEMBER_UPLOAD_TOP_PHOTO_DELETED,
		"img_path": newPath,
	}
	if err = s.dao.EditPhotoByID(backed.ID, up, tx); err != nil {
		log.Error("TopPhoto Service: BackPhoto (%d) EditPhotoByID Error %+v", backed.ID, err)
		return
	}

	// VipAuditLog 审核记录更新
	if err = s.addVipAuditLog(backed, uname, reason, reasonDefault, tx); err != nil {
		log.Error("TopPhoto Service: BackPhoto (%d) addVipAuditLog Error %+v", backed.ID, err)
		return
	}

	// 更新上一张审核通过的头图为激活状态
	var (
		res       *model.MemberUploadTopPhoto
		lastPhoto *model.MemberUploadTopPhoto
		platfrom  []int
	)
	if backed.PlatFrom == model.MEMBER_UPLOAD_TOPPHOTO_FROM_IOS || backed.PlatFrom == model.MEMBER_UPLOAD_TOPPHOTO_FROM_ANDROID {
		platfrom = []int{1, 2}
	} else {
		platfrom = []int{3, 4}
	}

	if res, err = s.dao.GetLastPhoto(backed.MID, platfrom, backed.ID); err != nil {
		log.Error("TopPhoto Service: BackPhoto (%d) GetLastPhoto Error %+v", backed.ID, err)
		return
	}

	if res != nil && res.ID != 0 {
		lastPhoto = res
		if err = s.processLastPhoto(lastPhoto, backed, tx); err != nil {
			log.Error("TopPhoto Service: BackPhoto processLastPhoto %+v error %v", lastPhoto, err)
			return
		}
	}

	//头图被驳回系统通知
	msg := &model.NotifySendInit{
		Mc:       model.NOTIFY_SEND_MC,
		Title:    "空间自定义头图强制退回通知",
		DataType: model.NOTIFY_DATA_TYPE,
		Context:  "对不起，您上传的头图存在违规行为，未通过原因:" + reasonDefault,
		MIDList:  backed.MID,
	}
	var notifyResRaw *sysMsgGRPC.AsyncSendUserNotifyResp
	if notifyResRaw, err = s.dao.SendNotify(msg); err != nil || notifyResRaw.ErrorCount != 0 {
		log.Error("TopPhoto Service: BackPhoto (%d) SendNotify Error %+v", backed.ID, err)
		return
	}

	//账号封锁
	if toBlock.AccountBlock != 0 {
		if err = s.accountBlock(backed.MID, toBlock, uname, uid); err != nil {
			log.Error("TopPhoto Service: BackPhoto (%d) accountBlock (%d) Error %+v", backed.ID, backed.MID, err)
			return
		}
	}
	//节操扣除
	if toBlock.Moral != 0 {
		if err = s.delMoral(backed.MID, toBlock, uname); err != nil {
			log.Error("TopPhoto Service: BackPhoto (%d) delMoral Error %+v", backed.ID, err)
			return
		}
	}

	eg.Go(func(ctx context.Context) (err error) {
		// 移动图片到private
		if !strings.Contains(backed.ImgPath, "http") {
			imgPath = "/" + backed.ImgPath
		} else {
			imgPath = backed.ImgPath
		}
		if bfsRes, err = s.dao.BFSMove(imgPath, uname, "private"); err != nil || bfsRes.Code != 0 {
			log.Error("TopPhoto Service: BackPhoto (%d) BFSMove Error %+v", backed.ID, err)
		}

		//删除缓存
		var resRaw *model.ResRaw
		if resRaw, err = s.dao.PurgeCache(backed.MID); err != nil || resRaw.Code != 0 {
			log.Error("TopPhoto Service: BackPhoto (%d) PurgeCache (%d) Error %+v", backed.ID, backed.MID, err)
		}

		//行为日志
		if err = s.addBackActionLog(uname, uid, backed); err != nil {
			log.Error("TopPhoto Service: BackPhoto (%d) addBackActionLog Error %+v", backed.ID, err)
		}

		return nil
	})

	return
}

func (s *Service) processLastPhoto(lastPhoto *model.MemberUploadTopPhoto, backed *model.MemberUploadTopPhoto, tx *gorm.DB) (err error) {
	if lastPhoto == nil || backed == nil {
		return xecode.Error(xecode.RequestErr, "processLastPhoto参数为空")
	}

	var (
		vipInfo       *midrpc.VipInfo
		existTopPhoto *model.MemberTopPhoto
	)

	if lastPhoto.Deleted == 1 {
		updateMap := map[string]interface{}{
			"deleted": model.MEMBER_UPLOAD_TOP_PHOTO_NOTDELETED,
		}
		if err = s.dao.EditPhotoByID(lastPhoto.ID, updateMap, tx); err != nil {
			log.Error("TopPhoto Service: BackPhoto (%d) EditPhotoDeleteById (%d) Error %+v", backed.ID, lastPhoto.ID, err)
			return
		}
	}

	// 取消所有激活的头图
	if err = s.dao.EditByMidToNotActivated(backed.MID, tx); err != nil {
		log.Error("TopPhoto Service: BackPhoto (%d) EditByMidToNotActivate (%d) Error %+v", backed.ID, backed.MID, err)
		return
	}

	if vipInfo, err = s.dao.GetVipInfo(backed.MID); err != nil {
		log.Error("TopPhoto Service: BackPhoto (%d) GetVipInfo (%d) Error %+v", backed.ID, backed.MID, err)
		return
	}
	//如果存在 更新 否则新加
	if existTopPhoto, err = s.dao.GetByMidAndSid(backed.MID, lastPhoto.ID); err != nil {
		log.Error("TopPhoto Service: BackPhoto (%d) GetByMidAndSid Error %+v", backed.ID, err)
		return
	}

	if existTopPhoto != nil && existTopPhoto.ID != 0 {
		var (
			set = map[string]interface{}{
				"expire":       vipInfo.DueDate / 1000,
				"is_activated": 1,
			}
			params = map[string]interface{}{
				"mid": backed.MID,
				"sid": lastPhoto.ID,
			}
		)

		if err = s.dao.EditByParams(backed.MID, set, params, tx); err != nil {
			log.Error("TopPhoto Service: BackPhoto (%d) EditByParams Error %+v", backed.ID, err)
			return
		}
	} else {
		var platFrom int
		switch backed.PlatFrom {
		case model.MEMBER_UPLOAD_TOPPHOTO_FROM_IOS, model.MEMBER_UPLOAD_TOPPHOTO_FROM_ANDROID:
			platFrom = model.TOPPHOTO_PLATFORM_MOBILE
		case model.MEMBER_UPLOAD_TOPPHOTO_FROM_IPAD, model.MEMBER_UPLOAD_TOPPHOTO_FROM_WEB:
			platFrom = model.TOPPHOTO_PLATFORM_CLIENT
		default:
			platFrom = 0
		}

		toAdd := &model.MemberTopPhoto{
			MID:         backed.MID,
			SID:         lastPhoto.ID,
			Expire:      vipInfo.DueDate / 1000,
			IsActivated: 1,
			PlatFrom:    platFrom,
			ModifyTime:  time.Now().Format("2006-01-02 15:04:05"),
		}

		if err = s.dao.AddTopPhoto(toAdd, tx); err != nil {
			log.Error("TopPhoto Service: BackPhoto (%d) AddTopPhoto Error %+v", backed.ID, err)
			return
		}
	}
	return
}

func (s *Service) delMoral(mid int64, param *model.AccountBlockParam, uname string) (err error) {

	delMoralInit := &model.DelMoralParam{
		MID:        mid,
		Delta:      -param.Moral,
		Origin:     2,
		Reason:     "空间头图违规",
		ReasonType: 5,
		Operator:   uname,
		Remark:     "空间头图违规",
		IsNotify:   0,
	}

	if err = s.dao.DelMoral(delMoralInit); err != nil {
		log.Error("TopPhoto Service: delMoral (%d) DelMoral Error %+v", delMoralInit.MID, err)
		return
	}

	return
}

func (s *Service) accountBlock(mid int64, param *model.AccountBlockParam, uname string, uid int64) (err error) {
	var (
		blockForever     = model.ACCOUNT_BLOCK_TEMP
		creditForever    = model.CREDIT_BLOCK_INFO_TEMP
		blockTime        = param.BlockTime
		creditInfoAddRes *model.ResRaw
	)
	//log.Info("block params in req: blocktime: %d,  if forever: %d", blockTime, blockForever)
	//log.Info("block param: %v", param)

	if param.BlockTime == -1 {
		blockTime = 0
		blockForever = model.ACCOUNT_BLOCK_FOREVER
		creditForever = model.CREDIT_BLOCK_INFO_FOREVER
	}
	//log.Info("block params before block: blocktime: %d,  if forever: %d", blockTime, blockForever)
	// 封禁账户
	accountBlockParams := &model.AccountBlockInit{
		MID:       mid,
		Source:    model.ACCOUNT_BLOCK_SOURCE,
		Area:      model.ACCOUNT_BLOCK_AREA,
		Action:    blockForever,
		Duration:  blockTime * 24 * 60 * 60,
		StartTime: time.Now().Unix(),
		OpID:      uid,
		Operator:  uname,
		Reason:    model.ACCOUNT_BLOCK_REASON[param.ReasonType],
		Comment:   param.BlockRemark,
		Notify:    param.BlockNotify,
	}

	if err = s.dao.BlockAccount(accountBlockParams); err != nil {
		log.Error("TopPhoto Service: accountBlock BlockAccount (%d) Error %+v", mid, err)
		return
	}

	// 封禁记录
	creditInfoToAdd := &model.BlockInfoAdd{
		MID:            mid,
		BlockedDays:    blockTime,
		BlockedForever: creditForever,
		BlockedRemark:  param.BlockRemark,
		MoralNum:       param.Moral,
		OriginType:     model.CREDIT_INFO_ORIGIN_TYPE,
		PunishTime:     time.Now().Unix(),
		PunishType:     model.CREDIT_PUNISH_TYPE,
		ReasonType:     param.ReasonType,
		OperID:         uid,
		OperatorName:   uname,
	}

	if creditInfoAddRes, err = s.dao.CreditInfoAdd(creditInfoToAdd); err != nil || creditInfoAddRes.Code != 0 {
		log.Error("TopPhoto Service: accountBlock CreditInfoAdd (%d) Error %+v", mid, err)
		return
	}

	return
}

// getReasonDefault
func (s *Service) getReasonDefault(reason int, reasonDefault string) (reReasonDefault string, reReason int) {

	if reason != -1 {
		reReasonDefault = model.BACK_REASON_REFLECT[reason]
		reReason = reason
	} else {
		reReasonDefault = reasonDefault
		reReason = -1
	}

	return
}

// addVipAuditLog 审核记录上报
func (s *Service) addVipAuditLog(photo *model.MemberUploadTopPhoto, operator string, reason int, reasonDefault string, tx *gorm.DB) (err error) {
	if photo == nil {
		return xecode.Error(xecode.RequestErr, "addVipAuditLog参数为空")
	}
	if tx == nil {
		tx = s.dao.DB.Begin()
		defer func() {
			if err != nil {
				err = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}()
	}

	vipAuditLog := &model.VipAuditLog{
		TID:           photo.ID,
		MID:           photo.MID,
		Operator:      operator,
		Reason:        strconv.Itoa(reason),
		ReasonDefault: reasonDefault,
		Ctime:         time.Now().Format("2006-01-02 15:04:05"),
	}

	var auditLogExist *model.VipAuditLog
	if auditLogExist, err = s.dao.GetByTID(photo.ID); err != nil {
		log.Error("TopPhoto Service: addVipAuditLog GetByTID (%d) Error %+v", photo.ID, err)
		return
	}

	if auditLogExist != nil && auditLogExist.ID != 0 {
		if err = s.dao.EditByTID(vipAuditLog, tx); err != nil {
			log.Error("TopPhoto Service: addVipAuditLog EditByTID (%d) Error %+v", photo.ID, err)
			return err
		}
	} else {
		if err = s.dao.AddLog(vipAuditLog, tx); err != nil {
			log.Error("TopPhoto Service: addVipAuditLog AddLog (%d) Error %+v", photo.ID, err)
			return err
		}
	}

	return
}

// addPassActionLog 审核通过行为日志上报
func (s *Service) addPassActionLog(uname string, uid int64, obj *model.MemberUploadTopPhoto) (err error) {
	now := time.Now()

	params := &model.AuditLogInitParams{
		UName:    uname,
		UID:      uid,
		Business: model.BUSINESS_TOP_PHOTO_ADMIN,
		Type:     model.TOP_PHOTO_PASSED,
		OID:      obj.ID,
		Action:   model.ACTIONPASS,
		CTime:    now,
		Index:    nil,
		Content:  obj,
	}

	if err = s.dao.AddAuditLog(params); err != nil {
		log.Error("TopPhoto Service: addPassActionLog Error %+v", err)
		return err
	}

	return

}

// addRePassActionLog 驳回再通过行为日志上报
func (s Service) addRePassActionLog(unmae string, uid int64, obj *model.MemberUploadTopPhoto) (err error) {
	now := time.Now()

	params := &model.AuditLogInitParams{
		UName:    unmae,
		UID:      uid,
		Business: model.BUSINESS_TOP_PHOTO_ADMIN,
		Type:     model.TOP_PHOTO_PASSED,
		OID:      obj.ID,
		Action:   model.ACTIONREPASS,
		CTime:    now,
		Index:    nil,
		Content:  obj,
	}

	if err = s.dao.AddAuditLog(params); err != nil {
		log.Error("TopPhoto Service: addBackActionLog Error %+v", err)
		return err
	}

	return

}

// addBackActionLog 审核驳回行为日志上报
func (s Service) addBackActionLog(uname string, uid int64, obj *model.MemberUploadTopPhoto) (err error) {
	now := time.Now()

	params := &model.AuditLogInitParams{
		UName:    uname,
		UID:      uid,
		Business: model.BUSINESS_TOP_PHOTO_ADMIN,
		Type:     model.TOP_PHOTO_NOTPASSED,
		OID:      obj.ID,
		Action:   model.ACTIONBACK,
		CTime:    now,
		Index:    nil,
		Content:  obj,
	}
	if err = s.dao.AddAuditLog(params); err != nil {
		log.Error("TopPhoto Service: addBackActionLog Error %+v", err)
		return err
	}

	return
}

// AuditLogList 获取Audit日志列表
func (s *Service) AuditLogList(ctx context.Context, params *model.VipAuditLogSearch, pn int, ps int) (res *model.VipAuditLogRes, err error) {
	var (
		resRaw                   = make([]*model.VipAuditLogResRaw, 0)
		total                    int
		vipAuditLogList          []*model.VipAuditLog
		memberUploadTopPhotoList []*model.MemberUploadTopPhoto
		tids                     = make([]int64, 0)
		mids                     = make([]int64, 0)
		fans                     map[int64]int64
		eg                       = errgroup.WithContext(ctx)
	)

	if vipAuditLogList, total, err = s.dao.GetAuditInfosByPage(params, pn, ps); err != nil {
		log.Error("Service AuditLogList GetAuditInfosByPage Error +%v", err)
		return nil, err
	}

	for _, vipAuditLog := range vipAuditLogList {
		tids = append(tids, vipAuditLog.TID)
		mids = append(mids, vipAuditLog.MID)
	}

	eg.Go(func(ctx context.Context) (err error) {
		if memberUploadTopPhotoList, err = s.dao.GetTopPhotoInfosByTIDs(params, tids); err != nil {
			log.Error("TopPhoto Service: AuditLogList GetTopPhotoInfosByTIDs Error %+v", err)
		}
		return err
	})

	eg.Go(func(ctx context.Context) (err error) {
		if fans, err = s.dao.GetFans(mids); err != nil {
			log.Errorc(ctx, "Dao: AuditLogList GetFans Error %+v", err)
		}
		return err
	})

	if err = eg.Wait(); err != nil {
		log.Error("TopPhoto Service: AuditLogList eg.wait Error %+v", err)
		return nil, err
	}

	for _, vipAuditInfo := range vipAuditLogList {
		for _, photo := range memberUploadTopPhotoList {
			if photo.ID == vipAuditInfo.TID {
				_vipAuditLog := vipAuditInfo
				tmpResRaw := &model.VipAuditLogResRaw{
					MemberUploadTopPhoto: *photo,
					Reason:               _vipAuditLog.Reason,
					ReasonDefault:        _vipAuditLog.ReasonDefault,
					Ctime:                _vipAuditLog.Ctime,
					Operator:             _vipAuditLog.Operator,
					Fans:                 fans[photo.MID],
				}
				resRaw = append(resRaw, tmpResRaw)
			}
		}
	}

	pager := &model.Pager{
		CurrentPage: pn,
		PageSize:    ps,
		TotalItems:  total,
	}

	res = &model.VipAuditLogRes{
		Items: resRaw,
		Pager: pager,
	}

	return
}

// GetActionLogList 获取行为日志列表
func (s *Service) GetActionLogList(ids string, pn int, ps int) (res *model.ActionLogRes, err error) {
	var (
		searchRes *model.LogSearchResRawData
	)

	if searchRes, err = s.dao.GetActionLog(ids, pn, ps); err != nil {
		log.Error("TopPhoto Service: GetActionLogList GetActionLog Error %+v", err)
		return nil, err
	}

	res = &model.ActionLogRes{
		Items: searchRes.Result,
		Pager: model.Pager{
			CurrentPage: searchRes.Page.Num,
			PageSize:    searchRes.Page.Size,
			TotalItems:  searchRes.Page.Total,
		},
	}

	return
}

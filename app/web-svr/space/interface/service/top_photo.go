package service

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/web-svr/space/ecode"
	api "go-gateway/app/web-svr/space/interface/api/v1"
	"go-gateway/app/web-svr/space/interface/model"

	accwar "git.bilibili.co/bapis/bapis-go/account/service"
	arcapi "git.bilibili.co/bapis/bapis-go/archive/service"
	archiveapi "git.bilibili.co/bapis/bapis-go/archive/service"
	cfcgrpc "git.bilibili.co/bapis/bapis-go/content-flow-control/service"

	"github.com/pkg/errors"
)

const (
	_defaultSid int64 = 1
	//web
	_defaultPlatfrom int64 = 0
	_isActive        int64 = 1
	_isDisable       int64 = 1
	//
	_defaultExpire = 40 * 365 * 24 * 3600
	//
	_addDayExpire   = 24 * 3600
	_addMonthExpire = 24 * 3600 * 31
)

const (
	_split = "bfs/"
	//
	_splitLen = 2
	//
	_banPhotoMallID = 8
)

func (s *Service) loadPhotoMallList() {
	list, err := s.dao.PhotoMallList(context.Background())
	if err != nil {
		log.Error("loadPhotoMallList s.dao.PhotoMallList error(%v)", err)
		return
	}
	if len(list) == 0 {
		log.Warn("loadPhotoMallList len(list) == 0")
		return
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].SortNum > list[j].SortNum
	})
	log.Info("loadPhotoMallList success len(list):%d", len(list))
	s.photoMallList = list
}

func (s *Service) PhotoMallList(c context.Context, arg *api.PhotoMallListReq) (*api.PhotoMallListReply, error) {
	if arg.GetMobiapp() != "iphone" && arg.GetMobiapp() != "android" && arg.GetMobiapp() != "iphone_i" && arg.GetMobiapp() != "android_i" {
		return nil, xecode.RequestErr
	}
	if len(s.photoMallList) == 0 {
		return nil, xecode.NothingFound
	}
	var list []*api.PhotoMall
	for _, v := range s.photoMallList {
		if v == nil || v.Type == "" {
			continue
		}
		var (
			match bool
			img   string
		)
		if arg.GetMobiapp() == "iphone" || arg.GetMobiapp() == "iphone_i" {
			match = regPhotoMallIphone.MatchString(v.Type)
			img = v.IphoneImg
		} else {
			match = regPhotoMallAndroid.MatchString(v.Type)
			img = v.AndroidImg
		}
		if !match || img == "" {
			continue
		}
		list = append(list, &api.PhotoMall{
			Id:       v.Id,
			Name:     v.ProductName,
			Img:      img,
			NightImg: v.ThumbnailImg,
		})
	}
	if len(list) == 0 {
		return nil, xecode.NothingFound
	}
	if arg.Mid > 0 {
		arg.Mobiapp = mobiappConvert(arg.Mobiapp)
		photo, err := s.dao.WebTopPhoto(c, arg.Mid, arg.Mid, arg.Mobiapp, arg.Device)
		if err != nil {
			log.Error("TopPhoto WebTopPhoto mid:%d mobiapp:%s error(%v)", arg.Mid, arg.Mobiapp, err)
		}
		if photo == nil || photo.Sid == 1 {
			for _, v := range list {
				if v.GetId() == s.c.Rule.DftPhotoID {
					v.IsActivated = 1
					break
				}
			}
		} else {
			for _, v := range list {
				if v.GetId() == photo.Sid {
					v.IsActivated = 1
					break
				}
			}
		}
	}
	return &api.PhotoMallListReply{List: list}, nil
}

// nolint:gocognit
func (s *Service) SetTopPhoto(ctx context.Context, arg *api.SetTopPhotoReq) (*api.NoReply, error) {
	var err error
	if arg.Type == api.TopPhotoType_UNKNOWN {
		arg.Type = api.TopPhotoType_PIC
	}
	switch arg.Type {
	case api.TopPhotoType_ARCHIVE:
		var isWhite bool
		if isWhite, err = s.isTopPhotoWhite(ctx, arg.Mid); err != nil || !isWhite {
			return nil, xecode.AccessDenied
		}
		var arcReply *archiveapi.ArcReply
		arcReply, err = s.arcClient.Arc(ctx, &archiveapi.ArcRequest{Aid: arg.ID})
		if err != nil {
			log.Errorc(ctx, "SetTopPhoto s.arcClient.Arc aid:%d error:%+v arg:%+v", arg.ID, err, arg)
			return nil, err
		}
		var cfcItem []*cfcgrpc.ForbiddenItem
		if err := retry(func() (err error) {
			cfcItem, err = s.contentFlowControlInfo(ctx, arg.ID)
			return err
		}); err != nil {
			log.Error("日志告警 contentFlowControlInfo error:%v", err)
		}
		arcForbidden := model.ItemToArcForbidden(cfcItem)
		if arcReply == nil ||
			arcReply.Arc == nil ||
			!arcReply.Arc.IsNormal() ||
			arcReply.Arc.Videos <= 0 ||
			arcForbidden.NoRecommend ||
			arcReply.Arc.AttrVal(arcapi.AttrBitIsPGC) == arcapi.AttrYes ||
			arcReply.Arc.AttrVal(arcapi.AttrBitSteinsGate) == arcapi.AttrYes ||
			arcReply.Arc.AttrVal(arcapi.AttrBitUGCPay) == arcapi.AttrYes {
			return nil, ecode.NotAllowedArc
		}
		if isAuthor := func() bool {
			if arcReply.Arc.Author.Mid == arg.Mid {
				return true
			}
			if arcReply.Arc.AttrVal(arcapi.AttrBitIsCooperation) == arcapi.AttrYes {
				for _, v := range arcReply.Arc.StaffInfo {
					if v != nil && v.Mid == arg.Mid {
						return true
					}
				}
			}
			return false
		}(); !isAuthor {
			return nil, ecode.SpaceNotAuthor
		}
		// 裁剪封面
		var cutURL string
		cutURL, err = s.cutTopPhotoArcCover(ctx, arcReply.Arc.Pic)
		if err != nil {
			log.Errorc(ctx, "SetTopPhoto cutTopPhotoArcCover mid:%d aid:%d pic:%s error:%+v", arg.Mid, arg.ID, arcReply.Arc.Pic, err)
			return nil, ecode.CutCoverFail
		}
		_, err = s.dao.AddTopPhotoArc(ctx, arg.Mid, arg.ID, cutURL)
		if err != nil {
			log.Error("SetTopPhoto s.dao.SetTopPhoto mid:%d id:%d mobiapp:%s error(%v)", arg.Mid, arg.ID, arg.Mobiapp, err)
			return nil, err
		}
		// update cache
		s.cache.Do(ctx, func(ctx context.Context) {
			if err := retry(func() error {
				return s.dao.DelCacheTopPhotoArc(ctx, arg.Mid)
			}); err != nil {
				log.Error("%+v", err)
			}
		})
	case api.TopPhotoType_PIC:
		arg.Mobiapp = mobiappConvert(arg.Mobiapp)
		err = s.dao.SetTopPhoto(ctx, arg.Mid, arg.ID, arg.Mobiapp)
		if err != nil {
			log.Error("SetTopPhoto s.dao.SetTopPhoto mid:%d id:%d mobiapp:%s error(%v)", arg.Mid, arg.ID, arg.Mobiapp, err)
			return nil, err
		}
	default:
		return nil, xecode.RequestErr
	}
	return &api.NoReply{}, nil
}

// nolint:gomnd
func (s *Service) cutTopPhotoArcCover(ctx context.Context, cover string) (string, error) {
	width, height, _, err := s.dao.DecodeImageSize(ctx, cover)
	if err != nil {
		return "", err
	}
	resultRatio, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(16)/float64(9)), 64)
	nowRatio, _ := strconv.ParseFloat(fmt.Sprintf("%.2f", float64(width)/float64(height)), 64)
	if resultRatio == nowRatio {
		return cover, nil
	}
	resultURL := func() string {
		// 获取原图后缀
		suffix := path.Ext(cover)
		// 裁宽度
		if nowRatio > resultRatio {
			resultWidth := int64(float64(height) * float64(16) / float64(9))
			widthStart := (width - resultWidth) / 2
			return fmt.Sprintf("%s@%d-%d-%d-%da%s", cover, widthStart, 0, resultWidth, height, suffix)
		}
		// 裁高度
		resultHeight := int64(float64(width) * float64(9) / float64(16))
		heightStart := (height - resultHeight) / 2
		return fmt.Sprintf("%s@%d-%d-%d-%da%s", cover, 0, heightStart, width, resultHeight, suffix)
	}()
	newFile, err := s.dao.ReadURLContent(ctx, resultURL)
	if err != nil {
		return "", err
	}
	newURL, err := s.dao.UploadBFS(ctx, newFile, true)
	if err != nil {
		return "", err
	}
	return newURL, nil
}

// nolint:gocognit
func (s *Service) TopPhoto(ctx context.Context, arg *api.TopPhotoReq) (*api.TopPhotoReply, error) {
	var needArcBuild bool
	if (arg.Mobiapp == "android" && arg.Build >= s.c.Rule.TopPhotoArcBuild.Android) ||
		(arg.Mobiapp == "iphone" && arg.Device == _devicePhone && arg.Build > s.c.Rule.TopPhotoArcBuild.Iphone) {
		needArcBuild = true
	}
	// 针对国际版特殊处理，强行转成粉板获取大会员头图，由于space里面有版本控制，无法在app-interface层进行强转
	topPhotoMobiAPP := arg.Mobiapp
	if (arg.Mobiapp == "android_i" && arg.Build > s.c.Rule.DftPhotoBuild.AndroidI) || (arg.Mobiapp == "android_b") || (arg.Mobiapp == "android_hd") {
		topPhotoMobiAPP = "android"
	} else if arg.Mobiapp == "iphone_i" || arg.Mobiapp == "iphone_b" {
		topPhotoMobiAPP = "iphone"
	}
	eg := errgroup.WithContext(ctx)
	var (
		imageURL, nightImg string
		sid                int64
	)
	eg.Go(func(ctx context.Context) error {
		photo, photoErr := s.dao.WebTopPhoto(ctx, arg.Mid, arg.LoginMid, topPhotoMobiAPP, arg.Device)
		if photoErr != nil {
			log.Error("TopPhoto WebTopPhoto mid:%d mobiapp:%s error(%v)", arg.Mid, arg.Mobiapp, photoErr)
		}
		if photo == nil ||
			(arg.Mobiapp == "android" && photo.AndroidImg == "") ||
			(arg.Mobiapp == "iphone" && photo.IphoneImg == "") ||
			(arg.Mobiapp == "iphone_i" && photo.IphoneImg == "") ||
			(arg.Mobiapp == "android_i" && photo.AndroidImg == "") ||
			(arg.Mobiapp == "android_hd" && photo.AndroidImg == "") {
			photo = new(model.TopPhoto)
			if arg.Device != "pad" &&
				((arg.Mobiapp == "android" && arg.Build > s.c.Rule.DftPhotoBuild.Android) ||
					(arg.Mobiapp == "iphone" && arg.Build > s.c.Rule.DftPhotoBuild.Iphone) ||
					(arg.Mobiapp == "android_i" && arg.Build > s.c.Rule.DftPhotoBuild.AndroidI) ||
					(arg.Mobiapp == "iphone_i") ||
					(arg.Mobiapp == "android_hd")) {
				for _, v := range s.photoMallList {
					if v != nil && v.Id == s.c.Rule.DftPhotoID {
						photo = &model.TopPhoto{
							SImg:         v.SImg,
							LImg:         v.LImg,
							AndroidImg:   v.AndroidImg,
							IphoneImg:    v.IphoneImg,
							IpadImg:      v.IpadImg,
							ThumbnailImg: v.ThumbnailImg,
							Sid:          v.Id,
						}
					}
				}
			}
		}
		imageURL = photo.AndroidImg
		if arg.Mobiapp == "iphone" || arg.Mobiapp == "iphone_i" {
			imageURL = photo.IphoneImg
		}
		nightImg = photo.ThumbnailImg
		sid = photo.Sid
		return nil
	})
	var topArc *model.TopPhotoArc
	if needArcBuild {
		eg.Go(func(ctx context.Context) error {
			if isWhite, err := s.isTopPhotoWhite(ctx, arg.Mid); err != nil || !isWhite {
				needArcBuild = false
				return nil
			}
			var topArcErr error
			topArc, topArcErr = s.dao.TopPhotoArc(ctx, arg.Mid)
			if topArcErr != nil {
				log.Error("TopPhoto s.dao.TopPhotoArc mid:%d error:%+v", arg.Mid, topArcErr)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Errorc(ctx, "TopPhoto arg:%+v eg.Wait error:%+v", arg, err)
		return nil, err
	}
	if arg.Device == _devicePad {
		nightImg = ""
	}
	res := &api.TopPhotoReply{TopPhoto: &api.TopPhoto{
		ImgUrl:      imageURL,
		NightImgUrl: nightImg,
		Sid:         sid,
	}}
	if needArcBuild {
		res.TopPhotoArc = &api.TopPhotoArc{Show: true}
		if topArc != nil {
			res.TopPhotoArc.Aid = topArc.Aid
			res.TopPhotoArc.Pic = topArc.ImageUrl
		}
	}
	return res, nil
}

func (s *Service) TopPhotoArc(ctx context.Context, mid int64) (*api.TopPhotoArc, error) {
	isWhite, err := s.isTopPhotoWhite(ctx, mid)
	res := &api.TopPhotoArc{}
	if err != nil || !isWhite {
		return res, nil
	}
	res.Show = true
	topArc, err := s.dao.TopPhotoArc(ctx, mid)
	if err != nil {
		log.Error("TopPhotoArc s.dao.TopPhotoArc mid:%d error:%+v", mid, err)
		return nil, err
	}
	if topArc != nil {
		res.Aid = topArc.Aid
		res.Pic = topArc.ImageUrl
	}
	return res, nil
}

func (s *Service) TopPhotoArcCancel(ctx context.Context, arg *api.TopPhotoArcCancelReq) (*api.NoReply, error) {
	if isWhite, err := s.isTopPhotoWhite(ctx, arg.Mid); err != nil || !isWhite {
		return &api.NoReply{}, nil
	}
	if _, err := s.dao.TopPhotoArcCancel(ctx, arg.Mid); err != nil {
		log.Error("TopPhotoArcCancel s.dao.TopPhotoArcCancel mid:%d error:%v", arg.Mid, err)
		return nil, err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		if err := retry(func() error {
			return s.dao.DelCacheTopPhotoArc(ctx, arg.Mid)
		}); err != nil {
			log.Error("%+v", err)
		}
	})
	return &api.NoReply{}, nil
}

func (s *Service) ClearTopPhotoArc(ctx context.Context, msg string) error {
	var m struct {
		Table string `json:"table"`
		Old   struct {
			Mid      int64 `json:"mid"`
			Platform int   `json:"platfrom"`
		} `json:"old,omitempty"`
		New struct {
			Mid      int64 `json:"mid"`
			Platform int   `json:"platfrom"`
		} `json:"new,omitempty"`
	}
	if err := json.Unmarshal([]byte(msg), &m); err != nil || m.Table == "" {
		log.Error("ClearTopPhotoArc json.Unmarshal msg(%s) error(%v)", msg, err)
		return err
	}
	log.Info("ClearTopPhotoArc json.Unmarshal msg(%s)", msg)
	// 手机端头图变更时清视频头图
	if m.New.Platform != 2 || m.New.Mid <= 0 {
		return nil
	}
	return s.ClearTopPhotoArcByMid(ctx, m.New.Mid)
}

func (s *Service) ClearTopPhotoArcByMid(ctx context.Context, mid int64) error {
	hasArc := func() bool {
		topPhotoArc, err := s.dao.TopPhotoArc(ctx, mid)
		if err != nil {
			log.Error("ClearTopPhotoArcByMid TopPhotoArc mid:%d error:%+v", mid, err)
			return false
		}
		if topPhotoArc != nil && topPhotoArc.Aid > 0 {
			return true
		}
		return false
	}()
	// 已经无头图数据不用更新
	if !hasArc {
		return nil
	}
	if _, err := s.dao.TopPhotoArcCancel(ctx, mid); err != nil {
		log.Error("ClearTopPhotoArc s.dao.AddTopPhotoArc mid:%d error:%v", mid, err)
		return err
	}
	s.cache.Do(ctx, func(ctx context.Context) {
		if err := retry(func() error {
			return s.dao.DelCacheTopPhotoArc(ctx, mid)
		}); err != nil {
			log.Error("%+v", err)
		}
	})
	return nil
}

//func (s *Service) clearMemberUploadTopPhoto(ctx context.Context, msg string) error {
//	var m struct {
//		Table  string `json:"table"`
//		Action string `json:"action"`
//		Old    struct {
//			Mid      int64 `json:"mid"`
//			Platform int   `json:"platfrom"`
//			Deleted  int   `json:"deleted"`
//		} `json:"old,omitempty"`
//		New struct {
//			Mid      int64 `json:"mid"`
//			Platform int   `json:"platfrom"`
//			Deleted  int   `json:"deleted"`
//		} `json:"new,omitempty"`
//	}
//	if err := json.Unmarshal([]byte(msg), &m); err != nil || m.Table == "" {
//		log.Error("clearMemberUploadTopPhoto json.Unmarshal msg(%s) error(%v)", msg, err)
//		return err
//	}
//	if m.Action != "insert" {
//		return nil
//	}
//	// 只处理手机端变更
//	if (m.New.Platform != 1 && m.New.Platform != 2) || m.New.Mid <= 0 || m.New.Deleted != 0 {
//		return nil
//	}
//	return s.ClearTopPhotoArcByMid(ctx, m.New.Mid)
//}

func (s *Service) isTopPhotoWhite(ctx context.Context, mid int64) (bool, error) {
	whiteReq, err := s.dao.Whitelist(ctx, &api.WhitelistReq{Mid: mid})
	if err != nil {
		log.Error("TopPhotoArcCancel s.dao.Whitelist mid:%d error:%+v", mid, err)
		return false, err
	}
	if whiteReq == nil {
		log.Warn("TopPhotoArcCancel s.dao.Whitelist mid:%d req is nil", mid)
		return false, nil
	}
	return whiteReq.IsWhite, nil
}

func retry(callback func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		if err = callback(); err == nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return err
}

func (s *Service) contentFlowControlInfo(ctx context.Context, aid int64) ([]*cfcgrpc.ForbiddenItem, error) {
	ts := time.Now().Unix()
	params := url.Values{}
	params.Set("source", s.c.CfcSvrConfig.Source)
	params.Set("business_id", strconv.FormatInt(s.c.CfcSvrConfig.BusinessID, 10))
	params.Set("ts", strconv.FormatInt(ts, 10))
	params.Set("oid", strconv.FormatInt(aid, 10))
	req := &cfcgrpc.FlowCtlInfoReq{
		Oid:        aid,
		BusinessId: int32(s.c.CfcSvrConfig.BusinessID),
		Source:     s.c.CfcSvrConfig.Source,
		Sign:       getSign(params, s.c.CfcSvrConfig.Secret),
		Ts:         ts,
	}
	reply, err := s.cfcGRPC.Info(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if reply == nil {
		return nil, nil
	}
	return reply.ForbiddenItems, nil
}

func getSign(params url.Values, secret string) string {
	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	var buf bytes.Buffer
	buf.WriteString(tmp)
	buf.WriteString(secret)
	mh := md5.Sum(buf.Bytes())
	return hex.EncodeToString(mh[:])
}

// SetWebTopphoto .
func (s *Service) SetWebTopphoto(c context.Context, mid, sid int64) (err error) {
	var (
		mall     *model.PhotoMall
		topphoto *model.MemberTopphoto
	)
	for _, photoMall := range s.photoMallList {
		if sid == photoMall.Id {
			mall = photoMall
			break
		}
	}
	if mall == nil {
		return ecode.SpaceTPMallError
	}
	if topphoto, err = s.dao.MemberTopphoto(c, mid); err != nil {
		return
	}
	if topphoto != nil && topphoto.Sid == sid && topphoto.IsActivated == _isActive {
		return
	}
	if mall.IsDisable == _isDisable {
		return ecode.SpaceSetTPError
	}
	if mall.Price > 0 {
		return ecode.SpacePayTPError
	}
	defaultExpire := time.Now().Unix() + _defaultExpire
	if err = s.dao.MemSetTopPhoto(c, mid, sid, defaultExpire); err != nil {
		return
	}
	_ = s.cache.Do(c, func(ctx context.Context) {
		_ = s.ClearTopPhotoCache(c, mid, model.UploadTopPhotoWeb)
	})
	return nil
}

// MemPhotoMall .
func (s *Service) ClearTopPhotoCache(c context.Context, mid int64, platFrom int) (err error) {
	if err = s.dao.DelCacheMemberUploadTopphoto(c, mid, platFrom); err != nil {
		log.Error("ClearTopPhotoCache mid(%d) error(%v)", mid, err)
	}
	log.Info("ClearTopPhotoCache DelCacheMemberUploadTopphoto mid(%d) platFrom(%d) err(%v)", mid, platFrom, err)
	if err = s.dao.DelCacheMemTopphotoCache(c, mid); err != nil {
		log.Error("ClearTopPhotoCache mid(%d) error(%v)", mid, err)
	}
	log.Info("ClearTopPhotoCache DelCacheMemTopphotoCache mid(%d) platFrom(%d) err(%v)", mid, platFrom, err)
	return
}

// MemPhotoMall .
func (s *Service) MemPhotoMall(c context.Context, vmid int64) (*model.MemberPhotoMallIndex, error) {
	photo, err := s.dao.MemberTopphoto(c, vmid)
	if err != nil {
		return nil, errors.Wrapf(ecode.TopPhotoNotFound, "%s", err)
	}
	if photo == nil || (photo.Sid == 8 && photo.Platfrom == 0) {
		photo = &model.MemberTopphoto{
			Sid:      _defaultSid,
			Platfrom: _defaultPlatfrom,
			Expire:   time.Now().Unix() + 40*365*24*3600,
		}
	}
	res := &model.MemberPhotoMallIndex{
		SID:      photo.Sid,
		Expire:   photo.Expire,
		Platform: photo.Platfrom,
	}
	if photo.Platfrom == _defaultPlatfrom {
		//default top photo
		var isFound bool
		for _, photoMall := range s.photoMallList {
			if photoMall.Id == photo.Sid {
				res.SImg = photoMall.SImg
				res.LImg = photoMall.LImg
				res.AndroidImg = photoMall.AndroidImg
				res.IphoneImg = photoMall.IphoneImg
				res.IpadImg = photoMall.IpadImg
				res.ThumbnailImg = photoMall.ThumbnailImg
				isFound = true
				break
			}
		}
		if !isFound {
			return nil, errors.Wrapf(ecode.TopPhotoNotFound, "%s", err)
		}
		return res, nil
	}
	topUploadPhoto, err := s.dao.MemberUploadTopphotoByID(c, photo.Sid)
	if err != nil {
		return nil, errors.Wrapf(ecode.TopPhotoNotFound, "%s", err)
	}
	if topUploadPhoto != nil {
		res.SImg = topUploadPhoto.ImgPath
		res.LImg = topUploadPhoto.ImgPath
		res.AndroidImg = topUploadPhoto.ImgPath
		res.IphoneImg = topUploadPhoto.ImgPath
		res.ThumbnailImg = topUploadPhoto.ImgPath
	}
	return res, nil
}

// MemWebTopPhotoIndex .
func (s *Service) MemWebTopPhotoIndex(c context.Context, mid, vmid int64) (*model.MemberPhotoMallIndex, error) {
	var (
		topPhoto, upload       *model.MemberPhotoMallIndex
		topPhotoErr, uploadErr error
	)
	group := errgroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		//member topphotoTmp
		topPhoto, topPhotoErr = s.MemPhotoMall(c, vmid)
		return nil
	})
	if mid == vmid {
		group.Go(func(ctx context.Context) error {
			//主人态需要额外获取头图
			info, err := s.accClient.Vip3(c, &accwar.MidReq{Mid: mid})
			if err != nil {
				log.Error("service.MemWebTopPhotoIndex vip3 mid:%d, error:%v,", mid, err)
				return nil
			}
			//验证VIP状态
			if !info.IsValid() {
				return nil
			}
			uploadTmp, err := s.dao.MemberUploadTopphoto(c, mid, model.UploadTopPhotoWeb)
			if err != nil {
				uploadErr = errors.Wrapf(ecode.TopPhotoNotFound, "%s", err)
				return nil
			}
			if uploadTmp != nil {
				upload = &model.MemberPhotoMallIndex{
					SID:          uploadTmp.ID,
					SImg:         uploadTmp.ImgPath,
					LImg:         uploadTmp.ImgPath,
					ThumbnailImg: uploadTmp.ImgPath,
					Platform:     2,
				}
				if uploadTmp.Platfrom > model.UploadTopPhotoAndroid {
					upload.Platform = 1
				}
			}
			return nil
		})
	}
	if err := group.Wait(); err != nil {
		log.Error("Service.MemWebTopPhotoIndex error:%v", err)
	}
	if topPhotoErr != nil && uploadErr != nil {
		return nil, topPhotoErr
	}
	if upload != nil {
		return upload, nil
	}
	if topPhoto != nil {
		return topPhoto, nil
	}
	return nil, errors.Wrapf(ecode.TopPhotoNotFound, "mid:%d", mid)
}

// UploadTopPhoto .
func (s *Service) WebUploadTopPhoto(c context.Context, mid int64, photo string) (upload *model.MemberPhotoUpload, err error) {
	var (
		info *accwar.VipReply
		url  string
	)
	if info, err = s.accClient.Vip3(c, &accwar.MidReq{Mid: mid}); err != nil {
		log.Error("WebUploadTopPhoto accClient.Vip3(%d) error(%v)", mid, err)
		return
	}
	//验证VIP状态
	if !info.IsValid() {
		err = ecode.SpaceVIPError
		return
	}
	upload = &model.MemberPhotoUpload{}
	if url, err = s.UploadPhoto(c, mid, photo); err != nil {
		log.Error("WebUploadTopPhoto UploadPhoto mid(%d) error(%v)", mid, err)
		return
	}
	upload.ImgPath = url
	if _, err = s.dao.AddTopphotoUpload(c, mid, model.UploadTopPhotoWeb, url); err != nil {
		log.Error("WebUploadTopPhoto AddTopphotoUpload mid(%d) error(%v)", mid, err)
		return
	}
	_ = s.cache.Do(c, func(ctx context.Context) {
		_ = s.ClearTopPhotoCache(ctx, mid, model.UploadTopPhotoWeb)
	})
	return upload, nil
}

// UploadPhoto .
func (s *Service) UploadPhoto(c context.Context, mid int64, body string) (url string, err error) {
	//判断文件大小
	if len(body) > model.MemTopPhotoMaxUpload {
		err = ecode.SpaceTPPicLarge
		return
	}
	bodyByte, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		log.Error("UploadPhoto mid(%d) DecodeString error(%v)", mid, err)
		err = ecode.SpaceBase64Error
		return
	}
	//检测文件类型
	ftype := http.DetectContentType(bodyByte)
	if ftype != "image/jpeg" && ftype != "image/jpg" && ftype != "image/png" {
		err = ecode.SpaceTPPicError
		log.Error("UploadPhoto mid(%d) DetectContentType error(%v)", mid, err)
		return
	}
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
	image.RegisterFormat("png", "png", png.Decode, png.DecodeConfig)
	buf := bytes.NewBufferString(string(bodyByte))
	i, _, err := image.DecodeConfig(buf)
	if err != nil {
		log.Error("UploadPhoto image.Decode mid(%d) body(%s) error(%v)", mid, body, err)
		return
	}
	if i.Width < 1280 || i.Height < 200 {
		err = ecode.SpaceSizError
		log.Error("UploadPhoto mid(%d) wide (%+v) error(%v)", mid, i, err)
		return
	}
	url, err = s.dao.UploadBFS(c, bodyByte, false)
	if err != nil {
		err = ecode.SpaceTPPicError
		log.Error("UploadPhoto mid(%d) UploadBfs error(%v)", mid, err)
		return
	}
	urlSplit := strings.Split(url, _split)
	if len(urlSplit) != _splitLen {
		log.Error("UploadPhoto strings.Split mid(%d) url(%s) urlSplit(%s) error(%v)", mid, url, urlSplit, err)
		return
	}
	urlTmp := _split + urlSplit[1]
	return urlTmp, nil
}

func (s *Service) GetPhotoMallList(c context.Context) ([]*model.PhotoMallIndex, error) {
	if len(s.photoMallList) == 0 {
		return nil, errors.WithStack(ecode.PhotoMallEmptyError)
	}
	// 获取type包含"pc"的头图
	photoMallList := []*model.PhotoMallIndex{}
	for _, photoMall := range s.photoMallList {
		if photoMall.Id == _banPhotoMallID {
			continue
		}
		if strings.Contains(photoMall.Type, "pc") {
			photoMallTmp := &model.PhotoMallIndex{
				ID:           photoMall.Id,
				IsDisable:    photoMall.IsDisable,
				Price:        photoMall.Price,
				CoinType:     photoMall.CoinType,
				VipFree:      photoMall.VipFree,
				SortNum:      photoMall.SortNum,
				ProductName:  photoMall.ProductName,
				SImg:         photoMall.SImg,
				LImg:         photoMall.LImg,
				ThumbnailImg: photoMall.ThumbnailImg,
				//Expire:       0,
				//Had:          0,
			}
			photoMallList = append(photoMallList, photoMallTmp)
		}
	}
	if len(photoMallList) == 0 {
		return nil, errors.WithStack(ecode.PhotoMallEmptyError)
	}
	return photoMallList, nil
}

func (s *Service) PurgeCache(c context.Context, param *model.PurgeCacheParam) error {
	if param.ModifiedAttr == "purgePhotoCache" {
		return s.ClearTopPhotoCache(c, param.Mid, model.UploadTopPhotoWeb)
	}
	if param.ModifiedAttr == "updateVip" {
		return s.updateTopPhoto(c, param)
	}
	return nil
}
func (s *Service) updateTopPhoto(c context.Context, param *model.PurgeCacheParam) error {
	if param.BuyMonth < 1 && param.Days < 1 {
		return nil
	}
	addExpire := param.BuyMonth * _addMonthExpire
	if param.Days > 0 {
		addExpire = param.Days * _addDayExpire
	}
	topPhotoList, err := s.dao.GetMemberTopPhoto(c, param.Mid)
	if err != nil {
		return err
	}
	// 获取vipfree头图
	vipFreePhotoMall := make(map[int64]*model.PhotoMall)
	for _, photoMall := range s.photoMallList {
		if photoMall.VipFree == 1 && photoMall.IsDisable == 0 {
			vipFreePhotoMall[photoMall.Id] = photoMall
		}
	}
	nowTime := time.Now().Unix()
	for _, topPhoto := range topPhotoList {
		if _, ok := vipFreePhotoMall[topPhoto.Sid]; !ok && topPhoto.Expire < nowTime && topPhoto.Platfrom != 0 {
			continue
		}
		expire := nowTime + addExpire
		if topPhoto.Expire > nowTime {
			expire = topPhoto.Expire + addExpire
		}
		if err := s.dao.UpdateMemberTopPhotoExpire(c, topPhoto.ID, param.Mid, expire); err != nil {
			return err
		}
	}
	_ = s.cache.Do(c, func(ctx context.Context) {
		_ = s.ClearTopPhotoCache(c, param.Mid, model.UploadTopPhotoWeb)
	})
	return nil
}

// 繁体版转成对应粉板
func mobiappConvert(mobiapp string) string {
	if mobiapp == "iphone_i" {
		return "iphone"
	}
	if mobiapp == "android_i" {
		return "android"
	}
	return mobiapp
}

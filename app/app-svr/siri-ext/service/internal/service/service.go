package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"go-common/component/metadata/device"
	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-feed/interface-ng/card-schema/util/sets"
	pb "go-gateway/app/app-svr/siri-ext/service/api"
	"go-gateway/app/app-svr/siri-ext/service/internal/dao"
	"go-gateway/app/app-svr/siri-ext/service/internal/model"

	accgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	uparcapi "git.bilibili.co/bapis/bapis-go/up-archive/service"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/google/wire"
	"github.com/mozillazg/go-pinyin"
	"github.com/yanyiwu/gojieba"
)

var Provider = wire.NewSet(New, wire.Bind(new(pb.SiriExtServer), new(*Service)))

const (
	Word_A_Action  = "A_Action"
	Word_App_Name  = "App_Name"
	Word_B_Action  = "B_Action"
	Word_Spec_Name = "SpecName"
)

const (
	_spaceArchiveUri      = "bilibili://music/playlist/spacepage/%d?%s"
	_spaceArchivePS       = "20"
	_spaceArchiveOffset   = "0"
	_spaceArchiveOrder    = "time"
	_spaceArchiveDesc     = "1"
	_spaceArchivePageType = "1"
	_spaceArchiveOid      = "0"
)

const (
	_specNameWatchLater = "稍后再看"
	_specNameFav        = "收藏夹"
)

var (
	_watchLaterAlias = sets.NewString("稍后再")
)

var _builtinWords = map[string]sets.String{
	Word_A_Action:  sets.NewString("打开", "启动", "使用", "用", "开启", "开"),
	Word_App_Name:  sets.NewString("bilibili", "哔哩", "哔哩哔哩", "B站", "b站", "bilibili App", "哔哩哔哩App", "哔哩哔哩应用", "哔哩哔哩软件", "小破站"),
	Word_B_Action:  sets.NewString("播放", "开播", "播", "放映", "看"),
	Word_Spec_Name: sets.NewString(_specNameWatchLater, _specNameFav),
}

func init() {
	_builtinWords[Word_Spec_Name].Insert(_watchLaterAlias.List()...)
}

// Service service.
type Service struct {
	ac    *paladin.Map
	dao   dao.Dao
	jieba *gojieba.Jieba
}

func jiebaRuntimeDict() ([]string, bool) {
	dictDir := os.Getenv("JIEBA_DICT_DIR")
	if dictDir == "" {
		return nil, false
	}

	dictPaths := [5]string{
		path.Join(dictDir, "jieba.dict.utf8"),
		path.Join(dictDir, "hmm_model.utf8"),
		path.Join(dictDir, "user.dict.utf8"),
		path.Join(dictDir, "idf.utf8"),
		path.Join(dictDir, "stop_words.utf8"),
	}
	return dictPaths[:], true
}

// New new a service and return.
func New(d dao.Dao) (s *Service, cf func(), err error) {
	s = &Service{
		ac:  &paladin.TOML{},
		dao: d,
	}
	jiebaDicts, ok := jiebaRuntimeDict()
	if ok {
		log.Warn("Customized jieba dict: %+v", jiebaDicts)
	}
	s.jieba = gojieba.NewJieba(jiebaDicts...)
	for _, kSet := range _builtinWords {
		for _, w := range kSet.List() {
			s.jieba.AddWord(w)
		}
	}
	cf = s.Close
	err = paladin.Watch("application.toml", s.ac)
	return
}

// Ping ping the resource.
func (s *Service) Ping(ctx context.Context, e *empty.Empty) (*empty.Empty, error) {
	return &empty.Empty{}, s.dao.Ping(ctx)
}

// Close close the resource.
func (s *Service) Close() {
}

//nolint:unused
type resolvedParts struct {
	AAction string
	AppName string
	BAction string

	SpecName string

	Remain       string
	RemainPinyin string
}

//nolint:unused
func (rp resolvedParts) fulfill() bool {
	return rp.AAction != "" && rp.AppName != "" && rp.BAction != ""
}

//nolint:deadcode,unused
func resolveParts(parts []string) (*resolvedParts, bool) {
	resolved := &resolvedParts{}

	lastFindAt := 0
	for at, i := range parts {
		for p, kSet := range _builtinWords {
			if !kSet.Has(i) {
				continue
			}
			switch p {
			case Word_A_Action:
				resolved.AAction = i
			case Word_App_Name:
				resolved.AppName = i
			case Word_B_Action:
				resolved.BAction = i
			case Word_Spec_Name:
				resolved.SpecName = i
			default:
				continue
			}
			lastFindAt = at
		}
	}
	if resolved.fulfill() {
		if len(parts) <= (lastFindAt + 1) {
			return resolved, true
		}
		resolved.Remain = strings.Join(parts[lastFindAt+1:], "")
		resolved.RemainPinyin = castAsPinyin(resolved.Remain)
		return resolved, true
	}
	return nil, false
}

func castAsPinyin(in string) string {
	a := pinyin.NewArgs()
	a.Heteronym = true

	py := pinyin.Pinyin(in, a)
	parts := make([]string, 0, len(py))
	for _, i := range py {
		parts = append(parts, strings.Join(i, ""))
	}
	return strings.Join(parts, "")
}

func firstAccount(in *model.Suggest3) (*model.Sug, bool) {
	if len(in.Result) <= 0 {
		return nil, false
	}
	for _, i := range in.Result {
		if i.TermType == model.SuggestionAccount {
			return i, true
		}
	}
	return nil, false
}

type spaceEpisodicCtx struct {
	accInfo   *accgrpc.Info
	arcPassed *uparcapi.ArcPassedReply
}

//nolint:unparam
func (s *Service) makeSpaceEpisodicURL(seCtx *spaceEpisodicCtx) (string, error) {
	params := url.Values{}
	params.Set("offset", _spaceArchiveOffset)
	params.Set("desc", _spaceArchiveDesc)
	params.Set("oid", _spaceArchiveOid)
	params.Set("ps", _spaceArchivePS)
	params.Set("order", _spaceArchiveOrder)
	params.Set("page_type", _spaceArchivePageType)
	params.Set("user_name", seCtx.accInfo.Name)
	params.Set("playlist_intro", "UP主的全部视频")
	params.Set("total_count", strconv.FormatInt(seCtx.arcPassed.Total, 10))
	params.Set("sort_field", "1")
	return fmt.Sprintf(_spaceArchiveUri, seCtx.accInfo.Mid, params.Encode()), nil
}

func (s *Service) commandBySpecName(ctx context.Context, req *pb.ResolveCommandReq) (*pb.ResolveCommandReply, error) {
	switch {
	case req.Command == _specNameWatchLater || _watchLaterAlias.Has(req.Command):
		if req.Mid <= 0 {
			return nil, ecode.Error(ecode.RequestErr, "please login first")
		}
		if s.dao.UserToViewsIsEmpty(ctx, req.Mid) {
			return &pb.ResolveCommandReply{
				RedirectUrl: "bilibili://user_center/watchlater",
			}, nil
		}
		return &pb.ResolveCommandReply{
			RedirectUrl: fmt.Sprintf("bilibili://music/playlist/playpage/%d?oid=0&page_type=2&play_page=0&pprogress=-1", req.Mid),
		}, nil
	case req.Command == _specNameFav:
		if req.Mid <= 0 {
			return nil, ecode.Error(ecode.RequestErr, "please login first")
		}
		defFolder, err := s.dao.UserDefaultFavFolder(ctx, req.Mid)
		if err != nil {
			log.Error("Failed to get user default fav folder: %d: %+v", req.Mid, err)
			return &pb.ResolveCommandReply{RedirectUrl: "bilibili://main/favorite"}, nil
		}
		if defFolder == nil {
			log.Warn("The default fav folder is not exist on mid: %d", req.Mid)
			return &pb.ResolveCommandReply{RedirectUrl: "bilibili://main/favorite"}, nil
		}
		// 默认收藏夹
		return &pb.ResolveCommandReply{
			RedirectUrl: fmt.Sprintf("bilibili://music/playlist/playpage/%d?from_spmid=playlist.playlist-detail.0.0&from=playlist-detail&page_type=3", defFolder.Mlid),
		}, nil
	default:
		return nil, ecode.Errorf(ecode.RequestErr, "unable to handle spec name: %q", req.Command)
	}
}

//nolint:unused
func (s *Service) refineDevice(ctx context.Context, dst *pb.DeviceMeta) {
	device, ok := device.FromContext(ctx)
	if !ok {
		return
	}
	dst.Build = device.Build
	dst.MobiApp = device.RawMobiApp
	dst.Device = device.Device
	dst.Channel = device.Channel
	dst.Buvid = device.Buvid
	dst.Platform = device.RawPlatform
	//nolint:gosimple
	return
}

// ResolveCommand is
// func (s *Service) ResolveCommand(ctx context.Context, req *pb.ResolveCommandReq) (*pb.ResolveCommandReply, error) {
// 	parts := s.jieba.Cut(req.Command, true)
// 	resolved, ok := resolveParts(parts)
// 	if !ok {
// 		return nil, ecode.Errorf(ecode.RequestErr, "unable to resolve command parts: %+v", parts)
// 	}
// 	if resolved.SpecName != "" {
// 		return s.commandBySpecName(ctx, req, resolved)
// 	}
// 	s.refineDevice(ctx, &req.Device)
// 	suggest, err := s.dao.Suggest3(ctx, &model.SearchSuggestReq{
// 		Mid:       req.Mid,
// 		Platform:  req.Device.Platform,
// 		Buvid:     req.Device.Buvid,
// 		Term:      resolved.RemainPinyin,
// 		Device:    req.Device.Device,
// 		Build:     req.Device.Build,
// 		Highlight: 0,
// 		MobiApp:   req.Device.MobiApp,
// 		Now:       time.Now(),
// 	})
// 	if err != nil {
// 		log.Error("Failed to get search suggest: %+v", err)
// 		return nil, ecode.Error(ecode.RequestErr, "unable to get suggest")
// 	}
// 	accSug, ok := firstAccount(suggest)
// 	if !ok {
// 		return nil, ecode.Errorf(ecode.RequestErr, "unable to get suggest account from command: %q", req.Command)
// 	}

// 	seCtx := &spaceEpisodicCtx{}
// 	eg := errgroup.WithContext(ctx)
// 	eg.Go(func(ctx context.Context) error {
// 		accInfo, err := s.dao.AccountInfo3(ctx, accSug.Ref)
// 		if err != nil {
// 			return err
// 		}
// 		seCtx.accInfo = accInfo
// 		return nil
// 	})
// 	eg.Go(func(ctx context.Context) error {
// 		arcPassed, err := s.dao.ArcPassed(ctx, accSug.Ref)
// 		if err != nil {
// 			return err
// 		}
// 		seCtx.arcPassed = arcPassed
// 		return nil
// 	})
// 	if err := eg.Wait(); err != nil {
// 		return nil, err
// 	}

// 	redirectURL, err := s.makeSpaceEpisodicURL(seCtx)
// 	if err != nil {
// 		return nil, err
// 	}
// 	reply := &pb.ResolveCommandReply{
// 		RedirectUrl: redirectURL,
// 	}
// 	if req.Debug {
// 		reply.Debug = jsonify(map[string]interface{}{
// 			"req":      req,
// 			"cut":      parts,
// 			"resolved": resolved,
// 			"suggest":  suggest,
// 		})
// 	}
// 	return reply, nil
// }

func jsonify(in interface{}) string {
	bs, _ := json.Marshal(in)
	return string(bs)
}

func markFromSiri(in string) string {
	u, err := url.Parse(in)
	if err != nil {
		return in
	}
	q := u.Query()
	q.Set("siri", "1")
	u.RawQuery = q.Encode()
	return u.String()
}

func (s *Service) ResolveCommand(ctx context.Context, req *pb.ResolveCommandReq) (*pb.ResolveCommandReply, error) {
	req.Command = strings.TrimSpace(req.Command)
	if req.Command == "" {
		return nil, ecode.Errorf(ecode.RequestErr, "empty command")
	}
	if _builtinWords[Word_Spec_Name].Has(req.Command) {
		reply, err := s.commandBySpecName(ctx, req)
		if err != nil {
			return nil, err
		}
		reply.RedirectUrl = markFromSiri(reply.RedirectUrl)
		return reply, nil
	}

	suggestedAccount := struct {
		byPinyin  *model.Sug
		byCommand *model.Sug
	}{}
	eg := errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		cmdPinyin := castAsPinyin(req.Command)
		suggest, err := s.dao.Suggest3(ctx, &model.SearchSuggestReq{
			Mid:       req.Mid,
			Platform:  req.Device.Platform,
			Buvid:     req.Device.Buvid,
			Term:      cmdPinyin,
			Device:    req.Device.Device,
			Build:     req.Device.Build,
			Highlight: 0,
			MobiApp:   req.Device.MobiApp,
			Now:       time.Now(),
		})
		if err != nil {
			log.Error("Failed to get search suggest: %+v", err)
			return nil
		}
		accSug, ok := firstAccount(suggest)
		if !ok {
			return nil
		}
		suggestedAccount.byPinyin = accSug
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		suggest, err := s.dao.Suggest3(ctx, &model.SearchSuggestReq{
			Mid:       req.Mid,
			Platform:  req.Device.Platform,
			Buvid:     req.Device.Buvid,
			Term:      req.Command,
			Device:    req.Device.Device,
			Build:     req.Device.Build,
			Highlight: 0,
			MobiApp:   req.Device.MobiApp,
			Now:       time.Now(),
		})
		if err != nil {
			log.Error("Failed to get search suggest: %+v", err)
			return nil
		}
		accSug, ok := firstAccount(suggest)
		if !ok {
			return nil
		}
		suggestedAccount.byCommand = accSug
		return nil
	})
	if err := eg.Wait(); err != nil {
		log.Error("Failed to execute search request: %+v", err)
		return nil, err
	}
	if suggestedAccount.byCommand == nil && suggestedAccount.byPinyin == nil {
		return nil, ecode.Errorf(ecode.RequestErr, "unable to get suggest account from command: %q", req.Command)
	}

	accSug := suggestedAccount.byPinyin
	if accSug == nil {
		accSug = suggestedAccount.byCommand
	}
	seCtx := &spaceEpisodicCtx{}
	eg = errgroup.WithContext(ctx)
	eg.Go(func(ctx context.Context) error {
		accInfo, err := s.dao.AccountInfo3(ctx, accSug.Ref)
		if err != nil {
			return err
		}
		seCtx.accInfo = accInfo
		return nil
	})
	eg.Go(func(ctx context.Context) error {
		arcPassed, err := s.dao.ArcPassed(ctx, accSug.Ref)
		if err != nil {
			return err
		}
		seCtx.arcPassed = arcPassed
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	redirectURL, err := s.makeSpaceEpisodicURL(seCtx)
	if err != nil {
		return nil, err
	}
	reply := &pb.ResolveCommandReply{
		RedirectUrl: markFromSiri(redirectURL),
	}
	if req.Debug {
		reply.Debug = jsonify(map[string]interface{}{})
	}
	return reply, nil
}

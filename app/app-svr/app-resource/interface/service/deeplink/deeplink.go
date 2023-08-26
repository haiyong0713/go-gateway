package deeplink

import (
	"context"
	"encoding/base64"
	"net/url"
	"strings"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/dao/deeplink"
	model "go-gateway/app/app-svr/app-resource/interface/model/deeplink"
	gateecode "go-gateway/ecode"

	"github.com/pkg/errors"
)

type Service struct {
	c   *conf.Config
	dao *deeplink.Dao
}

const (
	_ChannelVIVO     = "vivo"
	_ChannelOPPO     = "oppo"
	_ChannelHWPPS    = "hwpps"
	_ChannelTengxun  = "tengxun"
	_ChannelBaidu    = "baidu"
	_ChannelKuaishou = "ksjsb"
)

var (
	BtnResolver = btnResolver{}
)

func New(conf *conf.Config) *Service {
	return &Service{
		c:   conf,
		dao: deeplink.New(conf),
	}
}

func (s *Service) DeepLinkHW(ctx context.Context, req *model.HWDeeplinkReq) (*model.HWDeeplinkMeta, error) {
	res, err := s.dao.DeepLinkHW(ctx, req)
	if err != nil {
		log.Error("s.dao.DeepLinkHW req(%+v), err(%+v)", req, err)
		return &model.HWDeeplinkMeta{}, nil
	}
	return &model.HWDeeplinkMeta{Deeplink: res}, nil
}

func (s *Service) DeepLinkAI(ctx context.Context, req *model.AiDeeplinkReq, buvid string) (*model.AiDeeplink, error) {
	metas, err := resolveInnerMetaInOriginLink(req.OriginLink)
	if err != nil {
		log.Error("resolveInnerMetaInOriginLink originLink=%s, error=%+v", req.OriginLink, err)
		return nil, nil
	}
	res, err := s.dao.DeepLinkAI(ctx, buvid, metas)
	if err != nil {
		if errors.Cause(err) == context.DeadlineExceeded || errors.Cause(err) == ecode.Deadline {
			// 超时需要特殊错误码提供客户端上报处理
			log.Error("s.dao.DeepLinkAI context deadline exceeded buvid=%s, req=%+v", buvid, req)
			return nil, gateecode.DeeplinkTimeoutErr
		}
		if errors.Cause(err) == deeplink.ErrTaishanResultNil {
			// 此处大概率不会有对应link，err不打日志
			return nil, nil
		}
		log.Error("s.dao.DeepLinkAI buvid=%s, req=%+v, error=%+v", buvid, req, err)
		return nil, err
	}
	return res, nil
}

func resolveInnerMetaInOriginLink(link string) (*model.AiDeeplinkMaterial, error) {
	// 解析原始link
	u, err := url.Parse(link)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 解码h5awaken
	h5awaken := strings.ReplaceAll(u.Query().Get("h5awaken"), " ", "+")
	bs, err := base64.StdEncoding.DecodeString(h5awaken)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 获取open_app_url和open_app_from_type
	subParams, err := url.ParseQuery(string(bs))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	// 解析open_app_from_type里的数据
	sourceName, accountId := parseSourceNameAndAccountId(subParams.Get("open_app_from_type"))
	innerId, innerType := parseInnerIdAndInnerType(link)
	return &model.AiDeeplinkMaterial{
		InnerId:    innerId,
		InnerType:  innerType,
		SourceName: sourceName,
		AccountId:  accountId,
	}, nil
}

func parseInnerIdAndInnerType(rawLink string) (string, int64) {
	const (
		_innerIdBvType = 1
		_prefixBvType  = "bilibili://video/"

		_innerIdSeasonType = 2
		_prefixSeasonType  = "bilibili://bangumi/season/"

		_innerIdLiveType = 3
		_prefixLiveType  = "bilibili://live/"

		_innerIdChannelType = 4
		_prefixChannelType  = "bilibili://pegasus/channel/v2/"
	)
	link := rawLink[:strings.Index(rawLink, "?")]
	if strings.HasPrefix(link, _prefixBvType) {
		return link[len(_prefixBvType):], _innerIdBvType
	}
	if strings.HasPrefix(link, _prefixSeasonType) {
		return link[len(_prefixSeasonType):], _innerIdSeasonType
	}
	if strings.HasPrefix(link, _prefixLiveType) {
		return link[len(_prefixLiveType):], _innerIdLiveType
	}
	if strings.HasPrefix(link, _prefixChannelType) {
		return link[len(_prefixChannelType):], _innerIdChannelType
	}
	log.Warn("parseInnerIdAndInnerType() unrecognized link=%s", link)
	return "", -1
}

func parseSourceNameAndAccountId(origin string) (string, string) {
	const (
		_minOriginParsedSplitNum = 2
	)
	arr := strings.Split(origin, "-")
	if len(arr) < _minOriginParsedSplitNum {
		return origin, origin
	}
	return arr[0], arr[len(arr)-1]
}

func (s *Service) BackButton(ctx context.Context, req *model.BackButtonReq) (*model.BackButtonMeta, error) {
	schema, err := url.Parse(req.Schema)
	if err != nil {
		return nil, ecode.Errorf(ecode.RequestErr, "不合法的schema:%q", req.Schema)
	}
	btnChannel := schema.Query().Get("btn_channel")
	switch btnChannel {
	case _ChannelVIVO:
		return BtnResolver.VIVO(btnChannel, req)
	case _ChannelOPPO:
		return BtnResolver.OPPO(btnChannel, req)
	case _ChannelHWPPS:
		return BtnResolver.HUAWEI(btnChannel, req)
	case _ChannelTengxun:
		return BtnResolver.TENGXUN(btnChannel, req)
	case _ChannelBaidu:
		return BtnResolver.BAIDU(btnChannel, req)
	case _ChannelKuaishou:
		return BtnResolver.KUAISHOU(btnChannel, req)
	default:
		return nil, ecode.Errorf(ecode.RequestErr, "未识别的渠道schema:%q", req.Schema)
	}
}

type BackButtonResolver func(string, *model.BackButtonReq) (*model.BackButtonMeta, error)

type BtnResolverConfig struct {
	BackName   string
	Color      string
	NoClose    bool
	Passed     bool
	DirectBack bool
	BtnSize    int64
}

var (
	_defConfig = map[string]BtnResolverConfig{
		_ChannelVIVO: {
			BackName:   "返回vivo",
			Color:      "#80000000",
			NoClose:    false,
			Passed:     false,
			DirectBack: false,
			BtnSize:    1,
		},
		_ChannelOPPO: {
			BackName:   "返回oppo",
			Color:      "#80000000",
			NoClose:    true,
			Passed:     false,
			DirectBack: false,
			BtnSize:    1,
		},
		_ChannelHWPPS: {
			BackName:   "返回华为",
			Color:      "#80000000",
			NoClose:    false,
			Passed:     false,
			DirectBack: false,
			BtnSize:    1,
		},
		_ChannelTengxun: {
			BackName:   "返回腾讯",
			Color:      "#80000000",
			NoClose:    false,
			Passed:     true,
			DirectBack: false,
			BtnSize:    1,
		},
		_ChannelBaidu: {
			BackName:   "百度",
			Color:      "#80000000",
			NoClose:    true,
			Passed:     true,
			DirectBack: false,
			BtnSize:    1,
		},
		_ChannelKuaishou: {
			BackName:   "返回快手",
			Color:      "#CCFF6699",
			NoClose:    false,
			Passed:     false,
			DirectBack: false,
			BtnSize:    1,
		},
	}
	_defBtn = BtnResolverConfig{
		BackName:   "返回原APP",
		Color:      "#80000000",
		NoClose:    false,
		Passed:     false,
		DirectBack: false,
	}
)

func getButtonConfig(channel string) *BtnResolverConfig {
	config, ok := _defConfig[channel]
	if !ok {
		config = _defBtn
	}
	dup := config
	return &dup
}

type btnResolver struct{}

func assertChannel(in, right string) error {
	if in != right {
		return errors.Errorf("channel mismatched: %q, %q", in, right)
	}
	return nil
}

type byCommonParam struct {
	BtnChannel      string
	ActuallyChannel string
	BackURLField    string
}

func (br btnResolver) byCommon(param *byCommonParam, req *model.BackButtonReq) (*model.BackButtonMeta, error) {
	if err := assertChannel(param.BtnChannel, param.ActuallyChannel); err != nil {
		return nil, err
	}
	config := getButtonConfig(param.BtnChannel)

	u, err := url.Parse(req.Schema)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	query := u.Query()
	backURL := ""
	if param.BackURLField != "" {
		backURL = query.Get(param.BackURLField)
	}
	btnMeta := &model.BackButtonMeta{
		BtnChannel: param.BtnChannel,
		BackURL:    backURL,
	}
	btnMeta.BackName = config.BackName
	btnMeta.Color = config.Color
	btnMeta.NoClose = config.NoClose
	btnMeta.Passed = config.Passed
	btnMeta.DirectBack = config.DirectBack
	btnMeta.BtnSize = config.BtnSize
	return btnMeta, nil
}

func (br btnResolver) HUAWEI(btnChannel string, req *model.BackButtonReq) (*model.BackButtonMeta, error) {
	return br.byCommon(&byCommonParam{btnChannel, _ChannelHWPPS, "backurl"}, req)
}
func (br btnResolver) VIVO(btnChannel string, req *model.BackButtonReq) (*model.BackButtonMeta, error) {
	return br.byCommon(&byCommonParam{btnChannel, _ChannelVIVO, "backurl"}, req)
}
func (br btnResolver) OPPO(btnChannel string, req *model.BackButtonReq) (*model.BackButtonMeta, error) {
	btnMeta, err := br.byCommon(&byCommonParam{btnChannel, _ChannelOPPO, "back_url"}, req)
	if err != nil {
		return nil, err
	}
	// set btn_name
	func() {
		u, err := url.Parse(req.Schema)
		if err != nil {
			return
		}
		query := u.Query()
		backName := query.Get("btn_name")
		if backName != "" {
			btnMeta.BackName = backName
		}
	}()
	return btnMeta, nil
}
func (br btnResolver) BAIDU(btnChannel string, req *model.BackButtonReq) (*model.BackButtonMeta, error) {
	btnMeta, err := br.byCommon(&byCommonParam{btnChannel, _ChannelBaidu, ""}, req)
	if err != nil {
		return nil, err
	}
	btnMeta.BackURL = "baiduboxapp://donothing"
	return btnMeta, nil
}
func (br btnResolver) TENGXUN(btnChanel string, req *model.BackButtonReq) (*model.BackButtonMeta, error) {
	if err := assertChannel(btnChanel, _ChannelTengxun); err != nil {
		return nil, err
	}
	config := getButtonConfig(btnChanel)

	u, err := url.Parse(req.Schema)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	query := u.Query()
	backURL := query.Get("back_url")
	backName := query.Get("back_name")
	if backName == "" {
		backName = config.BackName
	}

	btnMeta := &model.BackButtonMeta{
		BtnChannel: btnChanel,
		BackURL:    backURL,
	}
	btnMeta.BackName = backName
	btnMeta.Color = config.Color
	btnMeta.NoClose = config.NoClose
	btnMeta.Passed = config.Passed
	btnMeta.DirectBack = config.DirectBack
	btnMeta.BtnSize = config.BtnSize
	return btnMeta, nil
}

func (br btnResolver) KUAISHOU(btnChannel string, req *model.BackButtonReq) (*model.BackButtonMeta, error) {
	return br.byCommon(&byCommonParam{btnChannel, _ChannelKuaishou, "backurl"}, req)
}

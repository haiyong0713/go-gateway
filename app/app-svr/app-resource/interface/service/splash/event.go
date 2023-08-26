package splash

import (
	"bytes"
	"context"
	"math"
	"sort"
	"text/template"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/sync/errgroup.v2"
	"go-gateway/app/app-svr/app-resource/interface/conf"
	"go-gateway/app/app-svr/app-resource/interface/model/splash"

	accountgrpc "git.bilibili.co/bapis/bapis-go/account/service"
	garbgrpc "git.bilibili.co/bapis/bapis-go/garb/service"
	"github.com/pkg/errors"
)

const (
	_14days = time.Hour * 24 * 14
	// _greetingText = "<font color=\"#000000\" >注册<strong><font color=\"#ff5377\" >{{.passYears}}周年</font></strong>快乐！</font>"
	// _textText     = "<font color=\"#000000\" ><font color=\"#ff5377\" >{{.joinYear}}年{{.joinMonth}}月{{.joinDay}}日</font>，我们初次遇见<font color=\"#ff5377\" >{{.passDays}}</font>个日夜，相伴走过了许多风景下一个冬夏，也请多指教呀电波相连，我们的故事永不完结~</font>"
)

func (s *Service) EventSplashList(ctx context.Context, req *splash.EventSplashRequest) (*splash.EventSplashListReply, error) {
	reply := &splash.EventSplashListReply{
		EventSplash: []splash.EventSplash{},
	}
	if req.Mid <= 0 {
		return reply, nil
	}
	func() {
		event, err := s.registrationDateEvent(ctx, req)
		if err != nil {
			log.Error("Failed to get registration date event: %+v: %+v", req, err)
			return
		}
		if event == nil {
			return
		}
		reply.EventSplash = append(reply.EventSplash, event)
	}()
	return reply, nil
}

func isLeapYear(year int) bool {
	return year%400 == 0 ||
		year%100 != 0 &&
			year%4 == 0
}

func timeToDate(in time.Time) time.Time {
	year, month, day := in.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, in.Location())
}

func allDayTimeRange(year int, month time.Month, day int, loc *time.Location) (time.Time, time.Time) {
	start := time.Date(year, month, day, 0, 0, 0, 0, loc)
	end := start.AddDate(0, 0, 1)
	return start, end
}

func accountCardFromProfile(in *accountgrpc.Profile) *splash.AccountCard {
	accCard := &splash.AccountCard{
		Mid:     in.Mid,
		Uname:   in.Name,
		Face:    in.Face,
		Sign:    in.Sign,
		Level:   int64(in.Level),
		Vip:     in.Vip,
		Pendant: in.Pendant,
	}
	accCard.OfficialVerify.Desc = in.Official.Title
	accCard.OfficialVerify.Type = in.Official.Type
	return accCard
}

func renderTemplate(input string, data interface{}) string {
	tmpl, err := template.New("").Parse(input)
	if err != nil {
		return input
	}
	out := &bytes.Buffer{}
	if err := tmpl.Execute(out, data); err != nil {
		log.Error("Failed to render text: %q: %+v: %+v", input, data, err)
		return input
	}
	return out.String()
}

func monthDaySub(left, right time.Time) time.Duration {
	lDate := time.Date(left.Year(), left.Month(), left.Day(), 0, 0, 0, 0, left.Location())
	rDate := time.Date(left.Year(), right.Month(), right.Day(), 0, 0, 0, 0, left.Location())
	return lDate.Sub(rDate)
}

func (s *Service) bestSizeImageConfig(req *splash.EventSplashRequest) (*conf.RegistrationDateEventConfigItem, string) {
	const (
		_normalSize = float64(1600) / float64(2560)
		_fullSize   = float64(1280) / float64(2560)
		_padSize    = float64(2732) / float64(2048)
		_se640Size  = float64(640) / float64(1136)
		_p1080Size  = float64(1080) / float64(1920)
	)
	type splashScreen struct {
		Delta  float64 `json:"delta"`
		Screen string  `json:"screen"`
	}
	sizeConfig := map[string]*conf.RegistrationDateEventConfigItem{
		"normal": &s.c.RegistrationDateEventConfig.Normal,
		"full":   &s.c.RegistrationDateEventConfig.Full,
		"pad":    &s.c.RegistrationDateEventConfig.Pad,
		"se640":  &s.c.RegistrationDateEventConfig.SE640,
		"p1080":  &s.c.RegistrationDateEventConfig.P1080,
	}
	if req.ScreenWidth == 0 || req.ScreenHeight == 0 {
		return &s.c.RegistrationDateEventConfig.Normal, "normal"
	}
	size := float64(req.ScreenWidth) / float64(req.ScreenHeight)
	rankScreen := []splashScreen{
		{Delta: math.Abs(_normalSize - size), Screen: "normal"},
		{Delta: math.Abs(_fullSize - size), Screen: "full"},
		{Delta: math.Abs(_padSize - size), Screen: "pad"},
		{Delta: math.Abs(_se640Size - size), Screen: "se640"},
		{Delta: math.Abs(_p1080Size - size), Screen: "p1080"},
	}
	sort.Slice(rankScreen, func(i, j int) bool {
		return rankScreen[i].Delta < rankScreen[j].Delta
	})
	return sizeConfig[rankScreen[0].Screen], rankScreen[0].Screen
}

func (s *Service) registrationDateEvent(ctx context.Context, req *splash.EventSplashRequest) (splash.EventSplash, error) {
	if s.c.RegistrationDateEventConfig.Disable {
		return nil, errors.New("registration splash is disabled")
	}
	if req.Mid <= 0 {
		return nil, ecode.Errorf(ecode.RequestErr, "invalid mid: %+v", req)
	}
	profile, err := s.dao.AccountProfile(ctx, req.Mid)
	if err != nil {
		return nil, err
	}
	today := timeToDate(time.Now())
	joinTime := timeToDate(time.Unix(int64(profile.JoinTime), 0))
	joinYear, joinMonth, joinDay := joinTime.Date()
	passYears := today.Year() - joinYear
	if passYears <= 0 {
		return nil, err
	}
	//nolint:gomnd
	passDays := int(today.Sub(joinTime) / (24 * time.Hour))
	// 非闰年 2.29 变换为 3.1
	if joinMonth == time.February && joinDay == 29 {
		if !isLeapYear(today.Year()) {
			joinTime = joinTime.AddDate(0, 0, 1)
		}
	}
	if monthDaySub(today, joinTime) > _14days {
		return nil, nil
	}
	event := splash.NewRegistrationDateEvent()
	beginTime, endTime := allDayTimeRange(today.Year(), joinTime.Month(), joinTime.Day(), today.Location())
	event.BeginTime = beginTime.Unix()
	event.EndTime = endTime.Unix()
	// config related
	event.ShowTimes = s.c.RegistrationDateEventConfig.ShowTimes
	event.Duration = s.c.RegistrationDateEventConfig.Duration
	event.Logo = s.c.RegistrationDateEventConfig.LogoURL
	event.AccountCard = *accountCardFromProfile(profile)
	event.URI = s.c.RegistrationDateEventConfig.URI
	event.SkipButton = s.c.RegistrationDateEventConfig.SkipButton
	event.Param = s.c.RegistrationDateEventConfig.Param

	bestConfig, screen := s.bestSizeImageConfig(req)
	event.Screen = screen
	event.ResourceType = bestConfig.ResourceType
	event.Image = bestConfig.ImageURL
	event.VideoURI = bestConfig.VideoURI
	event.VideoHash = bestConfig.VideoHash
	if bestConfig.AccountCard.Enable {
		event.Element = append(event.Element, &splash.Element{
			Type:              "account_card",
			MaxWidthPX:        bestConfig.AccountCard.MaxWidthPX,
			PaddingTopPercent: bestConfig.AccountCard.PaddingTopPercent,
		})
	}
	if bestConfig.Greeting.Enable {
		event.Element = append(event.Element, &splash.Element{
			Type:              "greetings",
			MaxWidthPX:        bestConfig.Greeting.MaxWidthPX,
			PaddingTopPercent: bestConfig.Greeting.PaddingTopPercent,
			FontSize:          bestConfig.Greeting.Fontsize,
			Text: renderTemplate(bestConfig.Greeting.Text, map[string]interface{}{
				"passYears": passYears,
			}),
		})
	}
	if bestConfig.Text.Enable {
		event.Element = append(event.Element, &splash.Element{
			Type:              "text",
			MaxWidthPX:        bestConfig.Text.MaxWidthPX,
			PaddingTopPercent: bestConfig.Text.PaddingTopPercent,
			FontSize:          bestConfig.Text.Fontsize,
			Text: renderTemplate(bestConfig.Text.Text, map[string]interface{}{
				"joinYear":  joinYear,
				"joinMonth": int(joinMonth),
				"joinDay":   joinDay,
				"passDays":  passDays,
			}),
		})
	}
	return event, nil
}

func (s *Service) EventSplashList2(ctx context.Context, req *splash.EventSplashRequest) (*splash.EventSplashList2Reply, error) {
	reply := &splash.EventSplashList2Reply{
		EventList: []splash.EventSplashV2{},
	}
	ip := metadata.String(ctx, metadata.RemoteIP)
	eg := errgroup.WithContext(ctx)
	var esl *garbgrpc.EventSplashListReply
	eg.Go(func(ctx context.Context) (err error) {
		if esl, err = s.dao.EventSplash(ctx, req.Mid, ip, req.MobiApp); err != nil {
			return err
		}
		return nil
	})
	var accProfile *accountgrpc.Profile
	if req.Mid > 0 {
		eg.Go(func(ctx context.Context) (err error) {
			if accProfile, err = s.dao.AccountProfile(ctx, req.Mid); err != nil {
				log.Error("Failed to request AccountProfile: %d, %+v", req.Mid, err)
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Error("%+v", err)
		return reply, nil
	}
	if len(esl.GetList()) == 0 {
		return reply, nil
	}
	for _, v := range esl.GetList() {
		eventSplash := splash.EventSplashV2{
			ID:            v.Id,
			EventType:     v.EventType,
			BeginTime:     v.BeginTime,
			EndTime:       v.EndTime,
			Resources:     v.Resources,
			Elements:      v.Elements,
			ShowTimes:     v.ShowTimes,
			ShowSkip:      v.ShowSkip,
			Duration:      v.Duration,
			ShowCountDown: v.ShowCountdown,
			WifiDownload:  v.WifiDownload,
		}
		reply.EventList = append(reply.EventList, eventSplash)
	}
	if accProfile != nil {
		reply.Account = splash.Account{
			Mid:      accProfile.GetMid(),
			Uname:    accProfile.GetName(),
			Level:    accProfile.GetLevel(),
			Uimage:   accProfile.GetFace(),
			Birthday: int64(accProfile.GetBirthday()),
			JoinTime: int64(accProfile.GetJoinTime()),
		}
	}
	return reply, nil
}

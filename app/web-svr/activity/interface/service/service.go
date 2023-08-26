package service

import (
	"go-gateway/app/web-svr/activity/interface/conf"
	"go-gateway/app/web-svr/activity/interface/service/account"
	"go-gateway/app/web-svr/activity/interface/service/acg"
	"go-gateway/app/web-svr/activity/interface/service/act"
	"go-gateway/app/web-svr/activity/interface/service/appstore"
	"go-gateway/app/web-svr/activity/interface/service/archive"
	"go-gateway/app/web-svr/activity/interface/service/bind"
	"go-gateway/app/web-svr/activity/interface/service/bml"
	"go-gateway/app/web-svr/activity/interface/service/bnj"
	"go-gateway/app/web-svr/activity/interface/service/brand"
	"go-gateway/app/web-svr/activity/interface/service/bws"
	"go-gateway/app/web-svr/activity/interface/service/bwsonline"
	"go-gateway/app/web-svr/activity/interface/service/cards"
	"go-gateway/app/web-svr/activity/interface/service/college"
	"go-gateway/app/web-svr/activity/interface/service/cpc100"
	"go-gateway/app/web-svr/activity/interface/service/dubbing"
	"go-gateway/app/web-svr/activity/interface/service/esports"
	"go-gateway/app/web-svr/activity/interface/service/examination"
	"go-gateway/app/web-svr/activity/interface/service/fit"
	"go-gateway/app/web-svr/activity/interface/service/funny"
	gameholiday "go-gateway/app/web-svr/activity/interface/service/game_holiday"
	"go-gateway/app/web-svr/activity/interface/service/handwrite"
	"go-gateway/app/web-svr/activity/interface/service/invite"
	"go-gateway/app/web-svr/activity/interface/service/jsondata"
	"go-gateway/app/web-svr/activity/interface/service/kfc"
	"go-gateway/app/web-svr/activity/interface/service/knowledge"
	"go-gateway/app/web-svr/activity/interface/service/like"
	"go-gateway/app/web-svr/activity/interface/service/lol"
	lotteryv2 "go-gateway/app/web-svr/activity/interface/service/lottery"
	"go-gateway/app/web-svr/activity/interface/service/mission"
	newcards "go-gateway/app/web-svr/activity/interface/service/new_cards"
	"go-gateway/app/web-svr/activity/interface/service/newstar"
	"go-gateway/app/web-svr/activity/interface/service/newyear2021"
	"go-gateway/app/web-svr/activity/interface/service/olympic"
	"go-gateway/app/web-svr/activity/interface/service/page"
	"go-gateway/app/web-svr/activity/interface/service/preheat"
	rankv2 "go-gateway/app/web-svr/activity/interface/service/rank"
	rank "go-gateway/app/web-svr/activity/interface/service/rank_v3"
	"go-gateway/app/web-svr/activity/interface/service/remix"
	"go-gateway/app/web-svr/activity/interface/service/s10"
	"go-gateway/app/web-svr/activity/interface/service/springfestival2021"

	stockserver "go-gateway/app/web-svr/activity/interface/service/stock_server"
	"go-gateway/app/web-svr/activity/interface/service/summer_camp"
	"go-gateway/app/web-svr/activity/interface/service/system"
	"go-gateway/app/web-svr/activity/interface/service/task"
	"go-gateway/app/web-svr/activity/interface/service/timemachine"
	"go-gateway/app/web-svr/activity/interface/service/vogue"
	"go-gateway/app/web-svr/activity/interface/service/vote"
	"go-gateway/app/web-svr/activity/tools/lib/initialize"
)

var (
	LikeSvc               *like.Service
	MatchSvc              *like.Service
	BwsSvc                *bws.Service
	BnjSvc                *bnj.Service
	KfcSvc                *kfc.Service
	TmSvc                 *timemachine.Service
	AppstoreSvc           *appstore.Service
	VogueSvc              *vogue.Service
	BrandSvc              *brand.Service
	PreheatSvc            *preheat.Service
	HandWriteSvc          *handwrite.Service
	BwsOnlineSvc          *bwsonline.Service
	RemixSvc              *remix.Service
	LotterySvc            *lotteryv2.Service
	NewstarSvc            *newstar.Service
	GameHolidaySvc        *gameholiday.Service
	LolSvc                *lol.Service
	S10Svc                *s10.Service
	CollegeSvc            *college.Service
	InviteSvc             *invite.Service
	DubbingSvc            *dubbing.Service
	FunnySvc              *funny.Service
	AcgSvc                *acg.Service
	RankSvc               *rankv2.Service
	PageSvc               *page.Service
	SystemSvc             *system.Service
	NewYear2021Svc        *newyear2021.Service
	SpringFestival2021Svc *springfestival2021.Service
	CardSvc               *cards.Service
	ActSvc                *act.Service
	Rankv3Svc             *rank.Service
	EsportSvc             *esports.Service
	VoteSvr               *vote.Service
	ExaminationSvr        *examination.Service
	FitSvr                *fit.Service
	Cpc100Svr             *cpc100.Service
	ExternalBindSvr       *bind.Service
	TaskSvr               *task.Service
	KnowledgeSvr          *knowledge.Service
	CardV2Svr             *newcards.Service
	AccountSvr            *account.Service
	SummerCampSvr         *summer_camp.Service
	JsonDataSvr           *jsondata.Service
	BmlSvc                *bml.Service
	ArcSvc                *archive.Service
	MissionActivitySvr    *mission.Service
	StockSvr              *stockserver.Service
	OlympicSvr            *olympic.Service
)

func New(c *conf.Config) {
	S10Svc = s10.New(c)
	LikeSvc = like.New(c)
	MatchSvc = like.New(c)
	BwsSvc = bws.New(c)
	BnjSvc = bnj.New(c)
	KfcSvc = kfc.New(c)
	TmSvc = timemachine.New(c)
	AppstoreSvc = appstore.New(c)
	VogueSvc = vogue.New(c)
	BwsOnlineSvc = bwsonline.New(c)
	BmlSvc = bml.New(c)
	BrandSvc = brand.New(c)
	PreheatSvc = preheat.New(c)
	HandWriteSvc = handwrite.New(c)
	RemixSvc = remix.New(c)
	GameHolidaySvc = gameholiday.New(c)
	LotterySvc = lotteryv2.New(c)
	NewstarSvc = newstar.New(c)
	LolSvc = lol.New(c)
	initialize.New(college.New, func() {
		CollegeSvc = college.New(c)
	})
	InviteSvc = invite.New(c)
	DubbingSvc = dubbing.New(c)
	FunnySvc = funny.New(c)
	initialize.New(acg.New, func() {
		AcgSvc = acg.New(c)
	})
	RankSvc = rankv2.New(c)
	PageSvc = page.New(c)
	ArcSvc = archive.New(c)
	initialize.New(system.New, func() {
		SystemSvc = system.New(c)
	})
	NewYear2021Svc = newyear2021.New(c)
	initialize.New(springfestival2021.New, func() {
		SpringFestival2021Svc = springfestival2021.New(c)
	})
	initialize.New(cards.New, func() {
		CardSvc = cards.New(c)
	})
	ActSvc = act.New(c)
	initialize.New(esports.New, func() {
		EsportSvc = esports.New(c)
	})
	Rankv3Svc = rank.New(c)

	initialize.New(vote.New, func() {
		VoteSvr = vote.New(c)
	})
	initialize.New(examination.New, func() {
		ExaminationSvr = examination.New(c)
	})
	initialize.New(fit.New, func() {
		FitSvr = fit.New(c)
	})
	initialize.New(cpc100.New, func() {
		Cpc100Svr = cpc100.New(c)
	})
	initialize.New(bind.New, func() {
		ExternalBindSvr = bind.New(c)
	})
	initialize.New(examination.New, func() {
		TaskSvr = task.New(c)
	})
	initialize.New(knowledge.New, func() {
		KnowledgeSvr = knowledge.New(c)
	})
	initialize.New(examination.New, func() {
		CardV2Svr = newcards.New(c)
	})
	initialize.New(account.New, func() {
		AccountSvr = account.New(c)
	})
	initialize.New(summer_camp.New, func() {
		SummerCampSvr = summer_camp.New(c)
	})
	initialize.New(jsondata.New, func() {
		JsonDataSvr = jsondata.New(c)
	})
	initialize.New(mission.New, func() {
		MissionActivitySvr = mission.New(c)
	})
	initialize.New(stockserver.New, func() {
		StockSvr = stockserver.New(c)
	})
	initialize.New(olympic.New, func() {
		OlympicSvr = olympic.New(c)
	})
}

func Close() {
	if LikeSvc != nil {
		LikeSvc.Close()
	}
	if MatchSvc != nil {
		MatchSvc.Close()
	}
	if BnjSvc != nil {
		BnjSvc.Close()
	}
	if BwsSvc != nil {
		S10Svc.Close()
	}
	if KfcSvc != nil {
		KfcSvc.Close()
	}
	if TmSvc != nil {
		TmSvc.Close()
	}
	if AppstoreSvc != nil {
		AppstoreSvc.Close()
	}
	if BrandSvc != nil {
		BrandSvc.Close()
	}
	if HandWriteSvc != nil {
		HandWriteSvc.Close()
	}
	if RemixSvc != nil {
		RemixSvc.Close()
	}
	if LotterySvc != nil {
		LotterySvc.Close()
	}
	if GameHolidaySvc != nil {
		GameHolidaySvc.Close()
	}
	if LolSvc != nil {
		LolSvc.Close()
	}
	if S10Svc != nil {
		S10Svc.Close()
	}
	if CollegeSvc != nil {
		CollegeSvc.Close()
	}
	if InviteSvc != nil {
		InviteSvc.Close()
	}
	if DubbingSvc != nil {
		DubbingSvc.Close()
	}
	if FunnySvc != nil {
		FunnySvc.Close()
	}
	if AcgSvc != nil {
		AcgSvc.Close()
	}
	if RankSvc != nil {
		RankSvc.Close()
	}
	if SpringFestival2021Svc != nil {
		SpringFestival2021Svc.Close()
	}
	if CardSvc != nil {
		CardSvc.Close()
	}
	if ExaminationSvr != nil {
		ExaminationSvr.Close()
	}
	if CardV2Svr != nil {
		CardV2Svr.Close()
	}
	if AccountSvr != nil {
		AccountSvr.Close()
	}
	if SummerCampSvr != nil {
		SummerCampSvr.Close()
	}
	if JsonDataSvr != nil {
		JsonDataSvr.Close()
	}
	if ArcSvc != nil {
		ArcSvc.Close()
	}
}

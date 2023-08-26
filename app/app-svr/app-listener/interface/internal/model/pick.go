package model

import (
	"fmt"
	"sync"

	"go-common/library/log"
	v1 "go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"

	listenerSvc "git.bilibili.co/bapis/bapis-go/dynamic/service/listener"
)

type SinglePick struct {
	Pick *listenerSvc.CollectionGroup
}

type SingleCollection struct {
	PickTitle  string // 最外层的播单组标题
	PickId     int64
	Collection *listenerSvc.Collection
}

func (sc SingleCollection) ToV1PlayItems() []*v1.PlayItem {
	ret := make([]*v1.PlayItem, 0, len(sc.Collection.GetArchives()))
	for _, arc := range sc.Collection.GetArchives() {
		ret = append(ret, &v1.PlayItem{ItemType: PlayItemUGC, Oid: arc.Aid})
	}
	return ret
}

const (
	PickFeed   = "feed"
	PickDetail = "detail"
)

type PickContext struct {
	C    *conf.AppConfig
	once sync.Once
	From string            // 区分具体处理逻辑
	In   *SingleCollection // 输入的单个播单卡
	Out  *v1.PickCard      // 最终吐出的成品卡
	MidM map[int64]*MemberInfo
	ArcM map[int64]*ArchiveInfo
}

func (p *PickContext) Init() {
	p.once.Do(func() {
		p.MidM = make(map[int64]*MemberInfo)
		p.ArcM = make(map[int64]*ArchiveInfo)
	})
}

func (p *PickContext) GetPickFeedCard(in SingleCollection) *v1.PickCard {
	p.In = &in
	p.Out = new(v1.PickCard)
	p.runHandlers(p.handleBase, p.handleHeader, p.handleArchive, p.handleSeeMoreBtn)
	return p.Out
}

func (p *PickContext) GetPickDetailModules(in SingleCollection) []*v1.CardModule {
	p.In = &in
	p.Out = new(v1.PickCard)
	p.runHandlers(p.handleHeader, p.handleArchive)
	return p.Out.Modules
}

func (p *PickContext) runHandlers(hds ...func()) {
	for _, hf := range hds {
		hf()
	}
}

// 填充基本信息
func (p *PickContext) handleBase() {
	p.Out.PickId = p.In.PickId
	p.Out.CardId = p.In.Collection.Id
	p.Out.CardName = p.In.PickTitle
}

func (p *PickContext) handleHeader() {
	var text string
	switch p.From {
	case PickDetail:
		text = p.C.Res.Text.PickHeaderDetailBtn
	default:
		text = p.C.Res.Text.PickHeaderBtn
	}
	p.Out.Modules = append(p.Out.Modules, &v1.CardModule{
		ModuleType: v1.CardModuleType_Module_header,
		Module: &v1.CardModule_ModuleHeader{
			ModuleHeader: &v1.PkcmHeader{
				Title:   p.In.Collection.Title,
				Desc:    fmt.Sprintf(p.C.Res.Text.PickHeaderDesc, p.In.Collection.TotalNum, formatDuration(p.In.Collection.TotalTime)),
				BtnText: text,
				BtnIcon: p.C.Res.Icon.PickHeaderBtn,
				// btnUri暂时留空 依赖客户端默认行为
			},
		},
	})
}

func (p *PickContext) handleArchive() {
	for _, arc := range p.In.Collection.GetArchives() {
		ai, ok := p.ArcM[arc.Aid]
		if !ok || ai == nil {
			log.Warn("PickArchive(%d) info not found", arc.Aid)
			continue
		}
		p.Out.Modules = append(p.Out.Modules, &v1.CardModule{
			ModuleType: v1.CardModuleType_Module_archive,
			Module: &v1.CardModule_ModuleArchive{
				ModuleArchive: &v1.PkcmArchive{
					PickReason: arc.Recommend,
					Arc:        ai.ToV1PickArchive(p.In.PickId, p.In.Collection.Id),
				},
			},
		})
	}
}

func (p *PickContext) handleSeeMoreBtn() {
	if p.In.Collection.TotalNum > p.In.Collection.DisplayNum {
		p.Out.Modules = append(p.Out.Modules, &v1.CardModule{
			ModuleType: v1.CardModuleType_Module_cbtn,
			Module: &v1.CardModule_ModuleCbtn{
				ModuleCbtn: &v1.PkcmCenterButton{
					Title: p.C.Res.Text.PickSeeMoreBtn,
					// TODO: move to config
					IconTail: p.C.Res.Icon.PickSeeMoreBtn,
					// Uri: 目前留空依赖客户端默认行为
				},
			},
		})
	}
}

// 返回 x小时x分钟
//
//nolint:gomnd
func formatDuration(secs int64) string {
	ret := ""
	switch {
	case secs <= 60:
		ret = fmt.Sprintf("%d秒", secs)
	case secs <= 3600:
		ret = fmt.Sprintf("%d分钟", secs/60)
		if secs%60 != 0 {
			ret += fmt.Sprintf("%d秒", secs%60)
		}
	default:
		ret = fmt.Sprintf("%d小时", secs/3600)
		if secs%3600/60 != 0 {
			ret += fmt.Sprintf("%d分钟", secs%3600/60)
		}
	}
	return ret
}

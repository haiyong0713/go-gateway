package native

import (
	"context"
	"fmt"
	"runtime"
	"time"

	actgrpc "git.bilibili.co/bapis/bapis-go/activity/service"
	"github.com/jinzhu/gorm"
	"go-common/library/log"

	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"
	natpagegrpc "go-gateway/app/web-svr/native-page/interface/api"
)

func (d *Dao) AddPageFromNewact(c context.Context, st *actgrpc.Subject, fromType int) (int64, error) {
	page := &natmdl.PageParam{
		Title:      st.Name,
		Creator:    "system",
		Operator:   "system",
		Type:       natpagegrpc.NewactType,
		ForeignID:  st.ID,
		FromType:   fromType,
		State:      natpagegrpc.WaitForOnline,
		ShareImage: st.ActivityImage,
		ShareTitle: fmt.Sprintf("%s-%s", st.Stime.Time().Format("2006.01.02"), st.Etime.Time().Format("2006.01.02")),
		Attribute:  1 << natpagegrpc.AttrIsNotNightModule,
		BgColor:    "#0000",
	}
	pageIDRes, err := d.AddPage(c, page)
	if err != nil {
		return 0, err
	}
	pageID := pageIDRes.ID
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			const size = 64 << 10
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Error("Fail to add page from newact, sid=%d panic=%+v", st.ID, buf)
			return
		}
		if err != nil {
			if err1 := tx.Rollback().Error; err1 != nil {
				log.Error("Fail to rollback, sid=%d error=%+v", st.ID, err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("Fail to commit, sid=%d error=%+v", st.ID, err)
			return
		}
	}()
	newactHeader := &natmdl.ConfNewactHeader{Fid: st.ID}
	if err = d.addNewactHeader(tx, pageID, newactHeader, 1); err != nil {
		return pageID, err
	}
	newactAward := &natmdl.ConfNewactAward{Fid: st.ID}
	if err = d.addNewactAward(tx, pageID, newactAward, 2); err != nil {
		return pageID, err
	}
	if err = d.addNewactStatement(tx, pageID, &natmdl.ConfNewactStatement{Fid: st.ID, Type: natpagegrpc.StatementNewactTask}, 3); err != nil {
		return pageID, err
	}
	if err = d.addNewactStatement(tx, pageID, &natmdl.ConfNewactStatement{Fid: st.ID, Type: natpagegrpc.StatementNewactRule}, 4); err != nil {
		return pageID, err
	}
	if err = d.addNewactStatement(tx, pageID, &natmdl.ConfNewactStatement{Fid: st.ID, Type: natpagegrpc.StatementNewactDeclaration}, 5); err != nil {
		return pageID, err
	}
	bottomBtn := newactBottomBtnCfg(fromType)
	if err = d.addBottomButton(tx, pageID, bottomBtn, 1); err != nil {
		return pageID, err
	}
	pageAttrs := map[string]interface{}{
		"ver":   buildVer("admin"),
		"state": natmdl.OnlineState,
		"stime": time.Now().Format("2006-01-02 15:04:05"),
		"etime": time.Unix(2147356800, 0).Format("2006-01-02 15:04:05"),
	}
	if err = tx.Table(_tablePage).Where("id=?", pageID).Where("state=?", natmdl.WaitForOnline).Update(pageAttrs).Error; err != nil {
		log.Error("Fail to update native_page, id=%+v attrs=%+v error=%+v", pageID, pageAttrs, err)
		return pageID, err
	}
	return pageID, nil
}

func (d *Dao) addNewactHeader(tx *gorm.DB, pageID int64, cfg *natmdl.ConfNewactHeader, order int) error {
	if cfg == nil || tx == nil {
		return nil
	}
	module := &natmdl.NatModule{}
	module.ToNewactHeader(cfg, pageID, order, natpagegrpc.CommonPage, generateUkey(natmdl.UkeyPrefixNewact))
	if err := tx.Create(module).Error; err != nil {
		log.Error("Fail to save NewactHeader module, pageID=%+v error=%+v", pageID, err)
		return err
	}
	return nil
}

func (d *Dao) addNewactAward(tx *gorm.DB, pageID int64, cfg *natmdl.ConfNewactAward, order int) error {
	if cfg == nil || tx == nil {
		return nil
	}
	module := &natmdl.NatModule{}
	module.ToNewactAward(cfg, pageID, order, natpagegrpc.CommonPage, generateUkey(natmdl.UkeyPrefixNewact))
	if err := tx.Create(module).Error; err != nil {
		log.Error("Fail to save NewactAward module, pageID=%+v error=%+v", pageID, err)
		return err
	}
	return nil
}

func (d *Dao) addNewactStatement(tx *gorm.DB, pageID int64, cfg *natmdl.ConfNewactStatement, order int) error {
	if cfg == nil || tx == nil {
		return nil
	}
	module := &natmdl.NatModule{}
	module.ToNewactStatement(cfg, pageID, order, natpagegrpc.CommonPage, generateUkey(natmdl.UkeyPrefixNewact))
	if err := tx.Create(module).Error; err != nil {
		log.Error("Fail to save NewactStatement module, pageID=%+v error=%+v", pageID, err)
		return err
	}
	return nil
}

func (d *Dao) addBottomButton(tx *gorm.DB, pageID int64, cfg *natmdl.ConfBottomButton, order int) error {
	if cfg == nil || tx == nil {
		return nil
	}
	module := &natmdl.NatModule{}
	module.ToMbottomButton(cfg, pageID, order, natpagegrpc.CommonBaseModule, generateUkey(natmdl.UkeyPrefixNewact))
	if err := tx.Create(module).Error; err != nil {
		log.Error("Fail to save BottomButton module, pageID=%+v error=%+v", pageID, err)
		return err
	}
	for _, area := range cfg.Areas {
		click := &natmdl.Click{}
		click.ToClick(module.ID, area, cfg.Width, cfg.Height)
		if err := tx.Create(click).Error; err != nil {
			log.Error("Fail to create BottomButton Area, area=%+v error=%+v", area, err)
			return err
		}
	}
	return nil
}

func newactBottomBtnCfg(fromType int) *natmdl.ConfBottomButton {
	var link string
	switch fromType {
	case natpagegrpc.PageFromNewactCollect, natpagegrpc.PageFromNewactVote:
		link = "bilibili://uper/user_center/add_archive/?from=0&is_new_ui=1&relation_from=NAactivity"
	case natpagegrpc.PageFromNewactShoot:
		link = "bilibili://uper/user_center/add_archive/?from=1&is_new_ui=1&relation_from=NAactivity"
	}
	ukey := generateUkey(natmdl.UkeyPrefixNewact)
	return &natmdl.ConfBottomButton{
		Image:  "https://i0.hdslb.com/bfs/activity-plat/static/20211105/a694e9444289630c855c5b3cf6d9880a/zhXRI4gHzP.png",
		Width:  375,
		Height: 96,
		Areas: []*natmdl.Areas{
			{
				X:    15,
				Y:    13,
				W:    346,
				H:    44,
				Link: link,
				Type: 0,
				UKey: ukey,
				ID:   ukey,
			},
		},
	}
}

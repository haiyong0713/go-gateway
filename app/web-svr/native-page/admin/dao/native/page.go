package native

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	natmdl "go-gateway/app/web-svr/native-page/admin/model/native"
	"go-gateway/app/web-svr/native-page/interface/api"
	"go-gateway/pkg/idsafe/bvid"

	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"github.com/jinzhu/gorm"
	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/pkg/errors"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/xstr"
)

const (
	_tablePage        = "native_page"
	_overdue          = 0 // 失效
	_valid            = 1 // 有效
	_delete           = 2
	_userMax          = 40
	_module           = "native_module"
	_cilck            = "native_click"
	_act              = "native_act"
	_videoExt         = "native_video_ext"
	_dynamicExt       = "native_dynamic_ext"
	_mixtureExt       = "native_mixture_ext"
	_participationExt = "native_participation_ext"
	_pageExt          = "native_page_dyn"

	_updateSQL         = "UPDATE native_page SET stime=?,share_title=?,share_image=?,operator=?,share_url=?,spmid =?,etime=?,pc_url=?,another_title=?,share_caption=?,attribute=?,bg_color=? WHERE id=?"
	_upDynExtSQL       = "UPDATE native_page_dyn SET `validity`=?,`stime`=?,`square_title`=?,`small_card`=?,`big_card`=?,`tids`=? WHERE `pid`=?"
	_addDynExtSQL      = "INSERT INTO native_page_dyn(`validity`,`stime`,`square_title`,`small_card`,`big_card`,`tids`,`pid`) VALUES (?,?,?,?,?,?,?)"
	_skipSQL           = "UPDATE native_page SET operator=?,skip_url=?,share_title=?,share_image=?,stime=?,spmid =?,etime=?,pc_url=?,another_title=?,share_caption=?,attribute=? WHERE id=?"
	_delskipSQL        = "UPDATE native_page SET skip_url=? WHERE state!=2 AND id=?"
	_mixExtBatchAddSQL = "INSERT INTO native_mixture_ext(module_id,foreign_id,`state`,rank,m_type,reason) VALUES%s"
	_batchAddActSQL    = "INSERT INTO native_act(`module_id`,`state`,`rank`,`page_id`) VALUES %s"

	_defaultShareImage = "https://i0.hdslb.com/bfs/activity-plat/static/8347b7383c4a730528a82854f98b9b32/sYbPL4QDx9.png"
)

// isDynamicChoice.
func isDynamicChoice(selectType string) bool {
	// 精选模式
	return selectType == natmdl.Choice
}

// AddPage .
func (d *Dao) AddPage(c context.Context, natPage *natmdl.PageParam) (pageID *natmdl.AddReID, err error) {
	if err = d.DB.Create(natPage).Error; err != nil {
		log.Error("[AddPage] d.DB.Save(%v), error(%v)", natPage, err)
		return
	}
	pageID = &natmdl.AddReID{
		ID: natPage.ID,
	}
	return
}

// ModifyPage .
func (d *Dao) ModifyPage(c context.Context, id int64, arg map[string]interface{}) error {
	if err := d.DB.Table(_tablePage).Where("id=?", id).Update(arg).Error; err != nil {
		log.Error("ModifyPage d.DB.Table(%d,%v) error(%v)", id, arg, err)
		return err
	}
	return nil
}

// DelPage .
func (d *Dao) DelPage(c context.Context, id int64, user, offReason string) (err error) {
	//menutype不支持删除，避免误删 影响首页逻辑
	if err = d.DB.Table(_tablePage).
		Where("id=?", id).
		Where("type not in (?,?,?,?,?,?,?)", api.BottomType, natmdl.MenuType, natmdl.OgvType, natmdl.UgcType, natmdl.PlayerType, natmdl.SpaceType, api.LiveTabType).
		Update(map[string]interface{}{"state": _delete, "operator": user, "off_reason": offReason}).Error; err != nil {
		log.Error("[DelPage] d.DB.Save(), ID(%d) error(%v)", id, err)
	}
	return
}

func (d *Dao) delBegin(c context.Context, tx *gorm.DB, pageID int64, pTypes []int) (err error) {
	// 存在的话，先删除组件，再添加
	var mous []*natmdl.NatModule
	if mous, err = d.ModulesInfo(c, pageID, pTypes); err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
		} else {
			log.Error("d.moduleID() error(%v)", err)
		}
		return
	}
	var moduleID, inlineMIDs []int64
	for _, v := range mous {
		if v.ID == 0 {
			continue
		}
		moduleID = append(moduleID, v.ID)
		if v.Category == natmdl.InlineTabModule || v.Category == natmdl.SelectModule {
			inlineMIDs = append(inlineMIDs, v.ID)
		}
	}
	if len(inlineMIDs) > 0 {
		mixRly := make([]*natmdl.MixtureExt, 0)
		if err = tx.Table(_mixtureExt).Select("foreign_id").Where("module_id in (?)", inlineMIDs).Where("m_type=?", natmdl.MixturePageType).Where("state=?", _valid).Find(&mixRly).Error; err != nil {
			log.Error("tx.Update() error(%v)", err)
			return
		}
		var pageIDs []int64
		for _, v := range mixRly {
			if v == nil || v.ForeignID == 0 {
				continue
			}
			pageIDs = append(pageIDs, v.ForeignID)
		}
		if len(pageIDs) > 0 {
			if err = tx.Table(_tablePage).Where("id in (?)", pageIDs).Where("type=?", natmdl.InLineType).Update(map[string]interface{}{"state": _delete}).Error; err != nil {
				log.Error("[DelPage] tx.update(), ID(%d) error(%v)", pageIDs, err)
				return
			}
		}
	}
	if err = tx.Table(_module).Where("native_id=?", pageID).Where("p_type in (?)", pTypes).Where("state !=?", _overdue).Update(map[string]interface{}{"state": _overdue}).Error; err != nil {
		log.Error("tx.Update() error(%v)", err)
		return
	}
	for _, v := range moduleID {
		if err = tx.Table(_cilck).Where("module_id=?", v).Where("state !=?", _overdue).Update(map[string]interface{}{"state": _overdue}).Error; err != nil {
			log.Error("tx.Update() error(%v)", err)
			return
		}
		if err = tx.Table(_act).Where("module_id=?", v).Where("state !=?", _overdue).Update(map[string]interface{}{"state": _overdue}).Error; err != nil {
			log.Error("tx.Update() error(%v)", err)
			return
		}
		if err = tx.Table(_participationExt).Where("module_id=?", v).Where("state !=?", _overdue).Update(map[string]interface{}{"state": _overdue}).Error; err != nil {
			log.Error("tx.Update() error(%v)", err)
			return
		}
	}
	return
}

func (d *Dao) DynExtByPid(c context.Context, pid int64) (*natmdl.PageDyn, error) {
	rly := &natmdl.PageDyn{}
	if err := d.DB.Table(_pageExt).Where("pid=?", pid).First(rly).Error; err != nil {
		return nil, err
	}
	return rly, nil
}

func (d *Dao) UpDynExt(c context.Context, natPage *natmdl.EditParam) error {
	//先查询是否有pid对应的数据
	_, err := d.DynExtByPid(c, natPage.ID)
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Error("d.DynExtByPid(%d) error(%v)", natPage.ID, err)
		return err
	}
	stime := time.Unix(natPage.ValidStime, 0).Format("2006-01-02 15:04:05")
	if err == gorm.ErrRecordNotFound {
		//插入 pid是唯一索引
		if err = d.DB.Exec(_addDynExtSQL, natPage.Validity, stime, natPage.SquareTitle, natPage.SmallCard, natPage.BigCard, xstr.JoinInts(natPage.Tids), natPage.ID).Error; err != nil {
			log.Error("UpDynExt add %d error(%v)", natPage.ID, err)
			return err
		}
	} else {
		//更新
		if err = d.DB.Exec(_upDynExtSQL, natPage.Validity, stime, natPage.SquareTitle, natPage.SmallCard, natPage.BigCard, xstr.JoinInts(natPage.Tids), natPage.ID).Error; err != nil {
			log.Error("UpDynExt update %d error(%v)", natPage.ID, err)
			return err
		}
	}
	return nil
}

// UpdatePage .
func (d *Dao) UpdatePage(c context.Context, natPage *natmdl.EditParam, pType int) (err error) {
	var (
		tx     = d.DB.Begin()
		module []int64
	)
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("UpdatePage %v", r)
		}
		if err == nil {
			if err = tx.Commit().Error; err != nil {
				log.Error("commit(%+v)", err)
				return
			}
		} else {
			tx.Rollback()
		}
	}()
	var mous []*natmdl.NatModule
	if mous, err = d.ModulesInfo(c, natPage.ID, []int{pType}); err != nil {
		log.Error("moduleID error(%v)", err)
		return
	}
	for _, v := range mous {
		if v.ID > 0 {
			module = append(module, v.ID)
		}
	}
	if len(module) == 0 {
		err = ecode.Error(ecode.RequestErr, "该话题下没有一个有效的组件")
		return
	}
	if err = tx.Exec(_delskipSQL, "", natPage.ID).Error; err != nil {
		log.Error("[PageSkipUrl] d.DB.Exec(_delskipSQL) error(%v)", err)
		return
	}
	stime := time.Unix(natPage.Stime, 0).Format("2006-01-02 15:04:05")
	etime := time.Unix(natPage.Etime, 0).Format("2006-01-02 15:04:05")
	if err = tx.Exec(_updateSQL, stime, natPage.ShareTitle, natPage.ShareImage, natPage.UserName, natPage.ShareUrl, natPage.Spmid, etime, natPage.PcUrl, natPage.AnotherTitle, natPage.ShareCaption, natPage.Attribute, natPage.BgColor, natPage.ID).Error; err != nil {
		log.Error("[UpdatePage] d.DB.Update() error(%v)", err)
	}
	return
}

// PageSkipUrl
func (d *Dao) PageSkipUrl(c context.Context, param *natmdl.EditParam, pType int) (err error) {
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("UpdatePage %v", r)
		}
		if err == nil {
			if err = tx.Commit().Error; err != nil {
				log.Error("commit(%+v)", err)
				return
			}
		} else {
			tx.Rollback()
		}
	}()
	if err = d.delBegin(c, tx, param.ID, []int{pType}); err != nil {
		log.Error("[PageSkipUrl] delBegin pageID(%d) error(%v)", param.ID, err)
		return
	}
	stime := time.Unix(param.Stime, 0).Format("2006-01-02 15:04:05")
	etime := time.Unix(param.Etime, 0).Format("2006-01-02 15:04:05")
	if err = tx.Exec(_skipSQL, param.UserName, param.SkipUrl, param.ShareTitle, param.ShareImage, stime, param.Spmid, etime, param.PcUrl, param.AnotherTitle, param.ShareCaption, param.Attribute, param.ID).Error; err != nil {
		log.Error("[PageSkipUrl] d.DB.Exec(_skipSQL) error(%v)", err)
	}
	return
}

// PageByID
func (d *Dao) PageByID(c context.Context, id int64) (res *natmdl.FindRes, err error) {
	res = &natmdl.FindRes{}
	if err = d.DB.Table(_tablePage).Select("foreign_id,state,stime,type").Where("id=?", id).First(&res).Error; err != nil {
		log.Error("[PageByID] d.DB.Where() error(%v)", err)
	}
	return
}

// PageByFID .
func (d *Dao) PageByFID(c context.Context, foreignID int64, genre int) (pageRes []*natmdl.PageParam, err error) {
	if err = d.DB.Table(_tablePage).Where("foreign_id=?", foreignID).Where("type=?", genre).Where("state in (?,?)", natmdl.OnlineState, natmdl.WaitForOnline).Find(&pageRes).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = nil
			return
		}
		log.Error("[PageByFID] d.DB.Count() error(%v)", err)
	}
	return
}

// ModulesInfo .
func (d *Dao) ModulesInfo(c context.Context, nativeID int64, pType []int) (mou []*natmdl.NatModule, err error) {
	if err = d.DB.Table(_module).Where("native_id=?", nativeID).Where("p_type in (?)", pType).Where("state=?", _valid).Find(&mou).Error; err != nil {
		log.Error("ModuleID error(%v)", err)
	}
	return
}

// SearchPage .
func (d *Dao) SearchPage(c context.Context, param *natmdl.SearchParam) (res *natmdl.SearchRes, err error) {
	var (
		db       = d.DB.Table(_tablePage)
		pageInfo []*natmdl.NatPage
	)
	res = &natmdl.SearchRes{}
	if param.Title != "" {
		db = db.Where("title LIKE ?", "%"+param.Title+"%")
	}
	if param.Creator != "" {
		db = db.Where("creator=?", param.Creator)
	}
	if param.BeginTime != "" && param.EndTime != "" {
		db = db.Where("ctime>=?", param.BeginTime).Where("ctime<=?", param.EndTime)
	}
	if param.RelatedUid != 0 {
		db = db.Where("related_uid=?", param.RelatedUid)
	}
	if param.ActOrigin != "" {
		db = db.Where("act_origin=?", param.ActOrigin)
	}
	db = db.Where("type in (?)", param.Ptypes)
	db = db.Order("id DESC") // 前端要求列表倒序
	db = db.Where("state in (?)", param.States)
	if len(param.FromTypes) != 0 {
		db = db.Where("from_type in (?)", param.FromTypes)
	}
	db.Count(&res.Page.Total)
	if err = db.Offset((param.Pn - 1) * param.Ps).Limit(param.Ps).Find(&pageInfo).Error; err != nil {
		log.Error("[SearchPage] error(%v)", err)
		return
	}
	res.Item = pageInfo
	res.Page.Size = param.Ps
	res.Page.Num = param.Pn
	return
}

// FindPage .
func (d *Dao) FindPage(c context.Context, title string, id, defaulType int64, states []int64) (*natmdl.NatPage, error) {
	res := &natmdl.NatPage{}
	db := d.DB.Table(_tablePage)
	if id != 0 {
		db = db.Where("id=?", id)
	}
	if title != "" {
		db = db.Where("title=?", title).Where("type=?", defaulType)
	}
	if err := db.Where("state in (?)", states).First(&res).Error; err != nil {
		log.Error("[FindPage] error(%v)", err)
		return nil, err
	}
	return res, nil
}

func (d *Dao) FindPageByIds(c context.Context, ids []int64) (pages []*natmdl.NatPage, err error) {
	if err = d.DB.Table(_tablePage).Where("id in (?)", ids).Find(&pages).Error; err != nil {
		log.Error("[FindPageByIds] d.DB.Find(%v), error(%v)", ids, err)
	}
	return
}

func (d *Dao) FindPageById(c context.Context, id int64) (*natmdl.NatPage, error) {
	page := &natmdl.NatPage{}
	if err := d.DB.Table(_tablePage).Where("id = ?", id).First(&page).Error; err != nil {
		log.Error("Fail to find page, id=%+v error=%+v", id, err)
		return nil, err
	}
	return page, nil
}

// NatTagID .
func (d *Dao) NatTagID(c context.Context, title string) (tagID int64, err error) {
	var rly *tagrpc.TagReply
	if rly, err = d.tagGRPC.TagByName(c, &tagrpc.TagByNameReq{Tname: title}); err != nil {
		log.Error("d.tagGRPC.TagByName  error(%v)", err)
		return
	}
	if rly == nil || rly.Tag == nil {
		err = ecode.RequestErr
		return
	}
	tagID = rly.Tag.Id
	return
}

func (d *Dao) AddTag(c context.Context, tagName string) (*tagrpc.Tag, error) {
	req := &tagrpc.AddTagReq{Name: tagName}
	rly, err := d.tagGRPC.AddTag(c, req)
	if err != nil {
		log.Error("Fail to add tag, req=%+v error=%+v", req, err)
		return nil, err
	}
	if rly.GetTag() == nil {
		log.Error("Fail to add tag, tag is nil, req=%+v", req)
		return nil, errors.New("tag is nil")
	}
	return rly.GetTag(), nil
}

// Tags .
func (d *Dao) Tags(c context.Context, tids []int64) (map[int64]*tagrpc.Tag, error) {
	rly, err := d.tagGRPC.Tags(c, &tagrpc.TagsReq{Tids: tids})
	if err != nil {
		return nil, err
	}
	if rly == nil {
		return nil, ecode.NothingFound
	}
	return rly.Tags, nil

}
func rankModule(param string, child map[string]int) (idx int, err error) {
	if order, ok := child[param]; ok {
		idx = order
	} else {
		err = ecode.Errorf(ecode.NothingFound, "children给定组件在列表中找不到")
		log.Error("children给定组件在列表中找不到 error(%v)", err)
	}
	return
}

// CommitTsModule .
// nolint:gocognit
func (d *Dao) CommitTsModule(c context.Context, page *natmdl.NatPage, modules []*natmdl.NatTsModule, tsid int64, tsInfo *natmdl.NatTsPage) (err error) {
	tx := d.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("CommitTsModule %v", r)
		}
		if err == nil {
			if err = tx.Commit().Error; err != nil {
				log.Error("CommitTsModule(%+v)", err)
				return
			}
		} else {
			tx.Rollback()
		}
	}()
	// 存在的话，先删除组件，再添加
	if err = d.delBegin(c, tx, page.ID, []int{0, 1, 2, 3, 4}); err != nil {
		log.Error("[CommitModule] delBegin error(%v)", err)
		return
	}
	resources, err := d.getTsModResourceFromModules(c, modules)
	if err != nil {
		return
	}
	var (
		ct         int
		shareTitle string
	)
	for _, v := range modules {
		if v == nil {
			continue
		}
		ct++
		tmpPage := &api.NativeModule{Category: int64(v.Category)}
		mod := &natmdl.NatModule{}
		switch {
		case tmpPage.IsStatement(): // 文本组件
			if len(v.Remark) == 0 {
				continue
			}
			banner := &natmdl.StatementModule{Remark: v.Remark}
			shareTitle = v.Remark
			mod.ToStatement(banner, page.ID, v.Rank, v.PType, v.Ukey)
		case tmpPage.IsClick(): //自定义点击组件
			click := &natmdl.ConfClick{Image: v.Meta, Width: v.Width, Height: v.Length}
			mod.ToMclick(click, page.ID, v.Rank, v.PType, v.Ukey)
		case tmpPage.IsResourceID():
			resource := v.Trans2ConfResource()
			mod.ToMResource(resource, page.ID, v.Rank, v.PType, v.Ukey)
		case tmpPage.IsNewVideoID():
			resource := v.Trans2ConfArchive()
			mod.ToMArchive(resource, page.ID, v.Rank, v.PType, v.Ukey)
		case tmpPage.IsActCapsule():
			actCapsule := &natmdl.ConfActCapsule{Caption: v.Remark}
			mod.ToActCapsule(actCapsule, page.ID, v.Rank, v.PType, v.Ukey)
		case tmpPage.IsRecommend():
			rcmd := &natmdl.ConfRecommned{Category: v.Category, Num: v.Num}
			mod.ToMRecommend(rcmd, page.ID, v.Rank, v.PType, v.Ukey)
		case tmpPage.IsCarouselImg():
			carousel := &natmdl.ConfCarousel{Category: v.Category, ContentStyle: 1}
			mod.ToMCarousel(carousel, page.ID, v.Rank, v.PType, v.Ukey)
		default:
			continue
		}
		// 组件表插入 native_module
		if err = tx.Create(mod).Error; err != nil {
			log.Error("[SaveModule] tx.Create(mClick) error(%v)", err)
			return
		}
		err = func() error {
			modResources, ok := resources[v.ID]
			if !ok && !tmpPage.IsActCapsule() {
				return nil
			}
			switch {
			case tmpPage.IsResourceID(), tmpPage.IsNewVideoID():
				attrs := make([]*natmdl.ResourceIDs, 0, len(modResources))
				for _, v := range modResources {
					if v.ResourceType != natmdl.MixtureArcType || v.ResourceID == 0 {
						continue
					}
					attrs = append(attrs, &natmdl.ResourceIDs{Type: int(v.ResourceType), ID: v.ResourceID})
				}
				if err = d.batchAddMixtureExt(tx, mod.ID, attrs, nil); err != nil {
					return err
				}
			case tmpPage.IsActCapsule():
				pageIDs := make([]int64, 0, len(modResources))
				// 空间页第一位要固定展示当前活动
				pageIDs = append(pageIDs, page.ID)
				for _, v := range modResources {
					if v.ResourceType != natmdl.MixtureAct || v.ResourceID == 0 {
						continue
					}
					pageIDs = append(pageIDs, v.ResourceID)
				}
				if err = d.batchAddActPage(tx, mod.ID, pageIDs); err != nil {
					return err
				}
			case tmpPage.IsRecommend():
				attrs := make([]*natmdl.ResourceIDs, 0, len(modResources))
				reasons := make([]string, 0, len(modResources))
				for _, v := range modResources {
					if v.ResourceType != natmdl.MixtureUpType || v.ResourceID == 0 {
						continue
					}
					attrs = append(attrs, &natmdl.ResourceIDs{Type: int(v.ResourceType), ID: v.ResourceID})
					reasons = append(reasons, "活动发起人")
				}
				if err = d.batchAddMixtureExt(tx, mod.ID, attrs, reasons); err != nil {
					return err
				}
			case tmpPage.IsCarouselImg():
				if len(modResources) == 0 || modResources[0].Ext == "" {
					return nil
				}
				rawExt := &natmdl.ResourceExt{}
				if err := json.Unmarshal([]byte(modResources[0].Ext), rawExt); err != nil {
					log.Error("Fail to unmarshal ResourceExt, ext=%s error=%+v", modResources[0].Ext, err)
					return err
				}
				carousel := &natmdl.ConfCarousel{
					Category: v.Category,
					ImgList:  []*natmdl.CarouselImg{{ImgUrl: rawExt.ImgUrl, Length: rawExt.Length, Width: rawExt.Width}},
				}
				if err = d.AddCarouselMixtureExt(tx, carousel, mod.ID); err != nil {
					return err
				}
			}
			return nil
		}()
		if err != nil {
			return
		}
	}
	//动态列表，无限feed流，综合排序
	dy := &natmdl.ConfDynamic{DySort: 2, Attribute: 1, SourceID: page.ForeignID}
	dyMod := &natmdl.NatModule{}
	dyMod.ToMdynamic(dy, page.ID, ct, 0, generateUkey(natmdl.UkeyPrefixUP))
	// 组件表插入 native_module
	if err = tx.Create(dyMod).Error; err != nil {
		log.Error("[SaveModule] tx.Create(Dynamic) error(%v)", err)
		return
	}
	// 0表示全部，选择全部时，不支持其余多选模式
	dynamicExt := &natmdl.DynamicExt{}
	dynamicExt.ToDynamicExt(0, dyMod.ID, 0, false)
	if err = tx.Create(dynamicExt).Error; err != nil {
		log.Error("tx.Save(dynamicExt) error(%v)", err)
		return
	}
	// /动态列表，无限feed流，综合排序
	//版头组件
	head := &natmdl.HeadBase{Title: "发起", Attribute: 16}
	heMod := &natmdl.NatModule{}
	heMod.ToHead(head, page.ID, 1, 4, generateUkey(natmdl.UkeyPrefixUP))
	// 组件表插入 native_module
	if err = tx.Create(heMod).Error; err != nil {
		log.Error("[SaveModule] tx.Create(head) error(%v)", err)
		return
	}
	//投稿组件
	part := &natmdl.ConfParticipation{Button: []*natmdl.ParticipationExt{{MType: 0, Title: "参与话题讨论"}}}
	mod := &natmdl.NatModule{}
	mod.ToMPart(part, page.ID, 1, 2, generateUkey(natmdl.UkeyPrefixUP))
	// 组件表插入 native_module
	if err = tx.Create(mod).Error; err != nil {
		log.Error("[SaveModule] tx.Create(Participation) error(%v)", err)
		return
	}
	for k, button := range part.Button {
		button.ToParticipation(mod.ID, k)
		if err = tx.Create(button).Error; err != nil {
			log.Error("tx.Save(Participation) error(%v)", err)
			return
		}
	}
	// 审核通过
	var (
		upArg map[string]interface{}
	)
	ver := fmt.Sprintf("%d-%s", time.Now().UnixNano()/1e6, "up")
	shareImage := buildShareImage(tsInfo.ShareImage, page.ShareImage)
	if page.State == natmdl.WaitForCheck { // 首次审核
		upArg = map[string]interface{}{
			"pc_url": fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d", page.ID),
			"ver":    ver, "creator": "system", "state": natmdl.OnlineState, "share_image": shareImage, "share_title": shareTitle, "stime": time.Now().Format("2006-01-02 15:04:05"), "etime": time.Unix(2147356800, 0).Format("2006-01-02 15:04:05")}
		// state 防并发
		if err = tx.Table(_tablePage).Where("id=?", page.ID).Where("state=?", natmdl.WaitForCheck).Update(upArg).Error; err != nil {
			log.Error("_tablePage d.DB.Table(%d,%v) error(%v)", page.ID, upArg, err)
			return
		}
	} else {
		upArg = map[string]interface{}{"share_title": shareTitle, "ver": ver, "share_image": shareImage}
		// state 防并发
		if err = tx.Table(_tablePage).Where("id=?", page.ID).Update(upArg).Error; err != nil {
			log.Error("_tablePage d.DB.Table(%d,%v) error(%v)", page.ID, upArg, err)
			return
		}
	}
	// 更新ts_module
	if err = tx.Table(_tsPage).Where("id=?", tsid).Update(map[string]interface{}{"state": natmdl.TsOnline}).Error; err != nil {
		log.Error("_tsPage d.DB.Table(%d) error(%v)", tsid, err)
	}
	// 更新native_user_tab
	func() {
		userSpace, err := d.UserSpaceByMid(c, page.RelatedUid)
		if err != nil || userSpace == nil {
			return
		}
		// 已绑定其他活动
		if userSpace.PageId != page.ID || userSpace.State != api.USpaceWaitingOnline {
			return
		}
		success := false
		defer func() {
			state := api.USpaceOfflineAuditFail
			if success {
				state = api.USpaceOnline
			}
			_ = d.UpdateUserSpaceState(c, userSpace.Id, page.ID, state, api.USpaceWaitingOnline)
		}()
		success, err = d.UpActivityTab(c, page.RelatedUid, 1, userSpace.Title, page.ID)
		if err != nil {
			return
		}
	}()
	return
}

// CommitModule .
// nolint:gocognit
func (d *Dao) CommitModule(c context.Context, nativeID int64, portion *natmdl.JsonData, ver string) (err error) {
	tx := d.DB.Begin()
	childMap := make(map[string]int, len(portion.Structure.Root.Children))
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Error("CommitModule %v", r)
		}
		if err == nil {
			if err = tx.Commit().Error; err != nil {
				log.Error("commit(%+v)", err)
				return
			}
		} else {
			tx.Rollback()
		}
	}()
	// 为了方便给组件排序
	for idx, value := range portion.Structure.Root.Children {
		childMap[value] = idx
	}
	var (
		ptypes    []int
		moreBases = make([]*natmdl.BaseJson, 0)
	)
	if portion.Base != nil {
		moreBases = append(moreBases, portion.Base)
	}
	if len(portion.MoreBases) > 0 {
		moreBases = append(moreBases, portion.MoreBases...)
	}
	for _, v := range moreBases {
		ptypes = append(ptypes, v.PType)
	}
	ptypes = append(ptypes, portion.PType)
	// 存在的话，先删除组件，再添加
	if err = d.delBegin(c, tx, nativeID, ptypes); err != nil {
		log.Error("[CommitModule] delBegin error(%v)", err)
		return
	}
	for idx, v := range portion.Modules {
		if v != nil {
			var order int
			if order, err = rankModule(idx, childMap); err != nil {
				return
			}
			switch v.Type {
			case "reply":
				tl := &natmdl.ConfReply{}
				if err = json.Unmarshal(v.Config, tl); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := new(natmdl.NatModule)
				//无需导航标题
				mod.ToReply(tl, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(reply) error(%v)", err)
					return
				}
			case "ogv_season": //ogv剧集卡组件
				tl := &natmdl.ConfOgvSeason{}
				if err = json.Unmarshal(v.Config, tl); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := new(natmdl.NatModule)
				mod.ToOgvSeason(tl, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(ogv_season) error(%v)", err)
					return
				}
				switch mod.Category {
				case natmdl.OgvSeasonIDModule: //ogv剧集卡组件-id模式
					// 插入 native_mixture_ext 表
					if err = d.AddVideoMixtureExt(tx, tl.IDs, mod.ID); err != nil {
						return
					}
				}
			case "timeline": //时间轴组件
				tl := &natmdl.ConfTimeline{}
				if err = json.Unmarshal(v.Config, tl); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := new(natmdl.NatModule)
				mod.ToTimeline(tl, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(timeline) error(%v)", err)
					return
				}
				switch mod.Category {
				case natmdl.TimelineIDModule: //时间轴-id模式
					// 插入 native_mixture_ext 表
					if err = d.AddVideoMixtureExt(tx, tl.IDs, mod.ID); err != nil {
						return
					}
				}
			case "navigation":
				navi := &natmdl.Navigation{}
				if err = json.Unmarshal(v.Config, navi); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				if !navi.CheckColor() {
					err = ecode.Error(ecode.RequestErr, "非法color")
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToNavigation(navi, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(navigation) error(%v)", err)
					return
				}
			case "live":
				navi := &natmdl.ConfLive{}
				if err = json.Unmarshal(v.Config, navi); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToLive(navi, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(live) error(%v)", err)
					return
				}
			case "inline_tab": //页面tab组件
				tabTmp := &natmdl.InlineTab{}
				if err = json.Unmarshal(v.Config, tabTmp); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				pids := make(map[int64]string)
				checkTimes := make([]*natmdl.DefTime, 0)
				var rlyIDs []*natmdl.ResourceIDs
				for _, v := range tabTmp.IDs {
					if v.ID > 0 {
						var confStr string
						if v.DisplayType == 1 {
							confObj := &natmdl.ConfSet{
								DT:     v.DisplayType,
								DC:     v.DisplayCondition,
								Stime:  v.Stime,
								Tip:    v.Tip,
								UnLock: v.UnLock,
							}
							var setByte []byte
							if setByte, err = json.Marshal(confObj); err != nil {
								return
							}
							if len(setByte) > natmdl.MaxLen {
								err = ecode.Error(ecode.RequestErr, "json.Marshal非法锁定设置")
								return
							}
							confStr = string(setByte)
						}
						pids[v.ID] = confStr
						tmpIDs := &natmdl.ResourceIDs{ID: v.ID, Type: natmdl.MixturePageType}
						tmpIDs.Content = &natmdl.MixContent{
							Type:        v.Type,
							LocationKey: v.LocationKey,
							SI:          v.SelectImage,
							UnI:         v.UnImage,
							UnSI:        v.UnSelectImage,
							DefType:     v.DefType,
						}
						if v.DefType == api.DefTypeTiming {
							tmpIDs.Content.DStime = v.DStime
							tmpIDs.Content.DEtime = v.DEtime
							checkTimes = append(checkTimes, &natmdl.DefTime{DEtime: v.DEtime, DStime: v.DStime})
						}
						rlyIDs = append(rlyIDs, tmpIDs)
					}
				}
				//定时生效时间check
				if natmdl.DefCheckTime(checkTimes) {
					err = ecode.Error(ecode.RequestErr, "非法开始和结束时间")
					return
				}
				if len(pids) == 0 {
					err = ecode.Error(ecode.RequestErr, "非法pids")
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToInlineTab(tabTmp, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(inline_tab) error(%v)", err)
					return
				}
				//处理pageids
				for k, pv := range pids {
					if err = tx.Table(_tablePage).Where("id =?", k).Where("type=?", natmdl.InLineType).Update(map[string]interface{}{"state": _valid, "operator": "管理员", "conf_set": pv}).Error; err != nil {
						log.Error("[DelPage] tx.update(), ID(%+v) error(%v)", pids, err)
						return
					}
				}
				if err = d.AddVideoMixtureExt(tx, rlyIDs, mod.ID); err != nil {
					return
				}
			case "select":
				selectTmp := &natmdl.Select{}
				if err = json.Unmarshal(v.Config, selectTmp); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				var pids []int64
				var rlyIDs []*natmdl.ResourceIDs
				checkTimes := make([]*natmdl.DefTime, 0)
				for _, v := range selectTmp.IDs {
					if v.ID > 0 {
						pids = append(pids, v.ID)
						tmpIDs := &natmdl.ResourceIDs{ID: v.ID, Type: natmdl.MixturePageType}
						tmpIDs.Content = &natmdl.MixContent{
							Type:        v.Type,
							LocationKey: v.LocationKey,
							DefType:     v.DefType,
						}
						if v.DefType == api.DefTypeTiming {
							tmpIDs.Content.DStime = v.DStime
							tmpIDs.Content.DEtime = v.DEtime
							checkTimes = append(checkTimes, &natmdl.DefTime{DEtime: v.DEtime, DStime: v.DStime})
						}
						rlyIDs = append(rlyIDs, tmpIDs)
					}
				}
				//定时生效时间check
				if natmdl.DefCheckTime(checkTimes) {
					err = ecode.Error(ecode.RequestErr, "非法开始和结束时间")
					return
				}
				if len(pids) == 0 {
					err = ecode.Error(ecode.RequestErr, "非法pids")
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToSelect(selectTmp, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(select) error(%v)", err)
					return
				}
				//处理pageids
				if err = tx.Table(_tablePage).Where("id in (?)", pids).Where("type=?", natmdl.InLineType).Update(map[string]interface{}{"state": _valid, "operator": "管理员"}).Error; err != nil {
					log.Error("[DelPage] tx.update(), ID(%d) error(%v)", pids, err)
					return
				}
				if err = d.AddVideoMixtureExt(tx, rlyIDs, mod.ID); err != nil {
					return
				}
			case "banner":
				banner := &natmdl.BannerImage{}
				if err = json.Unmarshal(v.Config, banner); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToBanner(banner, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(banner) error(%v)", err)
					return
				}
			case "statement":
				banner := &natmdl.StatementModule{}
				if err = json.Unmarshal(v.Config, banner); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToStatement(banner, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(statement) error(%v)", err)
					return
				}
			case "single-dynamic":
				mod := &natmdl.NatModule{}
				mod.ToSingleDynamic(nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(single-dynamic) error(%v)", err)
					return
				}
			case "vote":
				vot := &natmdl.ConfVote{}
				mod := &natmdl.NatModule{}
				if err = json.Unmarshal(v.Config, vot); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod.ToMVote(vot, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(mVote) error(%v)", err)
					return
				}
				// native_click表插入
				for _, area := range vot.Areas {
					cliData := &natmdl.Click{}
					switch {
					case area.IsTypeVoteButton(), area.IsTypeVoteProgress(), area.IsTypeVoteUser():
					default:
						continue
					}
					cliData.ToVote(mod.ID, area)
					if err = tx.Create(cliData).Error; err != nil {
						log.Error("[SaveModule] tx.Create(mVote) error(%v)", err)
						return
					}
				}
			case "click":
				click := &natmdl.ConfClick{}
				mod := &natmdl.NatModule{}
				if err = json.Unmarshal(v.Config, click); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod.ToMclick(click, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(mClick) error(%v)", err)
					return
				}
				// native_click表插入
				for _, area := range click.Areas {
					cliData := &natmdl.Click{}
					cliData.ToClick(mod.ID, area, click.Width, click.Height)
					if err = tx.Create(cliData).Error; err != nil {
						log.Error("[SaveModule] tx.Create(mClick) error(%v)", err)
						return
					}
				}
			case "act":
				act := &natmdl.ConfAct{}
				mod := &natmdl.NatModule{}
				if err = json.Unmarshal(v.Config, act); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod.ToMact(act, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(act) error(%v)", err)
					return
				}
				// native_act插入
				for idx, v := range act.Acts {
					act := &natmdl.Act{}
					act.ToAct(mod.ID, v, idx)
					if err = tx.Create(act).Error; err != nil {
						log.Error("tx.Save(act) error(%v)", err)
						return
					}
				}
			case "act_capsule":
				actConf := &natmdl.ConfActCapsule{}
				if err = json.Unmarshal(v.Config, actConf); err != nil {
					log.Error("Fail to unmarshal act_capsule config, conf=%+v error=%+v", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToActCapsule(actConf, nativeID, order, portion.PType, v.ID)
				if err = tx.Create(mod).Error; err != nil {
					log.Error("Fail to create act_capsule module, conf=%+v error=%+v", v.Config, err)
					return
				}
				for i, v := range actConf.Acts {
					act := &natmdl.Act{}
					act.ToAct(mod.ID, v, i)
					if err = tx.Create(act).Error; err != nil {
						log.Error("Fail to create act_capsule items, act=%+v error=%+v", v, err)
						return
					}
				}
			case "dynamic":
				dy := &natmdl.ConfDynamic{}
				mod := &natmdl.NatModule{}
				if err = json.Unmarshal(v.Config, dy); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod.ToMdynamic(dy, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(Dynamic) error(%v)", err)
					return
				}
				// native_dynamic_ext插入
				if dy.Pattern == 0 {
					// 非精选类型
					if !isDynamicChoice(dy.SelectType) {
						//dy.SelectType 多选 按照逗号分隔
						var selectInts []int64
						selectArray := strings.Split(dy.SelectType, ",")
						for _, v := range selectArray {
							var tempInt int64
							if tempInt, err = strconv.ParseInt(v, 10, 64); err != nil {
								log.Error("tx.Save(strconv.ParseInt) error(%v)", err)
								return
							}
							if tempInt > 0 {
								selectInts = append(selectInts, tempInt)
							} else if tempInt == 0 {
								// 0表示全部，选择全部时，不支持其余多选模式
								selectInts = []int64{0}
								break
							}
						}
						if len(selectInts) == 0 {
							err = ecode.Error(ecode.RequestErr, "dynamic select type is empty")
							return
						}
						for _, val := range selectInts {
							dynamicExt := &natmdl.DynamicExt{}
							dynamicExt.ToDynamicExt(val, mod.ID, 0, false)
							if err = tx.Create(dynamicExt).Error; err != nil {
								log.Error("tx.Save(dynamicExt) error(%v)", err)
								return
							}
						}
					} else {
						dynamicExt := &natmdl.DynamicExt{}
						dynamicExt.ToDynamicExt(0, mod.ID, dy.ClassID, true)
						if err = tx.Create(dynamicExt).Error; err != nil {
							log.Error("tx.Save(dynamicExt) error(%v)", err)
							return
						}
					}
				}
				// native_video_ext插入
				if dy.Pattern == 1 {
					// 这里要求不同的顺序存多条记录
					var typeRank []int64
					if typeRank, err = xstr.SplitInts(dy.SortType); err != nil {
						log.Error("xstr.SplitInts error(%v)", err)
						return
					}
					for idx, v := range typeRank {
						videoExt := natmdl.VideoExt{}
						videoExt.ToVideoExt(mod.ID, v, idx, 0, "")
						if err = tx.Create(videoExt).Error; err != nil {
							log.Error("tx.Save(videoExt) error(%v)", err)
							return
						}
					}
				}
			case "archive", "new-archive":
				arc := new(natmdl.ConfArchive)
				if err = json.Unmarshal(v.Config, arc); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := new(natmdl.NatModule)
				mod.ToMArchive(arc, nativeID, order, portion.PType, v.ID)
				if mod.Category == natmdl.VideoDynModule || mod.Category == natmdl.NewVideoDynModule {
					// 获取话题ID
					if mod.FID, err = d.NatTagID(c, arc.TName); err != nil {
						return
					}
				}
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(Dynamic) error(%v)", err)
					return
				}
				switch mod.Category {
				case natmdl.NewVideoAvidModule: //新视频卡-id模式
					//插入 native_mixture_ext 表
					var IDs []*natmdl.ResourceIDs
					for _, v := range arc.IDs {
						if v.Type == natmdl.MixtureArcType && v.Bvid != "" {
							var aid int64
							if aid, err = bvid.BvToAv(v.Bvid); err != nil {
								log.Error("bvid.BvToAv %+v error(%v)", v, err)
								return
							}
							v.ID = aid
						}
						IDs = append(IDs, v)
					}
					if err = d.batchAddMixtureExt(tx, mod.ID, IDs, nil); err != nil {
						return
					}
				case natmdl.VideoAvidModule: //老视频卡-id模式
					// 插入 native_mixture_ext 表
					var aids []int64
					if len(arc.Bvids) > 0 {
						for _, v := range arc.Bvids {
							var aid int64
							if aid, err = bvid.BvToAv(v); err != nil {
								log.Error("bvid.BvToAv %s error(%v)", v, err)
								return
							}
							aids = append(aids, aid)
						}
					} else {
						aids = arc.ObjIDs
					}
					var IDs []*natmdl.ResourceIDs
					for _, v := range aids {
						IDs = append(IDs, &natmdl.ResourceIDs{ID: v, Type: natmdl.MixtureArcType})
					}
					if err = d.batchAddMixtureExt(tx, mod.ID, IDs, nil); err != nil {
						return
					}
				case natmdl.VideoActModule, natmdl.NewVideoActModule: //新老视频开act模式
					// 插入 native_video_ext 表
					var typeRank []int64
					if typeRank, err = xstr.SplitInts(arc.SortType); err != nil {
						log.Error("xstr.SplitInts error(%v)", err)
						return
					}
					for idx, v := range typeRank {
						videoExt := natmdl.VideoExt{}
						videoExt.ToVideoExt(mod.ID, v, idx, 0, "")
						if err = tx.Create(videoExt).Error; err != nil {
							log.Error("tx.Save(videoExt) error(%v)", err)
							return
						}
					}
				}
			case "resource":
				arc := new(natmdl.ConfResource)
				if err = json.Unmarshal(v.Config, arc); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := new(natmdl.NatModule)
				mod.ToMResource(arc, nativeID, order, portion.PType, v.ID)
				if mod.Category == natmdl.ResourceDynamicModule {
					// 获取话题ID
					if mod.FID, err = d.NatTagID(c, arc.TName); err != nil {
						return
					}
				}
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(Dynamic) error(%v)", err)
					return
				}
				switch mod.Category {
				case natmdl.ResourceIDModule:
					// 插入 native_mixture_ext 表
					if err = d.AddVideoMixtureExt(tx, arc.IDs, mod.ID); err != nil {
						return
					}
				case natmdl.ResourceDataOriginModule:
					if arc.RDBType == api.RDBLive {
						for idx, v := range arc.SortList {
							videoExt := natmdl.VideoExt{}
							videoExt.ToVideoExt(mod.ID, v.SortType, idx, v.Category, v.SortName)
							if err = tx.Create(videoExt).Error; err != nil {
								log.Error("tx.Save(videoExt) error(%v)", err)
								return
							}
						}
					}
				case natmdl.ResourceActModule:
					// 插入 native_video_ext 表
					var typeRank []int64
					if typeRank, err = xstr.SplitInts(arc.SortType); err != nil {
						log.Error("xstr.SplitInts error(%v)", err)
						return
					}
					for idx, v := range typeRank {
						videoExt := natmdl.VideoExt{}
						videoExt.ToVideoExt(mod.ID, v, idx, 0, "")
						if err = tx.Create(videoExt).Error; err != nil {
							log.Error("tx.Save(videoExt) error(%v)", err)
							return
						}
					}
				case natmdl.ResourceDynamicModule:
					dynamicExt := &natmdl.DynamicExt{}
					var tempInt int64
					if tempInt, err = strconv.ParseInt(arc.DynStyle, 10, 64); err != nil {
						log.Error("tx.Save(strconv.ParseInt) error(%v)", err)
						return
					}
					dynamicExt.ToDynamicExt(tempInt, mod.ID, 0, false)
					if err = tx.Create(dynamicExt).Error; err != nil {
						log.Error("tx.Save(dynamicExt) error(%v)", err)
						return
					}
				}
			case "reserve":
				rev := new(natmdl.ConfReserve)
				if err = json.Unmarshal(v.Config, rev); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				if len(rev.Sids) == 0 {
					continue
				}
				mod := &natmdl.NatModule{}
				mod.ToMReserve(rev, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(game) error(%v)", err)
					return
				}
				var rlyIDs []*natmdl.ResourceIDs
				for _, v := range rev.Sids {
					if v <= 0 {
						continue
					}
					tmpIDs := &natmdl.ResourceIDs{ID: v, Type: api.MixUpReserve}
					rlyIDs = append(rlyIDs, tmpIDs)
				}
				if len(rlyIDs) == 0 {
					continue
				}
				if err = d.AddVideoMixtureExt(tx, rlyIDs, mod.ID); err != nil {
					return
				}
			case "game":
				game := new(natmdl.ConfGame)
				if err = json.Unmarshal(v.Config, game); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				if len(game.Games) == 0 {
					continue
				}
				mod := &natmdl.NatModule{}
				mod.ToMGame(game, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(game) error(%v)", err)
					return
				}
				var rlyIDs []*natmdl.ResourceIDs
				for _, v := range game.Games {
					if v.GameID <= 0 {
						continue
					}
					tmpIDs := &natmdl.ResourceIDs{ID: v.GameID, Type: api.MixGame}
					if v.Content != "" {
						tmpIDs.Content = &natmdl.MixContent{Desc: v.Content}
					}
					rlyIDs = append(rlyIDs, tmpIDs)
				}
				if len(rlyIDs) == 0 {
					continue
				}
				if err = d.AddVideoMixtureExt(tx, rlyIDs, mod.ID); err != nil {
					return
				}
			case "recommend":
				recommend := new(natmdl.ConfRecommned)
				if err = json.Unmarshal(v.Config, recommend); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToMRecommend(recommend, nativeID, order, portion.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(Recomment) error(%v)", err)
					return
				}
				if recommend.Category == natmdl.RcmdSourceModule && recommend.SourceType == api.SourceTypeRank {
					// 插入 native_mixture_ext 表
					if err = d.AddVideoMixtureExt(tx, recommend.IDs, mod.ID); err != nil {
						return
					}
				} else {
					userMax := len(recommend.RecoUsers)
					if userMax == 0 {
						continue
					}
					//竖卡组件记数逻辑与横卡保持一致，一个组件算一个卡片，需要限定数量
					if userMax > _userMax {
						err = ecode.Errorf(ecode.RequestErr, "推荐用户单组件数量不能超过%d", _userMax)
						return
					}
					var (
						IDs     = make([]*natmdl.ResourceIDs, 0)
						reasons = make([]string, 0)
					)
					for _, v := range recommend.RecoUsers {
						IDs = append(IDs, &natmdl.ResourceIDs{ID: v.MID, Type: natmdl.MixtureUpType})
						reasons = append(reasons, v.Content)
					}
					if err = d.batchAddMixtureExt(tx, mod.ID, IDs, reasons); err != nil {
						return
					}
				}
			case "recommend_vertical":
				rcmd := new(natmdl.ConfRecommned)
				if err = json.Unmarshal(v.Config, rcmd); err != nil {
					log.Error("Fail to unmarshal ConfRecommned, ConfRecommned=%s error=%+v", v.Config, err)
					return
				}
				module := &natmdl.NatModule{}
				module.ToMRcmdVertical(rcmd, nativeID, order, portion.PType, v.ID)
				if err = tx.Create(module).Error; err != nil {
					log.Error("Fail to create native_module, error=%+v", err)
					return
				}
				userMax := len(rcmd.RecoUsers)
				//组件内卡片数量为0,不保存组件
				if userMax == 0 {
					continue
				}
				//竖卡组件记数逻辑与横卡保持一致，一个组件算一个卡片，需要限定数量
				if userMax > _userMax {
					err = ecode.Errorf(ecode.RequestErr, "推荐用户单组件数量不能超过%d", _userMax)
					return
				}
				var (
					IDs     = make([]*natmdl.ResourceIDs, 0, len(rcmd.RecoUsers))
					reasons = make([]string, 0, len(rcmd.RecoUsers))
				)
				for _, v := range rcmd.RecoUsers {
					IDs = append(IDs, &natmdl.ResourceIDs{ID: v.MID, Type: natmdl.MixtureRcmdVertical})
					reasons = append(reasons, buildRcmdVerticalExt(v))
				}
				if err = d.batchAddMixtureExt(tx, module.ID, IDs, reasons); err != nil {
					log.Error("Fail to batch create native_mixture_ext, mou_id=%d ids=%+v reasons=%+v error=%+v", module.ID, IDs, reasons, err)
					return
				}
			case "editor":
				editor := new(natmdl.ConfEditor)
				if err = json.Unmarshal(v.Config, editor); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				// 插入 native_module 表
				mod := &natmdl.NatModule{}
				mod.ToEditor(editor, nativeID, order, portion.PType, v.ID)
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(Editor) error(%v)", err)
					return
				}
				// 插入 native_mixture_ext 表
				if err = d.AddVideoMixtureExt(tx, editor.IDs, mod.ID); err != nil {
					return
				}
			case "carousel":
				carousel := new(natmdl.ConfCarousel)
				if err = json.Unmarshal(v.Config, carousel); err != nil {
					log.Error("Fail to unmarshal carousel, config=%s error=%+v", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToMCarousel(carousel, nativeID, order, portion.PType, v.ID)
				if err = tx.Create(mod).Error; err != nil {
					log.Error("Fail to create carousel, mod=%+v error=%+v", mod, err)
					return
				}
				if err = d.AddCarouselMixtureExt(tx, carousel, mod.ID); err != nil {
					log.Error("Fail to add carouselMixtureExt, carousel=%+v moduleID=%d err=%+v", carousel, mod.ID, err)
					return
				}
			case "icon":
				icon := new(natmdl.ConfIcon)
				if err = json.Unmarshal(v.Config, icon); err != nil {
					log.Error("Fail to unmarshal icon, config=%s error=%+v", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToMIcon(icon, nativeID, order, portion.PType, v.ID)
				if err = tx.Create(mod).Error; err != nil {
					log.Error("Fail to create icon, mod=%+v error=%+v", mod, err)
					return
				}
				if err = d.AddIconMixtureExt(tx, icon, mod.ID); err != nil {
					log.Error("Fail to add iconMixtureExt, icon=%+v moduleID=%d err=%+v", icon, mod.ID, err)
					return
				}
			case "progress":
				progress := new(natmdl.ConfProgress)
				if err = json.Unmarshal(v.Config, progress); err != nil {
					log.Error("Fail to unmarshal progress, config=%s error=%+v", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToProgress(progress, nativeID, order, portion.PType, v.ID)
				if err = tx.Create(mod).Error; err != nil {
					log.Error("Fail to create process, mod=%+v error=%+v", mod, err)
					return
				}
			case "match_medal":
				medal := new(natmdl.ConfMatchMedal)
				if err = json.Unmarshal(v.Config, medal); err != nil {
					log.Error("Fail to unmarshal ConfMatchMedal, config=%s error=%+v", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToMatchMedal(medal, nativeID, order, portion.PType, v.ID)
				if err = tx.Create(mod).Error; err != nil {
					log.Error("Fail to create match_medal, mod=%+v error=%+v", mod, err)
					return
				}
			case "match_event":
				event := new(natmdl.ConfMatchEvent)
				if err = json.Unmarshal(v.Config, event); err != nil {
					log.Error("Fail to unmarshal ConfMatchEvent, config=%s error=%+v", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToMatchEvent(event, nativeID, order, portion.PType, v.ID)
				if err = tx.Create(mod).Error; err != nil {
					log.Error("Fail to create match_event, mod=%+v error=%+v", mod, err)
					return
				}
				var ids []*natmdl.ResourceIDs
				for _, id := range event.EventIds {
					if id <= 0 {
						continue
					}
					ids = append(ids, &natmdl.ResourceIDs{ID: id, Type: api.MixMatchEvent})
				}
				if err = d.batchAddMixtureExt(tx, mod.ID, ids, nil); err != nil {
					return
				}
			}
		}
	}
	upPage := make(map[string]interface{})
	for _, ba := range moreBases {
		if ba == nil {
			continue
		}
		baMap := make(map[string]int, len(ba.Children))
		for idx, value := range ba.Children {
			baMap[value] = idx
		}
		for idx, v := range ba.Modules {
			if v == nil {
				continue
			}
			var order int
			if order, err = rankModule(idx, baMap); err != nil {
				return
			}
			switch v.Type {
			case "head":
				head := &natmdl.HeadBase{}
				if err = json.Unmarshal(v.Config, head); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToHead(head, nativeID, order, ba.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(head) error(%v)", err)
					return
				}
			case "participation":
				part := new(natmdl.ConfParticipation)
				if err = json.Unmarshal(v.Config, part); err != nil {
					log.Error("[Module] json.Unmarshal() json(%s) error(%v)", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToMPart(part, nativeID, order, ba.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("[SaveModule] tx.Create(Participation) error(%v)", err)
					return
				}
				for k, button := range part.Button {
					button.ToParticipation(mod.ID, k)
					if err = tx.Create(button).Error; err != nil {
						log.Error("tx.Save(Participation) error(%v)", err)
						return
					}
				}
			case "hover_button":
				hover := new(natmdl.ConfHoverButton)
				if err = json.Unmarshal(v.Config, hover); err != nil {
					log.Error("Fail to unmarshal ConfHoverButton, config=%s error=%+v", v.Config, err)
					return
				}
				mod := &natmdl.NatModule{}
				mod.ToHoverButton(hover, nativeID, order, ba.PType, v.ID)
				if err = tx.Create(mod).Error; err != nil {
					log.Error("Fail to create hover_button, mod=%+v error=%+v", mod, err)
					return
				}
			case "bottom_button":
				btn := &natmdl.ConfBottomButton{}
				mod := &natmdl.NatModule{}
				if err = json.Unmarshal(v.Config, btn); err != nil {
					log.Error("Fail to unmarshal ConfBottomButton, config=%s error=%+v", v.Config, err)
					return
				}
				mod.ToMbottomButton(btn, nativeID, order, ba.PType, v.ID)
				// 组件表插入 native_module
				if err = tx.Create(mod).Error; err != nil {
					log.Error("Fail to create BottomButton, mod=%+v error=%+v", mod, err)
					return
				}
				// native_click表插入
				for _, area := range btn.Areas {
					cliData := &natmdl.Click{}
					cliData.ToClick(mod.ID, area, btn.Width, btn.Height)
					if err = tx.Create(cliData).Error; err != nil {
						log.Error("Fail to create BottomButton Area, area=%+v error=%+v", area, err)
						return
					}
				}
			default:
				continue
			}
		}
	}
	//更新版本号
	upPage["ver"] = ver
	if err = tx.Table(_tablePage).Where("id=?", nativeID).Update(upPage).Error; err != nil {
		log.Error("_tablePage d.DB.Table(%d) error(%v)", nativeID, err)
	}
	return
}

func (d *Dao) AddVideoMixtureExt(tx *gorm.DB, ids []*natmdl.ResourceIDs, modID int64) error {
	var (
		err     error
		IDs     []*natmdl.ResourceIDs
		reasons []string
	)
	for _, v := range ids {
		var reason string
		if (v.Type == natmdl.MixtureArcType || v.Type == natmdl.MixtureFolder) && v.Bvid != "" {
			var aid int64
			if aid, err = bvid.BvToAv(v.Bvid); err != nil {
				log.Error("bvid.BvToAv %+v error(%v)", v, err)
				return err
			}
			v.ID = aid
		}
		if v.Type == natmdl.MixtureFolder || v.RcmdContent != nil {
			var tmp []byte
			folder := &natmdl.MixFolder{}
			if v.Type == natmdl.MixtureFolder {
				folder.Fid = v.Fid
			}
			if v.RcmdContent != nil {
				folder.RcmdContent = v.RcmdContent
			}
			if tmp, err = json.Marshal(folder); err != nil {
				log.Error("json.Marshal(%+v) error(%+v)", folder, err)
				return err
			}
			reason = string(tmp)
		} else if v.Content != nil {
			var tlmp []byte
			if tlmp, err = json.Marshal(v.Content); err != nil {
				log.Error("json.Marshal(%+v) error(%+v)", v.Content, err)
				return err
			}
			reason = string(tlmp)
		}
		//超过数据库大小限制
		if len([]byte(reason)) > natmdl.MaxLen {
			return errors.Wrapf(err, "native_mixture_ext.reason 超过2000")
		}
		IDs = append(IDs, v)
		reasons = append(reasons, reason)
	}
	if err = d.batchAddMixtureExt(tx, modID, IDs, reasons); err != nil {
		return err
	}
	return nil
}

func (d *Dao) AddCarouselMixtureExt(tx *gorm.DB, carousel *natmdl.ConfCarousel, moduleID int64) error {
	if carousel == nil {
		return nil
	}
	var (
		ids     []*natmdl.ResourceIDs
		reasons []string
	)
	switch carousel.Category {
	case natmdl.CarouselImgModule:
		ids = make([]*natmdl.ResourceIDs, 0, len(carousel.ImgList))
		reasons = make([]string, 0, len(carousel.ImgList))
		for _, v := range carousel.ImgList {
			if v == nil {
				continue
			}
			imgItem, err := json.Marshal(v)
			if err != nil {
				log.Error("Fail to marshal imgItem, imgItem=%+v error=%+v", v, err)
				continue
			}
			//超过数据库限制
			if len(imgItem) > natmdl.MaxLen {
				continue
			}
			ids = append(ids, &natmdl.ResourceIDs{Type: natmdl.MixtureCarouselImg})
			reasons = append(reasons, string(imgItem))
		}
	case natmdl.CarouselWordModule:
		ids = make([]*natmdl.ResourceIDs, 0, len(carousel.WordList))
		reasons = make([]string, 0, len(carousel.WordList))
		for _, v := range carousel.WordList {
			if v == nil {
				continue
			}
			wordItem, err := json.Marshal(v)
			if err != nil {
				log.Error("Fail to marshal wordItem, wordItem=%+v error=%+v", v, err)
				continue
			}
			// 超过数据库限制
			if len(wordItem) > natmdl.MaxLen {
				continue
			}
			ids = append(ids, &natmdl.ResourceIDs{Type: natmdl.MixtureCarouselWord})
			reasons = append(reasons, string(wordItem))
		}
	case natmdl.CarouselSourceModule:
		return nil
	default:
		log.Warn("unexpected category of carousel, category=%d", carousel.Category)
	}
	if err := d.batchAddMixtureExt(tx, moduleID, ids, reasons); err != nil {
		return err
	}
	return nil
}

func (d *Dao) AddIconMixtureExt(tx *gorm.DB, icon *natmdl.ConfIcon, moduleID int64) error {
	if icon == nil || len(icon.ImgList) == 0 {
		return nil
	}
	ids := make([]*natmdl.ResourceIDs, 0, len(icon.ImgList))
	reasons := make([]string, 0, len(icon.ImgList))
	for _, v := range icon.ImgList {
		if v == nil {
			continue
		}
		imgItem, err := json.Marshal(v)
		if err != nil {
			log.Error("Fail to marshal imgItem, wordItem=%+v error=%+v", v, err)
			continue
		}
		ids = append(ids, &natmdl.ResourceIDs{Type: natmdl.MixtureIconImg})
		reasons = append(reasons, string(imgItem))
	}
	if err := d.batchAddMixtureExt(tx, moduleID, ids, reasons); err != nil {
		return err
	}
	return nil
}

// SaveModule .
func (d *Dao) SaveModule(c context.Context, nativeID int64, portion *natmdl.JsonData, ver string) (err error) {
	var (
		pageInfo *natmdl.FindRes
	)
	if pageInfo, err = d.PageByID(c, nativeID); err != nil || pageInfo == nil {
		log.Error("[SaveModule] d.PageByID() NativeID(%d) error(%v)", nativeID, err)
		return
	}
	//审核态，草稿态不支持编辑up主发起活动
	if pageInfo.State == natmdl.WaitForCheck || pageInfo.State == natmdl.CheckOffline || pageInfo.State == natmdl.WaitForCommit {
		err = ecode.Error(ecode.RequestErr, "审核态，草稿态不支持编辑up主发起活动")
		return
	}
	return d.CommitModule(c, nativeID, portion, ver)
}

// SearchModule .
func (d *Dao) SearchModule(c context.Context, param *natmdl.SearchModule) (res *natmdl.ModuleRes, err error) {
	var (
		module []*natmdl.ModuleData
		db     = d.DB.Table(_module)
		mAll   = &natmdl.ModuleAll{}
	)
	res = &natmdl.ModuleRes{}
	if param.ID != 0 {
		// 该页面下的所有组件
		db = db.Where("native_id=?", param.ID).Where("state=?", _valid)
	} else if param.ModuleID != 0 {
		// 查询指定组件组件
		db = db.Where("id=?", param.ModuleID).Where("state=?", _valid)
	}
	if err = db.Find(&module).Error; err != nil {
		log.Error("[SearchModule] error(%v)", err)
		return
	}
	for _, v := range module {
		switch v.Category {
		case natmdl.ClickModule: // 自定义点击组件
			var (
				mClick = &natmdl.ModuleCli{}
				cli    []*natmdl.Click
			)
			if err = d.DB.Table(_cilck).Where("module_id=?", v.ID).Where("state=?", _valid).Find(&cli).Error; err != nil {
				log.Error("[SearchModule] d.DB.Table(_cilck) error(%v)", err)
				return
			}
			mClick.ModuleData = v
			mClick.Cli = cli
			mAll.Click = append(mAll.Click, mClick)
		case natmdl.DynmaicModule: // 动态模式
			var (
				mDy   = &natmdl.ModuleDy{}
				dyExt []*natmdl.DynamicExt
			)
			if err = d.DB.Table(_dynamicExt).Where("module_id=?", v.ID).Where("state=?", _valid).Find(&dyExt).Error; err != nil {
				log.Error("SearchModule] d.DB.Table(_dynamicExt) error(%v)", err)
				return
			}
			mDy.ModuleData = v
			mDy.Dy = dyExt
			mAll.DynamicExt = append(mAll.DynamicExt, mDy)
		case natmdl.VideoModule: // 视频模式
			var (
				mVideo   = &natmdl.ModuleVideo{}
				videoExt []*natmdl.VideoExt
			)
			if err = d.DB.Table(_videoExt).Where("module_id=?", v.ID).Where("state=?", _valid).Find(&videoExt).Error; err != nil {
				log.Error("SearchModule] d.DB.Table(_videoExt) error(%v)", err)
				return
			}
			mVideo.ModuleData = v
			mVideo.Video = videoExt
			mAll.VideoExt = append(mAll.VideoExt, mVideo)
		case natmdl.ActModule: // 相关活动
			var (
				mAct = &natmdl.ModuleAct{}
				act  []*natmdl.Act
			)
			if err = d.DB.Table(_act).Where("module_id=?", v.ID).Where("state=?", _valid).Find(&act).Error; err != nil {
				log.Error("SearchModule] d.DB.Table(_act) error(%v)", err)
				return
			}
			mAct.ModuleData = v
			mAct.Act = act
			mAll.Act = append(mAll.Act, mAct)
		case natmdl.ParticipationModule: // 参与用户组件
			var (
				mPart = &natmdl.ModuleParticipation{}
				part  []*natmdl.ParticipationExt
			)
			if err = d.DB.Table(_participationExt).Where("module_id=?", v.ID).Where("state=?", _valid).Find(&part).Error; err != nil {
				log.Error("SearchModule] d.DB.Table(_participationExt) error(%v)", err)
				return
			}
			mPart.ModuleData = v
			mPart.Part = part
			mAll.ParticipationExt = append(mAll.ParticipationExt, mPart)
		}
	}
	res.Item = mAll
	return
}

func (d *Dao) batchAddMixtureExt(db *gorm.DB, moduleID int64, IDs []*natmdl.ResourceIDs, reasons []string) (err error) {
	var (
		addSQLs []string
		addArgs []interface{}
	)
	if len(IDs) == 0 {
		return
	}
	for i, v := range IDs {
		addSQLs = append(addSQLs, "(?,?,?,?,?,?)")
		addArgs = append(addArgs, moduleID, v.ID, _valid, i+1, v.Type)
		if reasons == nil {
			addArgs = append(addArgs, "")
		} else {
			addArgs = append(addArgs, reasons[i])
		}
	}
	if err = db.Exec(fmt.Sprintf(_mixExtBatchAddSQL, strings.Join(addSQLs, ",")), addArgs...).Error; err != nil {
		log.Error("BatchAddMixtureExt db.Exec %v error(%v)", IDs, err)
	}
	return
}

func (d *Dao) batchAddActPage(db *gorm.DB, moduleID int64, pageIDs []int64) error {
	if len(pageIDs) == 0 {
		return nil
	}
	sqls := make([]string, 0, len(pageIDs))
	args := make([]interface{}, 0, len(pageIDs))
	for i, pid := range pageIDs {
		sqls = append(sqls, "(?,?,?,?)")
		args = append(args, moduleID, 1, i, pid)
	}
	if err := db.Exec(fmt.Sprintf(_batchAddActSQL, strings.Join(sqls, ",")), args...).Error; err != nil {
		log.Error("Fail to batch add native_act, pageIDs=%+v error=%+v", pageIDs, err)
		return err
	}
	return nil
}

func (d *Dao) getTsModResourceFromModules(c context.Context, modules []*natmdl.NatTsModule) (map[int64][]*natmdl.NatTsModuleResource, error) {
	modIDs := make([]int64, 0, len(modules))
	for _, v := range modules {
		if v.Category == natmdl.ResourceIDModule || v.Category == natmdl.NewVideoAvidModule || v.Category == natmdl.ActCapsuleModule ||
			v.Category == natmdl.RecommentModule || v.Category == natmdl.CarouselImgModule {
			modIDs = append(modIDs, v.ID)
		}
	}
	return d.TsModResources(c, modIDs)
}

func (d *Dao) AddPageFromTopicUpg(c context.Context, topic string, topicID, attribute int64, fromType int) (int64, error) {
	page := &natmdl.PageParam{
		Title:        topic,
		Creator:      "system",
		Operator:     "system",
		Type:         api.TopicActType,
		ForeignID:    topicID,
		FromType:     fromType,
		State:        api.WaitForOnline,
		ShareImage:   _defaultShareImage,
		ShareCaption: topic,
		Attribute:    attribute,
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
			log.Error("Fail to add page from topic_upg, topic=%+v panic=%+v", topic, buf)
			return
		}
		if err != nil {
			if err1 := tx.Rollback().Error; err1 != nil {
				log.Error("Fail to rollback, topic=%+v error=%+v", topic, err1)
			}
			return
		}
		if err = tx.Commit().Error; err != nil {
			log.Error("Fail to commit, topic=%+v error=%+v", topic, err)
			return
		}
	}()
	headCfg := &natmdl.HeadBase{}
	if err = d.addHeadModule(tx, pageID, headCfg); err != nil {
		return pageID, err
	}
	partCfg := &natmdl.ConfParticipation{Button: []*natmdl.ParticipationExt{{MType: 0, Title: "参与话题讨论"}}}
	if err = d.addParticipation(tx, pageID, partCfg); err != nil {
		return pageID, err
	}
	dyCfg := &natmdl.ConfDynamic{DySort: 2, Attribute: 1, SourceID: page.ForeignID}
	if err = d.addDynamic(tx, pageID, 1, dyCfg); err != nil {
		return pageID, err
	}
	pageAttrs := map[string]interface{}{
		"pc_url": fmt.Sprintf("https://www.bilibili.com/blackboard/dynamic/%d", pageID),
		"ver":    buildVer("admin"),
		"state":  natmdl.OnlineState,
		"stime":  time.Now().Format("2006-01-02 15:04:05"),
		"etime":  time.Unix(2147356800, 0).Format("2006-01-02 15:04:05"),
	}
	if err = tx.Table(_tablePage).Where("id=?", pageID).Where("state=?", natmdl.WaitForOnline).Update(pageAttrs).Error; err != nil {
		log.Error("Fail to update native_page, id=%+v attrs=%+v error=%+v", pageID, pageAttrs, err)
		return pageID, err
	}
	return pageID, nil
}

func (d *Dao) addHeadModule(tx *gorm.DB, pageID int64, cfg *natmdl.HeadBase) error {
	if cfg == nil || tx == nil {
		return nil
	}
	module := &natmdl.NatModule{}
	module.ToHead(cfg, pageID, 1, api.CommonBaseModule, generateUkey(natmdl.UkeyPrefixUpgrade))
	if err := tx.Create(module).Error; err != nil {
		log.Error("Fail to save head module, pageID=%+v error=%+v", pageID, err)
		return err
	}
	return nil
}

func (d *Dao) addParticipation(tx *gorm.DB, pageID int64, cfg *natmdl.ConfParticipation) error {
	if cfg == nil || tx == nil {
		return nil
	}
	module := &natmdl.NatModule{}
	module.ToMPart(cfg, pageID, 1, api.BasePage, generateUkey(natmdl.UkeyPrefixUpgrade))
	if err := tx.Create(module).Error; err != nil {
		log.Error("Fail to save participation module, pageID=%+v error=%+v", pageID, err)
		return err
	}
	for k, button := range cfg.Button {
		if button == nil {
			continue
		}
		button.ToParticipation(module.ID, k)
		if err := tx.Create(button).Error; err != nil {
			log.Error("Fail to save participation button, pageID=%+v moduleID=%+v error=%+v", pageID, module.ID, err)
			return err
		}
	}
	return nil
}

func (d *Dao) addDynamic(tx *gorm.DB, pageID int64, order int, cfg *natmdl.ConfDynamic) error {
	module := &natmdl.NatModule{}
	module.ToMdynamic(cfg, pageID, order, api.CommonPage, generateUkey(natmdl.UkeyPrefixUpgrade))
	if err := tx.Create(module).Error; err != nil {
		log.Error("Fail to save dynamic module, pageID=%+v error=%+v", pageID, err)
		return err
	}
	// 0表示全部，选择全部时，不支持其余多选模式
	dynamicExt := &natmdl.DynamicExt{}
	dynamicExt.ToDynamicExt(0, module.ID, 0, false)
	if err := tx.Create(dynamicExt).Error; err != nil {
		log.Error("Fail to save dynamic_ext, pageID=%+v moduleID=%+v error=%+v", pageID, module.ID, err)
		return err
	}
	return nil
}

func buildRcmdVerticalExt(user *natmdl.RecoUser) string {
	if user == nil {
		return ""
	}
	rawRcmdExt := &struct {
		Reason string `json:"reason"` //推荐理由
		URI    string `json:"uri"`    //链接
	}{
		Reason: user.Content,
		URI:    user.URI,
	}
	rcmdExt, err := json.Marshal(rawRcmdExt)
	if err != nil {
		log.Error("Fail to marshal rawRcmdExt, rawRcmdExt=%+v error=%+v", rawRcmdExt, err)
		return ""
	}
	return string(rcmdExt)
}

func buildVer(from string) string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano()/1e6, from)
}

func buildShareImage(upShare, pageShare string) string {
	if upShare != "" {
		return upShare
	}
	if pageShare != "" {
		return pageShare
	}
	return "https://i0.hdslb.com/bfs/activity-plat/static/8347b7383c4a730528a82854f98b9b32/sYbPL4QDx9.png"
}

func (d *Dao) FindPageByID(c context.Context, tagID, id, defaulType int64, states []int64) (*natmdl.NatPage, error) {
	res := &natmdl.NatPage{}
	db := d.DB.Table(_tablePage)
	if id != 0 {
		db = db.Where("id=?", id)
	}
	if tagID > 0 {
		db = db.Where("foreign_id=?", tagID).Where("type=?", defaulType)
	}
	if err := db.Where("state in (?)", states).First(&res).Error; err != nil {
		log.Error("[FindPageByID] error(%v)", err)
		return nil, err
	}
	return res, nil
}

func generateUkey(prefix string) string {
	id := func() string {
		if id, err := gonanoid.New(10); err == nil {
			return id
		}
		return strconv.FormatInt(time.Now().UnixNano(), 10)
	}()
	return fmt.Sprintf("%s_%s", prefix, id)
}

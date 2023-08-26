package service

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	artmdl "git.bilibili.co/bapis/bapis-go/article/model"
	artapi "git.bilibili.co/bapis/bapis-go/article/service"
	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	api "git.bilibili.co/bapis/bapis-go/platform/admin/act-plat"
	tunnelmdl "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"

	"go-common/library/ecode"
	xecode "go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	"go-common/library/net/netutil"
	"go-common/library/retry"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/activity/admin/model"
	"go-gateway/app/web-svr/activity/admin/model/stime"
	"go-gateway/app/web-svr/activity/admin/model/task"
	ecode2 "go-gateway/app/web-svr/activity/ecode"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

const (
	_taskBusinessID             = 1
	canotFindTagErr             = 16001
	noNeedUpgradeTagErr         = 176000
	notification4TunnelGroupFix = `预约活动ID(%v)Name(%s)操作人(%s): 预约人群包<font color=\"warning\">%v</font>。\n
>活动后台地址:http://activity-template.bilibili.co/source/reserve \n
>activity-admin修复url:<font color=\"info\">curl "%v"</font> \n
`
)

// GetArticleMetas from rpc .
func (s *Service) GetArticleMetas(c context.Context, aids []int64) (list map[int64]*artmdl.Meta, err error) {
	var (
		res *artapi.ArticleMetasReply
	)
	if res, err = s.artClient.ArticleMetas(c, &artapi.ArticleMetasReq{Ids: aids}); err != nil {
		log.Errorc(c, "s.ArticleMetas(%v) error(%v)", aids, err)
		return
	}
	list = res.Res
	return
}

// SubInfos .
func (s *Service) SubInfos(c context.Context, sids []int64) (res map[int64]*model.SubProtocol, err error) {
	var (
		list     []*model.ActSubject
		protocol []*model.ActSubjectProtocol
		mapPro   map[int64]*model.ActSubjectProtocol
	)
	db := s.DB.Where("id in (?)", sids)
	if err = db.Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, " db.Model(&model.ActSubject{}).Find() args(%v) error(%v)", sids, err)
		return
	}
	if len(list) > 0 {
		dbTwo := s.DB.Where("sid in (?)", sids)
		if err = dbTwo.Find(&protocol).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Errorc(c, " db.Model(&model.ActSubjectProtocol{}).Find() args(%v) error(%v)", sids, err)
			return
		}
		if len(protocol) > 0 {
			mapPro = make(map[int64]*model.ActSubjectProtocol, len(protocol))
			for _, v := range protocol {
				mapPro[v.Sid] = v
			}
		}
		res = make(map[int64]*model.SubProtocol, len(list))
		for _, v := range list {
			res[v.ID] = &model.SubProtocol{ActSubject: v}
			if _, ok := mapPro[v.ID]; ok {
				res[v.ID].Protocol = mapPro[v.ID]
			}
		}
	}
	return
}

// SubjectList get subject list .
func (s *Service) SubjectList(c context.Context, listParams *model.ListSub) (listRes *model.SubListRes, err error) {
	var (
		count int64
		list  []*model.ActSubject
	)
	db := s.DB

	if listParams.Keyword != "" {
		names := listParams.Keyword + "%"
		db = db.Where("`id` = ? or `name` like ? or `author` like ?", listParams.Keyword, names, names)
	}
	if listParams.Sctime != 0 {
		parseScime := time.Unix(listParams.Sctime, 0)
		db = db.Where("ctime >= ?", parseScime.Format("2006-01-02 15:04:05"))
	}
	if listParams.Ectime != 0 {
		parseEcime := time.Unix(listParams.Ectime, 0)
		db = db.Where("etime <= ?", parseEcime.Format("2006-01-02 15:04:05"))
	}
	if len(listParams.States) > 0 {
		db = db.Where("state in (?)", listParams.States)
	}
	if len(listParams.Types) > 0 {
		db = db.Where("type in (?)", listParams.Types)
	}
	if len(listParams.IDs) > 0 {
		db = db.Where("id in (?)", listParams.IDs)
	}
	if err = db.Offset((listParams.Page - 1) * listParams.PageSize).Limit(listParams.PageSize).Order("id desc").Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, " db.Model(&model.ActSubject{}).Find() args(%v) error(%v)", listParams, err)
		return
	}
	if err = db.Model(&model.ActSubject{}).Count(&count).Error; err != nil {
		log.Errorc(c, "db.Model(&model.ActSubject{}).Count() args(%v) error(%v)", listParams, err)
		return
	}

	listRes = &model.SubListRes{
		List:     list,
		Page:     listParams.Page,
		PageSize: listParams.PageSize,
		Count:    count,
	}
	return
}

// SubjectListAll get subject list .
func (s *Service) SubjectListAll(c context.Context, listParams *model.ListSub) (list []*model.ActSubject, err error) {
	db := s.DB
	if listParams.Sctime != 0 {
		parseScime := time.Unix(listParams.Sctime, 0)
		db = db.Where("stime <= ?", parseScime.Format("2006-01-02 15:04:05"))
	}
	if listParams.Ectime != 0 {
		parseEcime := time.Unix(listParams.Ectime, 0)
		db = db.Where("etime >= ?", parseEcime.Format("2006-01-02 15:04:05"))
	}
	if len(listParams.States) > 0 {
		db = db.Where("state in (?)", listParams.States)
	}
	if len(listParams.Types) > 0 {
		db = db.Where("type in (?)", listParams.Types)
	}
	if len(listParams.IDs) > 0 {
		db = db.Where("id in (?)", listParams.IDs)
	}
	if listParams.Name != "" {
		db = db.Where("name = (?)", listParams.Name)
	}
	if err = db.Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, " db.Model(&model.ActSubject{}).Find() args(%v) error(%v)", listParams, err)
		return
	}
	return
}

// VideoList .
func (s *Service) VideoList(c context.Context) (res []*model.ActSubjectResult, err error) {
	var (
		types    = []int{1, 4}
		list     []*model.ActSubject
		likeList []*model.Like
	)
	db := s.DB
	if err = db.Where("state = ?", 1).Where("type in (?)", types).Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "db.Model(&model.ActSubject{}).Where(state = ?, 1).Where(type in (?), %v).Find() error(%v)", types, err)
		return
	}
	listCount := len(list)
	if listCount == 0 {
		return
	}
	sids := make([]int64, 0, listCount)
	for _, value := range list {
		sids = append(sids, value.ID)
	}
	if err = db.Where("sid in (?)", sids).Where("wid > ?", 0).Find(&likeList).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "db.Model(&model.Like{}).Where(sid in (?), %v).Find() error(%v)", sids, err)
		return
	}
	hashList := make(map[int64][]int64)
	for _, value := range likeList {
		hashList[value.Sid] = append(hashList[value.Sid], value.Wid)
	}
	res = make([]*model.ActSubjectResult, 0, len(list))
	for _, value := range list {
		rs := &model.ActSubjectResult{
			ActSubject: value,
		}
		if v, ok := hashList[value.ID]; ok {
			rs.Aids = v
		}
		res = append(res, rs)
	}
	return
}

// AddActSubject .
func (s *Service) AddActSubject(c context.Context, params *model.AddList, tagType tagrpc.TagType) (res int64, err error) {
	if params.ScreenSet != 2 {
		params.ScreenSet = 1
	}
	protocol := &model.ActSubjectProtocol{
		Protocol:        params.Protocol,
		Types:           params.Types,
		Pubtime:         stime.FromString(params.Pubtime),
		Deltime:         stime.FromString(params.Deltime),
		Editime:         stime.FromString(params.Editime),
		Tags:            params.Tags,
		Hot:             params.Hot,
		BgmID:           params.BgmID,
		Oids:            params.Oids,
		ScreenSet:       params.ScreenSet,
		PasterID:        params.PasterID,
		InstepID:        params.InstepID,
		Award:           params.Award,
		AwardUrl:        params.AwardUrl,
		PriorityRegion:  params.PriorityRegion,
		RegionWeight:    params.RegionWeight,
		GlobalWeight:    params.GlobalWeight,
		WeightStime:     stime.FromString(params.WeightStime),
		WeightEtime:     stime.FromString(params.WeightEtime),
		TagShowPlatform: 127, //默认全选
	}
	actTime := &model.ActTimeConfig{
		Interval: params.Interval,
		Tlimit:   params.Tlimit,
		Ltime:    params.Ltime,
	}
	if params.Tags != "" {
		if err = s.addActivityTag(c, params.Tags, tagType); err != nil {
			log.Errorc(c, "s.addActivityTag(%s,) error(%v)", params.Tags, err)
			return 0, ecode2.ActivitySubjectTagDup
		}
	}
	actSub := &model.ActSubject{
		Oid:           params.Oid,
		Type:          params.Type,
		State:         params.State,
		Level:         params.Level,
		Flag:          params.Flag,
		Rank:          params.Rank,
		Stime:         stime.FromString(params.Stime),
		Etime:         stime.FromString(params.Etime),
		Lstime:        stime.FromString(params.Lstime),
		Letime:        stime.FromString(params.Letime),
		Uetime:        stime.FromString(params.Uetime),
		Ustime:        stime.FromString(params.Ustime),
		Name:          params.Name,
		Author:        params.Author,
		ActURL:        params.ActURL,
		Cover:         params.Cover,
		Dic:           params.Dic,
		H5Cover:       params.H5Cover,
		LikeLimit:     params.LikeLimit,
		AndroidURL:    params.AndroidURL,
		IosURL:        params.IosURL,
		ChildSids:     params.ChildSids,
		MonthScore:    params.MonthScore,
		YearScore:     params.YearScore,
		Contacts:      params.Contacts,
		ShieldFlag:    params.ShieldFlag,
		RelationID:    params.RelationID,
		Calendar:      params.Calendar,
		AuditPlatform: params.AuditPlatform,
	}
	if err = s.DB.Create(actSub).Error; err != nil {
		log.Errorc(c, "s.DB.Create(%v) error(%v)", actSub, err)
		return
	}
	protocol.Sid = actSub.ID
	if err = s.DB.Create(protocol).Error; err != nil {
		log.Errorc(c, "s.DB.Create(%v) error(%v)", protocol, err)
		return
	}
	if params.Type == model.ONLINEVOTE {
		actTime.Sid = actSub.ID
		if err = s.DB.Create(actTime).Error; err != nil {
			log.Errorc(c, "s.DB.Create(%v) error(%v)", actTime, err)
			return
		}
	}
	if params.Type == model.USERACTIONSTAT {
		_, err = s.platAdminClient.AddActivity(c, &api.ActivityReq{
			Name:        fmt.Sprint(actSub.ID),
			Description: actSub.Name,
			Contact:     actSub.Contacts,
			StartTime:   actSub.Stime.Time().Unix(),
			EndTime:     actSub.Etime.Time().Unix(),
			OpUser: &api.OpUser{
				Uname: actSub.Author,
				Refer: api.ReferActivityAdmin,
			},
		})
		if err != nil {
			log.Errorc(c, "s.platAdminClient.AddActivity error(%v)", err)
			return
		}
	}
	if isReserve(actSub.Type) {
		if err = s.TunnelAddGroup(c, actSub); err != nil {
			log.Errorc(c, "AddActSubject s.tunnelAddGroup actSub.ID(%d) error(%+v)", actSub.ID, err)
			err = ecode2.ActivityTunnelGroupErr
			return
		}
		log.Infoc(c, "AddActSubject s.tunnelAddGroup actSub.ID(%d) success name(%s) stime(%s) etime(%s) params.State(%d) success", actSub.ID, params.Name, params.Stime, params.Etime, params.State)
	}
	res = actSub.ID
	if err != nil {
		return
	}
	// 免审处理
	if params.AuditLevel == 1 && (actSub.Type == model.VIDEO || actSub.Type == model.VIDEOLIKE ||
		actSub.Type == model.VIDEO2 || actSub.Type == model.PHONEVIDEO || actSub.Type == model.SMALLVIDEO) {
		err = s.dao.SubjectAuditFeatureImport(c, actSub.ID, actSub.Author)
	}
	return
}

func (s *Service) AddNativePage(c context.Context, tags string) (res bool) {
	if err := retry.WithAttempts(c, "video_activity_native", 3, netutil.DefaultBackoffConfig, func(ctx context.Context) error {
		e := s.dao.AddNativePage(ctx, tags, metadata.String(c, metadata.RemoteIP))
		if xecode.EqualError(xecode.Int(noNeedUpgradeTagErr), e) {
			log.Errorc(c, "UpActSubject AddNativePage s.dao.AddNativePage activity actName(%s) noNeedUpgradeTagErr error(%+v)", tags, e)
			return nil
		}
		return e
	}); err != nil {
		if xecode.EqualError(xecode.Int(canotFindTagErr), err) {
			res = true
		}
		log.Errorc(c, "UpActSubject AddNativePage s.dao.AddNativePage activity actName(%s) error(%+v)", tags, err)
	}
	return res
}

func isReserve(tp int) bool {
	if tp == model.RESERVATION || tp == model.CLOCKIN || tp == model.USERACTIONSTAT {
		return true
	}
	return false
}

func (s *Service) TunnelAddGroup(ctx context.Context, actSub *model.ActSubject) (err error) {
	groupArg := &tunnelmdl.GroupReq{
		Source:      s.c.TunnelGroup.Source,
		Name:        strconv.FormatInt(actSub.ID, 10),
		Description: actSub.Name,
		StartTime:   actSub.Stime.Time().Format("2006-01-02 15:04:05"),
		EndTime:     actSub.Etime.Time().Format("2006-01-02 15:04:05"),
	}
	for i := 0; i < 3; i++ {
		if _, err = s.tunnelClient.AddGroup(ctx, groupArg); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		log.Errorc(ctx, "AddActSubject tunnelAddGroup s.tunnelClient.AddGroup actSub.ID(%d) error(%+v)", actSub.ID, err)
		notification4FixMsg := fmt.Sprintf(
			notification4TunnelGroupFix,
			actSub.ID,
			actSub.Name,
			actSub.Author,
			"创建失败"+err.Error(),
			fmt.Sprintf("http://127.0.0.1:7741/x/admin/activity/fix/tunnel/group/add?sid=%v", actSub.ID))
		if err = SendWeChatMessage(ctx, notification4FixMsg); err != nil {
			log.Errorc(ctx, "AddActSubject tunnelAddGroup SendWeChatMessage actSub.ID(%d) error(%+v)", actSub.ID, err)
		}
	}
	return
}

// addActivityTag ...
func (s *Service) addActivityTag(c context.Context, tagName string, tagType tagrpc.TagType) error {
	// 查询tag
	tag, err := s.dao.TagInfoByName(c, tagName)
	if err != nil && !xecode.EqualError(xecode.Int(canotFindTagErr), err) {
		log.Errorc(c, "s.dao.TagInfoByName err(%v)", err)
		return err
	}
	if tag == nil {
		if tag, err = s.dao.AddTagNew(c, tagName); err != nil {
			log.Errorc(c, "s.AddTagNew(%s,) error(%v)", tagName, err)
			return err
		}
	}
	// 如果tag存在，判断tag状态
	if tag.Type == model.TagTypeActivity || tag.Type == model.TagTypeUp {
		return nil
	}
	// 更新tag为活动tag
	err = s.dao.TagUpdateByID(c, tag.Id, tagType)
	if err != nil {
		log.Errorc(c, "s.dao.TagUpdateByID err(%v)", err)
		return err
	}
	return nil

}

// tagToNormal ...
func (s *Service) tagToNormal(c context.Context, tagName string) (err error) {
	// 查询tag
	tag, err := s.dao.TagInfoByName(c, tagName)
	if err != nil {
		log.Errorc(c, "s.dao.TagInfoByName err(%v)", err)
		return err
	}
	if tag == nil {
		err = errors.Wrapf(err, "s.dao.TagInfoByName not find", err)
		return
	}
	// 如果tag存在，判断tag状态
	if tag.Type != model.TagTypeActivity && tag.Type != model.TagTypeUp {
		return nil
	}
	// 更新tag为活动tag
	err = s.dao.TagUpdateByID(c, tag.Id, tagrpc.TagType_TypeUser)
	if err != nil {
		log.Errorc(c, "s.dao.TagUpdateByID err(%v)", err)
		return err
	}
	return nil

}

// OffileSubject 下线活动
func (s *Service) OffileSubject(c context.Context, sid []int64, userName string) (err error) {
	data := make(map[string]interface{})
	var etime stime.Time
	etime = stime.Time(time.Now())
	data["etime"] = etime
	data["author"] = userName
	if err = s.DB.Model(&model.ActSubject{}).Where("id in (?)", sid).Update(data).Error; err != nil {
		log.Errorc(c, "s.DB.Model(&model.ActSubject{}).Where(id in ?, %d).Update(%v) error(%v)", sid, data, err)
		return
	}
	return
}

// UpActSubject .
func (s *Service) UpActSubject(c context.Context, params *model.AddList, sid int64, data map[string]interface{}) (res int64, err error) {
	if params.ScreenSet != 2 {
		params.ScreenSet = 1
	}
	onlineData := &model.ActTimeConfig{
		Interval: params.Interval,
		Tlimit:   params.Tlimit,
		Ltime:    params.Ltime,
	}
	actSubject := new(model.ActSubject)
	if err = s.DB.Where("id = ?", sid).Last(actSubject).Error; err != nil {
		log.Errorc(c, "s.DB.Where(id = ?, %d).Last(%v) error(%v)", sid, actSubject, err)
		return
	}
	// 预约活动，判断不能修改推送类型
	if _, ok := data["flag"]; ok {
		yyType := reserveType()
		if _, ok := yyType[actSubject.Type]; ok {
			if err = checkReserve(params.Flag, actSubject.Flag); err != nil {
				return
			}
		}
	}

	if actSubject.Type == model.CLOCKIN {
		var (
			lstime int64
			letime int64
		)
		nowTs := time.Now().Unix()
		if v, ok := data["lstime"]; ok {
			lstime = v.(time.Time).Unix()
		}
		if v, ok := data["letime"]; ok {
			letime = v.(time.Time).Unix()
		}
		// 如果修改的是活动时间 默认打卡时间为0000-00-00 00:00:00的话 不做打卡时间调整
		if lstime > 0 && letime > 0 {
			// 开始时间小于结束时间
			if lstime >= letime {
				return 0, ecode.Error(ecode.RequestErr, "开始时间不能大于等于结束时间")
			}
			// 如果DB中没有打卡时间的话 可以随意配置开始时间 但是结束时间要大于当前时间
			if actSubject.Lstime.Time().Unix() < 0 && actSubject.Letime.Time().Unix() < 0 {
				if letime < nowTs {
					return 0, ecode.Error(ecode.RequestErr, "配置打卡时间，结束时间不能小于当前时间")
				}
			} else {
				// 如果DB中有打卡时间的话 活动开始前 时间随意修改 但是结束时间要比当前时间大
				if nowTs < actSubject.Lstime.Time().Unix() {
					if letime < nowTs {
						return 0, ecode.Error(ecode.RequestErr, "活动未开始，结束时间不能小于当前时间")
					}
				} else {
					// 活动开始后，开始时间无法修改，结束时间只能往后延
					if actSubject.Lstime.Time().Unix() != lstime {
						return 0, ecode.Error(ecode.RequestErr, "活动已开始，开始时间无法修改")
					}
					if letime < actSubject.Letime.Time().Unix() {
						return 0, ecode.Error(ecode.RequestErr, "活动已开始，结束时间无法提前")
					}
				}
			}
		}
	}

	// 打卡时间禁止修改
	//if actSubject.Type == model.CLOCKIN {
	//	if !actSubject.Lstime.Time().IsZero() {
	//		if _, ok := data["lstime"]; ok && stime.FromString(params.Lstime) != actSubject.Lstime {
	//			return 0, ecode.Error(ecode.RequestErr, "打卡时间不能修改")
	//		}
	//	}
	//	if !actSubject.Letime.Time().IsZero() {
	//		if _, ok := data["letime"]; ok && stime.FromString(params.Letime) != actSubject.Letime {
	//			return 0, ecode.Error(ecode.RequestErr, "打卡时间不能修改")
	//		}
	//	}
	//}
	if err = s.DB.Model(&model.ActSubject{}).Where("id = ?", sid).Update(data).Error; err != nil {
		log.Errorc(c, "s.DB.Model(&model.ActSubject{}).Where(id = ?, %d).Update(%v) error(%v)", sid, data, err)
		return
	}
	item := new(model.ActSubjectProtocol)
	if err = s.DB.Where("sid = ? ", sid).Last(item).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "s.DB.Where(sid = ? , %d).Last(%v) error(%v)", sid, item, err)
		return
	}
	var nativeNotice bool
	//item有值
	if item.ID > 0 {
		if params.Tags != "" {
			if item.Tags != params.Tags {
				if err = s.dao.AddTags(c, params.Tags, metadata.String(c, metadata.RemoteIP)); err != nil {
					log.Errorc(c, "s.AddTags(%s) error(%v)", params.Tags, err)
					return 0, ecode2.ActivitySubjectTagDup
				}
				// 调用Ponney话题自动升级接口
				if actSubject.IsVideoSource() {
					nativeNotice = s.AddNativePage(c, params.Tags)
				}
			}
		}
		if err = s.DB.Model(&model.ActSubjectProtocol{}).Where("id = ?", item.ID).Update(data).Error; err != nil {
			log.Errorc(c, "s.DB.Model(&model.ActSubjectProtocol{}).Where(id = ?, %d).Update(%v) error(%v)", item.ID, data, err)
			return
		}
	} else {
		protocolData := &model.ActSubjectProtocol{
			Protocol:        params.Protocol,
			Types:           params.Types,
			Pubtime:         stime.FromString(params.Pubtime),
			Deltime:         stime.FromString(params.Deltime),
			Editime:         stime.FromString(params.Editime),
			Hot:             params.Hot,
			BgmID:           params.BgmID,
			Oids:            params.Oids,
			ScreenSet:       params.ScreenSet,
			PasterID:        params.PasterID,
			Sid:             sid,
			InstepID:        params.InstepID,
			Award:           params.Award,
			AwardUrl:        params.AwardUrl,
			Tags:            params.Tags,
			PriorityRegion:  params.PriorityRegion,
			RegionWeight:    params.RegionWeight,
			GlobalWeight:    params.GlobalWeight,
			WeightStime:     stime.FromString(params.WeightStime),
			WeightEtime:     stime.FromString(params.WeightEtime),
			TagShowPlatform: params.TagShowPlatform,
		}
		if err = s.DB.Create(protocolData).Error; err != nil {
			log.Errorc(c, "s.DB.Create(%v) error(%v)", protocolData, err)
			return
		}
	}
	if actSubject.Type == model.ONLINEVOTE {
		onlineData.Sid = sid
		output := new(model.ActTimeConfig)
		if err = s.DB.Where("sid = ?", sid).Last(output).Error; err != nil && err != gorm.ErrRecordNotFound {
			log.Errorc(c, "s.DB.Where(sid = ?, %d).Last(%v) error(%v)", sid, output, err)
			return
		}
		if output.ID > 0 {
			if err = s.DB.Model(&model.ActTimeConfig{}).Where("id = ?", output.ID).Update(onlineData).Error; err != nil {
				log.Errorc(c, "s.DB.Model(&model.ActTimeConfig{}).Where(id = ?, %d).Update(%v) error(%v)", output.ID, onlineData, err)
				return
			}
		}
	}

	if actSubject.Type == model.USERACTIONSTAT {
		sTime, eTime := actSubject.Stime, actSubject.Etime
		if _, ok := data["stime"]; ok {
			sTime = stime.FromString(params.Stime)
		}
		if _, ok := data["etime"]; ok {
			eTime = stime.FromString(params.Etime)
		}
		contact := actSubject.Contacts
		if params.Contacts != "" {
			contact = params.Contacts
		}
		if sTime != actSubject.Stime || eTime != actSubject.Etime || contact != actSubject.Contacts {
			_, err = s.platAdminClient.UpdateActivity(c, &api.ActivityReq{
				Name:        fmt.Sprint(params.ID),
				Description: actSubject.Name,
				StartTime:   sTime.Time().Unix(),
				EndTime:     eTime.Time().Unix(),
				Contact:     contact,
				OpUser: &api.OpUser{
					Uname: params.Author,
					Refer: api.ReferActivityAdmin,
				},
			})
			if err != nil {
				log.Errorc(c, "s.platAdminClient.UpdateActivity error(%v)", err)
				return
			}
		}
	}

	//// 编辑 要触发线上刷新内存IDs 以及需要刷新单条缓存数据 具体值详情收拢到该地址 interface/model/like/like.go
	//req := &pb.InternalUpdateItemDataWithCacheReq{
	//	Typ:        2,
	//	ActionType: 1,
	//	Oid:        sid,
	//}
	//if _, err = s.actClient.InternalUpdateItemDataWithCache(c, req); err != nil {
	//	log.Errorc(c, "InternalUpdateItemDataWithCache Req(%v) Err(%v)", req, err)
	//	err = ecode.Error(ecode.RequestErr, "数据更新成功，线上数据同步失败，请编辑该条目，重新提交")
	//	return
	//}
	//
	//req1 := &pb.InternalSyncActSubjectInfoDB2CacheReq{
	//	From: "op",
	//}
	//_, err = s.actClient.InternalSyncActSubjectInfoDB2Cache(c, req1)
	//if err != nil {
	//	log.Errorc(c, "InternalSyncActSubjectInfoDB2Cache Req(%v) Err(%v)", req1, err)
	//	err = ecode.Error(ecode.RequestErr, "预约活动线上预热同步失败，请编辑打开该条目，重新保存")
	//	return
	//}

	if isReserve(params.Type) {
		if params.State == -1 {
			if err = s.TunnelDelGroup(c, sid, actSubject.Name, actSubject.Author); err != nil {
				log.Errorc(c, "UpActSubject s.tunnelDelGroup actSub.ID(%d) error(%+v)", sid, err)
			}
		} else {
			if err = s.TunnelUpGroup(c, sid, params.Name, params.Stime, params.Etime, params.Author); err != nil {
				log.Errorc(c, "UpActSubject s.tunnelUpGroup actSub.ID(%d) error(%+v)", sid, err)
			}
		}
		if err == nil {
			log.Infoc(c, "UpActSubject s.tunnelUpGroup actSub.ID(%d) name(%s) stime(%s) etime(%s) params.State(%d) success", sid, params.Name, params.Stime, params.Etime, params.State)
		}
	}
	if nativeNotice && err == nil {
		err = ecode2.ActivityNativePageError
	}
	res = sid
	return
}

func (s *Service) TunnelUpGroup(ctx context.Context, sid int64, description, stime, etime, author string) (err error) {
	groupArg := &tunnelmdl.GroupReq{
		Source:      s.c.TunnelGroup.Source,
		Name:        strconv.FormatInt(sid, 10),
		Description: description,
		StartTime:   stime,
		EndTime:     etime,
	}
	for i := 0; i < 3; i++ {
		if _, err = s.tunnelClient.UpdateGroup(ctx, groupArg); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		log.Errorc(ctx, "UpActSubject tunnelUpGroup s.tunnelClient.UpdateGroup actSub.ID(%d) error(%+v)", sid, err)
		notification4FixMsg := fmt.Sprintf(
			notification4TunnelGroupFix,
			sid,
			description,
			author,
			"修改失败"+err.Error(),
			fmt.Sprintf("http://127.0.0.1:7741/x/admin/activity/fix/tunnel/group/up?sid=%v", sid))
		if err = SendWeChatMessage(ctx, notification4FixMsg); err != nil {
			log.Errorc(ctx, "UpActSubject TunnelUpGroup SendWeChatMessage actSub.ID(%d) error(%+v)", sid, err)
		}
	}
	return
}

func (s *Service) TunnelDelGroup(ctx context.Context, sid int64, sname, author string) (err error) {
	groupDelReq := &tunnelmdl.GroupDelReq{
		Source: s.c.TunnelGroup.Source,
		Name:   strconv.FormatInt(sid, 10),
	}
	for i := 0; i < 3; i++ {
		if _, err = s.tunnelClient.DelGroup(ctx, groupDelReq); err == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if err != nil {
		log.Errorc(ctx, "UpActSubject TunnelDelGroup s.tunnelClient.DelGroup actSub.ID(%d) error(%+v)", sid, err)
		notification4FixMsg := fmt.Sprintf(
			notification4TunnelGroupFix,
			sid,
			sname,
			author,
			"删除失败"+err.Error(),
			fmt.Sprintf("http://127.0.0.1:7741/x/admin/activity/fix/tunnel/group/del?sid=%v", sid))
		if err = SendWeChatMessage(ctx, notification4FixMsg); err != nil {
			log.Errorc(ctx, "UpActSubject TunnelDelGroup SendWeChatMessage actSub.ID(%d) error(%+v)", sid, err)
		}
	}
	return
}

// SubProtocol .
func (s *Service) SubProtocol(c context.Context, sid int64) (res *model.ActSubjectProtocol, err error) {
	res = &model.ActSubjectProtocol{}
	if err = s.DB.Where("sid = ?", sid).First(res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "s.DB.Where(sid = %d ) error(%v)", sid, err)
	}
	return
}

// SubProtocolByTagName .
func (s *Service) SubProtocolByTagName(c context.Context, tagName string) (list []*model.ActSubjectProtocol, err error) {
	db := s.DB.Where("tags = ?", tagName)
	if err = db.Find(&list).Error; err != nil {
		log.Errorc(c, "SubProtocolByTagName db.Where(tagName:%v).Find error(%v)", tagName, err)
	}
	return
}

// TimeConf .
func (s *Service) TimeConf(c context.Context, sid int64) (res *model.ActTimeConfig, err error) {
	res = new(model.ActTimeConfig)
	if err = s.DB.Where("sid = ?", sid).First(res).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "actSrv.DB.Where(sid = ?, %d) error(%v)", sid, err)
	}
	return
}

func (s *Service) OptVideoList(c context.Context, arg *model.OptVideoListSub) (res []*model.SubProtocol, cnt int, err error) {
	var (
		subject   []*model.ActSubject
		protocol  []*model.ActSubjectProtocol
		sids      []int64
		mapSub    map[int64]*model.ActSubject
		mapProto  map[int64]*model.ActSubjectProtocol
		videoList = []int{model.VIDEO, model.VIDEOLIKE, model.VIDEO2, model.PHONEVIDEO, model.SMALLVIDEO, model.CLOCKIN}
	)
	db := s.dao.DB
	if arg.Keyword != "" {
		db = db.Where("`sid` = ?", arg.Keyword)
	}
	if arg.Types != "" {
		db = db.Where("`types` like '%"+arg.Types+"%'").Or("types=?", "")
	}
	if err = db.Model(&model.ActSubjectProtocol{}).
		Joins("INNER JOIN act_subject on act_subject.id = act_subject_protocol.sid and act_subject.state != -1 and act_subject.type in(?)", videoList).Count(&cnt).Error; err != nil {
		log.Errorc(c, "db.Model(&model.ActSubjectProtocol{}).Count() args(%v) error(%v)", arg, err)
		return
	}
	if cnt == 0 {
		return
	}
	if err = db.Offset((arg.Page-1)*arg.PageSize).Limit(arg.PageSize).Order("id DESC").
		Joins("INNER JOIN act_subject on act_subject.id = act_subject_protocol.sid and act_subject.state != -1 and act_subject.type in(?)", videoList).Find(&protocol).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "db.Where(state = 1).Order(id desc).Find(&protocol) error(%v)", err)
		return
	}
	if len(protocol) == 0 {
		return
	}
	mapProto = make(map[int64]*model.ActSubjectProtocol, 0)
	if arg.Types == "" {
		for _, v := range protocol {
			sids = append(sids, v.Sid)
			mapProto[v.Sid] = v
		}
	} else {
		for _, v := range protocol {
			tmp := strings.Split(v.Types, ",")
			if v.Types == "" {
				sids = append(sids, v.Sid)
				mapProto[v.Sid] = v
				continue
			}
			for _, val := range tmp {
				if val == arg.Types {
					sids = append(sids, v.Sid)
					mapProto[v.Sid] = v
				}
			}
		}
	}
	if len(sids) == 0 {
		log.Warn("no suit sid")
		return
	}
	if err = s.dao.DB.Where("id in (?)", sids).Where("state != -1").Find(&subject).Error; err != nil {
		log.Errorc(c, "db.Where(id in (?), sids).Find(&subject) error(%v)", err)
		return
	}
	if len(subject) == 0 {
		return
	}
	mapSub = make(map[int64]*model.ActSubject, len(subject))
	for _, v := range subject {
		mapSub[v.ID] = v
	}
	for _, v := range mapProto {
		if _, ok := mapSub[v.Sid]; ok {
			tmp := &model.SubProtocol{Protocol: v}
			tmp.ActSubject = mapSub[v.Sid]
			res = append(res, tmp)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ActSubject.ID > res[j].ActSubject.ID
	})
	return
}

func (s *Service) SubjectRules(c context.Context, sid, state int64) (list []*model.SubjectRule, err error) {
	db := s.dao.DB.Where("sid=?", sid)
	if state > 0 {
		db = db.Where(fmt.Sprintf("state = %d", state))
	} else {
		db = db.Where("state != 3")
	}
	if err = db.Find(&list).Error; err != nil {
		log.Errorc(c, "SubjectRules db.Where(sid:%d).Find error(%v)", sid, err)
	}
	return
}

func (s *Service) AddSubjectRule(c context.Context, rule *model.AddSubjectRuleArg) (err error) {
	if rule.Tags, err = s.checkSubjectRule(c, 0, rule.TypeIDs, rule.Tags); err != nil {
		return
	}
	sub := new(model.ActSubject)
	if err = s.dao.DB.Where("id = ?", rule.Sid).First(&sub).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "AddSubjectRules first subject sid:%d error(%v)", rule.Sid, err)
		return
	}
	if sub == nil || sub.Type != model.CLOCKIN {
		err = ecode.Error(ecode.RequestErr, "subject 类型不对")
		return
	}
	if sub.Lstime.Time().IsZero() || sub.Letime.Time().IsZero() {
		return ecode.Error(ecode.RequestErr, "请先配置打卡时间")
	}
	if !(rule.Stime < rule.Etime &&
		int64(rule.Stime) >= sub.Lstime.Time().Unix() &&
		int64(rule.Etime) <= sub.Letime.Time().Unix()) {
		return ecode.Error(ecode.RequestErr, "规则时间范围一定要在打卡时间范围内")
	}
	tx := s.dao.DB.Begin()
	taskAdd := &task.Task{
		Name:       "sub_rule",
		BusinessID: _taskBusinessID,
		ForeignID:  rule.Sid,
		Attribute:  (1 << task.AttrBitIsAutoGet) + (1 << task.AttrBitNoFinish) + (1 << task.AttrBitNewTable) + (1 << task.AttrBitDayCount),
		State:      task.TaskOffline,
		Stime:      rule.Stime,
		Etime:      rule.Etime,
	}
	if rule.State == model.SubRuleOnline {
		taskAdd.State = task.TaskOnline
	}

	if err = tx.Model(&task.Task{}).Create(taskAdd).Error; err != nil {
		log.Errorc(c, "AddSubjectRule task(%+v) create error(%v)", taskAdd, err)
		tx.Rollback()
		return
	}

	if err = tx.Model(&model.SubjectRule{}).Create(&model.SubjectRule{
		Sid:       rule.Sid,
		Category:  rule.Category,
		TypeIDs:   rule.TypeIDs,
		Tags:      rule.Tags,
		TaskID:    taskAdd.ID,
		State:     rule.State,
		Attribute: rule.Attribute,
		Stime:     rule.Stime,
		Etime:     rule.Etime,
	}).Error; err != nil {
		log.Errorc(c, "AddSubjectRules create %+v error(%v)", rule, err)
		tx.Rollback()
		return
	}
	err = tx.Commit().Error
	return
}

func (s *Service) SaveSubjectRule(c context.Context, rule *model.SaveSubjectRuleArg) (err error) {
	if rule.Tags, err = s.checkSubjectRule(c, rule.ID, rule.TypeIDs, rule.Tags); err != nil {
		return
	}
	sub := new(model.ActSubject)
	if err = s.dao.DB.Where("id = ?", rule.Sid).First(&sub).Error; err != nil && err != gorm.ErrRecordNotFound {
		log.Errorc(c, "SaveSubjectRule first subject sid:%d error(%v)", rule.Sid, err)
		return
	}
	if sub == nil || sub.Type != model.CLOCKIN {
		err = ecode.Error(ecode.RequestErr, "subject 类型不对")
		return
	}
	preRule := &model.SubjectRule{}
	if err = s.dao.DB.Model(&model.SubjectRule{}).Where("id=?", rule.ID).First(&preRule).Error; err != nil {
		log.Errorc(c, "SaveSubjectRule preRule id:%d error(%v)", rule.ID, err)
		return
	}
	if preRule.Sid != rule.Sid {
		err = ecode.Error(ecode.RequestErr, "sid 数据冲突")
		return
	}
	if !(rule.Stime < rule.Etime &&
		int64(rule.Stime) >= sub.Lstime.Time().Unix() &&
		int64(rule.Etime) <= sub.Letime.Time().Unix()) {
		return ecode.Error(ecode.RequestErr, "规则时间范围一定要在打卡时间范围内")
	}
	nowTs := time.Now().Unix()
	// 活动开始前 时间随意修改 但是结束时间不能小于现在时间
	if nowTs < int64(rule.Stime) {
		if int64(rule.Etime) < nowTs {
			return ecode.Error(ecode.RequestErr, "活动未开始，但是结束时间不能小于当前时间")
		}
	} else {
		// 老数据的话 时间默认是0000-00-00 这种情况开始时间要跟打卡开始时间来比对
		if int64(preRule.Stime) < 0 {
			if int64(rule.Stime) != sub.Lstime.Time().Unix() {
				return ecode.Error(ecode.RequestErr, "活动已经开始，开始时间不允许调整")
			}
		} else {
			// 活动开始后，开始时间无法调整
			if rule.Stime != preRule.Stime {
				return ecode.Error(ecode.RequestErr, "活动已经开始，开始时间不允许调整")
			}
		}
		// 老数据是0000-00-00 所以结束时间要跟打卡结束时间来比对
		if int64(preRule.Etime) < 0 {
			if int64(rule.Etime) < sub.Letime.Time().Unix() {
				return ecode.Error(ecode.RequestErr, "活动已经开始，无法选择提前结束打卡时间")
			}
		} else {
			// 结束时间不能提前
			if rule.Etime < preRule.Etime {
				return ecode.Error(ecode.RequestErr, "活动已经开始，无法选择提前结束打卡时间")
			}
		}
	}
	tx := s.dao.DB.Begin()
	if err = tx.Model(&model.SubjectRule{}).Where("id=?", rule.ID).Updates(map[string]interface{}{
		"category": rule.Category,
		"type_ids": rule.TypeIDs,
		"tags":     rule.Tags,
		"state":    rule.State,
		"stime":    rule.Stime,
		"etime":    rule.Etime,
	}).Error; err != nil {
		log.Errorc(c, "SaveSubjectRule updates %+v error(%v)", rule, err)
		tx.Rollback()
		return
	}
	// 同步task
	if preRule.TaskID > 0 {
		// task有效统计时间跟每条规则时间走
		err = tx.Model(&task.Task{}).Where("id=?", preRule.TaskID).Updates(map[string]interface{}{"stime": rule.Stime, "etime": rule.Etime}).Error
		if err != nil {
			log.Errorc(c, "SaveSubjectRule task rule %v update stime etime error(%v)", rule, err)
			tx.Rollback()
			return
		}
		if preRule.State != model.SubRuleOnline && rule.State == model.SubRuleOnline {
			err = tx.Model(&task.Task{}).Where("id=?", preRule.TaskID).Updates(map[string]interface{}{"state": task.TaskOnline}).Error
		} else if preRule.State == model.SubRuleOnline && rule.State != model.SubRuleOnline {
			err = tx.Model(&task.Task{}).Where("id=?", preRule.TaskID).Updates(map[string]interface{}{"state": task.TaskOffline}).Error
		}
		if err != nil {
			log.Errorc(c, "SaveSubjectRule task state update error(%v)", rule, err)
			tx.Rollback()
			return
		}
	}
	return tx.Commit().Error
}

func (s *Service) UpSubRuleState(c context.Context, id, state int64) (err error) {
	preRule := &model.SubjectRule{}
	if err = s.dao.DB.Model(&model.SubjectRule{}).Where("id=?", id).First(&preRule).Error; err != nil {
		log.Errorc(c, "SaveSubjectRule preRule id:%d error(%v)", id, err)
		return
	}
	tx := s.dao.DB.Begin()
	if err = tx.Model(&model.SubjectRule{}).Where("id=?", id).Updates(map[string]interface{}{"state": state}).Error; err != nil {
		log.Errorc(c, "UpSubRuleState Updates id(%d) state(%d) error(%v)", id, state, err)
		tx.Rollback()
		return
	}
	if preRule.TaskID > 0 {
		if preRule.State != model.SubRuleOnline && state == model.SubRuleOnline {
			err = tx.Model(&task.Task{}).Where("id=?", preRule.TaskID).Updates(map[string]interface{}{"state": task.TaskOnline}).Error
		} else if preRule.State == model.SubRuleOnline && state != model.SubRuleOnline {
			err = tx.Model(&task.Task{}).Where("id=?", preRule.TaskID).Updates(map[string]interface{}{"state": task.TaskOffline}).Error
		}
		if err != nil {
			log.Errorc(c, "UpSubRuleState task state update error(%v)", err)
			tx.Rollback()
			return
		}
	}
	err = tx.Commit().Error
	return
}

func (s *Service) SubjectRuleUserState(c context.Context, mid, sid int64) (res *model.SubRuleUserStateRes, err error) {
	rules, err := s.SubjectRules(c, sid, 0)
	if err != nil {
		return
	}
	var taskIDs []int64
	for _, v := range rules {
		if v != nil && v.TaskID > 0 {
			taskIDs = append(taskIDs, v.TaskID)
		}
	}
	stateMap := make(map[int64]*task.TaskUserState)
	if len(taskIDs) > 0 {
		var taskState []*task.TaskUserState
		if err = s.dao.DB.Table(task.TaskUserStateTable(sid)).Where("task_id IN(?)", taskIDs).Where("mid=?", mid).Find(&taskState).Error; err != nil {
			log.Errorc(c, "SubjectRuleUserState taskIDs(%+v) mid(%d) error(%v)", taskIDs, mid, err)
			return
		}
		for _, v := range taskState {
			if v != nil {
				stateMap[v.TaskID] = v
			}
		}
	}
	res = new(model.SubRuleUserStateRes)
	for _, v := range rules {
		if v == nil {
			continue
		}
		tmp := &model.SubRuleUserState{
			ID:     v.ID,
			Sid:    v.Sid,
			TaskID: v.TaskID,
		}
		if state, ok := stateMap[v.TaskID]; ok && state != nil {
			if v.IsDayCount() {
				tmp.Total = state.RoundCount
			} else {
				tmp.Total = state.Count
			}
		}
		res.List = append(res.List, tmp)
		res.Total += tmp.Total
	}
	return
}

func (s *Service) checkSubjectRule(c context.Context, id int64, typeIDs, tags string) (afTags string, err error) {
	typeIDArr, err := xstr.SplitInts(typeIDs)
	if err != nil {
		return "", ecode.Error(ecode.RequestErr, "type_ids 参数错误")
	}
	typeIDMap := make(map[int64]struct{}, len(typeIDArr))
	for _, typeID := range typeIDArr {
		if typeID == 0 {
			return "", ecode.Error(ecode.RequestErr, "type_ids 中有0")
		}
		if _, ok := typeIDMap[typeID]; ok {
			return "", ecode.Error(ecode.RequestErr, "type_ids 中有重复id")
		}
		typeIDMap[typeID] = struct{}{}
	}
	tagArr := strings.Split(tags, ",")
	//if len(tagArr) == 0 {
	//	return "", ecode.Error(ecode.RequestErr, "tags 参数错误")
	//}
	if len(tagArr) > 3 {
		return "", ecode.Error(ecode.RequestErr, "tags数量需小于3")
	}
	tagMap := make(map[string]struct{}, len(tags))
	if len(tagArr) == 0 {
		return
	}
	for i, tag := range tagArr {
		trimTag := strings.TrimSpace(tag)
		//if trimTag == "" {
		//	return "", ecode.Error(ecode.RequestErr, "tags内容不能为空")
		//}
		tagArr[i] = trimTag
		if _, ok := tagMap[trimTag]; ok {
			return "", ecode.Error(ecode.RequestErr, "tags内容中有重复tag")
		}
		tagMap[trimTag] = struct{}{}
	}
	//var actSubs []*model.ActSubject
	//now := time.Now()
	//if err = s.dao.DB.Model(&model.ActSubject{}).Where("type=?", model.CLOCKIN).Where("stime<=?", now).Where("etime>=?", now).Where("state=?", 1).Find(&actSubs).Error; err != nil {
	//	log.Errorc(c, "checkSubjectRule act subject error(%v)", err)
	//	return "", err
	//}
	//var sids []int64
	//for _, v := range actSubs {
	//	if v != nil && v.ID > 0 {
	//		sids = append(sids, v.ID)
	//	}
	//}
	//if len(sids) > 0 {
	//	var preRules []*model.SubjectRule
	//	if err = s.dao.DB.Model(&model.SubjectRule{}).Where("sid IN(?)", sids).Where("state=?", model.SubRuleOnline).Find(&preRules).Error; err != nil {
	//		log.Errorc(c, "checkSubjectRule pre rules sids(%v) error(%v)", sids, err)
	//		return "", err
	//	}
	//	for _, v := range preRules {
	//		if v != nil && v.ID != id {
	//			preTags := strings.Split(v.Tags, ",")
	//			for _, preTag := range preTags {
	//				for _, newTag := range tagArr {
	//					if preTag == newTag {
	//						return "", ecode.Error(ecode.RequestErr, fmt.Sprintf("tag【%s】名称冲突", newTag))
	//					}
	//				}
	//			}
	//		}
	//	}
	//}

	return strings.Join(tagArr, ","), nil
}

func checkReserve(newFlag, oldFlag int64) error {
	if newFlag>>model.FLAGRESERVEPUSH != oldFlag>>model.FLAGRESERVEPUSH {
		return ecode.Error(ecode.RequestErr, "预约活动不能修改推送")
	}
	return nil
}

func (s *Service) TunnelGroupAdd(ctx context.Context, sid int64) (err error) {
	var (
		actSubject *model.ActSubject
	)
	if actSubject, err = s.checkSubject(ctx, sid); err != nil {
		return
	}
	if err = s.TunnelAddGroup(ctx, actSubject); err != nil {
		log.Errorc(ctx, "TunnelGroupAdd s.tunnelAddGroup actSub.ID(%d) error(%+v)", actSubject.ID, err)
		return
	}
	log.Infoc(ctx, "TunnelGroupAdd s.tunnelAddGroup actSub.ID(%d) success", actSubject.ID)
	return
}

func (s *Service) TunnelGroupUp(ctx context.Context, sid int64) (err error) {
	var (
		actSubject *model.ActSubject
	)
	if actSubject, err = s.checkSubject(ctx, sid); err != nil {
		return
	}
	if err = s.TunnelUpGroup(ctx, sid, actSubject.Name, actSubject.Stime.Time().Format("2006-01-02 15:04:05"), actSubject.Etime.Time().Format("2006-01-02 15:04:05"), actSubject.Author); err != nil {
		log.Errorc(ctx, "TunnelGroupUp s.TunnelUpGroup actSub.ID(%d) error(%+v)", actSubject.ID, err)
		return
	}
	log.Infoc(ctx, "TunnelGroupUp s.TunnelUpGroup actSub.ID(%d) success", actSubject.ID)
	return
}

func (s *Service) TunnelGroupDel(ctx context.Context, sid int64) (err error) {
	var (
		actSubject *model.ActSubject
	)
	if actSubject, err = s.checkSubject(ctx, sid); err != nil {
		return
	}
	if err = s.TunnelDelGroup(ctx, sid, actSubject.Name, actSubject.Author); err != nil {
		log.Errorc(ctx, "TunnelGroupDel s.TunnelDelGroup actSub.ID(%d) error(%+v)", actSubject.ID, err)
		return
	}
	log.Infoc(ctx, "TunnelGroupDel s.TunnelDelGroup actSub.ID(%d) success", actSubject.ID)
	return
}

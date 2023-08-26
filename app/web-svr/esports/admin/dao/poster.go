package dao

import (
	"context"

	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/admin/model"
	"go-gateway/app/web-svr/esports/ecode"
)

const _posterTableName = "match_poster"

// 新增海报
func (d *Dao) CreatePoster(_ context.Context, poster *model.Poster) (err error) {
	return d.DB.Table(_posterTableName).Model(&model.Poster{}).Create(poster).Error
}

// 删除海报
func (d *Dao) DeletePoster(_ context.Context, posterId int64) (err error) {
	return d.DB.Table(_posterTableName).
		Where("id = ? AND is_deprecated = 0", posterId).
		Update("is_deprecated", 1).Error
}

// 编辑海报
func (d *Dao) EditPoster(_ context.Context, poster *model.Poster) (err error) {
	attrsToUpdate := map[string]interface{}{
		"position_order": poster.PositionOrder,
		"bg_image":       poster.BgImage,
		"contest_id":     poster.ContestID,
		"created_by":     poster.CreatedBy,
	}
	return d.DB.Table(_posterTableName).
		Where("id = ? AND is_deprecated = 0", poster.ID).
		Updates(attrsToUpdate).Error
}

// 切换海报上下线
func (d *Dao) TogglePoster(_ context.Context, posterId int64, onlineStatus int32) (err error) {
	return d.DB.Table(_posterTableName).
		Where("id = ? AND is_deprecated = 0", posterId).
		Update("online_status", onlineStatus).Error
}

// 设定中心位置海报
func (d *Dao) CenterPoster(_ context.Context, posterId int64) (err error) {
	return d.DB.Table(_posterTableName).
		Where("id = ? AND is_deprecated = 0", posterId).
		Update("is_centeral", 1).Error
}

// 设定中心位置海报
func (d *Dao) UnCenterPoster(_ context.Context, posterId int64) (err error) {
	return d.DB.Table(_posterTableName).
		Where("id = ? AND is_deprecated = 0", posterId).
		Update("is_centeral", 0).Error
}

// 设定中心位置海报
func (d *Dao) UnCenterAllPoster(_ context.Context) (err error) {
	return d.DB.Table(_posterTableName).
		Where("is_deprecated = 0").
		Update("is_centeral", 0).Error
}

// 根据id查找海报配置
func (d *Dao) FindPosterById(_ context.Context, posterId int64) (result *model.Poster, err error) {
	result = &model.Poster{}
	err = d.DB.Table(_posterTableName).
		Where("id = ? AND is_deprecated = 0", posterId).
		First(result).Error

	return result, err
}

// 查找所有海报配置
func (d *Dao) GetPosterList(_ context.Context, pn int32, ps int32) (result []*model.Poster, err error) {
	err = d.DB.Table(_posterTableName).
		Where("is_deprecated = 0").
		Order("online_status desc").
		Order("position_order").
		Order("ctime").
		Offset((pn - 1) * ps).
		Limit(ps).
		Find(&result).Error
	return
}

// 查找所有海报配置数量
func (d *Dao) GetPosterCount(_ context.Context) (count int, err error) {
	err = d.DB.Table(_posterTableName).Model(&model.Poster{}).Where("is_deprecated = 0").Count(&count).Error
	return
}

// 仅查找有效的海报配置
func (d *Dao) GetEffectivePosterList(_ context.Context) (result []*model.Poster, err error) {
	err = d.DB.Table(_posterTableName).Model(&model.Poster{}).
		Where("is_deprecated = 0 AND online_status = 1").
		Order("position_order").
		Order("ctime").
		Find(&result).Error
	return result, err
}

func (d *Dao) DrawPost(c context.Context, cid int64, templateID int, materials []*model.Material) (picture string, err error) {
	param := url.Values{}
	p := struct {
		Materials []*model.Material `json:"materials"`
	}{
		Materials: materials,
	}
	str, _ := json.Marshal(p)
	log.Warn("materials:%+v", string(str))
	param.Set("materials", string(str))
	param.Set("template_id", strconv.Itoa(templateID))
	param.Set("oid", fmt.Sprintf("https://www.bilibili.com/h5/match/data/detail/%d", cid))
	param.Set("share_id", "general")
	param.Set("buvid", "111111")
	param.Set("platform", "android")
	resp := &struct {
		Code int `json:"code"`
		Data struct {
			Picture string `json:"picture"`
		}
	}{}
	if err = d.replyClient.Post(c, d.genPostURL, "", param, &resp); err != nil {
		log.Error("d.DrawPost param(%+v) error(%+v)", d.genPostURL+"?"+param.Encode(), err)
		err = ecode.EsportsDrawPost
		return
	}
	if resp.Code != 0 || resp.Data.Picture == "" {
		log.Error("d.DrawPost param(%+v) resp.Code(%+v)", d.genPostURL+"?"+param.Encode(), resp.Code)
		err = ecode.EsportsDrawPost
		return
	}
	picture = resp.Data.Picture
	log.Warn("get s10 post success param:%+v data:%+v", d.genPostURL+"?"+param.Encode(), resp.Data)
	return
}

func (d *Dao) SavePost(c context.Context, cid int64, name, picture string) (err error) {
	saveParam := url.Values{}
	saveParam.Set("placard_id", fmt.Sprintf("%d_%s", cid, name))
	saveParam.Set("url", picture)
	resp := &struct {
		Code int `json:"code"`
		Data struct {
			Success bool `json:"success"`
		}
	}{}
	if err = d.replyClient.Post(c, d.savePostURL, "", saveParam, &resp); err != nil {
		log.Error("d.SavePost saveParam(%+v) error(%+v)", d.savePostURL+"?"+saveParam.Encode(), err)
		err = ecode.EsportsDrawPost
		return
	}
	if resp.Code != 0 || !resp.Data.Success {
		log.Error("d.SavePost param(%+v) resp.Code(%+v) success(%v)", d.savePostURL+"?"+saveParam.Encode(), resp.Code, resp.Data.Success)
		err = ecode.EsportsDrawPost
		return
	}
	log.Warn("save s10 post success placard_id(%s) param:%+v", fmt.Sprintf("%d_%s", cid, name), d.savePostURL+"?"+saveParam.Encode())
	return
}

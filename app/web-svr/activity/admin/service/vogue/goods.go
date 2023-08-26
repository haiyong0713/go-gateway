package vogue

import (
	"context"
	"strconv"
	"strings"

	"go-common/library/log"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

var (
	goodsTagToStr = map[int]string{
		1: "猛男必备",
		2: "仙女好物",
		3: "B站周边",
	}
	goodsAttrRealToStr = map[int]string{
		0: "虚拟",
		1: "实物",
	}
	goodsAttrSoldOutToStr = map[int]string{
		0: "未售罄",
		1: "已售罄",
	}
)

// List get goods information list
func (s *Service) ListGoods(c context.Context) (rsp voguemdl.GoodsListRsp, err error) {
	var (
		list []*voguemdl.GoodsData
	)
	if list, err = s.dao.GoodsList(c); err != nil {
		log.Error("[ListGoods] s.dao.GoodsList error(%v)", err)
		return
	}
	for _, item := range list {
		item.Tags, err = item.ExtractTags()
		if err != nil {
			log.Error("[ListGoods] item.ExtractTags error(%v), item(%v)", err, item)
			return
		}
		item.Type = item.AttrVal(voguemdl.GoodsAttrReal)
		item.SoldOut = item.AttrVal(voguemdl.GoodsAttrSellOut)
		item.LeftStock = item.ExtractLeftStock()
	}
	rsp.List = list
	return
}

// Add goods
func (s *Service) AddGoods(c context.Context, request *voguemdl.GoodsAddParam) (err error) {
	request.Type = request.TagsToType()
	if err = s.dao.AddGoods(c, request); err != nil {
		log.Error("[AddGoods] s.dao.AddGoods(%v) error(%v)", request, err)
	}
	return
}

// Delete goods
func (s *Service) DeleteGoods(c context.Context, id int) (err error) {
	if err = s.dao.DelGoods(c, id); err != nil {
		log.Error("[DelGoods] s.dao.DelGoods() ID(%d) error(%v)", id, err)
	}
	return
}

// Modify goods
func (s *Service) ModifyGoods(c context.Context, request *voguemdl.GoodsModifyParam) (err error) {
	request.Type = request.TagsToType()
	if err = s.dao.ModifyGoods(c, request); err != nil {
		log.Error("[ModifyGoods] s.dao.ModifyGoods(%v) error(%v)", request, err)
	}
	return
}

// Set soldOut
func (s *Service) SetSoldOutGoods(c context.Context, id int) (err error) {
	if err = s.dao.SetSoldOutGoods(c, id); err != nil {
		log.Error("[SetSoldOutGoods] s.dao.SetSoldOutGoods() ID(%d) error(%v)", id, err)
	}
	return
}

// Export goods
func (s *Service) ExportGoods(c context.Context) (result [][]string, err error) {
	var (
		list []*voguemdl.GoodsData
	)
	if list, err = s.dao.GoodsList(c); err != nil {
		log.Errorc(c, "[ExportGoods] s.dao.GoodsList error(%v)", err)
		return
	}
	for _, item := range list {
		item.Tags, err = item.ExtractTags()
		if err != nil {
			log.Errorc(c, "[ExportGoods] item.ExtractTags error(%v), item(%v)", err, item)
			return
		}
		var tagNames []string
		for _, tag := range item.Tags {
			tagName := goodsTagToStr[tag]
			tagNames = append(tagNames, tagName)
		}
		item.Type = item.AttrVal(voguemdl.GoodsAttrReal)
		item.SoldOut = item.AttrVal(voguemdl.GoodsAttrSellOut)
		result = append(result, []string{strconv.Itoa(item.ID), item.Name, goodsAttrRealToStr[item.Type], strings.Join(tagNames, ","), strconv.Itoa(item.Score), strconv.Itoa(item.Stock), strconv.Itoa(item.Stock - item.Send), strconv.Itoa(item.Want), goodsAttrSoldOutToStr[item.SoldOut]})
	}
	return
}

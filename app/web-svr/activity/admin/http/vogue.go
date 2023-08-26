package http

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	voguemdl "go-gateway/app/web-svr/activity/admin/model/vogue"
)

// 查询配置
func configList(c *bm.Context) {
	c.JSON(vogueSrv.ConfigList(c))
}

// 修改配置
func modifyConfig(c *bm.Context) {
	var (
		arg = map[string]string{}
	)

	for k, values := range c.Request.Form {
		for _, v := range values {
			arg[k] = v
		}
	}

	c.JSON(nil, vogueSrv.ModifyConfig(c, arg))
}

// 商品添加
func goodsAdd(c *bm.Context) {
	var (
		err error
		arg = &voguemdl.GoodsAddParam{}
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, vogueSrv.AddGoods(c, arg))
}

// 商品删除
func goodsDel(c *bm.Context) {
	var (
		err error
		arg = new(struct {
			ID int `form:"id" validate:"required"`
		})
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, vogueSrv.DeleteGoods(c, arg.ID))
}

// 商品列表
func goodsList(c *bm.Context) {
	c.JSON(vogueSrv.ListGoods(c))
}

// 商品修改
func goodsModify(c *bm.Context) {
	var (
		err error
		arg = &voguemdl.GoodsModifyParam{}
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, vogueSrv.ModifyGoods(c, arg))
}

// 商品置为售罄
func goodsSoldOut(c *bm.Context) {
	var (
		err error
		arg = new(struct {
			ID int `form:"id" validation:"required"`
		})
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(nil, vogueSrv.SetSoldOutGoods(c, arg.ID))
}

// 商品列表导出
func goodsExport(c *bm.Context) {
	var (
		err     error
		infoStr [][]string
	)
	if infoStr, err = vogueSrv.ExportGoods(c); err != nil {
		log.Error("vogueSrv.ExportGoods failed. error(%v)", err)
		c.JSON(nil, err)
		return
	}

	exportCsvParam := &voguemdl.ExportCsvParam{
		FileNameFormat: "时尚活动商品列表.csv",
		Header:         []string{"商品id", "商品名", "商品类型", "商品品类", "兑换积分", "累计库存", "剩余库存", "想要人数", "售罄状态"},
		Result:         infoStr,
	}
	exportCsv(c, exportCsvParam)
}

// 管理后台 - 时尚活动管理 - 活动参与进度 - 兑换进度 - 列表
func prizeList(c *bm.Context) {
	arg := new(voguemdl.PrizeSearch)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(vogueSrv.ListPrizes(c, arg))
}

// 管理后台 - 时尚活动管理 - 活动参与进度 - 兑换进度 - 导出
func prizeExport(c *bm.Context) {
	var (
		err     error
		infoStr [][]string
	)

	arg := new(voguemdl.PrizeExportSearch)
	if err := c.Bind(arg); err != nil {
		return
	}
	if infoStr, err = vogueSrv.ExportPrizes(c, arg); err != nil {
		log.Error("vogueSrv.ExportPrizes failed. error(%v)", err)
		c.JSON(nil, err)
		return
	}

	exportCsvParam := &voguemdl.ExportCsvParam{
		FileNameFormat: "时尚活动兑换进度列表.csv",
		Header:         []string{"昵称", "UID", "兑换商品名称", "商品类型", "兑换时间", "兑换耗时", "消耗积分", "收货信息", "是否存在异常"},
		Result:         infoStr,
	}
	exportCsv(c, exportCsvParam)
}

// 管理平台 - 时尚活动管理 - 活动参与进度 - 积分进度 - 列表
func creditList(c *bm.Context) {
	arg := new(voguemdl.CreditSearch)
	if err := c.Bind(arg); err != nil {
		return
	}
	//c.JSON(&voguemdl.CreditListRsp{
	//	List: make([]*voguemdl.CreditData, 0),
	//	Page: &voguemdl.Page{
	//		Num:   0,
	//		Size:  0,
	//		Total: 0,
	//	},
	//}, nil)
	c.JSON(vogueSrv.ListCredits(c, arg))

}

// 管理平台 - 时尚活动管理 - 活动参与进度 - 积分进度 - 导出
func creditListExport(c *bm.Context) {
	var (
		err     error
		infoStr [][]string
	)
	if infoStr, err = vogueSrv.ExportCredits(c); err != nil {
		log.Error("vogueSrv.ExportCredits failed. error(%v)", err)
		c.JSON(nil, err)
		return
	}

	exportCsvParam := &voguemdl.ExportCsvParam{
		FileNameFormat: "时尚活动积分进度列表.csv",
		Header:         []string{"昵称", "UID", "累计积分", "剩余积分", "看视频积分", "好友邀请积分", "提交礼物时间", "提交礼物名称", "礼物要求积分", "是否存在异常"},
		Result:         infoStr,
	}
	exportCsv(c, exportCsvParam)

}

// creditListGenerate 积分列表异步导出
func creditListGenerate(c *bm.Context) {
	var (
		err error
	)
	if err = vogueSrv.GenerateExportCreditsTask(c); err != nil {
		log.Error("vogueSrv.GenerateExportCreditsTask failed. error(%v)", err)
		c.JSON(nil, err)
		return
	}
	c.JSON(nil, err)
}

// creditListAvailable 获取最新一次导出结果对应信息（创建时间、导出成功时间）
func creditListAvailable(c *bm.Context) {
	c.JSON(vogueSrv.ExportAsyncData(c))
}

// creditListDownload 下载最新一次导出结果
func creditListDownload(c *bm.Context) {
	var (
		err        error
		exportData *voguemdl.CreditExportData
		bs         []byte
	)

	if exportData, err = vogueSrv.ExportAsyncData(c); err != nil {
		log.Error("[creditListExportAsyncDownload]vogueSrv.ExportAsyncData failed. error(%v)", err)
		c.JSON(nil, err)
		return
	}
	if exportData.FilePath == "" {
		log.Error("[creditListExportAsyncDownload]exportData.FilePath empty")
		c.JSON(nil, ecode.Error(ecode.RequestErr, "没有已导出的文件，请重新点导出"))
		return
	}
	bs, err = ioutil.ReadFile(exportData.FilePath)
	if err != nil {
		log.Error("[creditListExportAsyncDownload]ioutil.ReadFile error(%+v)", err)
		c.JSON(nil, ecode.Error(ecode.RequestErr, "导出的文件已过期，请重新点导出"))
		return
	}
	_, fileName := filepath.Split(exportData.FilePath)
	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	c.String(200, string(bs))
}

// 管理平台 - 时尚活动管理 - 活动参与进度 - 积分进度 - 积分明细 - 列表
func creditDetail(c *bm.Context) {
	arg := new(voguemdl.CreditDetailSearch)
	if err := c.Bind(arg); err != nil {
		return
	}
	c.JSON(vogueSrv.ListCreditsDetail(c, arg))
}

// 管理平台 - 时尚活动管理 - 活动参与进度 - 积分进度 - 积分明细 - 导出
func creditDetailExport(c *bm.Context) {
	var (
		err     error
		infoStr [][]string
	)
	arg := new(voguemdl.CreditDetailSearch)
	if err := c.Bind(arg); err != nil {
		return
	}
	if infoStr, err = vogueSrv.ExportCreditsDetail(c, arg); err != nil {
		log.Error("vogueSrv.ExportCreditsDetail failed. error(%v)", err)
		c.JSON(nil, err)
		return
	}

	fileNameParams := make([]interface{}, 1)
	fileNameParams[0] = arg.Uid

	exportCsvParam := &voguemdl.ExportCsvParam{
		FileNameFormat: "时尚活动用户(%d)积分进度详情.csv",
		FileNameParams: fileNameParams,
		Header:         []string{"时间", "明细类型", "方式", "积分变化", "累计积分", "视频信息", "好友信息"},
		Result:         infoStr,
	}
	exportCsv(c, exportCsvParam)
}

// 管理平台 - 时尚活动管理 - 活动参与进度 - 中奖进度 - 列表
func winningList(c *bm.Context) {
	var (
		err error
		arg = new(voguemdl.CreditSearch)
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(vogueSrv.WinningList(c, arg))
}

// 管理平台 - 时尚活动管理 - 活动参与进度 - 中奖进度 - 导出
func winningListExport(c *bm.Context) {
	var (
		err     error
		arg     = new(voguemdl.CreditSearch)
		infoStr [][]string
	)
	if err = c.Bind(arg); err != nil {
		return
	}

	if infoStr, err = vogueSrv.ExportWinningList(c, arg); err != nil {
		log.Error("vogueSrv.ExportWinningList failed. error(%v)", err)
		c.JSON(nil, err)
		return
	}

	exportCsvParam := &voguemdl.ExportCsvParam{
		FileNameFormat: "时尚活动中奖进度列表.csv",
		Header:         []string{"昵称", "UID", "中奖商品名称", "中奖时间", "收货信息", "是否存在异常"},
		Result:         infoStr,
	}
	exportCsv(c, exportCsvParam)
}

// 管理平台 - 时尚活动管理 - 微信封禁状态查询
func weChatBlockStatus(c *bm.Context) {
	var (
		err error
		arg = new(voguemdl.WeChatCheckReq)
	)
	if err = c.Bind(arg); err != nil {
		return
	}
	c.JSON(vogueSrv.WeChatBlockStatus(c, arg))
}

// exportCsv
func exportCsv(c *bm.Context, exportCsvParam *voguemdl.ExportCsvParam) {
	fileName := fmt.Sprintf(exportCsvParam.FileNameFormat, exportCsvParam.FileNameParams...)
	b := &bytes.Buffer{}
	b.WriteString("\xEF\xBB\xBF")
	wr := csv.NewWriter(b)
	wr.Write(exportCsvParam.Header)
	for i := 0; i < len(exportCsvParam.Result); i++ {
		wr.Write(exportCsvParam.Result[i])
	}
	wr.Flush()
	c.Writer.Header().Set("Content-Type", "text/csv")
	c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment;filename=%s", fileName))
	tet := b.String()
	c.String(200, tet)
}

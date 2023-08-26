package dynamicV2

import (
	"fmt"
	"strings"

	api "go-gateway/app/app-svr/app-dynamic/interface/api/v2"
	mdlv2 "go-gateway/app/app-svr/app-dynamic/interface/model/dynamicV2"
)

// nolint:deadcode,varcheck
const (
	_foldTextPublish  = "展开%d条相关动态"
	_foldTextUnite    = "展开%d条相关动态"
	_foldTypeFrequent = "转发了%d次此动态"
	_foldTextLimit    = "%d 条动态被折叠"
)

// nolint:gocognit
func (s *Service) procFold(list *mdlv2.FoldList, dynCtx *mdlv2.DynamicContext, _ *mdlv2.GeneralParam) []*api.DynamicItem {
	var rsp []*api.DynamicItem
	if dynCtx.ResFolds == nil || len(dynCtx.ResFolds) == 0 {
		for _, item := range list.List {
			rsp = append(rsp, item.Item)
		}
	} else {
		ignore := make(map[string]struct{})
		for index, item := range list.List {
			if _, ignoreOK := ignore[item.Item.Extend.DynIdStr]; ignoreOK {
				continue
			}
			foldItem, ok := dynCtx.ResFolds[item.Item.Extend.DynIdStr]
			if !ok {
				rsp = append(rsp, item.Item)
				continue
			}
			foldArr := []*mdlv2.FoldItem{item}
			for i := index + 1; i < len(list.List); i++ {
				foldItemTmp, ok := dynCtx.ResFolds[list.List[i].Item.Extend.DynIdStr]
				if !ok || foldItem.Group != foldItemTmp.Group {
					continue
				}
				foldArr = append(foldArr, list.List[i])
			}
			//nolint:exhaustive
			switch foldItem.FoldType {
			case api.FoldType_FoldTypeLimit: // 受限折叠：原地全部折叠
				show := foldArr[0]
				ignore[show.Item.Extend.DynIdStr] = struct{}{}
				var dynIds []string
				dynIds = append(dynIds, show.Item.Extend.DynIdStr)
				for i := 1; i < len(foldArr); i++ {
					dynIds = append(dynIds, foldArr[i].Item.Extend.DynIdStr)
					ignore[foldArr[i].Item.Extend.DynIdStr] = struct{}{}
				}
				show.Item.CardType = api.DynamicType_fold
				show.Item.Extend.OrigDynType = api.DynamicType_fold
				module := &api.Module{
					ModuleType: api.DynModuleType_module_fold,
					ModuleItem: &api.Module_ModuleFold{
						ModuleFold: &api.ModuleFold{
							FoldType: foldItem.FoldType,
							Text:     s.getMixFoldText(foldItem.FoldType, len(dynIds), foldItem.Statement),
							FoldIds:  strings.Join(dynIds, ","),
						},
					},
				}
				show.Item.Modules = []*api.Module{module}
				rsp = append(rsp, show.Item)
			case api.FoldType_FoldTypeUnite: // 联合投稿折叠 第一个展示 其他折在第一个后面
				fallthrough
			case api.FoldType_FoldTypeFrequent: // 转发超频折叠 第一个展示 其他折在第一个后面
				if len(foldArr) == 1 {
					rsp = append(rsp, foldArr[0].Item)
					break
				}
				one := foldArr[0]
				ignore[one.Item.Extend.DynIdStr] = struct{}{}
				var (
					dynIds   []string
					users    []*api.UserInfo
					foldMids = make(map[int64]struct{})
				)
				for i := 1; i < len(foldArr); i++ {
					dynIds = append(dynIds, foldArr[i].Item.Extend.DynIdStr)
					for _, module := range foldArr[i].Item.Modules {
						switch module.ModuleType {
						case api.DynModuleType_module_author:
							mdlAuthor := module.ModuleItem.(*api.Module_ModuleAuthor)
							if _, ok := foldMids[mdlAuthor.ModuleAuthor.Author.Mid]; ok {
								continue
							}
							users = append(users, mdlAuthor.ModuleAuthor.Author)
							foldMids[mdlAuthor.ModuleAuthor.Author.Mid] = struct{}{}
						}
					}
					ignore[foldArr[i].Item.Extend.DynIdStr] = struct{}{}
				}
				module := &api.Module{
					ModuleType: api.DynModuleType_module_fold,
					ModuleItem: &api.Module_ModuleFold{
						ModuleFold: &api.ModuleFold{
							FoldType:  foldItem.FoldType,
							Text:      s.getMixFoldText(foldItem.FoldType, len(dynIds), foldItem.Statement),
							FoldIds:   strings.Join(dynIds, ","),
							FoldUsers: users,
						},
					},
				}
				one.Item.Modules = append(one.Item.Modules, module)
				rsp = append(rsp, one.Item)
			case api.FoldType_FoldTypePublish: // 发布超频折叠 第一个展示 其他折在第一个后面
				// nolint:gomnd
				if len(foldArr) < 3 {
					for _, item := range foldArr {
						rsp = append(rsp, item.Item)
						ignore[item.Item.Extend.DynIdStr] = struct{}{}
					}
					break
				}
				one := foldArr[0]
				ignore[one.Item.Extend.DynIdStr] = struct{}{}
				var dynIds []string
				for i := 1; i < len(foldArr); i++ {
					dynIds = append(dynIds, foldArr[i].Item.Extend.DynIdStr)
					ignore[foldArr[i].Item.Extend.DynIdStr] = struct{}{}
				}
				module := &api.Module{
					ModuleType: api.DynModuleType_module_fold,
					ModuleItem: &api.Module_ModuleFold{
						ModuleFold: &api.ModuleFold{
							FoldType: foldItem.FoldType,
							Text:     s.getMixFoldText(foldItem.FoldType, len(dynIds), foldItem.Statement),
							FoldIds:  strings.Join(dynIds, ","),
						},
					},
				}
				one.Item.Modules = append(one.Item.Modules, module)
				rsp = append(rsp, one.Item)
			}
		}
	}
	return rsp
}

func (s *Service) getMixFoldText(foldType api.FoldType, num int, statement string) string {
	if foldType == api.FoldType_FoldTypePublish {
		return fmt.Sprintf(_foldTextPublish, num)
	}
	if foldType == api.FoldType_FoldTypeFrequent {
		return fmt.Sprintf(_foldTypeFrequent, num)
	}
	if foldType == api.FoldType_FoldTypeUnite {
		return fmt.Sprintf(_foldTextUnite, num)
	}
	if foldType == api.FoldType_FoldTypeLimit {
		return statement
	}
	return ""
}

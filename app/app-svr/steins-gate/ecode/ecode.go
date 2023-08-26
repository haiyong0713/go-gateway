package ecode

import (
	xecode "go-common/library/ecode"
)

var (
	SteinsCidNotMatch          = xecode.New(99000) // 节点Cid信息不匹配
	NonValidGraph              = xecode.New(99001) // 互动视频：当前稿件无可用剧情图
	NotSteinsGateArc           = xecode.New(99002) // 当前稿件不是互动视频
	GraphInvalid               = xecode.New(99003) // 剧情图被修改已失效
	GraphAidEmpty              = xecode.New(99004) // 剧情图缺少aid参数
	GraphNotOwner              = xecode.New(99005) // 你不是该稿件的作者
	GraphAidAttrErr            = xecode.New(99006) // 该稿件类型不是互动视频
	GraphScriptEmpty           = xecode.New(99007) // 剧情图缺少scpirt数据
	GraphNodeCntErr            = xecode.New(99008) // 剧情图节点数错误
	GraphNodeCidEmpty          = xecode.New(99009) // 剧情图节点缺少cid
	GraphNodeNameErr           = xecode.New(99010) // 剧情图节点名称长度不对
	GraphNodeNameExist         = xecode.New(99011) // 剧情图节点名称重复
	GraphDefaultNodeErr        = xecode.New(99012) // 剧情图有多个默认节点
	GraphEdgeCntErr            = xecode.New(99013) // 剧情图节点分支选项数错误
	GraphLackStartNode         = xecode.New(99014) // 剧情图缺少开始节点
	GraphEdgeNameErr           = xecode.New(99015) // 剧情图节点名称长度不对
	GraphDefaultEdgeErr        = xecode.New(99016) // 剧情图节点多个默认分支选项
	GraphFilterHitErr          = xecode.New(99017) // 剧情图有内容命中敏感词
	GraphNodeOtypeErr          = xecode.New(99018) // 剧情图节点类型错误
	GraphNodeCircle            = xecode.New(99019) // 剧情图节点有回环结构
	GraphEdgeToNodeErr         = xecode.New(99020) // 剧情图节点分支无到达节点
	GraphFilterErr             = xecode.New(99021) // 请求过滤词服务失败
	GraphShowTimeEdgeErr       = xecode.New(99022) // 剧情图直连只支持一个选项
	GraphArcStateErr           = xecode.New(99023) // 该稿件暂未过审，请耐心等待稿件过审后再提交
	GraphPageWidthErr          = xecode.New(99024) // 暂不支持竖屏视频，请更换含有【%s】的剧情后再进行提交
	GraphRegVarsErr            = xecode.New(99025) // 剧情图数值相关参数错误
	GraphRegVarsLenErr         = xecode.New(99026) // 剧情图数值数量错误
	GraphRegVarsTypeErr        = xecode.New(99027) // 剧情图数值类型错误
	GraphRegVarsRangeErr       = xecode.New(99028) // 剧情图常规数值超过区间
	GraphRegNormalVarsLenErr   = xecode.New(99029) // 剧情图常规数值数量错误
	GraphRegRandomVarsLenErr   = xecode.New(99030) // 剧情图随机数值数量错误
	GraphRegVarsIDRepeat       = xecode.New(99031) // 剧情图数值ID重复
	GraphRegVarsNameRepeat     = xecode.New(99032) // 剧情图数值名称重复
	GraphRegVarsNameLenErr     = xecode.New(99033) // 剧情图数值名称长度过长
	GraphVideoupArcErr         = xecode.New(99034) // 剧情图获取稿件数据失败
	GraphEdgeCondLenErr        = xecode.New(99035) // 剧情图判断条件数量错误
	GraphEdgeAttrLenErr        = xecode.New(99036) // 剧情图数值变化数量错误
	GraphEdgeAttrVarIDNone     = xecode.New(99037) // 剧情图数值变化缺少id
	GraphEdgeAttrVarIDErr      = xecode.New(99038) // 剧情图数值变化id错误
	GraphEdgeAttrTypeNone      = xecode.New(99039) // 剧情图数值变化类型不存在
	GraphEdgeCondVarIDNone     = xecode.New(99040) // 剧情图判断条件缺少id
	GraphEdgeCondVarIDErr      = xecode.New(99041) // 剧情图判断条件id错误
	GraphEdgeCondTypeNone      = xecode.New(99042) // 剧情图判断条件类型不存在
	GraphAttributeErr          = xecode.New(99043) // 剧情图数值变化数据错误
	GraphConditionErr          = xecode.New(99044) // 剧情图判断条件数据错误
	GraphPreviewStateErr       = xecode.New(99045) // 当前稿件暂未通过审核
	GraphNodeIDErr             = xecode.New(99046) // 剧情图节点ID错误
	GraphVarIDErr              = xecode.New(99047) // 剧情图数值ID错误
	GraphEdgeAttrRepeat        = xecode.New(99048) // 剧情图数值变化有重复数值
	GraphEdgeAttrTypeErr       = xecode.New(99049) // 剧情图数值变化仅支持普通变量
	GraphEdgeCondVarLenErr     = xecode.New(99050) // 剧情图常规数值仅支持%d个判断条件
	GraphEdgeCondRandRangeErr  = xecode.New(99051) // 剧情图随机数值判断条件超出数值范围
	GraphEdgeCondExclusion     = xecode.New(99052) // 剧情图判断条件冲突
	GraphEdgeToNodeNotFound    = xecode.New(99053) // 剧情图节点到达节点不存在
	GraphScriptTooLong         = xecode.New(99054) // 剧情图script长度太长
	GraphHiddenVarRecordNilErr = xecode.New(99055) // 隐藏变量存档为空
	GraphNodeIDExist           = xecode.New(99056) // 剧情图节点ID重复
	GraphCidNotDispatched      = xecode.New(99057) // 请等待视频转码分发完成后再编辑定点位
	GraphEdgeTextAlignErr      = xecode.New(99058) // 不支持的点定位位置
	GraphOgvNotAllowed         = xecode.New(99059) // Ogv视频不允许播放
	GraphGetDimensionErr       = xecode.New(99060) // 获取视频分辨率失败
	GraphSendAuditErr          = xecode.New(99061) // 提交审核失败
	GraphSkinNotFound          = xecode.New(99062) // 使用了不存在的皮肤
	GraphLoopRecordErr         = xecode.New(99063) // 开环存档错误
	GraphRestrictBuildErr      = xecode.New(99064) // 中插/表达式剧情树无法通过nodeinfo请求报错（用于web端，国际版，小程序等）
	ExprBadToken               = xecode.New(99070) // &|!等双元运算符不连贯，变量名错误
	ExprUnexpectedChar         = xecode.New(99071) // 出现字符使得表达式无法识别
	// ExprVarNotDeclared         = xecode.New(99073) // 变量未声明即使用
	ExprMissRightParenthesis = xecode.New(99074) // 缺少右括号
	ExprPrimaryExpected      = xecode.New(99075)
	ExprDivideByZero         = xecode.New(99076) // 除以0
	AidBvidNil               = xecode.New(99077) // 请输入aid/bvid
	BvidIllegal              = xecode.New(99078) // bvid非法

)

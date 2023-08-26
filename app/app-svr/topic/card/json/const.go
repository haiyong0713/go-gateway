package jsonwebcard

type TopicCardType string

const (
	TopicCardTypeDynamicAv      = TopicCardType("DYNAMIC_TYPE_AV")
	TopicCardTypeDynamicForward = TopicCardType("DYNAMIC_TYPE_FORWARD")
	TopicCardTypeDynamicDraw    = TopicCardType("DYNAMIC_TYPE_DRAW")
	TopicCardTypeDynamicWord    = TopicCardType("DYNAMIC_TYPE_WORD")
	TopicCardTypeDynamicArticle = TopicCardType("DYNAMIC_TYPE_ARTICLE")
	TopicCardTypeDynamicCommon  = TopicCardType("DYNAMIC_TYPE_COMMON")
	TopicCardTypeDynamicPGC     = TopicCardType("DYNAMIC_TYPE_PGC")
	TopicCardTypeFold           = TopicCardType("FOLD")
)

type CardType string

const (
	CardDynamicTypeAv      = CardType("DYNAMIC_TYPE_AV")
	CardDynamicTypeForward = CardType("DYNAMIC_TYPE_FORWARD")
	CardDynamicTypeDraw    = CardType("DYNAMIC_TYPE_DRAW")
	CardDynamicTypeWord    = CardType("DYNAMIC_TYPE_WORD")
	CardDynamicTypeArticle = CardType("DYNAMIC_TYPE_ARTICLE")
	CardDynamicTypeCommon  = CardType("DYNAMIC_TYPE_COMMON")
	CardDynamicTypePGC     = CardType("DYNAMIC_TYPE_PGC")
)

type FoldType string

const (
	FoldDynamicTypePublish  = FoldType("FOLD_TYPE_PUBLISH")
	FoldDynamicTypeFrequent = FoldType("FOLD_TYPE_FREQUENT")
	FoldDynamicTypeUnite    = FoldType("FOLD_TYPE_UNITE")
	FoldDynamicTypeLimit    = FoldType("FOLD_TYPE_LIMIT")
)

type AuthorType string

const (
	AuthorTypeNormal = AuthorType("AUTHOR_TYPE_NORMAL")
	AuthorTypePgc    = AuthorType("AUTHOR_TYPE_PGC")
)

type RichTextNodeType string

const (
	RichTextNodeTypeText       = RichTextNodeType("RICH_TEXT_NODE_TYPE_TEXT")
	RichTextNodeTypeAt         = RichTextNodeType("RICH_TEXT_NODE_TYPE_AT")
	RichTextNodeTypeLottery    = RichTextNodeType("RICH_TEXT_NODE_TYPE_LOTTERY")
	RichTextNodeTypeVote       = RichTextNodeType("RICH_TEXT_NODE_TYPE_VOTE")
	RichTextNodeTypeTopic      = RichTextNodeType("RICH_TEXT_NODE_TYPE_TOPIC")
	RichTextNodeTypeGoods      = RichTextNodeType("RICH_TEXT_NODE_TYPE_GOODS")
	RichTextNodeTypeBv         = RichTextNodeType("RICH_TEXT_NODE_TYPE_BV")
	RichTextNodeTypeAv         = RichTextNodeType("RICH_TEXT_NODE_TYPE_AV")
	RichTextNodeTypeEmoji      = RichTextNodeType("RICH_TEXT_NODE_TYPE_EMOJI")
	RichTextNodeTypeUser       = RichTextNodeType("RICH_TEXT_NODE_TYPE_USER")
	RichTextNodeTypeCv         = RichTextNodeType("RICH_TEXT_NODE_TYPE_CV")
	RichTextNodeTypeVc         = RichTextNodeType("RICH_TEXT_NODE_TYPE_VC")
	RichTextNodeTypeWeb        = RichTextNodeType("RICH_TEXT_NODE_TYPE_WEB")
	RichTextNodeTypeTaobao     = RichTextNodeType("RICH_TEXT_NODE_TYPE_TAOBAO")
	RichTextNodeTypeMail       = RichTextNodeType("RICH_TEXT_NODE_TYPE_MAIL")
	RichTextNodeTypeOgvSeason  = RichTextNodeType("RICH_TEXT_NODE_TYPE_SEASON")
	RichTextNodeTypeOgvEp      = RichTextNodeType("RICH_TEXT_NODE_TYPE_EP")
	RichTextNodeTypeSearchWord = RichTextNodeType("RICH_TEXT_NODE_TYPE_WORD")
)

type MajorType string

const (
	MajorTypeArchive = MajorType("MAJOR_TYPE_ARCHIVE")
	MajorTypeDraw    = MajorType("MAJOR_TYPE_DRAW")
	MajorTypeArticle = MajorType("MAJOR_TYPE_ARTICLE")
	MajorTypeCommon  = MajorType("MAJOR_TYPE_COMMON")
	MajorTypePGC     = MajorType("MAJOR_TYPE_PGC")
)

type MediaType string

const (
	MediaTypeUgc = MediaType("MediaTypeUgc")
)

type DrawTagType string

const (
	DrawTagTypeCommon = DrawTagType("DRAW_TAG_TYPE_COMMON")
	DrawTagTypeGoods  = DrawTagType("DRAW_TAG_TYPE_GOODS")
	DrawTagTypeUser   = DrawTagType("DRAW_TAG_TYPE_USER")
	DrawTagTypeTopic  = DrawTagType("DRAW_TAG_TYPE_TOPIC")
	DrawTagTypeLbs    = DrawTagType("DRAW_TAG_TYPE_LBS")
)

type AdditionalType string

const (
	AdditionalTypeVote    = AdditionalType("ADDITIONAL_TYPE_VOTE")
	AdditionalTypeReserve = AdditionalType("ADDITIONAL_TYPE_RESERVE")
	AdditionalTypeGoods   = AdditionalType("ADDITIONAL_TYPE_GOODS")
)

type ThreePointType string

const (
	ThreePointItemTopicIrrelevant = ThreePointType("THREE_POINT_TOPIC_IRRELEVANT")
	ThreePointItemReport          = ThreePointType("THREE_POINT_TOPIC_REPORT")
)

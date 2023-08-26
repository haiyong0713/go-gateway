package model

type ModuleType string

func (mt ModuleType) String() string {
	return string(mt)
}

type CardType string

func (ct CardType) String() string {
	return string(ct)
}

const (
	// 组件类型：proto
	ModuleTypeEditor          = ModuleType("editor_recommend_module")
	ModuleTypeParticipation   = ModuleType("participation_module")
	ModuleTypeHeader          = ModuleType("header_module")
	ModuleTypeDynamic         = ModuleType("dynamic_module")
	ModuleTypeLive            = ModuleType("live_module")
	ModuleTypeCarouselImg     = ModuleType("carousel_img_module")
	ModuleTypeCarouselWord    = ModuleType("carousel_word_module")
	ModuleTypeResource        = ModuleType("resource_module")
	ModuleTypeGame            = ModuleType("game_module")
	ModuleTypeVideo           = ModuleType("video_module")
	ModuleTypeRcmd            = ModuleType("recommend_module")
	ModuleTypeRcmdVertical    = ModuleType("recommend_vertical_module")
	ModuleTypeRelact          = ModuleType("relativeact_module")
	ModuleTypeRelactCapsule   = ModuleType("relativeact_capsule_module")
	ModuleTypeStatement       = ModuleType("statement_module")
	ModuleTypeIcon            = ModuleType("icon_module")
	ModuleTypeVote            = ModuleType("vote_module")
	ModuleTypeReserve         = ModuleType("reserve_module")
	ModuleTypeTimeline        = ModuleType("timeline_module")
	ModuleTypeOgv             = ModuleType("ogv_module")
	ModuleTypeNavigation      = ModuleType("navigation_module")
	ModuleTypeReply           = ModuleType("reply_module")
	ModuleTypeTab             = ModuleType("tab_module")
	ModuleTypeNewactHeader    = ModuleType("newact_header_module")
	ModuleTypeNewactAward     = ModuleType("newact_award_module")
	ModuleTypeNewactStatement = ModuleType("newact_statement_module")
	ModuleTypeProgress        = ModuleType("progress_module")
	ModuleTypeSelect          = ModuleType("select_module")
	ModuleTypeClick           = ModuleType("click_module")
	ModuleTypeHoverButton     = ModuleType("hover_button_module")
	ModuleTypeBottomButton    = ModuleType("bottom_button_module")
	// 组件卡片类型：proto
	CardTypeEditor                 = CardType("editor_recommend_card")
	CardTypeParticipation          = CardType("participation_card")
	CardTypeHeader                 = CardType("header_card")
	CardTypeDynamic                = CardType("dynamic_card")
	CardTypeText                   = CardType("text_card")
	CardTypeTextTitle              = CardType("text_title_card")
	CardTypeImageTitle             = CardType("image_title_card")
	CardTypeDynamicMore            = CardType("dynamic_more_card")
	CardTypeDynamicActMore         = CardType("dynamic_act_more_card")
	CardTypeLive                   = CardType("live_card")
	CardTypeCarouselImg            = CardType("carousel_img_card")
	CardTypeCarouselWord           = CardType("carousel_word_card")
	CardTypeResource               = CardType("resource_card")
	CardTypeResourceMore           = CardType("resource_more_card")
	CardTypeGame                   = CardType("game_card")
	CardTypeVideo                  = CardType("video_card")
	CardTypeVideoMore              = CardType("video_more_card")
	CardTypeRcmd                   = CardType("recommend_card")
	CardTypeRcmdVertical           = CardType("recommend_vertical_card")
	CardTypeRelact                 = CardType("relativeact_card")
	CardTypeRelactCapsule          = CardType("relativeact_capsule_card")
	CardTypeStatement              = CardType("statement_card")
	CardTypeIcon                   = CardType("icon_card")
	CardTypeVote                   = CardType("vote_card")
	CardTypeReserve                = CardType("reserve_card")
	CardTypeTimelineHead           = CardType("timeline_head_card")
	CardTypeTimelineEventText      = CardType("timeline_event_text_card")
	CardTypeTimelineEventImage     = CardType("timeline_event_image_card")
	CardTypeTimelineEventImagetext = CardType("timeline_event_imagetext_card")
	CardTypeTimelineEventResource  = CardType("timeline_event_resource_card")
	CardTypeTimelineMore           = CardType("timeline_more_card")
	CardTypeTimelineUnfold         = CardType("timeline_unfold_card")
	CardTypeOgvOne                 = CardType("ogv_one_card")
	CardTypeOgvThree               = CardType("ogv_three_card")
	CardTypeOgvMore                = CardType("ogv_more_card")
	CardTypeNavigation             = CardType("navigation_card")
	CardTypeReply                  = CardType("reply_card")
	CardTypeTab                    = CardType("tab_card")
	CardTypeSelect                 = CardType("select_card")
	CardTypeNewactHeader           = CardType("newact_header_card")
	CardTypeNewactAward            = CardType("newact_award_card")
	CardTypeNewactStatement        = CardType("newact_statement_card")
	CardTypeProgress               = CardType("progress_card")
	CardTypeClick                  = CardType("click_card")
	CardTypeHoverButton            = CardType("hover_button_card")
)
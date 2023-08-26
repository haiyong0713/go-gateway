package card

import (
	"fmt"

	"go-gateway/app/app-svr/app-card/interface/model/stat"

	"github.com/pkg/errors"
)

const (
	ResourceArchive        = "archive"
	ResourceSeason         = "season"
	ResourceEpisode        = "episode"
	ResourceRoom           = "room"
	ResourceShopping       = "shopping"
	ResourceAudio          = "audio"
	ResourceArticle        = "article"
	ResourcePicture        = "picture"
	ResourceTag            = "tag"
	ResourceItems          = "items"
	ResourceMoe            = "moe"
	ResourceBangumiUpdate  = "bangumi_update"
	ResourceBangumiRemind  = "bangumi_remind"
	ResourceRoomGroup      = "room_group"
	ResourceOperateCard    = "operate_card"
	ResourceVipInfo        = "vip_info"
	ResourceTunnelFeedCard = "feed_card"
	ResourceTopic          = "topic"
	ResourceAccount        = "account"
	ResourceDynamic        = "dynamic"
	ResourceAI             = "ai"
	ResourceOption         = "option"
	ResourceBannerInline   = "banner_inline"
	ResourceInlineAv       = "inline_av"
	ResourceInlinePGC      = "inline_pgc"
	ResourceInlineLive     = "inline_live"
)

type errResourceNotExist struct {
	Resource string
	ID       int64
}
type errUnexpectedGoto struct {
	Goto string
}
type errInsufficientClientVersion struct{}
type errEmptyOP struct{}
type errInvalidResource struct {
	Resource string
	ID       int64
}
type errUnexpectedCardGoto struct {
	CardGoto string
}
type errUnexpectedResourceType struct {
	ActualType string
}

func (e errInsufficientClientVersion) Error() string {
	return "errInsufficientClientVersion"
}

func (e errEmptyOP) Error() string {
	return "errEmptyOP"
}

func (e errResourceNotExist) Error() string {
	return fmt.Sprintf("errResourceNotExist: [%s %d]", e.Resource, e.ID)
}

func (e errUnexpectedGoto) Error() string {
	return fmt.Sprintf("errUnexpectedGoto: %s", e.Goto)
}

func (e errUnexpectedResourceType) Error() string {
	return fmt.Sprintf("errUnexpectedResourceType: %s", e.ActualType)
}

func (e errUnexpectedCardGoto) Error() string {
	return fmt.Sprintf("errUnexpectedCardGoto: %s", e.CardGoto)
}

func (e errInvalidResource) Error() string {
	return fmt.Sprintf("errInvalidResource: [%s %d]", e.Resource, e.ID)
}

// StatBuildCardErr is
func StatBuildCardErr(err error, rowType, gt, jumpGt, cardType string) {
	var reason string
	switch errObj := errors.Cause(err).(type) {
	case *errResourceNotExist:
		reason = fmt.Sprintf("%s_resource_not_exist", errObj.Resource)
	case *errUnexpectedGoto:
		reason = "unexpected_goto"
	case *errInsufficientClientVersion:
		reason = "insufficient_client_version"
	case *errEmptyOP:
		reason = "empty_op"
	case *errInvalidResource:
		reason = fmt.Sprintf("invalid_%s", errObj.Resource)
	case *errUnexpectedCardGoto:
		reason = "unexpected_card_goto"
	case *errUnexpectedResourceType:
		reason = "unexpected_resource_type"
	default:
		reason = "unexpected_build_card_error"
	}
	stat.MetricDiscardCardTotal.Inc(rowType, gt, jumpGt, cardType, reason)
}

func newEmptyOPErr() error {
	return errors.WithStack(&errEmptyOP{})
}
func newInsufficientClientVersionErr(format string, args ...interface{}) error {
	return errors.Wrapf(&errInsufficientClientVersion{}, format, args...)
}
func newInvalidResourceErr(resource string, id int64, format string, args ...interface{}) error {
	return errors.Wrapf(&errInvalidResource{
		Resource: resource,
		ID:       id,
	}, format, args...)
}
func newResourceNotExistErr(resource string, id int64) error {
	return errors.WithStack(&errResourceNotExist{
		Resource: resource,
		ID:       id,
	})
}
func newUnexpectedResourceTypeErr(input interface{}, format string, args ...interface{}) error {
	return errors.Wrapf(&errUnexpectedResourceType{
		ActualType: fmt.Sprintf("%T", input),
	}, format, args...)
}

//nolint:unparam
func newUnexpectedCardGotoErr(input string, format string, args ...interface{}) error {
	return errors.Wrapf(&errUnexpectedCardGoto{CardGoto: input}, format, args...)
}

//nolint:unparam
func newUnexpectedGotoErr(input string, format string, args ...interface{}) error {
	return errors.Wrapf(&errUnexpectedGoto{Goto: input}, format, args...)
}

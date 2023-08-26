package act

import "go-gateway/app/web-svr/native-page/interface/api"

const (
	_tabPageModule = "page_module"
	_tabUrlModule  = "direct_module"
)

type TabReply struct {
	ErrLimit *ErrLimit `json:"err_limit,omitempty"`
	Tab      *Tab      `json:"tab,omitempty"`
}

// Tab
type Tab struct {
	BgType        int32        `json:"bg_type,omitempty"`
	BgImg         string       `json:"bg_img,omitempty"`
	BgColor       string       `json:"bg_color,omitempty"`
	IconType      int32        `json:"icon_type,omitempty"`
	ActiveColor   string       `json:"active_color,omitempty"`
	InactiveColor string       `json:"inactive_color,omitempty"`
	Items         []*TabModule `json:"items,omitempty"`
}

type TabModule struct {
	Goto        string `json:"goto,omitempty"` //category
	ActiveImg   string `json:"active_img,omitempty"`
	InactiveImg string `json:"inactive_img,omitempty"`
	Pid         int64  `json:"pid,omitempty"`
	URL         string `json:"url,omitempty"`
	Select      bool   `json:"select,omitempty"` //是否选中
	ShareOrigin string `json:"share_origin,omitempty"`
	TabID       int64  `json:"tab_id,omitempty"`
	TabModuleID int64  `json:"tab_module_id,omitempty"`
	Title       string `json:"title,omitempty"`
	TopicName   string `json:"topic_name,omitempty"`
	ForeignID   int64  `json:"foreign_id,omitempty"`
}

// 错误提示信息
type ErrLimit struct {
	Code    int     `json:"code,omitempty"`
	Message string  `json:"message,omitempty"`
	Button  *Button `json:"button,omitempty"`
}

// Button .
type Button struct {
	Title string `json:"title,omitempty"`
	Link  string `json:"link,omitempty"`
}

func (out *TabModule) FormatTabModule(in *api.NativeTabModule) bool {
	out.Title = in.Title
	switch {
	case in.IsTabPage():
		if in.Pid <= 0 {
			return false
		}
		out.Goto = _tabPageModule
		out.Pid = in.Pid
	case in.IsTabUrl():
		if in.URL == "" {
			return false
		}
		out.Goto = _tabUrlModule
		out.URL = in.URL
	default:
		return false
	}
	out.ActiveImg = in.ActiveImg
	out.InactiveImg = in.InactiveImg
	out.TabID = in.TabID
	// 版本覆盖率85%以上才下发新tab聚合页面地址 OriginTab
	out.ShareOrigin = SimpleTab
	out.TabModuleID = in.ID
	return true
}

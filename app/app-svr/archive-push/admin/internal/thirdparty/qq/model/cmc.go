package model

import "regexp"

const (
	DefaultSSource = "bilibili"
)

type ModifyPGCAdminMode int

const (
	ModifyModeModify   = ModifyPGCAdminMode(0)
	ModifyModeWithdraw = ModifyPGCAdminMode(1)
)

type BasePGCAdminCType string

const (
	CTypeInfo  BasePGCAdminCType = "1"
	CTypeVideo BasePGCAdminCType = "2"
)

// SyncUserAction 用户认证/绑定状态回流行为类型
type SyncUserAction string

const (
	SyncUserActionBindStart   SyncUserAction = "bindStart"
	SyncUserActionBindSuccess SyncUserAction = "bindSuccess"
	SyncUserActionBindReject  SyncUserAction = "bindReject"
	SyncUserActionBindCancel  SyncUserAction = "bindCancel"
)

type SyncArchiveStatus string

const (
	SyncArchiveStatusReceived SyncArchiveStatus = "received"
	SyncArchiveStatusPassed   SyncArchiveStatus = "passed"
	SyncArchiveStatusRejected SyncArchiveStatus = "rejected"
)

var (
	RegPushPGCExistingBVID = regexp.MustCompile("find duplicate sOriginID already exist, docid: (\\d+)")
)

type BasePGCAdminQuery struct {
	CType string `form:"ctype" json:"ctype"` // 1-PGC资讯；2-PGC视频
}

// pushPGCAdmin
type PushPGCAdminReq struct {
	SCreater        string `form:"sCreater" json:"sCreater" validate:"required"`
	SCreaterHeader  string `form:"sCreaterHeader" json:"sCreaterHeader"`
	STitle          string `form:"sTitle" json:"sTitle" validate:"required"`
	SExt9           string `form:"sExt9" json:"sExt9" validate:"required"`
	SIMG            string `form:"sIMG" json:"sIMG"`
	SDESC           string `form:"sDesc" json:"sDesc"`
	SAuthor         string `form:"sAuthor" json:"sAuthor"`
	IType           string `form:"iType" json:"iType"`
	ISubType        string `form:"iSubType" json:"iSubType"`
	SOriginID       string `form:"sOriginID" json:"sOriginID"`
	SURL            string `form:"sUrl" json:"sUrl"`
	SCreated        string `form:"sCreated" json:"sCreated"`
	SCreatedOther   string `form:"sCreatedOther" json:"sCreatedOther"`
	SOuterUserID    string `form:"sOuterUserId" json:"sOuterUserId"`
	SHelperAuthorID string `form:"sHelperAuthorId" json:"sHelperAuthorId"`
	STagIDs         string `form:"sTagIds" json:"sTagIds"`
	STagsOther      string `form:"sTagsOther" json:"sTagsOther"`
	SSource         string `form:"sSource" json:"sSource"`
	SVID            string `form:"sVID" json:"sVID"`
	ITime           string `form:"iTime" json:"iTime"`
	IFrom           string `form:"iFrom" json:"iFrom"` // 11=b站
	SVideoSize      string `form:"sVideoSize" json:"sVideoSize"`
}

type PushPGCAdminReply struct {
	Status int                    `json:"status"`
	MSG    string                 `json:"msg"`
	Data   *PushPGCAdminReplyData `json:"data"`
}

type PushPGCAdminReplyData struct {
	DocID    string `json:"docid"`
	SCreated string `json:"sCreated"`
}

// modifyPGCAdmin
type ModifyPGCAdminQuery struct {
	*BasePGCAdminQuery
	ID   string `form:"id" json:"id"`     // ModifyPGCAdmin用。内容docid
	Mode int    `form:"mode" json:"mode"` // ModifyPGCAdmin用。0-更新 （仅待审核状态内容允许更新）；1-下架 默认更新，下架时不需要传postdata
}

type ModifyPGCAdminReq struct {
	SCreater        string `form:"sCreater" json:"sCreater,omitempty"`
	SCreaterHeader  string `form:"sCreaterHeader" json:"sCreaterHeader,omitempty"`
	STitle          string `form:"sTitle" json:"sTitle,omitempty"`
	SExt9           string `form:"sExt9" json:"sExt9,omitempty"`
	SIMG            string `form:"sIMG" json:"sIMG,omitempty"`
	SDESC           string `form:"sDesc" json:"sDesc,omitempty"`
	SAuthor         string `form:"sAuthor" json:"sAuthor,omitempty"`
	IType           string `form:"iType" json:"iType,omitempty"`
	ISubType        string `form:"iSubType" json:"iSubType,omitempty"`
	SOriginID       string `form:"sOriginID" json:"sOriginID,omitempty"`
	SURL            string `form:"sUrl" json:"sUrl,omitempty"`
	SCreated        string `form:"sCreated" json:"sCreated,omitempty"`
	SCreatedOther   string `form:"sCreatedOther" json:"sCreatedOther,omitempty"`
	SOuterUserID    string `form:"sOuterUserId" json:"sOuterUserId,omitempty"`
	SHelperAuthorID string `form:"sHelperAuthorId" json:"sHelperAuthorId,omitempty"`
	STagIDs         string `form:"sTagIds" json:"sTagIds,omitempty"`
	STagsOther      string `form:"sTagsOther" json:"sTagsOther,omitempty"`
	SSource         string `form:"sSource" json:"sSource,omitempty"`
	SContent        string `form:"sContent" json:"sContent,omitempty"` // 资讯用。视频不用此字段
	SVID            string `form:"sVID" json:"sVID,omitempty"`
	ITime           string `form:"iTime" json:"iTime,omitempty"`
	IFrom           string `form:"iFrom" json:"iFrom,omitempty"` // 11=b站
	SVideoSize      string `form:"sVideoSize" json:"sVideoSize,omitempty"`
}

type ModifyPGCAdminReply struct {
	Status int                      `json:"status"`
	MSG    string                   `json:"msg"`
	Data   *ModifyPGCAdminReplyData `json:"data"`
}

type ModifyPGCAdminReplyData struct {
	ID string `json:"id"`
}

// detailAdmin
type DetailAdminQuery struct {
	*BasePGCAdminQuery
	ID string `form:"id" json:"id"` // ModifyPGCAdmin用。内容docid
}

type DetailAdminReq struct {
	SCreater        string `form:"sCreater" json:"sCreater"`
	SCreaterHeader  string `form:"sCreaterHeader" json:"sCreaterHeader"`
	STitle          string `form:"sTitle" json:"sTitle"`
	SExt9           string `form:"sExt9" json:"sExt9"`
	SIMG            string `form:"sIMG" json:"sIMG"`
	SDESC           string `form:"sDesc" json:"sDesc"`
	SAuthor         string `form:"sAuthor" json:"sAuthor"`
	IType           string `form:"iType" json:"iType"`
	ISubType        string `form:"iSubType" json:"iSubType"`
	SOriginID       string `form:"sOriginID" json:"sOriginID"`
	SURL            string `form:"sUrl" json:"sUrl"`
	SCreated        string `form:"sCreated" json:"sCreated"`
	SCreatedOther   string `form:"sCreatedOther" json:"sCreatedOther"`
	SOuterUserID    string `form:"sOuterUserId" json:"sOuterUserId"`
	SHelperAuthorID string `form:"sHelperAuthorId" json:"sHelperAuthorId"`
	STagIDs         string `form:"sTagIds" json:"sTagIds"`
	STagsOther      string `form:"sTagsOther" json:"sTagsOther"`
	SSource         string `form:"sSource" json:"sSource"`
	SContent        string `form:"sContent" json:"sContent"` // 资讯用。视频不用此字段
	SVID            string `form:"sVID" json:"sVID"`
	ITime           string `form:"iTime" json:"iTime"`
	IFrom           string `form:"iFrom" json:"iFrom"` // 11=b站
	SVideoSize      string `form:"sVideoSize" json:"sVideoSize"`
}

type DetailAdminReply struct {
	Status int                     `json:"status"`
	MSG    string                  `json:"msg"`
	Data   []*DetailAdminReplyData `json:"data"`
}

type DetailAdminReplyData struct {
	STitle           string        `json:"sTitle"`
	SDesc            string        `json:"sDesc"`
	ILD              int           `json:"iId"`
	IBiz             int           `json:"iBiz"`
	IStatus          int           `json:"iStatus"`
	SExt9            int           `json:"sExt9"`
	SIMG             string        `json:"sIMG"`
	SIDxTime         string        `json:"sIdxTime"`
	SCreated         string        `json:"sCreated"`
	IInfoType        int           `json:"iInfoType"`
	SAuthor          string        `json:"sAuthor"`
	IDocID           string        `json:"iDocID"`
	STagIDs          []int         `json:"sTagIds"`
	SVID             string        `json:"sVID"`
	SUrl             string        `json:"sUrl"`
	ITime            int           `json:"sTime"`
	AuthorID         string        `json:"authorID"`
	ITotalPlay       int           `json:"iTotalPlay"`
	SContent         string        `json:"sContent"`
	SCoverList       []interface{} `json:"sCoverList"`
	SPics            []string      `json:"sPics"`
	SPicSizes        []string      `json:"sPicSizes"`
	SPicTypes        []string      `json:"sPicTypes"`
	SGifHead         []string      `json:"sGifHead"`
	IFrom            int           `json:"iFrom"`
	STagInfo         string        `json:"sTagInfo"`
	SIsVerticalVideo string        `json:"sIsVerticalVideo"`
	SVideoHeight     string        `json:"sVideoHeight"`
	SVideoWidth      string        `json:"sVideoWidth"`
	SVideoFileSize   string        `json:"sVideoFileSize"`
	SIMGSize         string        `json:"sIMGSize"`
	SIMGType         string        `json:"sIMGType"`
	SOrgID           string        `json:"sOrgId"`
	SOuterHelperID   string        `json:"sOuterHelperId"`
	Cache            int           `json:"cache"`
}

// userContentListAdmin
type UserContentListAdminQuery struct {
	*BasePGCAdminQuery
	Creater  string `form:"creater" json:"creater"`   // 查询用。作者ID
	Page     int    `form:"page" json:"page"`         // 页码。默认为1
	PageSize int    `form:"pagesize" json:"pagesize"` // 每页内容数，默认10
}

type UserContentListAdminReply struct {
	Status int                              `json:"status"`
	MSG    string                           `json:"msg"`
	Data   []*UserContentListAdminReplyData `json:"data"`
}

type UserContentListAdminReplyData struct {
	STitle           string        `json:"sTitle"`
	SDesc            string        `json:"sDesc"`
	ILD              int           `json:"iId"`
	IBiz             int           `json:"iBiz"`
	IStatus          int           `json:"iStatus"`
	SExt9            int           `json:"sExt9"`
	SIMG             string        `json:"sIMG"`
	SIDxTime         string        `json:"sIdxTime"`
	SCreated         string        `json:"sCreated"`
	IInfoType        int           `json:"iInfoType"`
	SAuthor          string        `json:"sAuthor"`
	IDocID           string        `json:"iDocID"`
	STagIDs          []int         `json:"sTagIds"`
	SVID             string        `json:"sVID"`
	SUrl             string        `json:"sUrl"`
	ITime            int           `json:"sTime"`
	AuthorID         string        `json:"authorID"`
	ITotalPlay       int           `json:"iTotalPlay"`
	SContent         string        `json:"sContent"`
	SCoverList       []interface{} `json:"sCoverList"`
	SPics            []string      `json:"sPics"`
	SPicSizes        []string      `json:"sPicSizes"`
	SPicTypes        []string      `json:"sPicTypes"`
	SGifHead         []string      `json:"sGifHead"`
	IFrom            int           `json:"iFrom"`
	STagInfo         string        `json:"sTagInfo"`
	SIsVerticalVideo string        `json:"sIsVerticalVideo"`
	SVideoHeight     string        `json:"sVideoHeight"`
	SVideoWidth      string        `json:"sVideoWidth"`
	SVideoFileSize   string        `json:"sVideoFileSize"`
	SIMGSize         string        `json:"sIMGSize"`
	SIMGType         string        `json:"sIMGType"`
	SOrgID           string        `json:"sOrgId"`
	SOuterHelperID   string        `json:"sOuterHelperId"`
	Cache            int           `json:"cache"`
}

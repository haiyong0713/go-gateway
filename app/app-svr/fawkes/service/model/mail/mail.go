package mail

import (
	"mime/multipart"
	xtime "time"
)

// Type for mail
type Type uint8

// Mail types
const (
	TypeTextPlain Type = iota
	TypeTextHTML
)

// Attach def.
type Attach struct {
	Name        string
	File        multipart.File
	ShouldUnzip bool
}

// Mail def.
type Mail struct {
	ToAddresses  []*Address `json:"to_addresses"`
	CcAddresses  []*Address `json:"cc_addresses"`
	BccAddresses []*Address `json:"bcc_addresses"`
	Subject      string     `json:"subject"`
	Body         string     `json:"body"`
	Type         Type       `json:"type"`
}

// Address def.
type Address struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

// Sender def
type Sender struct {
	Host    string
	Port    int
	Address string
	Pwd     string
	Name    string
}

type Attribution struct {
	AppKey     string `json:"app_key"`
	FuncModule string `json:"func_module"`
}

type SenderConfig struct {
	Id         int64      `json:"id"`
	AppKey     string     `json:"app_key"`
	FuncModule string     `json:"func_module"`
	Host       string     `json:"host,omitempty"`
	Port       int        `json:"port,omitempty"`
	Address    string     `json:"address"`
	Pwd        string     `json:"pwd,omitempty"`
	Name       string     `json:"name"`
	Operator   string     `json:"operator"`
	Ctime      xtime.Time `json:"ctime"`
	Mtime      xtime.Time `json:"mtime"`
}

type AppMailWithModule struct {
	AppKey       string `json:"app_key"`
	FuncModule   string `json:"func_module"`
	SenderId     int64  `json:"sender_id"`
	SenderName   string `json:"sender_name"`
	ReceiverId   int64  `json:"receiver_id"`
	ReceiverName string `json:"receiver_name"`
}

const (
	CINotifyGroupMail   = "CI_NOTIFY_GROUP"
	CDReleaseNotifyMail = "CD_RELEASE_NOTIFY"
	HotfixFinishJobMail = "HOTFIX_FINISH_JOB"

	ReceiverWithAll = -1
	ReceiverWithTo  = 0
	ReceiverWithCC  = 1
	ReceiverWithBCC = 2

	AddressSuffix = "@bilibili.com"
)

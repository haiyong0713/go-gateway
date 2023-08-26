package component

// Mail types
const (
	TypeTextPlain Type = iota
	TypeTextHTML
)

// Mail def.
type Mail struct {
	ToAddresses  []*Address `json:"to_addresses"`
	CcAddresses  []*Address `json:"cc_addresses"`
	BccAddresses []*Address `json:"bcc_addresses"`
	Subject      string     `json:"subject"`
	Body         string     `json:"body"`
	Type         Type       `json:"type"`
}

type EmailInfo struct {
	Host         string
	Port         int
	Address      string
	Pwd          string
	Name         string
	ToAddress    []*Address
	CcAddress    []*Address
	BccAddresses []*Address
}

// Address def.
type Address struct {
	Address string `json:"address"`
	Name    string `json:"name"`
}

// Type for mail
type Type uint8

// Base 基础数据
type Base struct {
	Host    string
	Port    int
	Address string
	Pwd     string
	Name    string
}

// Attach def.
type Attach struct {
	Name string
	File string
}

type CorpWeChat struct {
	MentionUserIDs  []string
	MentionUserTels []string
	WebhookUrl      string
}

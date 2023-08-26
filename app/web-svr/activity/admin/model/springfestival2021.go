package model

import "time"

// ActSpringRelation ...
type ActSpringRelation struct {
	ID      int64     `form:"id" json:"id" gorm:"column:id"`
	Mid     int64     `form:"mid" json:"mid" gorm:"column:mid"`
	Invitee int64     `form:"invitee" json:"invitee" gorm:"column:invitee"`
	Ctime   time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime   time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// ActSpringRelationReply ...
type ActSpringRelationReply struct {
	List []*ActSpringRelation   `json:"list"`
	Page map[string]interface{} `json:"page"`
}

// ActSpringCardsNums ...
type ActSpringCardsNums struct {
	ID        int64     `form:"id" json:"id" gorm:"column:id"`
	Mid       int64     `form:"mid" json:"mid" gorm:"column:mid"`
	Card1     int64     `form:"card_1" json:"card_1" gorm:"column:card_1"`
	Card1Used int64     `form:"card_1_used" json:"card_1_used" gorm:"column:card_1_used"`
	Card2     int64     `form:"card_2" json:"card_2" gorm:"column:card_2"`
	Card2Used int64     `form:"card_2_used" json:"card_2_used" gorm:"column:card_2_used"`
	Card3     int64     `form:"card_3" json:"card_3" gorm:"column:card_3"`
	Card3Used int64     `form:"card_3_used" json:"card_3_used" gorm:"column:card_3_used"`
	Card4     int64     `form:"card_4" json:"card_4" gorm:"column:card_4"`
	Card4Used int64     `form:"card_4_used" json:"card_4_used" gorm:"column:card_4_used"`
	Card5     int64     `form:"card_5" json:"card_5" gorm:"column:card_5"`
	Card5Used int64     `form:"card_5_used" json:"card_5_used" gorm:"column:card_5_used"`
	State     int       `form:"state" json:"state" gorm:"column:state"`
	Compose   int64     `form:"compose" json:"compose" gorm:"column:compose"`
	Ctime     time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime     time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// ActYouthCardsNums ...
type ActYouthCardsNums struct {
	ID        int64     `form:"id" json:"id" gorm:"column:id"`
	Mid       int64     `form:"mid" json:"mid" gorm:"column:mid"`
	Card1     int64     `form:"card_1" json:"card_1" gorm:"column:card_1"`
	Card1Used int64     `form:"card_1_used" json:"card_1_used" gorm:"column:card_1_used"`
	Card2     int64     `form:"card_2" json:"card_2" gorm:"column:card_2"`
	Card2Used int64     `form:"card_2_used" json:"card_2_used" gorm:"column:card_2_used"`
	Card3     int64     `form:"card_3" json:"card_3" gorm:"column:card_3"`
	Card3Used int64     `form:"card_3_used" json:"card_3_used" gorm:"column:card_3_used"`
	Card4     int64     `form:"card_4" json:"card_4" gorm:"column:card_4"`
	Card4Used int64     `form:"card_4_used" json:"card_4_used" gorm:"column:card_4_used"`
	Card5     int64     `form:"card_5" json:"card_5" gorm:"column:card_5"`
	Card5Used int64     `form:"card_5_used" json:"card_5_used" gorm:"column:card_5_used"`
	Card6     int64     `form:"card_6" json:"card_6" gorm:"column:card_6"`
	Card6Used int64     `form:"card_6_used" json:"card_6_used" gorm:"column:card_6_used"`
	Card7     int64     `form:"card_7" json:"card_7" gorm:"column:card_7"`
	Card7Used int64     `form:"card_7_used" json:"card_7_used" gorm:"column:card_7_used"`
	Card8     int64     `form:"card_8" json:"card_8" gorm:"column:card_8"`
	Card8Used int64     `form:"card_8_used" json:"card_8_used" gorm:"column:card_8_used"`
	Card9     int64     `form:"card_9" json:"card_9" gorm:"column:card_9"`
	Card9Used int64     `form:"card_9_used" json:"card_9_used" gorm:"column:card_9_used"`
	State     int       `form:"state" json:"state" gorm:"column:state"`
	Compose   int64     `form:"compose" json:"compose" gorm:"column:compose"`
	Ctime     time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime     time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// ActComposeCardLog ...
type ActComposeCardLog struct {
	ID       int64     `form:"id" json:"id" gorm:"column:id"`
	Mid      int64     `form:"mid" json:"mid" gorm:"column:mid"`
	Activity string    `form:"activity" json:"activity" gorm:"column:activity"`
	Ctime    time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime    time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// ActSpringCardsNumsReply ...
type ActSpringCardsNumsReply struct {
	List []*ActSpringCardsNums  `json:"list"`
	Page map[string]interface{} `json:"page"`
}

// ActYouthCardsNumsReply ...
type ActYouthCardsNumsReply struct {
	List []*ActYouthCardsNums   `json:"list"`
	Page map[string]interface{} `json:"page"`
}

// ActComposeLogReply ...
type ActComposeLogReply struct {
	Count int64 `json:"count"`
}

// ActSpringSendCardLog ...
type ActSpringSendCardLog struct {
	ID          int64     `form:"id" json:"id" gorm:"column:id"`
	Mid         int64     `form:"mid" json:"mid" gorm:"column:mid"`
	CardID      int64     `form:"card_id" json:"card_id" gorm:"column:card_id"`
	ReceiverMid int64     `form:"receiver_mid" json:"receiver_mid" gorm:"column:receiver_mid"`
	Ctime       time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime       time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// ActSpringSendCardLogReply ...
type ActSpringSendCardLogReply struct {
	List []*ActSpringSendCardLog `json:"list"`
	Page map[string]interface{}  `json:"page"`
}

// ActSpringComposeCardLog ...
type ActSpringComposeCardLog struct {
	ID    int64     `form:"id" json:"id" gorm:"column:id"`
	Mid   int64     `form:"mid" json:"mid" gorm:"column:mid"`
	Ctime time.Time `json:"ctime" gorm:"column:ctime" time_format:"2006-01-02 15:04:05"`
	Mtime time.Time `json:"mtime" gorm:"column:mtime" time_format:"2006-01-02 15:04:05"`
}

// ActSpringComposeCardLogReply ...
type ActSpringComposeCardLogReply struct {
	List []*ActSpringComposeCardLog `json:"list"`
	Page map[string]interface{}     `json:"page"`
}

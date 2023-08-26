package pcdn

import (
	"strings"
	"time"
)

type Business string

var (
	MOD    = Business("modmanager")
	Bender = Business("bender")
)

// Files  pcdn文件列表
type Files struct {
	ID        int64     `gorm:"column:id" db:"id" json:"id" form:"id"`                                 //  自增id
	Rid       string    `gorm:"column:rid" db:"rid" json:"rid" form:"rid"`                             //  rid
	Url       string    `gorm:"column:url" db:"url" json:"url" form:"url"`                             //  文件url
	Md5       string    `gorm:"column:md5" db:"md5" json:"md5" form:"md5"`                             //  md5
	Size      int64     `gorm:"column:size" db:"size" json:"size" form:"size"`                         //  文件大小
	Business  string    `gorm:"column:business" db:"business" json:"business" form:"business"`         //  业务模块
	VersionId string    `gorm:"column:version_id" db:"version_id" json:"version_id" form:"version_id"` //  版本号
	Ctime     time.Time `gorm:"column:ctime" db:"ctime" json:"ctime" form:"ctime"`                     //  创建时间
	Mtime     time.Time `gorm:"column:mtime" db:"mtime" json:"mtime" form:"mtime"`                     //  修改时间
}

type QueryLog struct {
	ID        int64     `gorm:"column:id" db:"id" json:"id" form:"id"`                                 //  自增id
	VersionId string    `gorm:"column:version_id" db:"version_id" json:"version_id" form:"version_id"` //  版本号
	Zone      string    `gorm:"column:zone" db:"zone" json:"zone" form:"zone"`                         //  zone
	Ctime     time.Time `gorm:"column:ctime" db:"ctime" json:"ctime" form:"ctime"`                     //  创建时间
	Mtime     time.Time `gorm:"column:mtime" db:"mtime" json:"mtime" form:"mtime"`                     //  修改时间
}

type ListResp struct {
	Resource      []*Item `json:"resource"`
	LatestVersion string  `json:"latest_version"`
}

type Item struct {
	Rid     string `json:"key"`
	Url     string `json:"url"`
	Md5     string `json:"md5"`
	Size    int64  `json:"size"`
	Popular int64  `json:"popular"`
}

func VersionId(t time.Time) string {
	var tArr []string
	tStr := t.Format("2006-01-02 15:04:05")
	split := strings.Split(tStr, " ")
	year, time1 := split[0], split[1]

	for _, y := range strings.Split(year, "-") {
		tArr = append(tArr, y)
	}
	for _, ts := range strings.Split(time1, ":") {
		tArr = append(tArr, ts)
	}
	return strings.Join(tArr, "")
}

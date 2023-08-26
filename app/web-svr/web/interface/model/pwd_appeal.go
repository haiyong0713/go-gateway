package model

import (
	"fmt"
	"strconv"
	"time"
)

type PwdAppeal struct {
	ID          int64  `json:"id"`
	Mid         int64  `json:"mid"`
	DeviceToken string `json:"device_token"`
	Mobile      int64  `json:"mobile"`
	Mode        int64  `json:"mode"`
	State       int64  `json:"state"`
	UploadKey   string `json:"upload_key"`
}

func GenerateAppealUploadKey(key string) string {
	return fmt.Sprintf("%s-%s", key, strconv.FormatInt(time.Now().UnixNano()/1e6, 10))
}

type AddPwdAppealReq struct {
	Mobile      int64  `json:"mobile" form:"mobile" validate:"required"`
	Captcha     string `json:"captcha" form:"captcha" validate:"required"`
	Mode        int64  `json:"mode" form:"mode" validate:"required"`
	DeviceToken string `json:"device_token" form:"device_token"`
	UploadKey   string `json:"upload_key" form:"upload_key" validate:"required"`
	Pwd         string `json:"pwd" form:"pwd"`
}

type UploadPwdAppealRly struct {
	UploadKey string `json:"upload_key"`
}

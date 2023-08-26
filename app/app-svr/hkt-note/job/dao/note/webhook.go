package note

import (
	"fmt"
	"net/http"
)

func (d *Dao) SendMarkdown(key, text string) (err error) {
	client := d.restyClient
	r := client.R().SetBody(map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": text,
		},
	})
	r.SetQueryParam("key", key)
	resp, err := r.Post("https://qyapi.weixin.qq.com/cgi-bin/webhook/send")
	if err != nil {
		return
	}
	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("code: %d", resp.StatusCode())
		return
	}
	return
}

func (d *Dao) SendFile(key, mediaID string) (err error) {
	client := d.restyClient
	r := client.R().SetBody(map[string]interface{}{
		"msgtype": "file",
		"file": map[string]string{
			"media_id": mediaID,
		},
	})
	r.SetQueryParam("key", key)
	resp, err := r.Post("https://qyapi.weixin.qq.com/cgi-bin/webhook/send")
	if err != nil {
		return
	}
	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("code: %d", resp.StatusCode())
		return
	}
	return
}

func (d *Dao) UploadFile(key, path string) (mediaID string, err error) {

	client := d.restyClient
	r := client.R()
	r.SetFile("", path)
	r.SetQueryParam("key", key)
	r.SetQueryParam("type", "file")
	r.SetResult(&UploadFileResp{})
	resp, err := r.Post("https://qyapi.weixin.qq.com/cgi-bin/webhook/upload_media")
	if err != nil {
		return
	}
	if resp.StatusCode() != http.StatusOK {
		err = fmt.Errorf("code: %d", resp.StatusCode())
		return
	}
	re := resp.Result().(*UploadFileResp)
	if re.Errcode != 0 {
		err = fmt.Errorf("code: %v err: %v", re.Errcode, re.Errmsg)
		return
	}
	mediaID = re.MediaID
	return
}

type UploadFileResp struct {
	Errcode   int    `json:"errcode"`
	Errmsg    string `json:"errmsg"`
	Type      string `json:"type"`
	MediaID   string `json:"media_id"`
	CreatedAt string `json:"created_at"`
}

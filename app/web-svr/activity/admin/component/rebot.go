package component

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

func Rebot(content string) {

	robot := "https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=0a4c9217-5eb6-4fde-a4c5-29a3272a7625"
	m := make(map[string]interface{})
	m["msgtype"] = "text"
	md := make(map[string]interface{})
	md["mentioned_list"] = []string{"005273", "@all"}
	md["content"] = content
	md["mentioned_mobile_list"] = []string{"18016385260", "18600417059", "15868820770"}
	m["text"] = md
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	err := encoder.Encode(m)
	if err != nil {
		return
	}
	str := string(buffer.Bytes())
	str = strings.ReplaceAll(str, `\\\`, `\`)
	str = strings.ReplaceAll(str, `\\n`, `\n`)

	resp, err := http.Post(robot, "application/json", strings.NewReader(string(str)))

	if err != nil {
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
}

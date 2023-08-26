package note

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"unicode/utf8"

	"go-common/library/log"
)

const (
	AuditSkip = -1 // 无需更新审核状态
	AuditPass = 0
	AuditFail = 1
)

type FilterData struct {
	Level int64  `json:"level"`
	Msg   string `json:"msg"`
}

type ContentBody struct {
	Insert     interface{} `json:"insert"` // 只有string类型的正文
	Attributes interface{} `json:"attributes,omitempty"`
}

type ContentAttr struct {
	Link string `json:"link"`
}

type ContentInsert struct {
	ImageUpload struct {
		Url string `json:"url,omitempty"`
	} `json:"imageUpload,omitempty"`
	Video string `json:"video,omitempty"`
}

func ToBody(data string) string { // 去css的正文
	bodyArr := make([]*ContentBody, 0)
	if err := json.Unmarshal([]byte(data), &bodyArr); err != nil {
		log.Error("noteWarn ToBody data(%s) error(%v)", data, err)
		return ""
	}
	res := make([]string, 0)
	for _, b := range bodyArr {
		if reflect.TypeOf(b.Insert).Name() == "string" {
			str := fmt.Sprintf("%v", b.Insert)
			res = append(res, strings.Replace(str, "\n", " ", -1))
		}
	}
	return strings.Join(res, " ")
}

func ReplaceSensitive(cont string, words []string) (string, error) {
	arr := make([]*ContentBody, 0)
	if err := json.Unmarshal([]byte(cont), &arr); err != nil {
		log.Error("noteWarn ReplaceSensitive all(%s) error(%v)", cont, err)
		return "", err
	}
	res := make([]*ContentBody, 0, len(arr))
	for _, a := range arr {
		if reflect.TypeOf(a.Insert).Name() != "string" {
			res = append(res, a)
			continue
		}
		str := fmt.Sprintf("%v", a.Insert)
		a.Insert = ReplaceInStr(str, words)
		res = append(res, a)
	}
	contByte, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(contByte), nil
}

func FilterInvalid(cont string) (string, error) {
	arr := make([]*ContentBody, 0)
	if err := json.Unmarshal([]byte(cont), &arr); err != nil {
		log.Error("noteWarn ReplaceAndFilter all(%s) error(%v)", cont, err)
		return "", err
	}
	res := make([]*ContentBody, 0, len(arr))
	for _, a := range arr {
		log.Warn("filtertest a(%+v) attr(%+v) insert(%+v)", a, a.Attributes, a.Insert)
		attrNeedFilter := func() bool {
			attrCont := &ContentAttr{}
			attrbs, e1 := json.Marshal(a.Attributes)
			if e1 != nil {
				log.Warn("noteWarn ReplaceAndFilter all(%s) err(%+v)", cont, e1)
				return false
			}
			e1 = json.Unmarshal(attrbs, &attrCont)
			if e1 != nil || attrCont == nil {
				log.Warn("noteWarn ReplaceAndFilter all(%s) err(%+v)", cont, e1)
				return false
			}
			if attrCont.Link != "" {
				r := regexp.MustCompile(`(https?:\/\/(.+?\.)?(b23.tv|bili22.cn|bili33.cn|bili23.cn|bili2233.cn|dl.hdslb.com|acg.tv|biligame.com|bilibili.com|game.bilibili.com)(\/[A-Za-z0-9\-\._~:\/\?#\[\]@!$&'\(\)\*\+,;\=]*)?)`)
				match := r.MatchString(attrCont.Link)
				if !match {
					log.Warn("noteWarn ReplaceAndFilter all(%s) a(%s) hit invalid link,skip", cont, attrCont.Link)
					return true
				}
			}
			return false
		}()
		if attrNeedFilter {
			continue
		}
		if reflect.TypeOf(a.Insert).Name() == "string" {
			res = append(res, a)
			continue
		}
		insertCont := &ContentInsert{}
		bs, e := json.Marshal(a.Insert)
		if e != nil {
			log.Warn("noteWarn ReplaceAndFilter all(%s) err(%+v)", cont, e)
			res = append(res, a)
			continue
		}
		e = json.Unmarshal(bs, &insertCont)
		if e != nil || insertCont == nil {
			log.Warn("noteWarn ReplaceAndFilter all(%s) err(%+v)", cont, e)
			res = append(res, a)
			continue
		}
		if insertCont.Video != "" {
			log.Warn("noteWarn ReplaceAndFilter all(%s) a(%+v) hit video,skip", cont, a)
			continue
		}
		if insertCont.ImageUpload.Url != "" && !strings.Contains(insertCont.ImageUpload.Url, "i0.hdslb.com/bfs/note") && !strings.Contains(insertCont.ImageUpload.Url, "api.bilibili.com/x/note/image") {
			log.Warn("noteWarn ReplaceAndFilter all(%s) a(%+v) hit invalid img,skip", cont, a)
			continue
		}
		res = append(res, a)
		continue
	}
	contByte, err := json.Marshal(res)
	if err != nil {
		return "", err
	}
	return string(contByte), nil
}

func ReplaceInStr(cont string, words []string) string {
	for _, w := range words {
		re, err := regexp.Compile("[^\\sS\\pP]{0,3}" + w + "[^\\sS\\pP]{0,3}")
		if err != nil {
			log.Error("noteError ReplaceInStr cont(%s) w(%s) err(%+v)", cont, w, err)
			continue
		}
		hitLen := utf8.RuneCountInString(string(re.Find([]byte(cont))))
		rp := make([]string, 0, hitLen)
		for i := 0; i < hitLen; i++ {
			rp = append(rp, "*")
		}
		cont = re.ReplaceAllString(cont, strings.Join(rp, ""))
	}
	return cont
}

func ToSensitiveStr(data []*FilterData) []string {
	res := make([]string, 0, len(data))
	for _, d := range data {
		res = append(res, d.Msg)
	}
	return res
}

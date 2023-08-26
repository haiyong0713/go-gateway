package rewards

import (
	"crypto/md5"
	"encoding/hex"
	"sort"
	"strings"

	"go-common/library/log"
)

// Values .
type Values map[string]string

// Set sets the key to value. It replaces any existing
// values.
func (v Values) Set(key string, value string) {
	v[key] = value
}

// Encode encodes the values into .
func (v Values) Encode() string {
	if v == nil {
		return ""
	}
	var buf strings.Builder
	keys := make([]string, 0, len(v))
	for k := range v {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		vs := v[k]
		if buf.Len() > 0 {
			buf.WriteByte('&')
		}
		buf.WriteString(k)
		buf.WriteByte('=')
		buf.WriteString(vs)
	}
	return buf.String()
}

// Sign .
func (v Values) Sign(token string) (Values, error) {
	if len(v) == 0 {
		return nil, nil
	}
	tmp := v.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}
	tmp = tmp + "&token=" + token
	log.Warn("sign:%s", tmp)
	mh := md5.Sum([]byte(tmp))
	v["sign"] = hex.EncodeToString(mh[:])
	return v, nil
}

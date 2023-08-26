package game

import (
	"bytes"
	"context"
	"crypto/des" // #nosec
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/conf"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

type EntryClient struct {
	client *http.Client
	secret string
	desKey string
}

func NewEntryClient(cfg *conf.EntryGameClientConfig) *EntryClient {
	c := &EntryClient{
		client: &http.Client{Timeout: time.Duration(cfg.Timeout)},
		secret: cfg.Secret,
		desKey: cfg.DesKey,
	}
	return c
}

func (client *EntryClient) Get(c context.Context, uri string, params url.Values, res interface{}) (err error) {
	type Res struct {
		Code    int
		Message string
		Data    string
	}

	query, err := client.sign(params)
	if err != nil {
		log.Error("client.sign params(%+v) error(%v)", params, err)
		return err
	}

	url := fmt.Sprintf("%v?%v", uri, query)
	resp, err := client.client.Get(url)
	if err != nil {
		log.Error("client.Get url(%v) error(%v)", url, err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("ioutil.ReadAll body(%+v) error(%v)", resp.Body, err)
		return err
	}

	r := &Res{}
	if err = json.Unmarshal(body, r); err != nil {
		log.Error("json.Unmarshal body(%+v) error(%v)", body, err)
		return err
	}

	if r.Code != ecode.OK.Code() || r.Data == "" {
		log.Error("client.Get params(%+v) error(%v) r(%+v)", params, err, r)
		return fmt.Errorf(util.ErrorDataNull)
	}

	decrypted, err := client.decryptDES(r.Data)
	if err != nil {
		log.Error("client.decryptDES data(%v) error(%v)", r.Data, err)
		return err
	}
	log.Info("decrypted string (%v)", string(decrypted))

	if string(decrypted) == "" {
		log.Error("client.Get empty data(%v)", decrypted)
		return fmt.Errorf(util.ErrorDataNull)
	}

	if err = json.Unmarshal(decrypted, res); err != nil {
		log.Error("json.Unmarshal decrypted(%v) error(%v)", string(decrypted), err)
		return err
	}

	return
}

func (client *EntryClient) sign(params url.Values) (query string, err error) {
	if params == nil {
		params = url.Values{}
	}
	if params.Get("secret") != "" {
		return "", errors.New("utils http get must not have parameter secret")
	}

	tmp := params.Encode()
	if strings.IndexByte(tmp, '+') > -1 {
		tmp = strings.Replace(tmp, "+", "%20", -1)
	}

	var b bytes.Buffer
	b.WriteString(tmp)
	b.WriteString(client.secret)
	mh := md5.Sum(b.Bytes())

	// query
	var qb bytes.Buffer
	qb.WriteString(tmp)
	qb.WriteString("&sign=")
	qb.WriteString(hex.EncodeToString(mh[:]))
	query = qb.String()
	return
}

func (client *EntryClient) decryptDES(data string) (ret []byte, err error) {
	dataBytes, err := hex.DecodeString(data)
	if err != nil {
		return nil, err
	}
	keyBytes := []byte(client.desKey)[:8]

	block, err := des.NewCipher(keyBytes)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	if len(dataBytes)%bs != 0 {
		return nil, errors.New("crypto/cipher: input not full blocks")
	}

	ret = make([]byte, len(dataBytes))
	dst := ret
	for len(dataBytes) > 0 {
		block.Decrypt(dst, dataBytes[:bs])
		dataBytes = dataBytes[bs:]
		dst = dst[bs:]
	}

	// PKCS5UnPadding
	length := len(ret)
	padding := int(ret[length-1])
	ret = ret[:(length - padding)]
	return
}

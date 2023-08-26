package offer

import (
	"context"
	iaes "crypto/aes"
	icipher "crypto/cipher"
	"crypto/md5"
	"encoding/hex"
	"time"

	"go-common/library/log"
	"go-gateway/app/app-svr/app-wall/interface/conf"
	callbackdao "go-gateway/app/app-svr/app-wall/interface/dao/callback"
	offerdao "go-gateway/app/app-svr/app-wall/interface/dao/offer"
	"go-gateway/app/app-svr/app-wall/interface/dao/padding"
	"go-gateway/app/app-svr/app-wall/interface/model"

	"github.com/pkg/errors"
)

type Service struct {
	c     *conf.Config
	dao   *offerdao.Dao
	cbdao *callbackdao.Dao
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c:     c,
		dao:   offerdao.New(c),
		cbdao: callbackdao.New(c),
	}
	return
}

// Click
func (s *Service) Click(c context.Context, ip uint32, cid, mac, idfa, cb string, now time.Time) error {
	_, err := s.dao.InClick(c, ip, cid, mac, idfa, cb, now)
	return err
}

// ANClick
func (s *Service) ANClick(c context.Context, channel, imei, androidid, mac, cb string, ip uint32, now time.Time) error {
	_, err := s.dao.InANClick(c, channel, imei, androidid, mac, cb, ip, now)
	return err
}

func (s *Service) ANActive(c context.Context, imei, androidid, mac string, now time.Time) (err error) {
	const (
		_typeUnactive = 0
		_typeActive   = 1
	)
	//gdt
	gdtImei := model.GdtIMEI(imei)
	count, err := s.dao.ANActive(c, androidid, imei, gdtImei)
	if err != nil {
		return
	}
	if count > 0 {
		log.Warn("ANActive androidid(%s) imei(%s) gdtImei(%s) already count(%d) activated", androidid, imei, gdtImei, count)
		return
	}
	id, channel, cb, typ, err := s.dao.ANCallback(c, androidid, imei, gdtImei)
	if err != nil {
		return
	}
	if id == 0 {
		log.Warn("ANActive androidid(%s) imei(%s) gdtImei(%s) not exist", androidid, imei, gdtImei)
		return
	}
	if typ != _typeUnactive {
		log.Warn("ANActive androidid(%s) imei(%s) gdtImei(%s) already activated", androidid, imei, gdtImei)
		return
	}
	if err = s.newCallback(c, model.TypeAndriod, channel, imei, gdtImei, cb, now); err != nil {
		err = errors.Wrapf(err, "%d", id)
		return
	}
	if _, err = s.dao.ANClickAct(c, id, _typeActive); err != nil {
		return
	}
	return
}

func (s *Service) newCallback(c context.Context, appType, channel string, imei, gdtImei, cb string, now time.Time) (err error) {
	const (
		_eventActive = "0"
	)
	switch channel {
	case model.ChannelToutiao:
		err = s.cbdao.ToutiaoCallback(c, cb, _eventActive)
		// gdt
	case model.ChannelShike:
		err = s.cbdao.ShikeCallback(c, imei, cb, now)
	case model.ChannelDontin:
		err = s.cbdao.DontinCallback(c, imei, cb)
	default:
		if _, ok := model.ChannelGdt[channel]; ok {
			if appID, ok := model.AppIDGdt[appType]; ok {
				err = s.cbdao.GdtCallback(c, appID, appType, channel, gdtImei, cb, now)
			}
		} else {
			log.Warn("channel(%s) undefined", channel)
		}
	}
	return
}

// nolint:gomnd
func (s *Service) Active(c context.Context, ip uint32, mid, rmac, mac, idfa, device string, now time.Time) (err error) {
	var (
		count int
	)
	res, err := s.dao.InActive(c, ip, mid, rmac, mac, idfa, device, now)
	if err != nil {
		log.Error("s.dao.InActive(%d, %s, %s, %s, %s, %s) error(%v) or row==0", ip, mid, rmac, mac, idfa, device, err)
		return
	}
	if res == 0 {
		log.Warn("wall active(%s) exist", idfa)
		return
	}
	if rmac != "" {
		count, err = s.dao.RMacCount(c, rmac)
		if err != nil {
			log.Error("s.dao.RMacCount(%s) error(%d)", rmac, err)
			return
		}
		if count > 10 {
			log.Warn("wallService.RMacCount(%s) count(%d) more than 10", rmac, count)
			return
		}
	}
	return s.callback(c, model.TypeIOS, idfa, now)
}

func (s *Service) Exists(c context.Context, idfa string) (exist bool, err error) {
	exist, err = s.dao.Exists(c, idfa)
	if err != nil {
		log.Error("s.dao.Exists(%s) error(%v)", idfa, err)
	}
	return
}

// callback
// nolint:gomnd
func (s *Service) callback(c context.Context, appType, idfa string, now time.Time) (err error) {
	if len(idfa) <= 20 {
		return
	}
	// only gdt
	allIdfa := idfa[0:8] + "-" + idfa[8:12] + "-" + idfa[12:16] + "-" + idfa[16:20] + "-" + idfa[20:]
	bs := md5.Sum([]byte(allIdfa))
	gdtIdfa := hex.EncodeToString(bs[:])
	// get callback
	channel, cb, err := s.dao.Callback(c, idfa, gdtIdfa, now)
	if err != nil {
		log.Error("s.dao.Callback(%s) error(%v)", idfa, err)
		return
	}
	if cb == "" {
		log.Info("callback idfa(%s) callback url is not exists", idfa)
		return
	}
	switch channel {
	case model.ChannelShike:
		_ = s.cbdao.ShikeCallback(c, idfa, cb, now)
		_, _ = s.dao.UpIdfaActive(c, allIdfa, idfa, now)
	case model.ChannelDontin:
		_ = s.cbdao.DontinCallback(c, idfa, cb)
		_, _ = s.dao.UpIdfaActive(c, allIdfa, idfa, now)
	default:
		if _, ok := model.ChannelGdt[channel]; ok {
			if appID, ok := model.AppIDGdt[appType]; ok {
				if err = s.cbdao.GdtCallback(c, appID, appType, channel, gdtIdfa, cb, now); err != nil {
					_, _ = s.dao.UpIdfaActive(c, allIdfa, gdtIdfa, now)
				}
			}
		} else {
			log.Warn("channel(%s) undefined", channel)
		}
	}
	return
}

// CBCDecrypt aes cbc decrypt.
func (s *Service) CBCDecrypt(src, key, iv []byte, p padding.Padding) ([]byte, error) {
	var (
		ErrAesSrcSize = errors.New("ciphertext too short")
		ErrAesIVSize  = errors.New("iv size is not a block size")
	)
	// check src
	if len(src) < iaes.BlockSize || len(src)%iaes.BlockSize != 0 {
		log.Info("check src src(%v) blockSize (%v)", src, iaes.BlockSize)
		log.Info("len(src) < iaes.BlockSize (%v)", len(src) < iaes.BlockSize)
		log.Info("len(src)iaes.BlockSize != 0 (%v)", len(src)%iaes.BlockSize != 0)
		return nil, ErrAesSrcSize
	}
	// check iv
	if len(iv) != iaes.BlockSize {
		log.Info("check iv iv(%v) blockSize (%v)", iv, iaes.BlockSize)
		return nil, ErrAesIVSize
	}
	block, err := iaes.NewCipher(key)
	if err != nil {
		log.Error("iaes.NewCipher err(%v)", err)
		return nil, err
	}
	mode := icipher.NewCBCDecrypter(block, iv)
	decryptText := make([]byte, len(src))
	mode.CryptBlocks(decryptText, src)
	if p == nil {
		return decryptText, nil
	} else {
		return p.Unpadding(decryptText, iaes.BlockSize)
	}
}

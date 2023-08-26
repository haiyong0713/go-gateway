package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-gateway/app/app-svr/app-thirdsdk/interface/internal/model"

	"github.com/pkg/errors"
)

func (s *Service) UserBindSync(ctx context.Context, param *model.UserBindParam, ip string) error {
	var ipWhitelist map[string][]string
	if err := s.ac.Get("ipWhitelist").UnmarshalTOML(&ipWhitelist); err != nil {
		return err
	}
	ips := ipWhitelist[param.Platform]
	if len(ips) != 0 {
		var ok bool
		for _, val := range ips {
			if ip == val {
				ok = true
				break
			}
		}
		if !ok {
			return ecode.Error(ecode.RequestErr, "公网IP不在白名单内")
		}
	}
	var vendorIDm map[string]string
	if err := s.ac.Get("vendorID").UnmarshalTOML(&vendorIDm); err != nil {
		return err
	}
	vendorID, ok := vendorIDm[param.Platform]
	if !ok {
		return ecode.Error(ecode.RequestErr, "platform字段值错误")
	}
	if err := s.dao.UserBindSync(ctx, vendorID, param); err != nil {
		log.Error("%+v", err)
		return errors.Cause(err)
	}
	return nil
}

func (s *Service) ArcStatusSync(ctx context.Context, param *model.ArcStatusParam, ip string) error {
	var ipWhitelist map[string][]string
	if err := s.ac.Get("ipWhitelist").UnmarshalTOML(&ipWhitelist); err != nil {
		return err
	}
	ips := ipWhitelist[param.Platform]
	if len(ips) != 0 {
		var ok bool
		for _, val := range ips {
			if ip == val {
				ok = true
				break
			}
		}
		if !ok {
			return ecode.Error(ecode.RequestErr, "公网IP不在白名单内")
		}
	}
	var vendorIDm map[string]string
	if err := s.ac.Get("vendorID").UnmarshalTOML(&vendorIDm); err != nil {
		return err
	}
	vendorID, ok := vendorIDm[param.Platform]
	if !ok {
		return ecode.Error(ecode.RequestErr, "platform字段值错误")
	}
	if err := s.dao.ArcStatusSync(ctx, vendorID, param); err != nil {
		log.Error("%+v", err)
		return errors.Cause(err)
	}
	return nil
}

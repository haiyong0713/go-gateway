package service

import (
	"go-gateway/app/app-svr/archive-push/admin/internal/model"
	"go-gateway/app/app-svr/archive-push/ecode"
)

func (s *Service) GetVendorByID(id int64) (res model.ArchivePushVendor, err error) {
	res = model.ArchivePushVendor{}
	if id == 0 {
		err = ecode.VendorNotFound
		return
	}

	for _, vendor := range model.DefaultVendors {
		if vendor.ID == id {
			res = vendor
		}
	}

	if res.ID == 0 {
		err = ecode.VendorNotFound
	}

	return
}

func (s *Service) GetOauthAppKeyByVendorID(vendorID int64) (key string, err error) {
	switch vendorID {
	case model.DefaultVendors[0].ID, model.DefaultVendors[1].ID:
		key = s.qqDAO.Cfg.CMC.Oauth2.ClientID
		return
	default:
		err = ecode.VendorNotFound
	}

	return
}

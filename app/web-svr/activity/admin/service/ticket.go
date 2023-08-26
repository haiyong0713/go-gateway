package service

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"go-common/library/log"
	"go-gateway/app/web-svr/activity/admin/model/ticket"
	"io"
)

// GetMd5String 生成32位md5字串
func GetMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// UniqueId 生成Guid字串
func UniqueId() string {
	b := make([]byte, 48)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return GetMd5String(base64.URLEncoding.EncodeToString(b))
}

func (s *Service) TicketCreate(c context.Context, req *ticket.ReqTicketCreate) ([]*ticket.Ticket, error) {
	tx := s.dao.DB.Begin()
	tickets := make([]*ticket.Ticket, 0, req.Num)
	for i := 0; i < req.Num; i++ {
		ti := &ticket.Ticket{
			Ticket: UniqueId(),
		}
		if err := tx.Model(&ticket.Ticket{}).Create(ti).Error; err != nil {
			log.Errorc(c, "TicketCreate tx.Model(&ticket.Ticket{}).Create(%v) err[%v]", *ti, err)
			if err := tx.Rollback().Error; err != nil {
				log.Errorc(c, "TicketCreate tx.Rollback() err[%v]", err)
			}
			return nil, err
		}
		tickets = append(tickets, ti)
	}
	return tickets, tx.Commit().Error
}

func (s *Service) TicketExport(c context.Context) ([]*ticket.Ticket, error) {
	tickets := make([]*ticket.Ticket, 0, 1000)
	return tickets, s.dao.DB.Limit(100000).Find(&tickets).Error
}

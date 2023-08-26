package mail

import (
	"context"
	"go-gateway/app/web-svr/activity/job/conf"
	"io"
	"os"
	"path/filepath"

	"go-common/library/log"
	"go-gateway/app/web-svr/activity/job/model/mail"
	mdlmail "go-gateway/app/web-svr/activity/job/model/mail"

	"gopkg.in/gomail.v2"
)

type Service struct {
	c *conf.Config
}

func New(c *conf.Config) (s *Service) {
	s = &Service{
		c: c,
	}
	return s
}

// SendMail send mail
func (s *Service) SendMail(c context.Context, m *mail.Mail, base *mail.Base, attach *mail.Attach) (err error) {
	var (
		toUsers  []string
		ccUsers  []string
		bccUsers []string
		msg      = gomail.NewMessage()
	)
	msg.SetAddressHeader("From", base.Address, base.Name) // 发件人
	for _, ads := range m.ToAddresses {
		toUsers = append(toUsers, msg.FormatAddress(ads.Address, ads.Name))
	}

	for _, ads := range m.CcAddresses {
		ccUsers = append(ccUsers, msg.FormatAddress(ads.Address, ads.Name))
	}

	for _, ads := range m.BccAddresses {
		bccUsers = append(bccUsers, msg.FormatAddress(ads.Address, ads.Name))
	}
	msg.SetHeader("To", toUsers...)
	msg.SetHeader("Subject", m.Subject) // 主题

	if len(ccUsers) > 0 {
		msg.SetHeader("Cc", ccUsers...)
	}
	if len(bccUsers) > 0 {
		msg.SetHeader("Bcc", bccUsers...)
	}

	if m.Type == mail.TypeTextHTML {
		msg.SetBody("text/html", m.Body)
	} else {
		msg.SetBody("text/plain", m.Body)
	}
	// 附件处理
	if attach != nil {
		tmpSavePath := filepath.Join(os.TempDir(), "mail_tmp")
		err = os.MkdirAll(tmpSavePath, 0755)
		if err != nil {
			log.Error("os.MkdirAll error(%v)", err)
			return
		}
		destFilePath := filepath.Join(tmpSavePath, attach.Name)
		destFile, cErr := os.Create(destFilePath)
		if cErr != nil {
			log.Error("os.Create(%s) error(%v)", destFilePath, cErr)
			return cErr
		}
		defer os.RemoveAll(tmpSavePath)
		orginFile, err := os.Open(attach.File)
		if err != nil {
			log.Error("os.Open(%s) error(%v)", attach.File, err)
			return err
		}
		defer orginFile.Close()

		_, err = io.Copy(destFile, orginFile)
		if err != nil {
			log.Error("io.Copy() error(%v)", err)
			return err
		}
		msg.Attach(destFilePath)

	}
	d := gomail.NewDialer(
		base.Host,
		base.Port,
		base.Address,
		base.Pwd,
	)
	if err = d.DialAndSend(msg); err != nil {
		log.Error("Send mail Fail(%v) diff(%s)", msg, err)
		return
	}

	return
}

func (s *Service) SendAttachMail(c context.Context, to []*mdlmail.Address, subject, content string, attach *mail.Attach) error {
	return s.SendMail(c, &mdlmail.Mail{
		ToAddresses: to,
		Subject:     subject,
		Type:        mdlmail.TypeTextPlain,
		Body:        content,
	}, &mdlmail.Base{
		Host:    s.c.Mail.Host,
		Port:    s.c.Mail.Port,
		Address: s.c.Mail.Address,
		Pwd:     s.c.Mail.Pwd,
		Name:    s.c.Mail.Name,
	}, attach)
}

func (s *Service) SendTextMail(c context.Context, to []*mdlmail.Address, subject, content string) error {
	return s.SendMail(c, &mdlmail.Mail{
		ToAddresses: to,
		Subject:     subject,
		Type:        mdlmail.TypeTextPlain,
		Body:        content,
	}, &mdlmail.Base{
		Host:    s.c.Mail.Host,
		Port:    s.c.Mail.Port,
		Address: s.c.Mail.Address,
		Pwd:     s.c.Mail.Pwd,
		Name:    s.c.Mail.Name,
	}, nil)
}

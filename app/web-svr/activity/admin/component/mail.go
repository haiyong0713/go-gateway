package component

import (
	"encoding/csv"
	"io"
	"os"
	"path/filepath"

	"go-gateway/app/web-svr/activity/admin/model/component"

	"go-common/library/log"

	"github.com/pkg/errors"
	"gopkg.in/gomail.v2"
)

// CreateCsvAndSend 创建csv文件并发送
func CreateCsvAndSend(filePath, fileName string, subject string, categoryHeader []string, data [][]string, mailInfo *component.EmailInfo) error {
	err := os.MkdirAll(filePath, 0755)
	if err != nil {
		log.Error("os.MkdirAll error(%v)", err)
		return err
	}
	f, err := os.Create(filePath + fileName)
	if err != nil {
		err = errors.Wrapf(err, "s.createCsv")
		log.Error("s.createCsv: error(%v) fileName %v", err, fileName)
	}
	defer os.RemoveAll(filePath)
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	w.Write(categoryHeader)
	w.WriteAll(data) //写入数据
	err = mailFile(subject, component.TypeTextHTML, &component.Attach{Name: fileName, File: filePath + fileName}, mailInfo)
	if err != nil {
		log.Error("s.mailFile: error(%v) fileName %v", err, fileName)

	}
	return nil
}

func CreateSignleColCsvAndSend(filePath, fileName string, subject string, data []string, mailInfo *component.EmailInfo) error {
	err := os.MkdirAll(filePath, 0755)
	if err != nil {
		log.Error("os.MkdirAll error(%v)", err)
		return err
	}
	f, err := os.Create(filePath + fileName)
	if err != nil {
		err = errors.Wrapf(err, "s.createCsv")
		log.Error("s.createCsv: error(%v) fileName %v", err, fileName)
	}
	defer os.RemoveAll(filePath)
	defer f.Close()
	f.WriteString("\xEF\xBB\xBF")
	w := csv.NewWriter(f)
	for _, v := range data {
		w.Write([]string{v})
	}
	w.Flush()
	err = mailFile(subject, component.TypeTextHTML, &component.Attach{Name: fileName, File: filePath + fileName}, mailInfo)
	if err != nil {
		log.Error("s.mailFile: error(%v) fileName %v", err, fileName)

	}
	return nil
}

// mailFile 邮件发送
func mailFile(subject string, mailType component.Type, attach *component.Attach, emailInfo *component.EmailInfo) error {
	base := &component.Base{
		Host:    emailInfo.Host,
		Port:    emailInfo.Port,
		Address: emailInfo.Address,
		Pwd:     emailInfo.Pwd,
		Name:    emailInfo.Name,
	}
	mail := &component.Mail{
		ToAddresses:  emailInfo.ToAddress,
		CcAddresses:  emailInfo.CcAddress,
		BccAddresses: emailInfo.BccAddresses,
		Subject:      subject,
		Type:         mailType,
	}
	return SendMail(mail, base, attach)
}

// SendMail send mail
func SendMail(m *component.Mail, base *component.Base, attach *component.Attach) (err error) {
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

	if m.Type == component.TypeTextHTML {
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

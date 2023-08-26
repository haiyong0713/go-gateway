package fawkes

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go-common/library/database/sql"
	"go-common/library/database/xsql"

	mailmdl "go-gateway/app/app-svr/fawkes/service/model/mail"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
	"go-gateway/app/app-svr/fawkes/service/tools/utils"

	gomail "gopkg.in/gomail.v2"
)

const (
	_appMailtoList = `SELECT uname FROM app_mailto WHERE app_key=? %s`
	_addMailto     = `INSERT INTO app_mailto (app_key,func_module,uname,type) VALUES (?,?,?,?)`
	_delMailto     = `DELETE FROM app_mailto WHERE app_key=? AND func_module=? AND uname=? AND type=?`

	_addMailConfig           = `INSERT INTO app_mail_config (app_key,func_module,host,port,address,pwd,name,operator) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE host=VALUES(host),port=VALUES(port),address=VALUES(address),pwd=VALUES(pwd),operator=VALUES(operator)`
	_delMailConfig           = `DELETE FROM app_mail_config WHERE id=?`
	_updateMailConfig        = `UPDATE app_mail_config SET app_key=?,func_module=?,host=?,port=?,address=?,pwd=?,name=?,operator=? WHERE id=?`
	_appMailConfigList       = `SELECT id,app_key,func_module,host,port,address,pwd,name,operator,ctime,mtime FROM app_mail_config WHERE app_key=? %s`
	_appMailConfigWithModule = `SELECT id,app_key,func_module,host,port,address,pwd,name,operator,ctime,mtime FROM app_mail_config WHERE app_key=? AND func_module=?`

	_appMailList = `SELECT c.app_key,c.func_module,c.id as sender_id,c.name as sender_name,t.id as receiver_id,t.uname as receiver_name FROM (app_mailto as t INNER JOIN app_mail_config as c ON c.app_key=t.app_key AND c.func_module=t.func_module) WHERE c.app_key=? %s`
)

// AppMailtoList get mailto list with appkey
func (d *Dao) AppMailtoList(c context.Context, appKey, funcModule string, receiverType int64) (mailList []string, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if receiverType != mailmdl.ReceiverWithAll {
		sqlAdd += " AND type=?"
		args = append(args, receiverType)
	}
	if funcModule != "" {
		sqlAdd += " AND func_module=?"
		args = append(args, funcModule)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_appMailtoList, sqlAdd), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var uname string
		if err = rows.Scan(&uname); err != nil {
			return
		}
		mailList = append(mailList, uname)
	}
	err = rows.Err()
	return
}

// TxAppMailtoAdd add user to mailto list
func (d *Dao) TxAppMailtoAdd(tx *sql.Tx, appKey, funcModule, uname string, receiverType int64) (r int64, err error) {
	res, err := tx.Exec(_addMailto, appKey, funcModule, uname, receiverType)
	if err != nil {
		return
	}
	r, err = res.RowsAffected()
	return
}

// TxAppMailtoDel delete user from mailto list
func (d *Dao) TxAppMailtoDel(tx *sql.Tx, appKey, funcModule, uname string, receiverType int64) (r int64, err error) {
	res, err := tx.Exec(_delMailto, appKey, funcModule, uname, receiverType)
	if err != nil {
		log.Error("TxAppMailtoDel %v", err)
		return
	}
	r, err = res.RowsAffected()
	return
}

// SendMailCommon send mail
func (d *Dao) SendMailCommon(c context.Context, m *mailmdl.Mail, attach *mailmdl.Attach, sender *mailmdl.Sender) (err error) {
	var (
		toUsers  []string
		ccUsers  []string
		bccUsers []string
		msg      = gomail.NewMessage()
	)
	if m == nil {
		log.Errorc(c, "mail is nil")
		return
	}
	if sender == nil {
		log.Errorc(c, "sender is nil")
		return
	}
	msg.SetAddressHeader("From", sender.Address, sender.Name) // 发件人
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

	if m.Type == mailmdl.TypeTextHTML {
		msg.SetBody("text/html", m.Body)
	} else {
		msg.SetBody("text/plain", m.Body)
	}

	// 附件处理
	if attach != nil {
		tmpSavePath := filepath.Join(os.TempDir(), "mail_tmp")
		err = os.MkdirAll(tmpSavePath, 0755)
		if err != nil {
			log.Errorc(c, "os.MkdirAll error(%v)", err)
			return
		}
		destFilePath := filepath.Join(tmpSavePath, attach.Name)
		destFile, cErr := os.Create(destFilePath)
		if cErr != nil {
			log.Errorc(c, "os.Create(%s) error(%v)", destFilePath, cErr)
			return cErr
		}
		defer os.RemoveAll(tmpSavePath)
		if _, err = io.Copy(destFile, attach.File); err != nil {
			log.Errorc(c, "io.Copy error(%v)", err)
			return
		}
		// 如果 zip 文件需要解压以后放在邮件附件中
		if attach.ShouldUnzip && strings.HasSuffix(attach.Name, ".zip") {
			unzipFilePath := filepath.Join(tmpSavePath, "unzip")
			err = os.MkdirAll(tmpSavePath, 0755)
			if err != nil {
				log.Errorc(c, "os.MkdirAll error(%v)", err)
				return
			}
			err = utils.Unzip(destFilePath, unzipFilePath)
			if err != nil {
				log.Errorc(c, "unzip(%s, %s) error(%v)", destFilePath, unzipFilePath, err)
				return
			}
			err = filepath.Walk(unzipFilePath, func(path string, f os.FileInfo, err error) error {
				if err != nil {
					log.Errorc(c, "filepath.Walk error(%v)", err)
					return err
				}
				if f == nil {
					errMsg := "found no file"
					err = fmt.Errorf(errMsg)
					log.Error(errMsg)
					return err
				}
				if f.IsDir() {
					return nil
				}
				msg.Attach(path)
				return err
			})
		} else {
			msg.Attach(destFilePath)
		}
	}
	dialer := gomail.NewDialer(
		sender.Host,
		sender.Port,
		sender.Address,
		sender.Pwd,
	)
	if err = dialer.DialAndSend(msg); err != nil {
		log.Errorc(c, "Send mail Fail(%v) diff(%s)", msg, err)
		return
	}
	return
}

// SendMail send mail
func (d *Dao) SendMail(c context.Context, m *mailmdl.Mail, attach *mailmdl.Attach, appKey, funcModule string) (err error) {
	var (
		sender       *mailmdl.Sender
		senderConfig *mailmdl.SenderConfig
	)
	if senderConfig, err = d.AppMailConfigWithModule(c, appKey, funcModule); err != nil {
		return
	}
	if senderConfig == nil {
		sender = &mailmdl.Sender{
			Host:    d.c.Mail.AppBuilder.Host,
			Port:    d.c.Mail.AppBuilder.Port,
			Address: d.c.Mail.AppBuilder.Address,
			Pwd:     d.c.Mail.AppBuilder.Pwd,
			Name:    d.c.Mail.AppBuilder.Name}

	} else {
		sender = &mailmdl.Sender{
			Host:    senderConfig.Host,
			Port:    senderConfig.Port,
			Address: senderConfig.Address,
			Pwd:     senderConfig.Pwd,
			Name:    senderConfig.Name}
	}
	if err = d.SendMailCommon(c, m, attach, sender); err != nil {
		log.Errorc(c, "Send mail Fail %v", err)
	}
	return
}

func (d *Dao) SendMailSample(c context.Context, subject, body, mailto, mailcc string, attach *mailmdl.Attach, sender *mailmdl.Sender) (err error) {
	var (
		toAddresses []*mailmdl.Address
		ccAddresses []*mailmdl.Address
	)
	for _, address := range strings.Split(mailto, ",") {
		toAddresses = append(toAddresses, &mailmdl.Address{
			Address: address,
		})
	}
	if mailcc != "" {
		for _, address := range strings.Split(mailcc, ",") {
			ccAddresses = append(ccAddresses, &mailmdl.Address{
				Address: address,
			})
		}
	}
	mail := &mailmdl.Mail{
		ToAddresses: toAddresses,
		CcAddresses: ccAddresses,
		Subject:     subject,
		Body:        body,
		Type:        mailmdl.TypeTextHTML,
	}
	err = d.SendMailCommon(c, mail, attach, sender)
	return
}

func (d *Dao) AppMailConfigAdd(c context.Context, appKey, funcModule, host, address, pwd, name, operator string, port int) (err error) {
	_, err = d.db.Exec(c, _addMailConfig, appKey, funcModule, host, port, address, pwd, name, operator)
	return
}

func (d *Dao) AppMailConfigDel(c context.Context, id int64) (err error) {
	_, err = d.db.Exec(c, _delMailConfig, id)
	return
}

func (d *Dao) AppMailConfigUpdate(c context.Context, appKey, funcModule, host, address, pwd, name, operator string, port int, id int64) (err error) {
	_, err = d.db.Exec(c, _updateMailConfig, appKey, funcModule, host, port, address, pwd, name, operator, id)
	return
}

func (d *Dao) AppMailConfigList(c context.Context, appKey, funcModule string) (res []*mailmdl.SenderConfig, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if funcModule != "" {
		sqlAdd += " AND func_module=? "
		args = append(args, funcModule)
	}
	rows, err := d.db.Query(c, fmt.Sprintf(_appMailConfigList, sqlAdd), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "%v", err)
		return
	}
	err = rows.Err()
	return
}

func (d *Dao) AppMailConfigWithModule(c context.Context, appKey, funcModule string) (res *mailmdl.SenderConfig, err error) {
	rows, err := d.db.Query(c, _appMailConfigWithModule, appKey, funcModule)
	if err != nil {
		return
	}
	defer rows.Close()
	var list []*mailmdl.SenderConfig
	if err = xsql.ScanSlice(rows, &list); err != nil {
		log.Errorc(c, "ScanSlice Error %v", err)
		return
	}
	if len(list) != 1 {
		return
	}
	err = rows.Err()
	res = list[0]
	return
}

func (d *Dao) AppMailList(c context.Context, appKey, funcModule string) (res []*mailmdl.AppMailWithModule, err error) {
	var (
		sqlAdd string
		args   []interface{}
	)
	args = append(args, appKey)
	if funcModule != "" {
		sqlAdd += "AND c.func_module=?"
		args = append(args, funcModule)
	}
	sqlAdd += " ORDER BY t.ctime DESC"
	rows, err := d.db.Query(c, fmt.Sprintf(_appMailList, sqlAdd), args...)
	if err != nil {
		return
	}
	defer rows.Close()
	if err = xsql.ScanSlice(rows, &res); err != nil {
		log.Errorc(c, "ScanSlice error(%v)", err)
		return
	}
	err = rows.Err()
	return
}

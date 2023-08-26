package pwd_appeal

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"go-common/library/ecode"
	"go-common/library/log"

	"go-gateway/app/app-svr/app-feed/admin/api"
	model "go-gateway/app/app-svr/app-feed/admin/model/pwd_appeal"
	"go-gateway/app/app-svr/app-feed/admin/util"
	xecode "go-gateway/app/app-svr/app-feed/ecode"
)

const (
	_exportPs = 100
)

func (s *Service) List(c context.Context, req *model.ListReq) (*model.ListRly, error) {
	appeals, total, err := s.dao.SearchPwdAppeal(req, true)
	if err != nil {
		return nil, ecode.ServerErr
	}
	var delAppeals []*model.PwdAppeal
	for _, appeal := range appeals {
		if pwd, ok := s.EncryptedPwd[appeal.Pwd]; ok {
			appeal.Pwd = pwd
		}
		if needDel(appeal) {
			delAppeal := *appeal
			appeal.UploadKey = ""
			delAppeals = append(delAppeals, &delAppeal)
		}
	}
	if len(delAppeals) > 0 {
		_ = s.worker.Do(c, func(ctx context.Context) {
			for _, appeal := range delAppeals {
				if err := s.boss.Delete(appeal.UploadKey); err != nil {
					continue
				}
				_ = s.dao.DelUploadKey(appeal.ID)
			}
		})
	}
	return &model.ListRly{
		List: appeals,
		Page: &model.Page{Num: req.Pn, Size: req.Ps, Total: total},
	}, nil
}

func (s *Service) Photo(c context.Context, req *model.PhotoReq) ([]byte, error) {
	buf, err := s.boss.DownloadBuffer(req.UploadKey)
	if err != nil {
		return nil, ecode.ServerErr
	}
	photo, err := util.AesDecrypt(buf, []byte(s.cfg.PwdAppeal.EncryptKey))
	if err != nil {
		log.Errorc(c, "Fail to decrypt photo, upload_key=%s error=%+v", req.UploadKey, err)
		return nil, ecode.ServerErr
	}
	return photo, nil
}

func (s *Service) Pass(c context.Context, req *model.PassReq, userid int64, username string) error {
	appeal, err := s.dao.PwdAppeal(req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ecode.NothingFound
		}
		return ecode.ServerErr
	}
	if appeal.State != model.StatePending {
		return xecode.PwdAppealProcessed
	}
	// 如果申诉页已上报密码，则已申诉页为准
	encrypted, err := func() (string, error) {
		if appeal.Pwd == "" {
			h := md5.New()
			h.Write([]byte(req.Pwd))
			return hex.EncodeToString(h.Sum(nil)), nil
		}
		if s.EncryptedPwd[appeal.Pwd] != req.Pwd {
			return "", ecode.Error(ecode.RequestErr, "密码与申诉上报不一致")
		}
		return appeal.Pwd, nil
	}()
	if err != nil {
		return err
	}
	smsCfg, ok := s.cfg.PwdAppeal.SmsCfg[strconv.FormatInt(appeal.Mode, 10)]
	if !ok {
		return xecode.PwdAppealSmsCfgError
	}
	return func() (err error) {
		defer func() {
			_ = s.worker.Do(c, func(ctx context.Context) {
				_, _ = s.dao.CreatePwdAppealLog(&model.PwdAppealLog{
					AppealID:    req.ID,
					OperatorUid: userid,
					Operator:    username,
					Content:     passLogContent(err, req.Pwd),
				})
			})
		}()
		// 发送短信
		msg := fmt.Sprintf(`{"password":"%s"}`, req.Pwd)
		if err = s.sms.SendSms(c, appeal.Mobile, smsCfg.PassTcode, msg); err != nil {
			return xecode.PwdAppealSendFail
		}
		// 更新数据
		if err = s.dao.PassAppeal(appeal.ID, encrypted, username); err != nil {
			return xecode.PwdAppealUpdateDBFail
		}
		return nil
	}()
}

func (s *Service) Reject(c context.Context, req *model.RejectReq, userid int64, username string) error {
	appeal, err := s.dao.PwdAppeal(req.ID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return ecode.NothingFound
		}
		return ecode.ServerErr
	}
	if appeal.State != model.StatePending {
		return xecode.PwdAppealProcessed
	}
	smsCfg, ok := s.cfg.PwdAppeal.SmsCfg[strconv.FormatInt(appeal.Mode, 10)]
	if !ok {
		return xecode.PwdAppealSmsCfgError
	}
	return func() (err error) {
		defer func() {
			_ = s.worker.Do(c, func(ctx context.Context) {
				_, _ = s.dao.CreatePwdAppealLog(&model.PwdAppealLog{
					AppealID:    req.ID,
					OperatorUid: userid,
					Operator:    username,
					Content:     rejectLogContent(err, req.Reason),
				})
			})
		}()
		// 发送短信
		msg := fmt.Sprintf(`{"reason":"%s","url":"%s"}`, req.Reason, smsCfg.AppealUrl)
		if err = s.sms.SendSms(c, appeal.Mobile, smsCfg.RejectTcode, msg); err != nil {
			return xecode.PwdAppealSendFail
		}
		// 更新数据
		if err = s.dao.RejectAppeal(appeal.ID, req.Reason, username); err != nil {
			return xecode.PwdAppealUpdateDBFail
		}
		return nil
	}()
}

func (s *Service) Export(c context.Context, req *model.ExportReq) ([]byte, error) {
	req.Pn = 1
	req.Ps = _exportPs
	appeals := make([]*model.PwdAppeal, 0, s.cfg.PwdAppeal.ExportLimit)
	for i := int64(0); i < s.cfg.PwdAppeal.ExportLimit; i += req.Ps {
		list := func() []*model.PwdAppeal {
			defer time.Sleep(5 * time.Millisecond)
			if list, _, err := s.dao.SearchPwdAppeal(&req.ListReq, false); err == nil {
				return list
			}
			return nil
		}()
		appeals = append(appeals, list...)
		if int64(len(list)) < req.Ps {
			break
		}
		req.Pn++
	}
	bf := &bytes.Buffer{}
	bf.WriteString("\xEF\xBB\xBF")
	bf.WriteString(buildCsv(appeals))
	return bf.Bytes(), nil
}

func (s *Service) CreatePwdAppeal(req *api.CreatePwdAppealReq) (*api.CreatePwdAppealRly, error) {
	pendingID, err := s.dao.PendingPwdAppeal(req.Mobile)
	if err != nil {
		return nil, ecode.ServerErr
	}
	if pendingID > 0 {
		return nil, xecode.PwdAppealPendingExist
	}
	id, err := s.dao.CreatePwdAppeal(&model.PwdAppeal{
		Mid:         req.Mid,
		DeviceToken: req.DeviceToken,
		Mobile:      req.Mobile,
		Mode:        req.Mode,
		State:       model.StatePending,
		UploadKey:   req.UploadKey,
		Pwd:         req.Pwd,
	})
	if err != nil {
		return nil, err
	}
	return &api.CreatePwdAppealRly{Id: id}, nil
}

func passLogContent(err error, pwd string) string {
	var content string
	switch err {
	case xecode.PwdAppealSendFail:
		content = fmt.Sprintf("短信通知失败，密码=%s", pwd)
	case xecode.PwdAppealUpdateDBFail:
		content = fmt.Sprintf("短信通知成功，更新数据库失败，密码=%s", pwd)
	case nil:
		content = fmt.Sprintf("审核通过，密码=%s", pwd)
	default:
		content = "未知错误"
	}
	return content
}

func rejectLogContent(err error, reason string) string {
	var content string
	switch err {
	case xecode.PwdAppealSendFail:
		content = fmt.Sprintf("短信通知失败，驳回理由=%s", reason)
	case xecode.PwdAppealUpdateDBFail:
		content = fmt.Sprintf("短信通知成功，更新数据库失败，驳回理由=%s", reason)
	case nil:
		content = fmt.Sprintf("审核驳回，驳回理由=%s", reason)
	default:
		content = "未知错误"
	}
	return content
}

func buildCsv(appeals []*model.PwdAppeal) string {
	if len(appeals) == 0 {
		return ""
	}
	items := make([]string, 0, len(appeals))
	items = append(items, fmt.Sprintf(
		"%s,%s,%s,%s,%s,%s,%s,%s,%s",
		"申诉单ID", "申诉类型", "用户ID", "设备ID", "联系方式", "申诉时间", "申诉状态", "操作时间", "操作人",
	))
	for _, v := range appeals {
		ctime := v.Ctime.Time().Format("2006-01-02 15:04:05")
		mtime := v.Mtime.Time().Format("2006-01-02 15:04:05")
		items = append(items, fmt.Sprintf(
			"%d,%s,%d,%s,%d,%s,%s,%s,%s",
			v.ID, model.ModeMap[v.Mode], v.Mid, v.DeviceToken, v.Mobile, ctime, model.StateMap[v.State], mtime, v.Operator,
		))
	}
	return strings.Join(items, "\n")
}

func needDel(appeal *model.PwdAppeal) bool {
	if appeal.State == model.StatePending {
		return false
	}
	if appeal.UploadKey == "" {
		return false
	}
	return appeal.Ctime.Time().AddDate(0, 0, 7).Before(time.Now())
}

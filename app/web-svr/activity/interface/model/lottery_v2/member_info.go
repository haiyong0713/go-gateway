package lottery

import (
	"go-gateway/app/web-svr/activity/ecode"
)

// MemberInfo 用户信息
type MemberInfo struct {
	Mid            int64
	Name           string
	Silence        int32
	Level          int32
	JoinTime       int32
	VipType        int32
	VipStatus      int32
	TelStatus      int32
	Identification int32
	IP             string
	MonthVip       bool
	NeverVip       bool
	AnnualVip      bool
	ValidIP        bool
	SpyScore       int32
	Percentage     int32
	DeviceID       string
	IsCartoonNew   bool
}

// IsValidIP ...
func (m *MemberInfo) IsValidIP() error {
	if !m.ValidIP {
		return ecode.ActivityLotteryValidIP
	}
	return nil
}

// IsNotVip 不是大会员
func (m *MemberInfo) IsNotVip() error {
	if m.VipType != 0 && m.VipStatus == 1 {
		return ecode.ActivityLotteryVip
	}
	return nil
}

// IsVip 是大会员
func (m *MemberInfo) IsVip() error {
	if m.VipType == 0 || m.VipStatus != 1 {
		return ecode.ActivityNotVip
	}
	return nil
}

// IsMonthVip 是否月度大会员
func (m *MemberInfo) IsMonthVip() error {
	if !m.MonthVip {
		return ecode.ActivityLotteryNotMonthVip
	}
	return nil
}

// IsAnnualVip 是否年度度大会员
func (m *MemberInfo) IsAnnualVip() error {
	if !m.AnnualVip {
		return ecode.ActivityLotteryNotAnnualVip
	}
	return nil
}

// IsNewVip 是新大会员
func (m *MemberInfo) IsNewVip() error {
	if !m.NeverVip {
		return ecode.ActivityLotteryNotNewVip
	}
	return nil
}

// IsOldVip 是老大会员
func (m *MemberInfo) IsOldVip() error {
	if m.NeverVip {
		return ecode.ActivityLotteryNotOldVip
	}
	return nil
}

// IsSilence 是否被禁言
func (m *MemberInfo) IsSilence() error {
	if m.Silence == silenceForbid {
		return ecode.ActivityMemberBlocked
	}
	return nil
}

// LevelLimit 等级限制
func (m *MemberInfo) LevelLimit(level int) error {
	if m.Level < int32(level) {
		return ecode.ActivityLotteryLevelLimit
	}
	return nil
}

// RegStimeLimit 注册时间开始限制
func (m *MemberInfo) RegStimeLimit(regTimeStime int64) error {
	if int64(m.JoinTime) > regTimeStime {
		return ecode.ActivityLotteryRegisterEarlyLimit
	}
	return nil
}

// RegEtimeLimit 注册时间结束限制
func (m *MemberInfo) RegEtimeLimit(regTimeEtime int64) error {
	if int64(m.JoinTime) < regTimeEtime {
		return ecode.ActivityLotteryRegisterLastLimit
	}
	return nil
}

// VipCheck 大会员限制
func (m *MemberInfo) VipCheck(memberVipCheck int) error {
	switch memberVipCheck {
	case vipCheck: // vip专享
		if m.VipType == 0 || m.VipStatus != 1 {
			return ecode.ActivityNotVip
		}
	case monthVip: // 月度大会员
		if m.VipType == 0 || m.VipStatus != 1 {
			return ecode.ActivityNotMonthVip
		}
	case yearVip: // 年度大会员
		if m.VipType != 2 || m.VipStatus != 1 {
			return ecode.ActivityNotYearVip
		}
	}
	return nil
}

// AccountCheck 账号限制
func (m *MemberInfo) AccountCheck(accountCheck int) error {
	switch accountCheck {
	case telValid: // 手机验证
		if m.TelStatus != 1 {
			return ecode.ActivityTelValid
		}
	case identifyValid: // 实名验证
		if m.Identification != 1 {
			return ecode.ActivityIdentificationValid
		}
	}
	return nil
}

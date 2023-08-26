package model

import (
	acpAPI "git.bilibili.co/bapis/bapis-go/account/service/account_control_plane"
)

// BlockStatus 封禁状态 0. 未封禁 1. 永久封禁 2. 限时封禁
type BlockStatus uint8

const (
	// BlockStatusFalse 未封禁
	BlockStatusFalse BlockStatus = iota
	// BlockStatusForever 永久封禁
	BlockStatusForever
	// BlockStatusLimit 限时封禁
	BlockStatusLimit
)

const (
	//主站封禁角色
	MainBlockRole = "silence"
	// 全站封禁角色
	AllBlockRole = "all_block"
)

// BlockInfo 封禁信息
type BlockInfo struct {
	MID         int64       `json:"mid"`
	BlockStatus BlockStatus `json:"status"`     // status 封禁状态 0. 未封禁 1. 永久封禁 2. 限时封禁
	StartTime   int64       `json:"start_time"` // 开始封禁时间 unix time 未封禁为 -1
	EndTime     int64       `json:"end_time"`   // 结束封禁时间 unix time 永久封禁为 -1
}

func (b *BlockInfo) FromControlRoleToBlockInfo(hasControlRoleReply *acpAPI.HasControlRoleReply) {
	b.MID = hasControlRoleReply.Mid
	b.BlockStatus = BlockStatusFalse
	b.AssignStatus(hasControlRoleReply.ControlRoleStatus[MainBlockRole])
	b.AssignStatus(hasControlRoleReply.ControlRoleStatus[AllBlockRole])
}

func (b *BlockInfo) AssignStatus(roleStatus *acpAPI.RoleStatus) {
	if roleStatus == nil || b.BlockStatus == BlockStatusForever {
		return
	}
	if !roleStatus.HasRole {
		return
	}
	if roleStatus.IsExpirable {
		if roleStatus.ExpireAt < b.EndTime {
			return
		}
		b.StartTime = roleStatus.LastEffectivedControlAt
		b.BlockStatus = BlockStatusLimit
		b.EndTime = roleStatus.ExpireAt
		return
	}
	b.BlockStatus = BlockStatusForever
	b.StartTime = roleStatus.LastEffectivedControlAt
	b.EndTime = -1
}

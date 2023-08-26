package bws

import (
	accapi "git.bilibili.co/bapis/bapis-go/account/service"
	"go-common/library/ecode"
)

func (b *BluetoothUpInfo) BluetoothUpInfoChange(up *BluetoothUp, acc *accapi.Card) error {
	if up == nil || acc == nil {
		return ecode.RequestErr
	}
	b.Mid = acc.Mid
	b.Name = acc.Name
	b.Face = acc.Face
	b.Key = up.Key
	b.Desc = up.Desc
	return nil
}

func (b *BluetoothUpInfo) BluetoothUpChange(acc *accapi.Card) error {
	if acc == nil {
		return ecode.RequestErr
	}
	b.Mid = acc.Mid
	b.Name = acc.Name
	b.Face = acc.Face
	return nil
}

package cards

import (
	"context"
	"go-common/library/log"
)

// GetInviteTokenToMid 根据token返回mid
func (d *Dao) GetInviteTokenToMid(c context.Context, token, activity string) (mid int64, err error) {
	mid, err = d.InviteTokenToMid(c, token, activity)
	if err != nil {
		log.Errorc(c, "d.InviteTokenToMid token(%s) activity(%s) err(%v)", token, activity, err)
	}
	if err == nil {
		return mid, nil
	}
	inviterMid, err := d.InviteTokenToMidDB(c, token, activity)
	if err != nil {
		log.Errorc(c, "d.InviteTokenToMidDB(c, %d) err(%v)", mid, err)
		return inviterMid, err
	}
	err = d.AddInviteTokenToMid(c, token, inviterMid, activity)
	if err != nil {
		log.Errorc(c, "d.AddInviteTokenToMid mid(%d) token(%v) activity(%s) err(%v)", inviterMid, token, activity, err)
	}
	return inviterMid, nil
}

// GetInviteMidToToken 根据mid返回token
func (d *Dao) GetInviteMidToToken(c context.Context, mid int64, activity string) (token string, err error) {
	token, err = d.InviteMidToToken(c, mid, activity)
	if err != nil {
		log.Errorc(c, "d.InviteMidToToken err(%v)", err)
	}
	if err == nil {
		return token, nil
	}
	inviterToken, err := d.InviteMidToTokenDB(c, mid, activity)
	if err != nil {
		log.Errorc(c, "d.InviteMidToTokenDB(c, %d) err(%v)", mid, err)
		return inviterToken, err
	}
	if inviterToken != "" {
		err = d.AddInviteMidToToken(c, mid, inviterToken, activity)
		if err != nil {
			log.Errorc(c, "d.AddInviteMidToToken mid(%d)token(%v) err(%v)", mid, inviterToken, err)
		}
	}
	return inviterToken, nil
}

// GetMidInviter 根据mid返回token
func (d *Dao) GetMidInviter(c context.Context, mid int64, activity string) (inviter int64, err error) {
	inviter, err = d.MidInviter(c, mid, activity)
	if err != nil {
		log.Errorc(c, "d.InviteMidToToken err(%v)", err)
	}
	if err == nil {
		return inviter, nil
	}
	inviterMid, err := d.MidInviterDB(c, mid, activity)
	if err != nil {
		log.Errorc(c, "d.MidNums(c, %d) err(%v)", mid, err)
		return inviterMid, err
	}
	if inviterMid > 0 {
		err = d.AddMidInviter(c, mid, inviterMid, activity)
		if err != nil {
			log.Errorc(c, "d.AddInviteMidToToken mid(%d)token(%d) err(%v)", mid, inviterMid, err)
		}
	}
	return inviterMid, nil
}

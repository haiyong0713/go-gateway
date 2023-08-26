//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by deepcopy-gen. DO NOT EDIT.

package api

import (
	like "go-gateway/app/web-svr/activity/interface/model/like"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ActSubjectProtocol) DeepCopyInto(out *ActSubjectProtocol) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ActSubjectProtocol.
func (in *ActSubjectProtocol) DeepCopy() *ActSubjectProtocol {
	if in == nil {
		return nil
	}
	out := new(ActSubjectProtocol)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyAsIntoActSubjectProtocol is an autogenerated deepcopy function, copying the receiver, writing into like.ActSubjectProtocol.
func (in *ActSubjectProtocol) DeepCopyAsIntoActSubjectProtocol(out *like.ActSubjectProtocol) {
	out.ID = in.ID
	out.Sid = in.Sid
	out.Protocol = in.Protocol
	out.Mtime = in.Mtime
	out.Ctime = in.Ctime
	out.Types = in.Types
	out.Tags = in.Tags
	out.Pubtime = in.Pubtime
	out.Deltime = in.Deltime
	out.Editime = in.Editime
	out.Hot = in.Hot
	out.BgmID = in.BgmID
	out.PasterID = in.PasterID
	out.Oids = in.Oids
	out.ScreenSet = in.ScreenSet
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	out.XXX_sizecache = in.XXX_sizecache
	return
}

// DeepCopyFromActSubjectProtocol is an autogenerated deepcopy function, copying the receiver, writing into like.ActSubjectProtocol.
func (out *ActSubjectProtocol) DeepCopyFromActSubjectProtocol(in *like.ActSubjectProtocol) {
	out.ID = in.ID
	out.Sid = in.Sid
	out.Protocol = in.Protocol
	out.Mtime = in.Mtime
	out.Ctime = in.Ctime
	out.Types = in.Types
	out.Tags = in.Tags
	out.Pubtime = in.Pubtime
	out.Deltime = in.Deltime
	out.Editime = in.Editime
	out.Hot = in.Hot
	out.BgmID = in.BgmID
	out.PasterID = in.PasterID
	out.Oids = in.Oids
	out.ScreenSet = in.ScreenSet
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	out.XXX_sizecache = in.XXX_sizecache
	return
}

// DeepCopyAsActSubjectProtocol is an autogenerated deepcopy function, copying the receiver, creating a new like.ActSubjectProtocol.
func (in *ActSubjectProtocol) DeepCopyAsActSubjectProtocol() *like.ActSubjectProtocol {
	if in == nil {
		return nil
	}
	out := new(like.ActSubjectProtocol)
	in.DeepCopyAsIntoActSubjectProtocol(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Item) DeepCopyInto(out *Item) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Item.
func (in *Item) DeepCopy() *Item {
	if in == nil {
		return nil
	}
	out := new(Item)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyAsIntoItem is an autogenerated deepcopy function, copying the receiver, writing into like.Item.
func (in *Item) DeepCopyAsIntoItem(out *like.Item) {
	out.ID = in.ID
	out.Sid = in.Sid
	out.Type = in.Type
	out.Mid = in.Mid
	out.Wid = in.Wid
	out.State = in.State
	out.StickTop = in.StickTop
	out.Ctime = in.Ctime
	out.Mtime = in.Mtime
	return
}

// DeepCopyFromItem is an autogenerated deepcopy function, copying the receiver, writing into like.Item.
func (out *Item) DeepCopyFromItem(in *like.Item) {
	out.ID = in.ID
	out.Wid = in.Wid
	out.Ctime = in.Ctime
	out.Sid = in.Sid
	out.Type = in.Type
	out.Mid = in.Mid
	out.State = in.State
	out.StickTop = in.StickTop
	out.Mtime = in.Mtime
	return
}

// DeepCopyAsItem is an autogenerated deepcopy function, copying the receiver, creating a new like.Item.
func (in *Item) DeepCopyAsItem() *like.Item {
	if in == nil {
		return nil
	}
	out := new(like.Item)
	in.DeepCopyAsIntoItem(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Subject) DeepCopyInto(out *Subject) {
	*out = *in
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Subject.
func (in *Subject) DeepCopy() *Subject {
	if in == nil {
		return nil
	}
	out := new(Subject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyAsIntoSubjectItem is an autogenerated deepcopy function, copying the receiver, writing into like.SubjectItem.
func (in *Subject) DeepCopyAsIntoSubjectItem(out *like.SubjectItem) {
	out.ID = in.ID
	out.Oid = in.Oid
	out.Type = in.Type
	out.State = in.State
	out.Stime = in.Stime
	out.Etime = in.Etime
	out.Ctime = in.Ctime
	out.Mtime = in.Mtime
	out.Name = in.Name
	out.Author = in.Author
	out.ActURL = in.ActURL
	out.Lstime = in.Lstime
	out.Letime = in.Letime
	out.Cover = in.Cover
	out.Dic = in.Dic
	out.Flag = in.Flag
	out.Uetime = in.Uetime
	out.Ustime = in.Ustime
	out.Level = in.Level
	out.H5Cover = in.H5Cover
	out.Rank = in.Rank
	out.LikeLimit = in.LikeLimit
	out.AndroidURL = in.AndroidURL
	out.IosURL = in.IosURL
	out.DailyLikeLimit = in.DailyLikeLimit
	out.DailySingleLikeLimit = in.DailySingleLikeLimit
	out.UpLevel = in.UpLevel
	out.UpScore = in.UpScore
	out.UpUetime = in.UpUetime
	out.UpUstime = in.UpUstime
	out.FanLimitMax = in.FanLimitMax
	out.FanLimitMin = in.FanLimitMin
	out.MonthScore = in.MonthScore
	out.YearScore = in.YearScore
	out.ChildSids = in.ChildSids
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	out.XXX_sizecache = in.XXX_sizecache
	return
}

// DeepCopyFromSubjectItem is an autogenerated deepcopy function, copying the receiver, writing into like.SubjectItem.
func (out *Subject) DeepCopyFromSubjectItem(in *like.SubjectItem) {
	out.ID = in.ID
	out.Oid = in.Oid
	out.Type = in.Type
	out.State = in.State
	out.Stime = in.Stime
	out.Etime = in.Etime
	out.Ctime = in.Ctime
	out.Mtime = in.Mtime
	out.Name = in.Name
	out.Author = in.Author
	out.ActURL = in.ActURL
	out.Lstime = in.Lstime
	out.Letime = in.Letime
	out.Cover = in.Cover
	out.Dic = in.Dic
	out.Flag = in.Flag
	out.Uetime = in.Uetime
	out.Ustime = in.Ustime
	out.Level = in.Level
	out.H5Cover = in.H5Cover
	out.Rank = in.Rank
	out.LikeLimit = in.LikeLimit
	out.AndroidURL = in.AndroidURL
	out.IosURL = in.IosURL
	out.DailyLikeLimit = in.DailyLikeLimit
	out.DailySingleLikeLimit = in.DailySingleLikeLimit
	out.UpLevel = in.UpLevel
	out.UpScore = in.UpScore
	out.UpUetime = in.UpUetime
	out.UpUstime = in.UpUstime
	out.FanLimitMax = in.FanLimitMax
	out.FanLimitMin = in.FanLimitMin
	out.MonthScore = in.MonthScore
	out.YearScore = in.YearScore
	out.ChildSids = in.ChildSids
	out.AuditPlatform = in.AuditPlatform
	out.XXX_NoUnkeyedLiteral = in.XXX_NoUnkeyedLiteral
	if in.XXX_unrecognized != nil {
		in, out := &in.XXX_unrecognized, &out.XXX_unrecognized
		*out = make([]byte, len(*in))
		copy(*out, *in)
	}
	out.XXX_sizecache = in.XXX_sizecache
	return
}

// DeepCopyAsSubjectItem is an autogenerated deepcopy function, copying the receiver, creating a new like.SubjectItem.
func (in *Subject) DeepCopyAsSubjectItem() *like.SubjectItem {
	if in == nil {
		return nil
	}
	out := new(like.SubjectItem)
	in.DeepCopyAsIntoSubjectItem(out)
	return out
}

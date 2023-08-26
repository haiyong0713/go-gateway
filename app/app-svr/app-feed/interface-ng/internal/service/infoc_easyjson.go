// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package service

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService(in *jlexer.Lexer, out *showInfoc) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService(out *jwriter.Writer, in showInfoc) {
	out.RawByte('{')
	first := true
	_ = first
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v showInfoc) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v showInfoc) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *showInfoc) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *showInfoc) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService(l, v)
}
func easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService1(in *jlexer.Lexer, out *ShowListItem) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.ID = string(in.String())
		case "pos":
			out.Pos = int8(in.Int8())
		case "style":
			out.Style = string(in.String())
		case "items":
			if in.IsNull() {
				in.Skip()
				out.Items = nil
			} else {
				in.Delim('[')
				if out.Items == nil {
					if !in.IsDelim(']') {
						out.Items = make([]Item, 0, 1)
					} else {
						out.Items = []Item{}
					}
				} else {
					out.Items = (out.Items)[:0]
				}
				for !in.IsDelim(']') {
					var v1 Item
					(v1).UnmarshalEasyJSON(in)
					out.Items = append(out.Items, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService1(out *jwriter.Writer, in ShowListItem) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.String(string(in.ID))
	}
	{
		const prefix string = ",\"pos\":"
		out.RawString(prefix)
		out.Int8(int8(in.Pos))
	}
	{
		const prefix string = ",\"style\":"
		out.RawString(prefix)
		out.String(string(in.Style))
	}
	{
		const prefix string = ",\"items\":"
		out.RawString(prefix)
		if in.Items == nil && (out.Flags&jwriter.NilSliceAsEmpty) == 0 {
			out.RawString("null")
		} else {
			out.RawByte('[')
			for v2, v3 := range in.Items {
				if v2 > 0 {
					out.RawByte(',')
				}
				(v3).MarshalEasyJSON(out)
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ShowListItem) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ShowListItem) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ShowListItem) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ShowListItem) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService1(l, v)
}
func easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService2(in *jlexer.Lexer, out *ShowList) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "section":
			(out.Section).UnmarshalEasyJSON(in)
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService2(out *jwriter.Writer, in ShowList) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"section\":"
		out.RawString(prefix[1:])
		(in.Section).MarshalEasyJSON(out)
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ShowList) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ShowList) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ShowList) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ShowList) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService2(l, v)
}
func easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService3(in *jlexer.Lexer, out *Item) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeString()
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "id":
			out.ID = int64(in.Int64())
		case "pos":
			out.Pos = int8(in.Int8())
		case "type":
			out.Type = int8(in.Int8())
		case "goto":
			out.Goto = string(in.String())
		case "source":
			out.Source = string(in.String())
		case "tid":
			out.Tid = int64(in.Int64())
		case "customized_title":
			out.CustomizedTitle = string(in.String())
		case "customized_cover":
			out.CustomizedCover = string(in.String())
		case "is_gifcover":
			out.IsGifCover = int8(in.Int8())
		case "dynamic_cover":
			out.DynamicCover = int32(in.Int8())
		case "av_feature":
			out.AvFeature = string(in.String())
		case "url":
			out.URL = string(in.String())
		case "rcmd_reason":
			out.RcmdReason = string(in.String())
		case "live_pendent":
			out.LivePendent = string(in.String())
		case "hash":
			out.Hash = string(in.String())
		case "is_ad_loc":
			out.IsAdLoc = bool(in.Bool())
		case "resource_id":
			out.ResourceID = int64(in.Int64())
		case "source_id":
			out.SourceID = int64(in.Int64())
		case "creative_id":
			out.CreativeID = int64(in.Int64())
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService3(out *jwriter.Writer, in Item) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.Int64(int64(in.ID))
	}
	{
		const prefix string = ",\"pos\":"
		out.RawString(prefix)
		out.Int8(int8(in.Pos))
	}
	{
		const prefix string = ",\"type\":"
		out.RawString(prefix)
		out.Int8(int8(in.Type))
	}
	{
		const prefix string = ",\"goto\":"
		out.RawString(prefix)
		out.String(string(in.Goto))
	}
	{
		const prefix string = ",\"source\":"
		out.RawString(prefix)
		out.String(string(in.Source))
	}
	{
		const prefix string = ",\"tid\":"
		out.RawString(prefix)
		out.Int64(int64(in.Tid))
	}
	if in.CustomizedTitle != "" {
		const prefix string = ",\"customized_title\":"
		out.RawString(prefix)
		out.String(string(in.CustomizedTitle))
	}
	if in.CustomizedCover != "" {
		const prefix string = ",\"customized_cover\":"
		out.RawString(prefix)
		out.String(string(in.CustomizedCover))
	}
	if in.IsGifCover != 0 {
		const prefix string = ",\"is_gifcover\":"
		out.RawString(prefix)
		out.Int8(int8(in.IsGifCover))
	}
	if in.DynamicCover != 0 {
		const prefix string = ",\"dynamic_cover\":"
		out.RawString(prefix)
		out.Int8(int8(in.DynamicCover))
	}
	{
		const prefix string = ",\"av_feature\":"
		out.RawString(prefix)
		out.String(string(in.AvFeature))
	}
	{
		const prefix string = ",\"url\":"
		out.RawString(prefix)
		out.String(string(in.URL))
	}
	{
		const prefix string = ",\"rcmd_reason\":"
		out.RawString(prefix)
		out.String(string(in.RcmdReason))
	}
	if in.LivePendent != "" {
		const prefix string = ",\"live_pendent\":"
		out.RawString(prefix)
		out.String(string(in.LivePendent))
	}
	if in.Hash != "" {
		const prefix string = ",\"hash\":"
		out.RawString(prefix)
		out.String(string(in.Hash))
	}
	if in.IsAdLoc {
		const prefix string = ",\"is_ad_loc\":"
		out.RawString(prefix)
		out.Bool(bool(in.IsAdLoc))
	}
	if in.ResourceID != 0 {
		const prefix string = ",\"resource_id\":"
		out.RawString(prefix)
		out.Int64(int64(in.ResourceID))
	}
	if in.SourceID != 0 {
		const prefix string = ",\"source_id\":"
		out.RawString(prefix)
		out.Int64(int64(in.SourceID))
	}
	if in.CreativeID != 0 {
		const prefix string = ",\"creative_id\":"
		out.RawString(prefix)
		out.Int64(int64(in.CreativeID))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Item) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService3(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Item) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson5b4bb239EncodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService3(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Item) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService3(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Item) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson5b4bb239DecodeGoGatewayAppAppSvrAppFeedInterfaceNgInternalService3(l, v)
}
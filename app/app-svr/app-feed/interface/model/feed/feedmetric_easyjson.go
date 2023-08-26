// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package feed

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

func easyjsonDc16a580DecodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed(in *jlexer.Lexer, out *OpenAppURLParam) {
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
		case "jump":
			out.Jump = string(in.String())
		case "type":
			out.Type_ = string(in.String())
		case "id":
			out.ID = string(in.String())
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
func easyjsonDc16a580EncodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed(out *jwriter.Writer, in OpenAppURLParam) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"jump\":"
		out.RawString(prefix[1:])
		out.String(string(in.Jump))
	}
	{
		const prefix string = ",\"type\":"
		out.RawString(prefix)
		out.String(string(in.Type_))
	}
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix)
		out.String(string(in.ID))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v OpenAppURLParam) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDc16a580EncodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v OpenAppURLParam) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDc16a580EncodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *OpenAppURLParam) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDc16a580DecodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *OpenAppURLParam) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDc16a580DecodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed(l, v)
}
func easyjsonDc16a580DecodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed1(in *jlexer.Lexer, out *FeedAppListParam) {
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
		case "mid":
			out.Mid = int64(in.Int64())
		case "buvid":
			out.Buvid = string(in.String())
		case "mobi_app":
			out.MobiApp = string(in.String())
		case "device":
			out.Device = string(in.String())
		case "platform":
			out.Platform = string(in.String())
		case "build":
			out.Build = int(in.Int())
		case "ip":
			out.IP = string(in.String())
		case "ua":
			out.Ua = string(in.String())
		case "referer":
			out.Referer = string(in.String())
		case "origin":
			out.Origin = string(in.String())
		case "cdn_ip":
			out.CdnIp = string(in.String())
		case "channel":
			out.Channel = string(in.String())
		case "brand":
			out.Brand = string(in.String())
		case "model":
			out.Model = string(in.String())
		case "osver":
			out.Osver = string(in.String())
		case "applist":
			out.Applist = string(in.String())
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
func easyjsonDc16a580EncodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed1(out *jwriter.Writer, in FeedAppListParam) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"mid\":"
		out.RawString(prefix[1:])
		out.Int64(int64(in.Mid))
	}
	{
		const prefix string = ",\"buvid\":"
		out.RawString(prefix)
		out.String(string(in.Buvid))
	}
	{
		const prefix string = ",\"mobi_app\":"
		out.RawString(prefix)
		out.String(string(in.MobiApp))
	}
	{
		const prefix string = ",\"device\":"
		out.RawString(prefix)
		out.String(string(in.Device))
	}
	{
		const prefix string = ",\"platform\":"
		out.RawString(prefix)
		out.String(string(in.Platform))
	}
	{
		const prefix string = ",\"build\":"
		out.RawString(prefix)
		out.Int(int(in.Build))
	}
	{
		const prefix string = ",\"ip\":"
		out.RawString(prefix)
		out.String(string(in.IP))
	}
	{
		const prefix string = ",\"ua\":"
		out.RawString(prefix)
		out.String(string(in.Ua))
	}
	{
		const prefix string = ",\"referer\":"
		out.RawString(prefix)
		out.String(string(in.Referer))
	}
	{
		const prefix string = ",\"origin\":"
		out.RawString(prefix)
		out.String(string(in.Origin))
	}
	{
		const prefix string = ",\"cdn_ip\":"
		out.RawString(prefix)
		out.String(string(in.CdnIp))
	}
	{
		const prefix string = ",\"channel\":"
		out.RawString(prefix)
		out.String(string(in.Channel))
	}
	{
		const prefix string = ",\"brand\":"
		out.RawString(prefix)
		out.String(string(in.Brand))
	}
	{
		const prefix string = ",\"model\":"
		out.RawString(prefix)
		out.String(string(in.Model))
	}
	{
		const prefix string = ",\"osver\":"
		out.RawString(prefix)
		out.String(string(in.Osver))
	}
	{
		const prefix string = ",\"applist\":"
		out.RawString(prefix)
		out.String(string(in.Applist))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v FeedAppListParam) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDc16a580EncodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v FeedAppListParam) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDc16a580EncodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *FeedAppListParam) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDc16a580DecodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *FeedAppListParam) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDc16a580DecodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed1(l, v)
}
func easyjsonDc16a580DecodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed2(in *jlexer.Lexer, out *Discard) {
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
		case "goto":
			out.Goto = string(in.String())
		case "discard_reason":
			out.DiscardReason = int8(in.Int8())
		case "error":
			out.Error = string(in.String())
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
func easyjsonDc16a580EncodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed2(out *jwriter.Writer, in Discard) {
	out.RawByte('{')
	first := true
	_ = first
	if in.ID != 0 {
		const prefix string = ",\"id\":"
		first = false
		out.RawString(prefix[1:])
		out.Int64(int64(in.ID))
	}
	if in.Goto != "" {
		const prefix string = ",\"goto\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Goto))
	}
	if in.DiscardReason != 0 {
		const prefix string = ",\"discard_reason\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int8(int8(in.DiscardReason))
	}
	if in.Error != "" {
		const prefix string = ",\"error\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.String(string(in.Error))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v Discard) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonDc16a580EncodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v Discard) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonDc16a580EncodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *Discard) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonDc16a580DecodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *Discard) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonDc16a580DecodeGoGatewayAppAppSvrAppFeedInterfaceModelFeed2(l, v)
}

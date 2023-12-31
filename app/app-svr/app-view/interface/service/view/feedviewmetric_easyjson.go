// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package view

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

func easyjson1c7829fbDecodeGoGatewayAppAppSvrAppViewInterfaceServiceView(in *jlexer.Lexer, out *ShowListSectionItem) {
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
			out.Pos = int64(in.Int64())
		case "goto":
			out.Goto = string(in.String())
		case "source":
			out.Source = string(in.String())
		case "av_feature":
			out.AVFeature = string(in.String())
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
func easyjson1c7829fbEncodeGoGatewayAppAppSvrAppViewInterfaceServiceView(out *jwriter.Writer, in ShowListSectionItem) {
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
		out.Int64(int64(in.Pos))
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
		const prefix string = ",\"av_feature\":"
		out.RawString(prefix)
		out.String(string(in.AVFeature))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ShowListSectionItem) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson1c7829fbEncodeGoGatewayAppAppSvrAppViewInterfaceServiceView(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ShowListSectionItem) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson1c7829fbEncodeGoGatewayAppAppSvrAppViewInterfaceServiceView(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ShowListSectionItem) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson1c7829fbDecodeGoGatewayAppAppSvrAppViewInterfaceServiceView(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ShowListSectionItem) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson1c7829fbDecodeGoGatewayAppAppSvrAppViewInterfaceServiceView(l, v)
}
func easyjson1c7829fbDecodeGoGatewayAppAppSvrAppViewInterfaceServiceView1(in *jlexer.Lexer, out *ShowListSection) {
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
		case "from_item":
			out.FromItem = string(in.String())
		case "items":
			if in.IsNull() {
				in.Skip()
				out.Items = nil
			} else {
				in.Delim('[')
				if out.Items == nil {
					if !in.IsDelim(']') {
						out.Items = make([]*ShowListSectionItem, 0, 8)
					} else {
						out.Items = []*ShowListSectionItem{}
					}
				} else {
					out.Items = (out.Items)[:0]
				}
				for !in.IsDelim(']') {
					var v1 *ShowListSectionItem
					if in.IsNull() {
						in.Skip()
						v1 = nil
					} else {
						if v1 == nil {
							v1 = new(ShowListSectionItem)
						}
						(*v1).UnmarshalEasyJSON(in)
					}
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
func easyjson1c7829fbEncodeGoGatewayAppAppSvrAppViewInterfaceServiceView1(out *jwriter.Writer, in ShowListSection) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"id\":"
		out.RawString(prefix[1:])
		out.String(string(in.ID))
	}
	{
		const prefix string = ",\"from_item\":"
		out.RawString(prefix)
		out.String(string(in.FromItem))
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
				if v3 == nil {
					out.RawString("null")
				} else {
					(*v3).MarshalEasyJSON(out)
				}
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v ShowListSection) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson1c7829fbEncodeGoGatewayAppAppSvrAppViewInterfaceServiceView1(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ShowListSection) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson1c7829fbEncodeGoGatewayAppAppSvrAppViewInterfaceServiceView1(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ShowListSection) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson1c7829fbDecodeGoGatewayAppAppSvrAppViewInterfaceServiceView1(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ShowListSection) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson1c7829fbDecodeGoGatewayAppAppSvrAppViewInterfaceServiceView1(l, v)
}
func easyjson1c7829fbDecodeGoGatewayAppAppSvrAppViewInterfaceServiceView2(in *jlexer.Lexer, out *ShowList) {
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
func easyjson1c7829fbEncodeGoGatewayAppAppSvrAppViewInterfaceServiceView2(out *jwriter.Writer, in ShowList) {
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
	easyjson1c7829fbEncodeGoGatewayAppAppSvrAppViewInterfaceServiceView2(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v ShowList) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson1c7829fbEncodeGoGatewayAppAppSvrAppViewInterfaceServiceView2(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *ShowList) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson1c7829fbDecodeGoGatewayAppAppSvrAppViewInterfaceServiceView2(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *ShowList) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson1c7829fbDecodeGoGatewayAppAppSvrAppViewInterfaceServiceView2(l, v)
}

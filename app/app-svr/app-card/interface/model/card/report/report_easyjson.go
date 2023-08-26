// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package report

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

func easyjsonBd361432DecodeGoGatewayAppAppSvrAppCardInterfaceModelCardReport(in *jlexer.Lexer, out *DislikeReportData) {
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
		case "unique_id":
			out.UniqueID = string(in.String())
		case "material_id":
			out.MaterialID = int64(in.Int64())
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
func easyjsonBd361432EncodeGoGatewayAppAppSvrAppCardInterfaceModelCardReport(out *jwriter.Writer, in DislikeReportData) {
	out.RawByte('{')
	first := true
	_ = first
	if in.UniqueID != "" {
		const prefix string = ",\"unique_id\":"
		first = false
		out.RawString(prefix[1:])
		out.String(string(in.UniqueID))
	}
	if in.MaterialID != 0 {
		const prefix string = ",\"material_id\":"
		if first {
			first = false
			out.RawString(prefix[1:])
		} else {
			out.RawString(prefix)
		}
		out.Int64(int64(in.MaterialID))
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v DislikeReportData) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjsonBd361432EncodeGoGatewayAppAppSvrAppCardInterfaceModelCardReport(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v DislikeReportData) MarshalEasyJSON(w *jwriter.Writer) {
	easyjsonBd361432EncodeGoGatewayAppAppSvrAppCardInterfaceModelCardReport(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *DislikeReportData) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjsonBd361432DecodeGoGatewayAppAppSvrAppCardInterfaceModelCardReport(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *DislikeReportData) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjsonBd361432DecodeGoGatewayAppAppSvrAppCardInterfaceModelCardReport(l, v)
}
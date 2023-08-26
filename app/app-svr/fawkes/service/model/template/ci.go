package template

import cimdl "go-gateway/app/app-svr/fawkes/service/model/ci"

const (
	CIBuildInfoTemp_Mail = `<table width="100%%"><tr><td width="70" valign="top" align="right">App ID:</td><td>{{.AppID}}</td></tr>
<tr><td valign="top" align="right">App Name:</td><td>{{.AppName}}</td></tr>
<tr><td valign="top" align="right">App Key:</td><td>{{.AppKey}}</td></tr>
<tr><td valign="top" align="right">Version:</td><td>{{.Version}}({{.VersionCode}})</td></tr>
<tr><td valign="top" align="right">Internal Ver:</td><td>{{.InternalVersionCode}}</td></tr>
<tr><td valign="top" align="right">commit:</td><td>{{.Commit}}</td></tr>
<tr><td valign="top" align="right">构建号:</td><td>{{.GitlabJobID}}</td></tr>
<tr><td valign="top" align="right">job URL:</td><td>{{.GitlabJobURL}}</td></tr>
<tr><td valign="top" align="right">size:</td><td>{{.ReadableSize}}</td></tr>
<tr><td valign="top" align="right">archive:</td><td>{{.SaveURLDir}}</td></tr>
<tr><td></td><td width="50">{{.PkgURL}}</td></tr>
{{if .AddFileUrl}}<tr><td></td><td width="50">{{.FileURL}}</td></tr>{{else}}{{end}}
<tr><td></td><td><img src="{{.QrURLDir}}/install.png" /></td></tr>
<tr><td colspan="2">{{.ChangeLogHTML}}</td></tr>
</table>`

	CIBuildInfoTemp_WeChat = `{{if eq .Result "成功"}}【Fawkes】通知[{{.AppName}}] {{.GitName}} 分支{{.PackType}} 包打包{{if .Result}} {{.Result}} {{else}} 成功 {{end}}
App ID:{{.AppID}}
App Key:{{.AppKey}}
Version:{{.Version}}({{.VersionCode}})
Internal Ver:{{.InternalVersionCode}}
Commit:{{.Commit}}
构建号:{{.GitlabJobID}}
Job URL:{{.GitlabJobURL}}
Size:{{.ReadableSize}}
Archive:{{.SaveURLDir}}
{{if .PkgURL}}{{.PkgURL}}{{else}}{{end}}
{{if .AddFileUrl}}{{.FileURL}}{{else}}{{end}}
{{if .ChangeLog}}Change Log:{{.ChangeLog}}{{else}}{{end}}
{{else if eq .Result "失败"}}【Fawkes通知】{{.AppName}} [{{.GitName}} - {{.PackType}}] 打包失败

请注意，您的构建任务执行发生错误，已中断执行！

构建日志: {{.GitlabJobURL}}

{{if .GitlabJobID}}Fawkes构建记录：https://fawkes.bilibili.co/#/ci/list?app_key={{.AppKey}}&gl_job_id={{.GitlabJobID}}&order=id&sort=desc&pn=1{{end}}
{{else if eq .Result "取消"}}【Fawkes通知】{{.AppName}} [{{.GitName}} - {{.PackType}}] 打包任务已被取消

请注意，您的构建任务已被取消！

构建日志: {{.GitlabJobURL}}

{{if .GitlabJobID}}Fawkes构建记录：https://fawkes.bilibili.co/#/ci/list?app_key={{.AppKey}}&gl_job_id={{.GitlabJobID}}&order=id&sort=desc&pn=1{{end}}
{{end}}`

	APIInfo = `{{ range $key, $value := . }}
{{ $key }}:{{ range $key1, $value1 := $value}}
	{{ $key1 }}:{{ range $value1}}
		{{.DisplayName}}{{end}}{{end}}{{end}}`
)

type CIBuildInfoTemplate struct {
	AppID               string
	AppKey              string
	AppName             string
	Version             string
	VersionCode         int64
	InternalVersionCode int64
	GitlabJobID         int64
	GitName             string
	PackType            string
	Commit              string
	GitlabJobURL        string
	ReadableSize        string
	SaveURLDir          string
	PkgURL              string
	QrURLDir            string
	ChangeLog           string
	ChangeLogHTML       string
	FileURL             string
	AddFileUrl          bool
	Result              CiResult
}

type CiResult string

const (
	Failed    CiResult = "失败"
	Canceled  CiResult = "已取消"
	Running   CiResult = "开始执行"
	Succeeded CiResult = "成功"
)

func CiResultString(v int) CiResult {
	switch v {
	case cimdl.CIFailed:
		return Failed
	case cimdl.CICancel:
		return Canceled
	case cimdl.CIBuilding:
		return Running
	case cimdl.CISuccess:
		return Succeeded
	}
	return ""
}

package template

const (
	CDReleaseTemplate_Mail = `<div style="inline-height:28px;">Dear all：</div>
<div style="inline-height:28px;">{{.AppName}} {{.Version}}灰度已发布，请关注Crash及用户反馈。</div>
<div style="inline-height:28px">Version：{{.Version}}</div>
<div style="inline-height:28px">Build：{{.VersionCode}}</div>
<div style="inline-height:28px">灰度比例：{{.FlowSum}}%</div>
<br />
<div style="inline-height:28px;">更新文案由产品组提供如下：</div>
<div style="inline-height:28px;white-space:pre-line;">{{.UpgradeContentHTML}}</div>
`

	CDReleaseTemplate_WeChat = `【班车信息同步】{{.AppName}} {{.Version}}灰度已发布，请关注Crash及用户反馈。
version: {{.Version}}
build: {{.VersionCode}}
灰度比例: {{.FlowSum}}%`

	IOSCDReleaseTemplate_Mail = `<div style="inline-height:28px;font-family:'Microsoft YaHei';">Dear all，</div>
<br />
<div style="inline-height:28px;font-family:'Microsoft YaHei';">【班车信息同步】{{.AppName}} {{.Version}} {{if eq .PackType 9}}TestFlight已发布{{else}}已全量发布市场{{end}}，请关注Crash及用户反馈。</div>
<div style="inline-height:28px;font-family:'Microsoft YaHei';">Version：{{.Version}}</div>
<div style="inline-height:28px;font-family:'Microsoft YaHei';">Build：{{.VersionCode}}</div>
<div style="inline-height:28px;font-family:'Microsoft YaHei';">{{if eq .PackType 9}}灰度比例：{{.TFFlowSum}}%{{else}}{{end}}</div>`
	IOSCDReleaseTemplate_WeChat = `【班车信息同步】{{.AppName}} {{.Version}} {{if eq .PackType 9}}TestFlight已发布{{else}}已全量发布市场{{end}}，请关注Crash及用户反馈。
version: {{.Version}}
build: {{.VersionCode}}
{{if eq .PackType 9}}灰度比例: {{.TFFlowSum}}%{{else}}{{end}}`
)

type CDReleaseTemplate struct {
	AppName            string
	Version            string
	VersionCode        int64
	PackType           int8
	UpgradeContent     string
	UpgradeContentHTML string
	FlowSum            int
	TFFlowSum          int
}

package template

const ModTrafficNotify = `MOD{{.OperateType}}带宽预估
【成本】：{{if eq .Cost 1 }}低{{else if eq .Cost 2}}中等{{else if eq .Cost 3}}高{{else if eq .Cost 4}}极高{{else}}未知{{end}}
【下载量】: 预计提升{{.Percentage}} [{{.DownloadSizeOnline}}/5min] -> [{{.DownloadSizeOnlineTotal}}/5min]
【带宽】: [{{.DownloadCDNBandwidthOnline}}/s] -> [{{.DownloadCDNBandwidthTotal}}/s]
【计算方式】: [预估下载量:{{.DownloadSizeEstimate}}] = [预估下载次数:{{.DownloadCount}}]x[文件体积:{{.AvgFileSize}}]
【Mod信息】: 应用-{{.AppKey}} 资源池-{{.PoolName}} 资源-{{.ModName}} 版本-{{.VerNum}} 优先级-{{if .IsManual}}手动{{else}}自动{{end}} 操作人-{{.Operator}}
【Fawkes发布记录】: {{.ModUrl}}
{{if .Doc}} 可参考文档{{.Doc}}获取更多发布建议，{{end}}如有疑问请联系@Fawkes小姐姐。
`

const ModTrafficNotifyErr = `MOD发布带宽预估【计算失败】
[错误信息]：{{.ErrorMsg}}
[Mod信息]: 应用-{{.AppKey}} 资源池-{{.PoolName}} 资源-{{.ModName}}
[Fawkes发布记录]: {{.ModUrl}}`

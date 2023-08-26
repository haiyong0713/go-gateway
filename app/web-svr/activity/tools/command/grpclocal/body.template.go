package main

var (
	_bodyTemplate = `
func (s *{{.Type}}) {{.Method}}({{.Params}}) ({{.Results}}) {
	if local{{.Service}}Server == nil {
		panic("Call InitLocal{{.Service}}Server First")
	}
	grpclocal.ServerLogging(ctx, "/"+_{{.Service}}_serviceDesc.ServiceName+"/{{.Method}}", in.String(), func() error {
		{{.Rets}} = local{{.Service}}Server.{{.Method}}({{.Vars}})
		return err
	})
	return
}
`
)

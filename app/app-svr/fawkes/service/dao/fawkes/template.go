package fawkes

import (
	"bytes"
	"text/template"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

func (d *Dao) TemplateAlter(data interface{}, alterTemplate string) (content string, err error) {
	var (
		tmpl *template.Template
		buf  bytes.Buffer
	)
	if tmpl, err = template.New("").Parse(alterTemplate); err != nil {
		log.Error("templateAlter parse %v", err)
		return
	}
	if err = tmpl.Execute(&buf, data); err != nil {
		log.Error("templateAlter execute %v", err)
		return
	}
	content = buf.String()
	return
}

func (d *Dao) TemplateAlterFunc(data interface{}, funcMap template.FuncMap, alterTemplate string) (content string, err error) {
	var (
		tmpl *template.Template
		buf  bytes.Buffer
	)
	if tmpl, err = template.New("").Funcs(funcMap).Parse(alterTemplate); err != nil {
		log.Error("templateAlter parse %v", err)
		return
	}
	if err = tmpl.Execute(&buf, data); err != nil {
		log.Error("templateAlter execute %v", err)
		return
	}
	content = buf.String()
	return
}

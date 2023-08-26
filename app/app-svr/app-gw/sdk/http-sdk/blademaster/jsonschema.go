package blademaster

import (
	"bytes"
	"encoding/json"
	"io"
	"net/url"
	"path"

	"go-common/library/conf/paladin"

	"github.com/pkg/errors"
	gjf "github.com/xeipuuv/gojsonreference"
	gjs "github.com/xeipuuv/gojsonschema"
)

type svenJSONLoader struct {
	source   string
	filename string
}

func (l *svenJSONLoader) JsonSource() interface{} {
	return l.source
}

func (l *svenJSONLoader) JsonReference() (gjf.JsonReference, error) {
	return gjf.NewJsonReference("#")
}

func (l *svenJSONLoader) LoaderFactory() gjs.JSONLoaderFactory {
	return &gjs.DefaultJSONLoaderFactory{}
}

func (l *svenJSONLoader) LoadJSON() (interface{}, error) {
	raw, err := paladin.Get(l.filename).String()
	if err != nil {
		return nil, err
	}
	return decodeJSONUsingNumber(bytes.NewReader([]byte(raw)))
}

func NewSvenJSONLoader(source string) (gjs.JSONLoader, error) {
	src, err := url.Parse(source)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &svenJSONLoader{
		source:   source,
		filename: path.Base(src.Path),
	}, nil
}

func decodeJSONUsingNumber(r io.Reader) (interface{}, error) {
	var document interface{}
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	err := decoder.Decode(&document)
	if err != nil {
		return nil, err
	}
	return document, nil
}

func NewJSONLoader(loader, ref string) (gjs.JSONLoader, error) {
	_, err := url.Parse(ref)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	switch loader {
	case "sven":
		return NewSvenJSONLoader(ref)
	case "reference":
		return gjs.NewReferenceLoader(ref), nil
	default:
		return nil, errors.Errorf("invalid json loader scheme: %s", loader)
	}
}

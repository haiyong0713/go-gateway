package asset

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func bindata_read(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	//nolint:gosec
	_, err = io.Copy(&buf, gz)
	gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	return buf.Bytes(), nil
}

var _metric_html = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x58\xcf\x4e\xe4\x46\x13\xbf\x23\xf1\x0e\xb5\xfe\x0e\x06\x09\x6c\xcf\x0c\xc3\xb0\xc6\x1e\xe9\x0b\x44\x28\x4a\x24\xb2\x81\x44\x49\x10\x8a\x7a\xdc\xcd\xb8\xc1\xe3\x76\xdc\xe5\x61\x10\x72\x0e\x7b\xcc\x21\x52\xce\x79\x80\x3c\x43\x5e\x28\x9b\xc7\x88\xda\x7f\xb0\x07\x6c\x03\x13\x50\xa4\x68\x7d\xd8\xed\xa9\x5f\x75\xfd\xaa\xaa\xab\xba\x9b\x76\xde\x1c\x1e\x1f\x9c\x7e\xf7\xe5\xa7\xe0\xe3\x2c\x18\xaf\xaf\x39\xc5\xff\x6a\xc4\x08\x1d\xaf\xaf\x01\x38\x33\x86\x04\x3c\x9f\xc4\x92\xa1\xab\x25\x78\xb1\xbd\xa7\xe5\x08\x72\x0c\xd8\xf8\xaf\x9f\xdf\x7f\x78\xff\xc7\x87\xdf\x7e\xfd\xf3\x97\xdf\x1d\x33\x97\x65\x70\xc0\xc3\x2b\x88\x59\xe0\x6a\x12\x6f\x02\x26\x7d\xc6\x50\x03\xbc\x89\x98\xab\x21\x5b\xa0\xe9\x49\xa9\x81\x1f\xb3\x0b\x57\xf3\x11\x23\x69\x9b\x26\xf1\x90\xcf\x39\xde\x18\x3e\x95\xc1\xc4\xf0\xc4\xcc\x9c\x04\xc4\xbb\x9a\x08\x12\x53\x53\x22\x41\xee\x99\x7d\xab\x6f\x59\x83\xde\x9e\x69\xbd\x1d\x8d\x46\xbb\xa3\x49\x9f\x8d\xde\xd2\x3d\x42\x2d\x32\xd8\x25\xa3\x41\xcf\xda\xdd\x23\x7b\x03\x3a\x32\x8f\xdf\x9d\x1c\x7e\xfb\xb9\x3c\x31\x14\x55\xee\x95\xf4\x62\x1e\x61\xdd\x8d\x4b\x32\x27\xb9\x54\x03\x19\x7b\xaf\xe6\xcc\x0e\x99\x1c\x1c\x9c\x1e\xfd\xdf\x37\x2e\xa5\x36\x76\xcc\x9c\xf3\xdf\xf6\xea\x68\x78\xfc\x53\xe4\xcd\xaf\xdb\x9d\x7a\x4d\x76\xef\xdd\xd5\x27\x27\x47\x27\x8d\xec\xaa\x68\xee\x97\x4b\x06\x01\xfc\x8f\x44\x11\xdc\xe6\x63\x80\x88\x50\xca\xc3\xa9\x0d\x7d\x2b\x5a\xc0\xd0\x8a\x16\xfb\x39\x94\xaa\x52\x56\x03\xbf\xd7\xa6\xdd\xa0\x6c\xf0\x79\xb2\x8d\x64\xa2\xd8\xe9\xd6\xd2\x4f\xff\xce\x8a\xcf\xf8\xd4\x47\x1b\xfa\xfd\x68\xf1\x86\xcf\x22\x11\x23\x09\xb1\x32\x05\xe0\x98\x59\x04\xaa\xad\xcc\xa2\x9b\xd6\xd7\x9c\x89\xa0\x37\x4a\x44\xf9\x1c\x38\x75\x35\x12\x45\x45\x61\x52\x3e\x1f\xf3\x50\x22\x09\x3d\x66\xc3\xed\x6d\x39\x4e\x53\xc7\x54\xd8\x9d\x52\x12\x21\x9f\x65\x2a\xf9\xe8\x81\x42\xc4\xa9\x42\x23\x4e\x97\x20\xbf\x37\xfe\x42\x78\x04\xb9\x08\x1d\xd3\xef\xe5\x42\x5e\x84\x66\x7b\x22\x48\x66\xa1\x74\xb5\x62\xa0\x81\x4d\x09\x12\x57\x93\x2c\x9e\xb3\xf8\x07\xf5\x43\x83\x89\x88\x29\x8b\x8b\x55\x70\x90\xcd\xa2\x80\x20\x03\x19\x08\xdc\x96\x9e\x50\x6b\x75\x0b\xb1\xb8\x86\x54\xcb\x84\xae\xd6\x5f\x2c\xca\x55\x03\xb8\xcd\x40\x23\x66\x32\x12\xa1\x64\x67\x19\x7a\x0e\x69\x5a\x18\x34\x4b\x8b\xcf\x61\xd8\xe9\x64\xd8\x79\x01\x86\x61\x27\xc3\xf0\x05\x18\x7a\x9d\x0c\xbd\x17\x60\x18\x74\x32\x0c\x5e\x80\x21\x09\xaf\x42\x71\x1d\x76\xb0\x94\x1a\xff\x94\x09\x05\x92\xa0\x83\x27\xc7\x5b\x59\x1c\xb3\xa8\xfa\xbb\xbe\xf8\x3a\x92\x18\x33\x32\x7b\x4e\x5f\x04\x45\x2f\x7d\xec\x8c\x8f\x9d\xf1\x5f\xec\x8c\xf2\xe4\x70\x6a\x87\x72\xc8\xae\xe1\x9b\x84\x6d\x14\xc7\x20\x0b\x6c\xd0\xd5\x59\xac\x6f\xe5\x02\xd5\x0a\x76\x75\xd4\x16\x2d\x63\xc3\x59\x29\x81\x0a\x54\x5f\x76\x5b\xb4\x41\xff\x5e\x84\xac\xb4\x91\x7f\x57\xec\xc6\x06\x3d\x24\xb3\x7b\xf2\x6b\x4e\xd1\xb7\xa1\x3f\xb4\x2a\x69\xba\xf5\x88\xf9\xaf\xd8\x8f\x09\x93\x28\x97\x4d\x91\x80\x4f\x43\x1b\x74\x8f\x85\xc8\xe2\x65\xcc\xf3\x79\x40\x63\x16\x2e\xb9\xfe\xc0\x7e\x9d\xe3\x54\x65\x76\xd9\x48\x15\x47\xdc\xe8\xc0\x63\x4e\x14\x17\x89\xf2\x3b\xaf\x47\xdc\x16\x68\xbe\xd0\xaf\x1a\x69\x7f\xb1\x78\x18\x87\x2a\xbf\x36\xec\x91\x18\xb7\xda\xa9\x76\x3a\xa8\x1a\xb1\xd5\xa9\x86\x1d\x54\x8d\xd8\xea\x54\xbd\x0e\xaa\x46\x6c\x75\xaa\x41\x07\x55\x23\xb6\x3a\x55\xb1\x85\xb5\xd2\xb5\xe2\xab\x53\xb6\xf4\x5c\x41\x88\xcd\xe8\x8a\x1d\x57\x0e\xcf\xef\x26\xd4\x2e\xc6\x36\x9c\x55\xf2\xa5\x8b\xc1\x12\x72\x77\xbd\xd7\x2b\xda\xe2\x36\x5f\x93\xa8\xeb\xbb\xae\x17\x7f\x49\x14\xd2\x19\x43\x5f\x50\x59\xdb\x5b\xa7\x0c\x0f\x33\xfb\x17\x49\xe8\x29\x3a\xd8\xd8\xac\x67\x2a\x60\x08\xe8\x13\x04\x17\xd0\xe7\xb2\x02\x92\x38\x00\x17\x74\xc3\x9c\x31\x8c\xb9\x27\x8d\x4b\x29\x42\x7d\xbf\x52\x20\x0b\x2e\xa4\x31\x65\xb8\x91\xc4\xc1\xe6\xfd\xf4\x01\x18\xe8\xb3\x70\xa3\xa2\x2d\x4f\x97\xcd\x87\x0b\x95\x05\x7d\x51\xa9\x18\x2a\x25\x86\x27\x28\x03\xd7\x75\xc1\x6a\x99\x02\x99\xe7\x46\x2d\xc1\xe0\xc2\xb2\x8d\xec\x9f\x32\xd3\xfb\x1d\x46\x96\x56\xa3\xd9\x4c\x52\xdc\xfe\xba\xcc\x94\x4b\xd7\x6c\xa1\x44\xbb\x2c\xe4\x0b\xdd\xe6\x81\xc2\xba\x66\x47\x9c\x36\x4f\x8d\x38\x6d\x9c\x97\x02\x0b\x24\x7b\x46\x7e\xcf\xce\x9f\x93\xc6\x6e\xed\x5a\xb6\x74\xfd\x49\x49\xe9\x56\xcb\xa3\x6f\xd1\x49\x1f\x0a\xd3\xa6\xb2\xf5\x08\x7a\x7e\xad\x6e\x59\x1c\x8b\xb8\xa5\x02\x3d\x11\x4a\x11\x30\x23\x10\xd3\x42\xaf\x91\xfa\xe9\x69\x7c\x4e\x12\x9f\x98\xc2\x27\x25\xf0\xb1\xf4\xa5\x55\x60\xe9\xbd\x2d\x47\x24\x21\x32\x5a\xdf\x55\x24\xc3\xcf\xd4\xa6\x39\x27\xc1\x86\xda\x54\x8c\x62\x13\xda\x82\x9e\x65\x59\x9b\xf5\xd7\x0f\xb5\x02\xb5\xf7\x1c\xc7\x2c\x1e\x3e\xb2\xd7\x90\xfc\x95\xf1\xef\x00\x00\x00\xff\xff\x7b\x91\xe8\x7c\x7f\x14\x00\x00")

func metric_html() ([]byte, error) {
	return bindata_read(
		_metric_html,
		"metric.html",
	)
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		return f()
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() ([]byte, error){
	"metric.html": metric_html,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for name := range node.Children {
		rv = append(rv, name)
	}
	return rv, nil
}

type _bintree_t struct {
	Func     func() ([]byte, error)
	Children map[string]*_bintree_t
}

var _bintree = &_bintree_t{nil, map[string]*_bintree_t{
	//nolint:gofmt
	"metric.html": {metric_html, map[string]*_bintree_t{}},
}}

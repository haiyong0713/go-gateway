package http

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"

	"go-common/library/log"
)

const _grpcHeaderSize = 5
const _encodeFlag = '\x01'

func unGzip(request *http.Request) bool {
	if request.Header.Get("grpc-encoding") != "gzip" {
		return false
	}
	if request.Body == nil {
		return false
	}
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Error("gzip.read request body(url: %v) err: %v", request.URL, err)
		return false
	}
	r := readerCloser{Reader: bytes.NewReader(body)}
	request.Body = r
	if len(body) == 0 || body[0] != _encodeFlag {
		return false
	}
	if len(body) < _grpcHeaderSize {
		return false
	}
	decodeBody, err := gzipDe(body[_grpcHeaderSize:])
	if err != nil {
		log.Error("gzip.de body(url: %v body:%v) err: %v", request.URL, body, err)
		return false
	}
	rawbody, err := BuildPacket(decodeBody)
	if err != nil {
		log.Error("gzip.en body(url: %v body:%v) err: %v", request.URL, decodeBody, err)
		return false
	}
	r = readerCloser{Reader: bytes.NewReader(rawbody)}
	request.Body = r
	request.Header.Set("Content-Length", fmt.Sprint(len(rawbody)))
	return true
}

type readerCloser struct {
	*bytes.Reader
}

func (r readerCloser) Close() error {
	return nil
}

func BuildPacket(message []byte) ([]byte, error) {
	buf := &bytes.Buffer{}
	lenBytes := make([]byte, 4)
	buf.WriteByte('\x00')
	binary.BigEndian.PutUint32(lenBytes, uint32(len(message)))
	buf.Write(lenBytes)
	buf.Write(message)
	return buf.Bytes(), nil
}

func gzipDe(in []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(in))
	if err != nil {
		var out []byte
		return out, err
	}
	defer reader.Close()
	return ioutil.ReadAll(reader)
}

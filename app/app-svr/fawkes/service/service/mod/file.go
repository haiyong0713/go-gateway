package mod

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const _bossBucket = "appstaticboss"

func (s *Service) fileUpload(ctx context.Context, filename, md5 string, fileData []byte) (string, error) {
	if len(fileData) == 0 {
		return "", errors.New("文件内容不能为空")
	}
	path := fmt.Sprintf("%s/%s", md5, filename)
	if _, err := s.boss.PutObject(ctx, _bossBucket, path, fileData); err != nil {
		return "", err
	}
	return fmt.Sprintf("/%s/%s", _bossBucket, path), nil
}

func parseFileData(data []byte) (contentType, md5Value string, size int64, err error) {
	h := md5.New()
	if _, err := io.Copy(h, bytes.NewReader(data)); err != nil {
		return "", "", 0, err
	}
	return http.DetectContentType(data), hex.EncodeToString(h.Sum(nil)), int64(len(data)), nil
}

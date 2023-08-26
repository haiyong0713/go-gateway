package unicom

import (
	"crypto/md5"
	"encoding/hex"
)

func (s *Service) FlowSign(timestamp string) (res string) {
	mh := md5.Sum([]byte(s.c.Unicom.Cpid + timestamp + s.c.Unicom.Password))
	return hex.EncodeToString(mh[:])
}

package view

import (
	"hash/crc32"
	"strconv"
	"strings"

	"go-common/library/log"

	"github.com/pkg/errors"
)

func (s *Service) matchNGBuilder(mid int64, buvid, feature string) bool {
	if s.c.NgSwitch.DisableAll {
		return false
	}
	if feature == "" {
		return false
	}
	if mid == 0 && buvid == "" {
		return false
	}
	policy, ok := s.c.NgSwitch.Feature[feature]
	if !ok {
		return false
	}
	if len(policy) == 0 {
		return true
	}
	for _, v := range policy {
		fn, err := parsePolicy(v)
		if err != nil {
			log.Error("Failed to parse policy: %+v", err)
			continue
		}
		if fn(mid, buvid) {
			return true
		}
	}
	return false
}

//nolint:gomnd
func parsePolicy(in string) (func(mid int64, buvid string) bool, error) {
	parts := strings.Split(in, ":")
	if len(parts) != 2 {
		return nil, errors.Errorf("Invalid policy: %q", in)
	}
	switch parts[0] {
	case "mid":
		matchMid, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return func(mid int64, buvid string) bool {
			return matchMid == mid
		}, nil
	case "buvid":
		matchBuvid := parts[1]
		return func(mid int64, buvid string) bool {
			return matchBuvid == buvid
		}, nil
	case "mid_mod":
		mmParts := strings.Split(parts[1], ",")
		if len(mmParts) != 2 {
			return nil, errors.Errorf("Invalid mid_mod policy: %q", parts[1])
		}
		mod, err := strconv.ParseInt(mmParts[0], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if mod == 0 {
			return nil, errors.Errorf("Invalid mid_mod policy: %q", parts[1])
		}
		pivtoal, err := strconv.ParseInt(mmParts[1], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return func(mid int64, buvid string) bool {
			return mid%mod <= pivtoal
		}, nil
	case "buvidcrc32_mod":
		bcmParts := strings.Split(parts[1], ",")
		if len(bcmParts) != 2 {
			return nil, errors.Errorf("Invalid mid_mod policy: %q", parts[1])
		}
		mod, err := strconv.ParseInt(bcmParts[0], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		if mod == 0 {
			return nil, errors.Errorf("Invalid mid_mod policy: %q", parts[1])
		}
		pivtoal, err := strconv.ParseInt(bcmParts[1], 10, 64)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return func(mid int64, buvid string) bool {
			return int64(crc32.ChecksumIEEE([]byte(buvid)))%mod < pivtoal
		}, nil
	default:
		return nil, errors.Errorf("Invalid policy: %q", in)
	}
}

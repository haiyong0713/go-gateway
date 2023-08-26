package service

import (
	"context"
	"strings"
)

func (s *Service) GetOnelink(c context.Context, mid int64, gid int64, source string) (string, error) {
	const (
		_play = 2
		_down = 0
	)
	rs, err := s.dao.GetOnelinkAccess(c, mid, gid)
	if err != nil {
		return "", err
	}
	if !(rs.Play == _play && rs.Down == _down) {
		return "", nil
	}

	var res string
	switch true {
	// 谷歌搜索
	case source == "search_google":
		res = "https://bilibili1.onelink.me/8pOD/fruxpl4i"
	// 分享
	case judgeBsource("share", source):
		res = "https://bilibili1.onelink.me/8pOD/cuwyzp7a"
	// 未知
	case source == "default":
		res = "https://bilibili1.onelink.me/8pOD/cnu4o02i"
	// 其他
	default:
		res = "https://bilibili1.onelink.me/8pOD/3o1pmdw0"
	}
	return res, nil
}

func judgeBsource(target string, source string) bool {
	return strings.HasPrefix(source, target)
}

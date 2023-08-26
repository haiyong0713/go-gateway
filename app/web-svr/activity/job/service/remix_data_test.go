package service

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRemixDataResult(t *testing.T) {
	Convey("test remix RemixDataResult", t, WithService(func(s *Service) {
		s.RemixDataResult()
	}))
}

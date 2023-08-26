package service

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDataResult(t *testing.T) {
	Convey("test handwrite DataResult", t, WithService(func(s *Service) {
		s.DataResult()
	}))
}

package article

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestConsumeArticleBinlog(t *testing.T) {
	Convey("consumeArticleBinlog", t, WithService(func(s *Service) {
		s.consumeArticleBinlog()
	}))
}

package article

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRetryArticleBinlog(t *testing.T) {
	Convey("retryArticleBinlog", t, WithService(func(s *Service) {
		s.retryArticleBinlog()
	}))
}

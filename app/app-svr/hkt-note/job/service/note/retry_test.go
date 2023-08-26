package note

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRetryDetailDB(t *testing.T) {
	Convey("retryDetailDB", t, WithService(func(s *Service) {
		s.retryDetailDB()
	}))
}

func TestRetryDetailDBDel(t *testing.T) {
	Convey("retryDetailDBDel", t, WithService(func(s *Service) {
		s.retryDetailDBDel()
	}))
}

func TestRetryArtDetailDB(t *testing.T) {
	Convey("retryArtDetailDB", t, WithService(func(s *Service) {
		s.retryArtDetailDB()
	}))
}

func TestRetryArtBinlog(t *testing.T) {
	Convey("retryArtBinlog", t, WithService(func(s *Service) {
		s.retryArtBinlog()
	}))
}

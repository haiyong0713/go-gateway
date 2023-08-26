package utils

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestHumanFileSize(t *testing.T) {
	Convey("", t, func() {
		input := 100000000000 * 2558908 / 0.1
		output := HumanFileSize(input)
		So(output, ShouldNotBeEmpty)
	})
}

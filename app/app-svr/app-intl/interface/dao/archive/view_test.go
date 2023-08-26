package archive

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestView(t *testing.T) {
	Convey(t.Name(), t, func() {
		d.View(context.Background(), 1)
	})
}
func TestDescription(t *testing.T) {
	Convey(t.Name(), t, func() {
		d.Description(context.Background(), 2)
	})
}

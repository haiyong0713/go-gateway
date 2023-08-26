package thread_sampler

import (
	"fmt"
	"testing"
)

func TestFloatEqual(t *testing.T) {
	a := 0.00001
	b := 0.00001
	fmt.Println("QQQ", floatEqual(a, b))
}

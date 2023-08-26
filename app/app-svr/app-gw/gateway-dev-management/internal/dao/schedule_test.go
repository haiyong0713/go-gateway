package dao

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelectValueByKey(t *testing.T) {
	res, err := d.SelectValueByKey(ctx(), "sre_schedule")
	assert.NoError(t, err)
	fmt.Println(res)
}

func TestUpdateValueByKey(t *testing.T) {
	err := d.UpdateValueByKey(ctx(), "sre_arrange", strconv.Itoa(0))
	assert.NoError(t, err)
}

func TestSelectConfigs(t *testing.T) {
	res, err := d.SelectConfigs(ctx())
	assert.NoError(t, err)
	for _, i := range res {
		fmt.Println(i)
	}
}

package request

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrParamRequired(t *testing.T) {
	ErrPR := NewErrParamRequired("d")
	assert.Equal(t, ErrPR.Code(), "ParamRequiredError")
	msgTest := "missing required field, d."
	assert.Equal(t, ErrPR.Message(), msgTest)
	assert.Equal(t, "ParamRequiredError: missing required field, d.", ErrPR.Error())
	assert.Equal(t, nil, ErrPR.OrigErr())
	assert.Equal(t, "d", ErrPR.Field())
	ErrPR.SetContext("ctx")
	assert.Equal(t, "ctx.d", ErrPR.Field())
	ErrPR.AddNestedContext("NestedContext")
	assert.Equal(t, ErrPR.errInvalidParam.nestedContext, "NestedContext")
	ErrPR.AddNestedContext("add")
	assert.Equal(t, ErrPR.errInvalidParam.nestedContext, "add.NestedContext")
}

func TestErrParamFormat(t *testing.T) {
	ErrPF := NewErrParamFormat("d", "UTF-8", "val")
	assert.Equal(t, ErrPF.Format(), "UTF-8")
	msgTest := "format UTF-8, val, d."
	assert.Equal(t, ErrPF.Message(), msgTest)
}

func TestMinValue(t *testing.T) {
	ErrPM := NewErrParamMinValue("d", 2.0)
	assert.Equal(t, ErrPM.MinValue(), 2.0)
	msgTest := "minimum field value of 2, d."
	assert.Equal(t, ErrPM.Message(), msgTest)
}

func TestErrParamMinAndMaxLen(t *testing.T) {
	ErrMinL := NewErrParamMinLen("d", 2)
	ErrMaxL := NewErrParamMaxLen("d", 4, "val")
	assert.Equal(t, ErrMinL.MinLen(), 2)
	assert.Equal(t, ErrMaxL.MaxLen(), 4)
	msgMinT := "minimum field size of 2, d."
	msgMaxT := "maximum size of 4, val, d."
	assert.Equal(t, ErrMaxL.Message(), msgMaxT)
	assert.Equal(t, ErrMinL.Message(), msgMinT)
}

func TestErrInvalidParams(t *testing.T) {
	errInvalidPra := &errInvalidParam{}
	errInvalidPra1 := &errInvalidParam{}
	ErrInvalidPra := &ErrInvalidParams{
		Context: "test",
	}
	nested := ErrInvalidParams{
		Context: "test2",
	}
	ErrInvalidPra.Add(errInvalidPra)
	nested.Add(errInvalidPra1)
	assert.Equal(t, errInvalidPra.context, "test")
	ErrInvalidPra.AddNested("nestedtest", nested)
	assert.Equal(t, errInvalidPra1.context, "test")
	assert.Equal(t, errInvalidPra1.nestedContext, "nestedtest")
	assert.Equal(t, ErrInvalidPra.Len(), 2)
	assert.Equal(t, ErrInvalidPra.Code(), "InvalidParameter")
	assert.Equal(t, ErrInvalidPra.Message(), "2 validation error(s) found.")
	s := ErrInvalidPra.Error()
	assert.Equal(t, s, "InvalidParameter: 2 validation error(s) found.\n- , test..\n- , test.nestedtest..\n")
}

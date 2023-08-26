package prettyecode

import (
	"fmt"
	"go-common/library/ecode"

	"github.com/pkg/errors"
)

type PrettyEcode struct {
	ecode.Codes
	inner error
}

func WithError(ecode ecode.Codes, err error) error {
	return &PrettyEcode{Codes: ecode, inner: err}
}

func New(ecode ecode.Codes, msg string) error {
	return &PrettyEcode{Codes: ecode, inner: errors.New(msg)}
}

func (pe *PrettyEcode) Message() string {
	return pe.inner.Error()
}

func (pe *PrettyEcode) Error() string {
	return pe.inner.Error()
}

func (pe *PrettyEcode) Format(s fmt.State, verb rune) {
	innerFmt, ok := pe.inner.(fmt.Formatter)
	if ok {
		innerFmt.Format(s, verb)
	}
}

func WithRawError(err error) error {
	switch err.(type) {
	case *ecode.Status:
		return err
	default:
		return &PrettyEcode{Codes: ecode.ServerErr, inner: err}
	}
}

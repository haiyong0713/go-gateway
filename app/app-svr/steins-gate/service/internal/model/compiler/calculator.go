package compiler

import (
	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/ecode"

	"github.com/pkg/errors"
)

const (
	floatTrue  = float64(1)
	floatFalse = float64(0)
)

// Calculator def.
type Calculator struct {
	Stream *TokenStream
	Values map[string]float64
}

// InitAndEval inits a calculator and eval the given expr's value
func (v *Calculator) InitAndEval(expr string, hvarRec map[string]float64) (result float64, err error) {
	v.Values = hvarRec
	v.Stream = &TokenStream{
		Input:        expr,
		CurrentToken: &Token{Kind: kEnd},
	}
	for {
		if err = v.Stream.GetNext(); err != nil {
			log.Error("%+v", err)
			return
		}
		cToken := v.Stream.CurrentToken
		if cToken.Kind == kEnd {
			break
		}
		if cToken.Kind == kPrint {
			continue
		}
		if result, err = v.LogicExpr(false); err != nil {
			log.Error("%+v", err)
			return
		}
		if nowKind := v.Stream.CurrentToken.Kind; nowKind != kPrint && nowKind != kEnd {
			err = errors.Wrapf(ecode.ExprUnexpectedChar, "currentToken %+v", v.Stream.CurrentToken)
			return
		}
	}
	return
}

// LogicExpr treats the AND/OR
func (v *Calculator) LogicExpr(readNext bool) (numValue float64, err error) {
	if numValue, err = v.Comp(readNext); err != nil {
		return
	}
	for {
		var (
			cToken = v.Stream.CurrentToken
			nextV  float64
		)
		if cToken.Kind != kAnd && cToken.Kind != kOr { // no and/or to treat
			return
		}
		if nextV, err = v.Comp(true); err != nil {
			return
		}
		//nolint:exhaustive
		switch cToken.Kind {
		case kAnd:
			numValue = BoolToFloat(numValue != floatFalse && nextV != floatFalse)
		case kOr:
			numValue = BoolToFloat(numValue != floatFalse || nextV != floatFalse)
		}
	}
	//nolint:govet
	return
}

// Comp treats the logic comparison
func (v *Calculator) Comp(readNext bool) (numValue float64, err error) {
	if numValue, err = v.NumExpr(readNext); err != nil {
		return
	}
	for {
		var (
			cToken = v.Stream.CurrentToken
			nextV  float64
		)
		if cToken.Kind != kEqual && cToken.Kind != kNotEqual && cToken.Kind != kGreater &&
			cToken.Kind != kGreaterEqual && cToken.Kind != kLess && cToken.Kind != kLessEqual { // no comp to treat
			return
		}
		if nextV, err = v.NumExpr(true); err != nil {
			return
		}
		//nolint:exhaustive
		switch cToken.Kind {
		case kEqual:
			numValue = BoolToFloat(numValue == nextV)
		case kNotEqual:
			numValue = BoolToFloat(numValue != nextV)
		case kGreater:
			numValue = BoolToFloat(numValue > nextV)
		case kGreaterEqual:
			numValue = BoolToFloat(numValue >= nextV)
		case kLess:
			numValue = BoolToFloat(numValue < nextV)
		case kLessEqual:
			numValue = BoolToFloat(numValue <= nextV)
		}
	}
	//nolint:govet
	return
}

// BoolToFloat treats the bool, true => 1 , false => 0
func BoolToFloat(v bool) float64 {
	if v {
		return floatTrue
	}
	return floatFalse
}

// NumExpr treats the addition and the subtraction
func (v *Calculator) NumExpr(readNext bool) (numValue float64, err error) {
	if numValue, err = v.Term(readNext); err != nil {
		return
	}
	for { // used to treat continuous calculation
		var (
			cToken = v.Stream.CurrentToken
			nextV  float64
		)
		if cToken.Kind != kPlus && cToken.Kind != kMinus { // no add/minus to treat
			return
		}
		if nextV, err = v.Term(true); err != nil {
			return
		}
		//nolint:exhaustive
		switch cToken.Kind {
		case kPlus:
			numValue = numValue + nextV
		case kMinus:
			numValue = numValue - nextV
		}
	}
	//nolint:govet
	return
}

// Trem treats the multiplication and the division
func (v *Calculator) Term(readNext bool) (numValue float64, err error) {
	if numValue, err = v.Prim(readNext); err != nil {
		return
	}
	for {
		var (
			cToken = v.Stream.CurrentToken
			nextV  float64
		)
		if cToken.Kind != kMul && cToken.Kind != kDiv { // no term to treat
			return
		}
		if nextV, err = v.Prim(true); err != nil { // div & multi needs to read the next token
			return
		}
		//nolint:exhaustive
		switch cToken.Kind {
		case kMul:
			numValue = numValue * nextV
		case kDiv:
			if nextV == 0 {
				err = ecode.ExprDivideByZero
				return
			}
			numValue = numValue / nextV
		}
	}
	//nolint:govet
	return
}

// Prim treats the basic unit of expr like +/-/!/var/number
func (v *Calculator) Prim(readNext bool) (numValue float64, err error) {
	if readNext { // by default, we read next token, but the first time we don't
		if err = v.Stream.GetNext(); err != nil {
			return
		}
	}
	var (
		cToken = v.Stream.CurrentToken
		nextV  float64
	)
	switch cToken.Kind {
	case kNumber:
		numValue = cToken.NumberValue
		if err = v.Stream.GetNext(); err != nil {
			return
		}
	case kName:
		varValue, ok := v.Values[cToken.StringValue]
		if !ok {
			v.Values[cToken.StringValue] = 0 // 初始化局部变量
			varValue = 0
		}
		numValue = varValue
		if err = v.Stream.GetNext(); err != nil {
			return
		}
		if nextToken := v.Stream.CurrentToken; nextToken.Kind == kAssign {
			if v.Values[cToken.StringValue], err = v.LogicExpr(true); err != nil {
				return
			}
		}
	case kPlus:
		if nextV, err = v.Prim(true); err != nil {
			return
		}
		numValue = +nextV
	case kMinus:
		if nextV, err = v.Prim(true); err != nil {
			return
		}
		numValue = -nextV
	case kNot:
		if nextV, err = v.Prim(true); err != nil {
			return
		}
		numValue = BoolToFloat(nextV == floatFalse) // reverse the logic result
	case kLp:
		if nextV, err = v.LogicExpr(true); err != nil {
			return
		}
		if nextToken := v.Stream.CurrentToken; nextToken.Kind != kRp {
			err = ecode.ExprMissRightParenthesis
			return
		}
		if err = v.Stream.GetNext(); err != nil {
			return
		}
		numValue = nextV
	default:
		err = ecode.ExprPrimaryExpected
	}
	return

}

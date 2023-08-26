package compiler

import (
	"fmt"
	"strconv"

	"go-common/library/log"

	"go-gateway/app/app-svr/steins-gate/ecode"

	"github.com/pkg/errors"
)

// TokenKind defines the token's kind
type TokenKind string

const (
	kAssign       = TokenKind("assign")
	kName         = TokenKind("name")
	kNumber       = TokenKind("number")
	kEnd          = TokenKind("end")
	kEqual        = TokenKind("equal")
	kNotEqual     = TokenKind("not_equal")
	kGreaterEqual = TokenKind("greater_equal")
	kLessEqual    = TokenKind("less_equal")

	kPrint   = TokenKind(';')
	kPlus    = TokenKind('+')
	kMinus   = TokenKind('-')
	kMul     = TokenKind('*')
	kDiv     = TokenKind('/')
	kAnd     = TokenKind('&')
	kGreater = TokenKind('>')
	kLess    = TokenKind('<')
	kOr      = TokenKind('|')
	kNot     = TokenKind('!')
	kLp      = TokenKind('(')
	kRp      = TokenKind(')')

	replacement = '_'
)

var errReachEnd = fmt.Errorf("ending")

type Token struct {
	Kind        TokenKind
	StringValue string
	NumberValue float64
}

// 词法分析
type TokenStream struct {
	Input        string
	CurrentToken *Token
	position     int
}

// Get cursor goes ahead and return the char
func (v *TokenStream) Get() (result byte, err error) {
	if v.position >= len(v.Input) {
		err = errReachEnd
		return
	}
	result = v.Input[v.position]
	v.position++
	return
}

// PutBack cursor returns one step
func (v *TokenStream) PutBack() {
	if v.position == 0 {
		return
	}
	v.position--
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func isAlpha(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z')
}

// TreatVarName 针对所有非法字符或者下划线全部替换为 _unicode_的格式
func TreatVarName(name string) (validName string) {
	var newName []rune
	for _, v := range name {
		if !isValidVar(byte(v), false) || v == replacement {
			placeholder := fmt.Sprintf("%d", v)
			newName = append(newName, replacement)
			newName = append(newName, []rune(placeholder)...)
			newName = append(newName, replacement)
			continue
		}
		newName = append(newName, v)
	}
	return string(newName)
}

// ReadNumber reads the numbers to parse to float64
func (v *TokenStream) ReadNumber() (res float64, err error) {
	var numStr []byte
	for {
		ch, reachEnd := v.Get()
		if reachEnd != nil {
			break
		}
		if ch != '.' && !isDigit(ch) {
			v.PutBack()
			break
		}
		numStr = append(numStr, ch)
	}
	if len(numStr) == 0 {
		return
	}
	if res, err = strconv.ParseFloat(string(numStr), 64); err != nil {
		err = errors.Wrapf(ecode.ExprUnexpectedChar, "numStr %s", numStr)
	}
	return
}

// ReadVarName reads the alpha and numbers to get the var's name
func (v *TokenStream) ReadVarName(ch byte) (name string, err error) {
	if !isValidVar(ch, true) {
		err = errors.Wrapf(ecode.ExprBadToken, "incomplete var name %s", []byte{ch})
		return
	}
	varName := []byte{ch}
	for {
		newCh, reachEnd := v.Get()
		if reachEnd != nil {
			break
		}
		if !isValidVar(newCh, false) {
			v.PutBack()
			break
		}
		varName = append(varName, newCh)
	}
	name = string(varName)
	return
}

// isSpace def
func isSpace(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r'
}

// isValidVar allows $a, _a, a, a123, starting with number is not allowed
func isValidVar(ch byte, first bool) bool {
	if ch == '$' || ch == '_' {
		return true
	}
	if isAlpha(ch) {
		return true
	}
	if !first && isDigit(ch) {
		return true
	}
	return false
}

// GetNext picks the next Token
func (v *TokenStream) GetNext() (err error) {
	var (
		ch, nextCh byte
		numValue   float64
		varName    string
	)
	for {
		if ch, err = v.Get(); err != nil {
			err = nil
			v.CurrentToken = &Token{Kind: kEnd}
			return
		}
		if !isSpace(ch) { // skip space
			break
		}
	}
	switch ch {
	case '\n', ';':
		v.CurrentToken = &Token{Kind: kPrint}
	case '*':
		v.CurrentToken = &Token{Kind: kMul}
	case '/':
		v.CurrentToken = &Token{Kind: kDiv}
	case '+':
		v.CurrentToken = &Token{Kind: kPlus}
	case '-':
		v.CurrentToken = &Token{Kind: kMinus}
	case '(':
		v.CurrentToken = &Token{Kind: kLp}
	case ')':
		v.CurrentToken = &Token{Kind: kRp}
	case '>':
		if nextCh, err = v.Get(); err == nil {
			if nextCh == '=' {
				v.CurrentToken = &Token{Kind: kGreaterEqual}
				return
			}
			v.PutBack() // if the next char is not =, put it back
			v.CurrentToken = &Token{Kind: kGreater}
		}
	case '<':
		if nextCh, err = v.Get(); err == nil {
			if nextCh == '=' {
				v.CurrentToken = &Token{Kind: kLessEqual}
				return
			}
			v.PutBack()
			v.CurrentToken = &Token{Kind: kLess}
		}
	case '=':
		if nextCh, err = v.Get(); err == nil {
			if nextCh == '=' {
				v.CurrentToken = &Token{Kind: kEqual}
				return
			}
			v.PutBack()
			v.CurrentToken = &Token{Kind: kAssign}
		}
	case '&':
		if nextCh, err = v.Get(); err == nil {
			if nextCh == '&' {
				v.CurrentToken = &Token{Kind: kAnd}
				return
			}
			v.PutBack()
			err = errors.Wrapf(ecode.ExprBadToken, "incomplete operator &")
			return
		}
	case '|':
		if nextCh, err = v.Get(); err == nil {
			if nextCh == '|' {
				v.CurrentToken = &Token{Kind: kOr}
				return
			}
			v.PutBack()
			err = errors.Wrapf(ecode.ExprBadToken, "incomplete operator |")
			return
		}
	case '!':
		if nextCh, err = v.Get(); err == nil {
			if nextCh == '=' {
				v.CurrentToken = &Token{Kind: kNotEqual}
				return
			}
			v.PutBack()
			v.CurrentToken = &Token{Kind: kNot}
			return
		}
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		v.PutBack()
		if numValue, err = v.ReadNumber(); err != nil {
			log.Error("%+v", err)
			return
		}
		v.CurrentToken = &Token{
			Kind:        kNumber,
			NumberValue: numValue,
		}
	default:
		if varName, err = v.ReadVarName(ch); err != nil {
			log.Error("%+v", err)
			return
		}
		v.CurrentToken = &Token{
			Kind:        kName,
			StringValue: varName,
		}
	}
	return

}

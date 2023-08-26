package compiler

import (
	"encoding/json"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCompilerisDigit(t *testing.T) {
	var (
		ch = byte('3')
	)
	convey.Convey("isDigit", t, func(ctx convey.C) {
		p1 := isDigit(ch)
		convey.Println(p1)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestCompilerTreat(t *testing.T) {
	var (
		ch = "1234@*&3333_"
	)
	convey.Convey("isDigit", t, func(ctx convey.C) {
		p1 := TreatVarName(ch)
		convey.Println(p1)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestCompilerisAlpha(t *testing.T) {
	var (
		ch = byte('Z')
	)
	convey.Convey("isAlpha", t, func(ctx convey.C) {
		p1 := isAlpha(ch)
		convey.Println(p1)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestCompilerGetNext(t *testing.T) {
	convey.Convey("GetNext", t, func(ctx convey.C) {
		v := new(TokenStream)
		v.Input = "1*2*3+4.6+$a3231312/2"
		for {
			if err := v.GetNext(); err != nil {
				break
			}
			str, _ := json.Marshal(v.CurrentToken)
			convey.Println("position: ", v.position, " str:", string(str))
		}
	})
}

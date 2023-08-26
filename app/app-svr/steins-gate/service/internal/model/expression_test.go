package model

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestModelExpressionEval(t *testing.T) {
	convey.Convey("ExpressionEval", t, func(ctx convey.C) {
		rec := &HiddenVarsRecord{
			Vars: make(map[string]*HiddenVar),
		}
		var1 := &HiddenVar{Value: 10.55}
		var1.ID = "v-12#@!#345"
		var2 := &HiddenVar{Value: 30.88}
		var2.ID = "v-a78#@!#Q8"
		var3 := &HiddenVar{Value: 99}
		var3.ID = "v-012#@!#3a"
		var4 := &HiddenVar{}
		var4.ID = "v-tes#@!#t12"
		rec.Vars[var1.ID] = var1
		rec.Vars[var2.ID] = var2
		rec.Vars[var3.ID] = var3
		rec.Vars[var4.ID] = var4
		err := rec.ExpressionEval(fmt.Sprintf("%s=(%s+10)/2-5+%s;%s=%s*2+%s;%s=1>=2||3<4&&!(4.44==4);%s=5.23*2",
			hvarIDToExpr(var1.ID),
			hvarIDToExpr(var2.ID),
			hvarIDToExpr(var1.ID),
			hvarIDToExpr(var2.ID),
			hvarIDToExpr(var1.ID),
			hvarIDToExpr(var2.ID),
			hvarIDToExpr(var3.ID),
			hvarIDToExpr(var4.ID),
		))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			str, _ := json.Marshal(rec.Vars)
			convey.Println(string(str))
			ctx.So(rec.Vars[var1.ID].Value, convey.ShouldEqual, 25.99)
			ctx.So(rec.Vars[var2.ID].Value, convey.ShouldEqual, 82.86)
			ctx.So(rec.Vars[var3.ID].Value, convey.ShouldEqual, 1)
			ctx.So(rec.Vars[var4.ID].Value, convey.ShouldEqual, 10.46)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestModelExpressionEval2(t *testing.T) {
	convey.Convey("ExpressionEval", t, func(ctx convey.C) {
		rec := &HiddenVarsRecord{
			Vars: make(map[string]*HiddenVar),
		}
		var1 := &HiddenVar{Value: 10.55}
		var1.ID = "v-score"
		rec.Vars[var1.ID] = var1
		err := rec.ExpressionEval("$a1=1;$a2=2;$a3=3*2;$score=$score+$a1+$a2+$a3")
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			str, _ := json.Marshal(rec.Vars)
			convey.Println(string(str))
			ctx.So(rec.Vars[var1.ID].Value, convey.ShouldEqual, 19.55)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

func TestModelhvarIDToExpr(t *testing.T) {
	var (
		varID = "v-12345"
	)
	convey.Convey("hvarIDToExpr", t, func(ctx convey.C) {
		p1 := hvarIDToExpr(varID)
		convey.Println(p1)
		ctx.Convey("Then p1 should not be nil.", func(ctx convey.C) {
			ctx.So(p1, convey.ShouldNotBeNil)
		})
	})
}

func TestModelAttrSyntaxToExpr(t *testing.T) {
	var (
		attr1 = &EdgeAttribute{
			VarID:  "v-1234",
			Action: "sub",
			Value:  2,
		}
	)
	convey.Convey("exprToHvarID", t, func(ctx convey.C) {
		attr := new(EdgeAttribute)
		attr.FromSyntax(attr1)
		str, _ := json.Marshal(attr)
		convey.Println(string(str))
	})
}

func TestModelConditionSyntaxToExpr(t *testing.T) {
	var (
		attr1 = &EdgeCondition{
			VarID:     "v-1234",
			Condition: "eq",
			Value:     2,
		}
	)
	convey.Convey("exprToHvarID", t, func(ctx convey.C) {
		cond := new(EdgeCondition)
		cond.FromSyntax(attr1)
		convey.Println(cond)
	})
}

func TestModelExpressionCondition(t *testing.T) {
	convey.Convey("ExpressionCondition", t, func(ctx convey.C) {
		rec := &HiddenVarsRecord{
			Vars: make(map[string]*HiddenVar),
		}
		var1 := &HiddenVar{Value: 10.55}
		var1.ID = "v-12#@!#345"
		var2 := &HiddenVar{Value: 30.88}
		var2.ID = "v-a78#@!#Q8"
		var3 := &HiddenVar{Value: 99}
		var3.ID = "v-012#@!#3a"
		var4 := &HiddenVar{}
		var4.ID = "v-tes#@!#t12"
		rec.Vars[var1.ID] = var1
		rec.Vars[var2.ID] = var2
		rec.Vars[var3.ID] = var3
		rec.Vars[var4.ID] = var4
		res, err := rec.ExpressionCondition(fmt.Sprintf("%s<%s",
			hvarIDToExpr(var1.ID),
			hvarIDToExpr(var2.ID),
		))
		ctx.Convey("Then err should be nil.", func(ctx convey.C) {
			ctx.So(res, convey.ShouldNotBeNil)
			ctx.So(err, convey.ShouldBeNil)
		})
	})
}

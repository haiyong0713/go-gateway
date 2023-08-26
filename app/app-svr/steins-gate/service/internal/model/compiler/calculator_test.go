package compiler

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestCompilerInitAndEval(t *testing.T) {
	convey.Convey("Normal Process", t, func(ctx convey.C) {
		v := new(Calculator)
		ctx.Convey("add/minus", func(ctx convey.C) {
			result, err := v.InitAndEval("1+2-4", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, -1)
		})
		ctx.Convey("multi", func(ctx convey.C) {
			result, err := v.InitAndEval("1.1*2", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 2.2)
		})
		ctx.Convey("add/minus/multi", func(ctx convey.C) {
			result, err := v.InitAndEval("1.1*2+3-1", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 4.2)
		})
		ctx.Convey("add/minus/multi/div", func(ctx convey.C) {
			result, err := v.InitAndEval("1.1*2+3-4/2", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 3.2)
		})
		ctx.Convey("parenthesis", func(ctx convey.C) {
			result, err := v.InitAndEval("1.1*(2+3-1)/2", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 2.2)
		})
		ctx.Convey("comparison <", func(ctx convey.C) {
			result, err := v.InitAndEval("1<2", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 1)
		})
		ctx.Convey("comparison >=", func(ctx convey.C) {
			result, err := v.InitAndEval("1>=2", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 0)
		})
		ctx.Convey("logic and or", func(ctx convey.C) {
			result, err := v.InitAndEval("1>=2 || 3<4 && 4.44==4.44", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 1)
		})
		ctx.Convey("logic and or not ", func(ctx convey.C) {
			result, err := v.InitAndEval("1>=2 || 3<4 && !(4.44==4)", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 1)
		})
		ctx.Convey("not equal ", func(ctx convey.C) {
			result, err := v.InitAndEval("3!=3", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 0)
		})
		ctx.Convey("number1 ", func(ctx convey.C) {
			result, err := v.InitAndEval(".23", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 0.23)
		})
		ctx.Convey("number2 ", func(ctx convey.C) {
			result, err := v.InitAndEval("12.+.23", nil)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(result, convey.ShouldEqual, 12.23)
		})
	})
	convey.Convey("Abnormal Process", t, func(ctx convey.C) {
		v := new(Calculator)
		ctx.Convey("illegal number", func(ctx convey.C) {
			_, err := v.InitAndEval("1+3.3.3", nil)
			convey.Println(err) // 99071
			ctx.So(err, convey.ShouldNotBeNil)
		})
		ctx.Convey("illegal operator1", func(ctx convey.C) {
			_, err := v.InitAndEval("1&2", nil)
			convey.Println(err) // 99070
			ctx.So(err, convey.ShouldNotBeNil)
		})
		ctx.Convey("illegal operator2", func(ctx convey.C) {
			_, err := v.InitAndEval("1|2", nil)
			convey.Println(err) // 99070
			ctx.So(err, convey.ShouldNotBeNil)
		})
		ctx.Convey("illegal operator3", func(ctx convey.C) {
			_, err := v.InitAndEval("1=2", nil)
			convey.Println(err) // 99071
			ctx.So(err, convey.ShouldNotBeNil)
		})
		ctx.Convey("divide by zero", func(ctx convey.C) {
			_, err := v.InitAndEval("1/(2-2)", nil)
			convey.Println(err) // 99076
			ctx.So(err, convey.ShouldNotBeNil)
		})
		ctx.Convey("parenthesis missing right part", func(ctx convey.C) {
			_, err := v.InitAndEval("1/(2-2", nil)
			convey.Println(err) // 99074
			ctx.So(err, convey.ShouldNotBeNil)
		})
		ctx.Convey("illegal char", func(ctx convey.C) {
			_, err := v.InitAndEval("(2+3.33)*5;#", nil)
			convey.Println(err) // 99074
			ctx.So(err, convey.ShouldNotBeNil)
		})
	})
	convey.Convey("Variable Related", t, func(ctx convey.C) {
		v := new(Calculator)
		vars := make(map[string]float64)
		vars["a1234"] = 50
		vars["d9231"] = 200
		vars["Q8393"] = 0
		vars["$1234"] = 0
		vars["$0123a"] = 0
		ctx.Convey("vars add/div/minus", func(ctx convey.C) {
			result, err := v.InitAndEval("a1234=(d9231+10)/2-5+a1234;d9231=a1234*2+d9231;Q8393=1>=2||3<4&&!(4.44==4);$1234=5.23*2;$0123a=1>=2||3<4&&!(4.44==4)", vars)
			convey.Println(result)
			ctx.So(err, convey.ShouldBeNil)
			ctx.So(vars["a1234"], convey.ShouldEqual, 150)
			ctx.So(vars["d9231"], convey.ShouldEqual, 500)
			ctx.So(vars["Q8393"], convey.ShouldEqual, 1)
			ctx.So(vars["$1234"], convey.ShouldEqual, 10.46)
			ctx.So(vars["$0123a"], convey.ShouldEqual, 1)
			convey.Println(v.Values)
		})
		ctx.Convey("vars not declared", func(ctx convey.C) {
			_, err := v.InitAndEval("a9=123", vars)
			ctx.So(err, convey.ShouldNotBeNil) // 99073
		})
		ctx.Convey("vars format", func(ctx convey.C) {
			_, err := v.InitAndEval("#a9=123", vars)
			ctx.So(err, convey.ShouldNotBeNil) // 99070
		})
	})
}

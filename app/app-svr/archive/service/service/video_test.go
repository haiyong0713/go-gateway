package service

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func Test_Description(t *testing.T) {
	Convey("Description", t, func() {
		desc, descV2, err := s.Description(context.TODO(), 480026702)
		fmt.Println(desc, descV2)
		So(err, ShouldBeNil)
	})
}

func Test_Descriptions(t *testing.T) {
	Convey("Descriptions", t, func() {
		desc, err := s.Descriptions(context.TODO(), []int64{480026702})
		fmt.Println(desc)
		So(err, ShouldBeNil)
	})
}

func Test_Page3(t *testing.T) {
	Convey("Page3", t, func() {
		ps, err := s.Page3(context.TODO(), 10098500)
		So(err, ShouldBeNil)
		for _, p := range ps {
			Printf("%+v\n\n", p)
			bs, _ := json.Marshal(p)
			Printf("%s\n\n", bs)
		}
	})
}

func Test_View3(t *testing.T) {
	Convey("View3", t, func() {
		v, err := s.View3(context.TODO(), 10097755, 0)
		So(err, ShouldBeNil)
		fmt.Println(v.Arc.IsSteinsGate())
		bs, _ := json.Marshal(v)
		fmt.Println(string(bs))
	})
}

func Test_SteinsGateView(t *testing.T) {
	Convey("Test_SteinsGateView", t, func() {
		v, err := s.SteinsGateView(context.TODO(), 10098500)
		So(err, ShouldBeNil)
		if v.Pages != nil {
			for _, p := range v.Pages {
				bs, _ := json.Marshal(p)
				Printf("%s\n\n", bs)
			}
		}

	})
}

func Test_Views3(t *testing.T) {
	Convey("Views3", t, func() {
		as, err := s.Views3(context.TODO(), []int64{10098500, 10097755, 10097694}, 0, "", "")
		fmt.Println(len(as))
		So(err, ShouldBeNil)
		for _, a := range as {
			fmt.Println("For Aid ", a.Aid)
			bs, _ := json.Marshal(a)
			fmt.Println(string(bs))
		}
	})
}

func Test_Video3(t *testing.T) {
	Convey("Video3", t, func() {
		v, err := s.Video3(context.TODO(), 10098500, 10109206)
		Printf("%+v\n\n\n", v)
		So(err, ShouldBeNil)
		bs, _ := json.Marshal(v)
		Printf("%s\n\n\n", bs)
	})
}

package model

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestModelHvarApplyAttrs(t *testing.T) {
	convey.Convey("HvarIDTransform Syntax Tree", t, func(ctx convey.C) {
		var (
			id1 = "v-i1234545#"
			id2 = "v-32131231#"
		)
		hvarRec := &HiddenVarsRecord{}
		hvarRec.Vars = make(map[string]*HiddenVar)
		hvarRec.Vars[id1] = &HiddenVar{
			Value: 333,
		}
		hvarRec.Vars[id1].ID = id1
		hvarRec.Vars[id2] = &HiddenVar{
			Value: 444,
		}
		hvarRec.Vars[id2].ID = id2
		attributes := []string{
			fmt.Sprintf(`[{"var_id":"%s","action":"add","value":5},{"var_id":"%s","action":"assign","value":10}]`, id1, id2),
		}
		convey.Println(attributes)
		hvarRec.ApplyAttrs(attributes)
		str, _ := json.Marshal(hvarRec)
		convey.Println(string(str))
	})
	convey.Convey("HvarIDTransform ExpressionEval", t, func(ctx convey.C) {
		var (
			id3 = "vi1234545"
			id4 = "v32131231"
		)
		hvarRec := &HiddenVarsRecord{}
		hvarRec.Vars = make(map[string]*HiddenVar)
		hvarRec.Vars[id3] = &HiddenVar{
			Value: 333,
		}
		hvarRec.Vars[id3].ID = id3
		hvarRec.Vars[id4] = &HiddenVar{
			Value: 444,
		}
		hvarRec.Vars[id4].ID = id4
		attributes := []string{
			fmt.Sprintf(`[{"action_type":1, "action":"$a1=1"}]`),
			fmt.Sprintf(`[{"action_type":1, "action":"$a2=2;$%s=$a1+$a2"}]`, id3),
		}
		convey.Println(attributes)
		hvarRec.ApplyAttrs(attributes)
		str, _ := json.Marshal(hvarRec)
		convey.Println(string(str))
	})

}

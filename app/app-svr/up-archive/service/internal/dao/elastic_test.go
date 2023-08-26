package dao

import (
	"testing"

	"go-common/library/database/elastic"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDao_EsQuery(t *testing.T) {
	var (
		mid     = 1212
		without []int64
	)
	Convey("RawArcPassed", t, func() {
		d := elastic.NewElastic(nil)
		r := d.NewRequest(businessKey("")).Index(_index)
		eq := []map[string]interface{}{
			{"mid": mid},
			{"staff_mid": mid},
		}
		//!(medl_id==59 && state=1)
		nots := []map[string]interface{}{
			{
				"archive_flow.meal_id": 59,
				"archive_flow.state":   1,
			},
		}
		for _, val := range without {
			switch val {
			case 1:
				eq = []map[string]interface{}{
					{"mid": mid},
				}
			case 2:
				r.WhereIn("up_from", 23)
				r.WhereNot(elastic.NotTypeIn, "up_from")
			case 3:
				//!(medl_id==59 && state=1) && !(medl_id==60 && state=1)
				nots = []map[string]interface{}{
					{
						"archive_flow.meal_id": 59,
						"archive_flow.state":   1,
					},
					{
						"archive_flow.meal_id": 60,
						"archive_flow.state":   1,
					},
				}
			default:
			}
		}
		comboEq := &elastic.Combo{}
		comboEq.ComboEQ(eq).MinEQ(1).MinAll(1)
		comboNots := &elastic.Combo{}
		comboNots.ComboNestedNots("archive_flow", nots).ComboNotExist([]string{"archive_flow"}).MinNested(1).MinAll(1)
		r.WhereCombo(comboEq, comboNots)
		Printf("%v", r.Params())
	})
}

func TestDao_EsQuery2(t *testing.T) {
	Convey("RawArcPassed", t, func() {
		d := elastic.NewElastic(nil)
		r := d.NewRequest("")
		nots := []map[string]interface{}{
			{
				"archive_flow.meal_id": 59,
				"archive_flow.state":   1,
			},
			{
				"archive_flow.meal_id": 60,
				"archive_flow.state":   1,
			},
		}
		comboNots := &elastic.Combo{}
		comboNots.ComboNestedNots("archive_flow", nots).ComboNotExist([]string{"archive_flow"}).MinNested(1).MinAll(1)
		r.WhereCombo(comboNots)
		Printf("%v", r.Params())
	})
}

func TestDao_EsQuery3(t *testing.T) {
	Convey("RawArcPassed", t, func() {
		e := elastic.NewElastic(nil)
		cmba := &elastic.Combo{}
		comboNots := []map[string]interface{}{}
		comboNots = append(comboNots, map[string]interface{}{
			"archive_flow.meal_id": 1,
			"archive_flow.state":   1,
		})
		cmba.ComboNestedNots("archive_flow", comboNots).ComboNotExist([]string{"subtitle"}).
			MinNested(1).MinAll(1)
		cmbB := &elastic.Combo{}
		cmbB.ComboEQ([]map[string]interface{}{
			{"mid": 88},
			{"staff_mid": 99},
		}).MinEQ(1).MinAll(1)
		r := e.NewRequest("").WhereCombo(cmba, cmbB)
		Printf("%v", r.Params())
	})
}

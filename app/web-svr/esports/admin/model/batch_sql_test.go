package model

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestArcBatchAddSQL(t *testing.T) {
	convey.Convey("TestArcBatchAddSQL", t, func(ctx convey.C) {
		sql, sqlParam := ArcBatchAddSQL([]int64{1, 2, 3})
		//fmt.Println(sqlParam)
		for _, v := range sqlParam {
			i := v.(int64)
			s := strconv.FormatInt(i, 10)
			sql = strings.Replace(sql, "?", s, 1)
		}
		fmt.Println(sql)
	})
}

func getValue(i interface{}) (s string) {
	switch v := i.(type) {
	case int:
		s = strconv.FormatInt(int64(v), 10)
	case int64:
		s = strconv.FormatInt(v, 10)
	case string:
		s = v
	case bool:
		s = strconv.FormatBool(v)
	default:
		s = fmt.Sprint(i)
	}
	return
}

func TestGidBatchAddSQL(t *testing.T) {
	convey.Convey("TestGidBatchAddSQL", t, func(ctx convey.C) {
		gid := &GIDMap{ID: 1, Type: 1, Oid: 1, Gid: 1, IsDeleted: 0}
		gid2 := &GIDMap{ID: 1, Type: 1, Oid: 1, Gid: 1, IsDeleted: 0}
		gid3 := &GIDMap{ID: 1, Type: 1, Oid: 1, Gid: 1, IsDeleted: 0}
		gids := make([]*GIDMap, 0)
		gids = append(gids, gid, gid2, gid3)
		sql, sqlParam := GidBatchAddSQL(gids)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchAddMachMapSQL(t *testing.T) {
	convey.Convey("TestBatchAddMachMapSQL", t, func(ctx convey.C) {
		p := &MatchMap{ID: 1, Mid: 1, Aid: 11, IsDeleted: 0}
		p2 := &MatchMap{ID: 2, Mid: 2, Aid: 22, IsDeleted: 0}
		p3 := &MatchMap{ID: 3, Mid: 3, Aid: 33, IsDeleted: 0}
		ps := make([]*MatchMap, 0)
		params := append(ps, p, p2, p3)
		sql, sqlParam := BatchAddMachMapSQL(params)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchAddTagMapSQL(t *testing.T) {
	convey.Convey("TestBatchAddTagMapSQL", t, func(ctx convey.C) {
		p := &TagMap{ID: 1, Tid: 1, Aid: 11, IsDeleted: 0}
		p2 := &TagMap{ID: 2, Tid: 2, Aid: 22, IsDeleted: 0}
		p3 := &TagMap{ID: 3, Tid: 3, Aid: 33, IsDeleted: 0}
		ps := make([]*TagMap, 0)
		params := append(ps, p, p2, p3)
		sql, sqlParam := BatchAddTagMapSQL(params)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchAddTeamMapSQL(t *testing.T) {
	convey.Convey("TestBatchAddTeamMapSQL", t, func(ctx convey.C) {
		p := &TeamMap{ID: 1, Tid: 1, Aid: 11, IsDeleted: 0}
		p2 := &TeamMap{ID: 2, Tid: 2, Aid: 22, IsDeleted: 0}
		p3 := &TeamMap{ID: 3, Tid: 3, Aid: 33, IsDeleted: 0}
		ps := make([]*TeamMap, 0)
		params := append(ps, p, p2, p3)
		sql, sqlParam := BatchAddTeamMapSQL(params)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchAddYearMapSQL(t *testing.T) {
	convey.Convey("TestBatchAddTeamMapSQL", t, func(ctx convey.C) {
		p := &YearMap{ID: 1, Year: 1, Aid: 11, IsDeleted: 0}
		p2 := &YearMap{ID: 2, Year: 2, Aid: 22, IsDeleted: 0}
		p3 := &YearMap{ID: 3, Year: 3, Aid: 33, IsDeleted: 0}
		ps := make([]*YearMap, 0)
		params := append(ps, p, p2, p3)
		sql, sqlParam := BatchAddYearMapSQL(params)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchAddCDataSQL(t *testing.T) {
	convey.Convey("TestBatchAddCDataSQL", t, func(ctx convey.C) {
		p := &ContestData{ID: 1, CID: 1, URL: "aaa", PointData: 555, IsDeleted: 0}
		p2 := &ContestData{ID: 2, CID: 2, URL: "bbb", PointData: 666, IsDeleted: 0}
		p3 := &ContestData{ID: 3, CID: 3, URL: "ccc", PointData: 777, IsDeleted: 0}
		ps := make([]*ContestData, 0)
		params := append(ps, p, p2, p3)
		sql, sqlParam := BatchAddCDataSQL(10, params)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchEditCDataSQL(t *testing.T) {
	convey.Convey("TestBatchAddCDataSQL", t, func(ctx convey.C) {
		p := &ContestData{ID: 1, CID: 1, URL: "aaa", PointData: 555, IsDeleted: 0}
		p2 := &ContestData{ID: 2, CID: 1, URL: "bbb", PointData: 666, IsDeleted: 0}
		p3 := &ContestData{ID: 3, CID: 1, URL: "ccc", PointData: 777, IsDeleted: 0}
		ps := make([]*ContestData, 0)
		params := append(ps, p, p2, p3)
		sql, sqlParam := BatchEditCDataSQL(params)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchAddModuleSQL(t *testing.T) {
	convey.Convey("TestBatchAddCDataSQL", t, func(ctx convey.C) {
		p := &Module{ID: 1, MaID: 1, Name: "aaa", Oids: "555", Status: 0}
		p2 := &Module{ID: 2, MaID: 1, Name: "bbb", Oids: "666", Status: 0}
		p3 := &Module{ID: 3, MaID: 1, Name: "ccc", Oids: "777", Status: 0}
		ps := make([]*Module, 0)
		params := append(ps, p, p2, p3)
		sql, sqlParam := BatchAddModuleSQL(100, params)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchAddActLiveSQL(t *testing.T) {
	convey.Convey("BatchAddActLiveSQL", t, func(ctx convey.C) {
		p := &Activelive{ID: 1, MaId: 1, Title: "aaa", LiveId: 555, IsDeleted: 0}
		p2 := &Activelive{ID: 2, MaId: 1, Title: "bbb", LiveId: 666, IsDeleted: 0}
		p3 := &Activelive{ID: 3, MaId: 1, Title: "ccc", LiveId: 777, IsDeleted: 0}
		ps := make([]*Activelive, 0)
		params := append(ps, p, p2, p3)
		sql, sqlParam := BatchAddActLiveSQL(100, params)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchEditActLiveSQL(t *testing.T) {
	convey.Convey("BatchAddActLiveSQL", t, func(ctx convey.C) {
		p := &Activelive{ID: 1, MaId: 1, Title: "aaa", LiveId: 555, IsDeleted: 0}
		p2 := &Activelive{ID: 2, MaId: 1, Title: "bbb", LiveId: 666, IsDeleted: 0}
		p3 := &Activelive{ID: 3, MaId: 1, Title: "ccc", LiveId: 777, IsDeleted: 0}
		ps := make([]*Activelive, 0)
		params := append(ps, p, p2, p3)
		sql, sqlParam := BatchEditActLiveSQL(params)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

func TestBatchEditTreeSQL(t *testing.T) {
	convey.Convey("BatchAddActLiveSQL", t, func(ctx convey.C) {
		p := make(map[int64]int64)
		p[1] = 1
		p[2] = 2
		p[3] = 3
		sql, sqlParam := BatchEditTreeSQL(p)
		for _, v := range sqlParam {
			sql = strings.Replace(sql, "?", getValue(v), 1)
		}
		fmt.Println(sql)
	})
}

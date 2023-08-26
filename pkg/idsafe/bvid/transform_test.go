package bvid

import (
	"io/ioutil"
	"strconv"
	"strings"

	"testing"
)

func makeCaseMap() (cases map[int64]string) {
	cases = make(map[int64]string)
	cases[327] = "xx411c7QG"
	cases[22712] = "Hx411c78B"
	cases[2147483647] = "sZ411p7j9"
	cases[0] = ""
	cases[36028799166447615] = ""
	return
}

func TestAvToBv(t *testing.T) {
	var (
		bs    []byte
		err   error
		avids []string
	)
	if bs, err = ioutil.ReadFile("../av.txt"); err != nil {
		t.Errorf("read file(av.txt) error(%v)", err)
		return
	}
	avids = strings.Split(string(bs), "\n")
	for _, idStr := range avids {
		var (
			aid int64
			bv  string
		)
		if aid, err = strconv.ParseInt(idStr, 10, 64); err != nil {
			t.Logf("Av(%s) To Bv(%s) error(%v)", idStr, bv, err)
			continue
		}
		bv, err = AvToBv(aid)
		t.Logf("Av(%d) To Bv(%s) error(%v)", aid, bv, err)
	}
}

func TestBvToAv(t *testing.T) {
	var (
		bs    []byte
		err   error
		bvids []string
	)
	if bs, err = ioutil.ReadFile("../bv.txt"); err != nil {
		t.Logf("read file(av.txt) error(%v)", err)
		return
	}
	bvids = strings.Split(string(bs), "\n")
	for _, idStr := range bvids {
		var aid int64
		aid, err = BvToAv(idStr)
		t.Logf("Bv(%s) To Av(%d) error(%v)", idStr, aid, err)
	}
}

package model

import "testing"

func TestCalculateIdx(t *testing.T) {
	data := []struct {
		Idx, Len, Size int32
		Head, Tail     int32
	}{
		{
			0, 2000, 2000,
			0, 2000,
		},
		{
			500, 1000, 2000,
			0, 1000,
		},
		{
			700, 2005, 2000,
			5, 2005,
		},
		{
			1000, 4000, 2000,
			500, 2500,
		},
		{
			1000, 2001, 2000,
			1, 2001,
		},
		{
			5, 30, 30,
			0, 30,
		},
		{
			5, 30, 20,
			0, 20,
		},
		{
			6, 30, 20,
			1, 21,
		},
	}
	for _, d := range data {
		if head, tail := CalculateIdx(d.Idx, d.Len, d.Size); head != d.Head || tail != d.Tail {
			t.Errorf("expecting Head(%d) Tail(%d) for Idx(%d) Len(%d) Size(%d), but got Head(%d), Tail(%d)",
				d.Head, d.Tail, d.Idx, d.Len, d.Size, head, tail)
		}
	}
}

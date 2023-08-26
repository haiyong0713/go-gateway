package common

func Int32SliceToInt64Slice(in []int32) []int64 {
	out := make([]int64, 0, len(in))
	for _, item := range in {
		out = append(out, int64(item))
	}
	return out
}

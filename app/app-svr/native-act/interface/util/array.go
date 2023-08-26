package util

func UniqueArray(arr []int64) []int64 {
	if len(arr) == 0 {
		return []int64{}
	}
	m := make(map[int64]struct{}, len(arr))
	uniq := make([]int64, 0, len(arr))
	for _, v := range arr {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		uniq = append(uniq, v)
	}
	return uniq
}

func UniqueArrayWithInt32(arr []int32) []int32 {
	if len(arr) == 0 {
		return []int32{}
	}
	m := make(map[int32]struct{}, len(arr))
	uniq := make([]int32, 0, len(arr))
	for _, v := range arr {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		uniq = append(uniq, v)
	}
	return uniq
}

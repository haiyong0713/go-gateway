package util

// Int64DuplicateRemoval .
func Int64DuplicateRemoval(args []int64) (res []int64) {
	argsMap := make(map[int64]struct{})
	for _, v := range args {
		if _, ok := argsMap[v]; !ok {
			argsMap[v] = struct{}{}
			res = append(res, v)
		}
	}
	return
}

func Int64ZeroRemoval(args []int64) (res []int64) {
	for _, v := range args {
		if v == 0 {
			continue
		}
		res = append(res, v)
	}

	return
}

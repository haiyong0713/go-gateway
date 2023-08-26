package rank

// Interface rank interface
type Interface interface {
	// pLen is the length of rank
	Len() int
	// TopLen is the length of top rank
	TopLen() int
	// Less reports whether the element with
	// index i should sort before the element with index j.
	Less(i, j int) bool
	// Swap swaps the elements with indexes i and j.
	Swap(i, j int)
	// Cut cut the length of data
	Cut(len int)
	// Append remainData to origin data
	Append(remainData Interface)
}

func heapify(data Interface, i, n int) {
	c1 := 2*i + 1
	c2 := 2*i + 2
	swap := i
	if c1 < n && data.Less(swap, c1) {
		swap = c1
	}
	if c2 < n && data.Less(swap, c2) {
		swap = c2
	}
	if swap != i {
		data.Swap(i, swap)
		heapify(data, swap, n)
	}
}

func buildHeap(data Interface) {
	lastNode := data.TopLen() - 1
	parent := (lastNode - 1) / 2
	for i := parent; i >= 0; i-- {
		heapify(data, i, data.TopLen())
	}
}

func remain(data Interface) {
	if data.Len() > data.TopLen() {
		for i := data.TopLen(); i < data.Len(); i++ {
			if data.Less(i, 0) {
				data.Swap(i, 0)
				heapify(data, 0, data.TopLen())
			}
		}
		data.Cut(data.TopLen())
	}
}

// Add add remain
func Add(data Interface, remainData Interface) {
	data.Append(remainData)
	remain(data)
	sort(data)
}
func sort(data Interface) {
	for i := data.TopLen() - 1; i >= 0; i-- {
		data.Swap(i, 0)
		heapify(data, 0, i)
	}
}

// Sort 排序
func Sort(data Interface) {
	buildHeap(data)
	remain(data)
	sort(data)
}

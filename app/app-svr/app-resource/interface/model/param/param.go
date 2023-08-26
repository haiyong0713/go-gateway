package param

// Param struct
type Param struct {
	Name      string `json:"-"`
	Value     string `json:"-"`
	Plat      int8   `json:"-"`
	Build     int    `json:"-"`
	Condition string `json:"-"`
}

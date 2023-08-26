package region

type Region struct {
	Rid      int32  `json:"-"`
	Reid     int32  `json:"-"`
	Name     string `json:"-"`
	Language string `json:"-"`
}

package v2

const (
	DetailViewBitAttrYes = 1

	DetailViewBitDrawFirst = 0
)

func (c *Config) IsDetailDrawFirst() bool {
	if c == nil {
		return false
	}
	return c.DetailViewBitAttr(DetailViewBitDrawFirst) == DetailViewBitAttrYes
}

func (c *Config) DetailViewBitAttr(bit uint) uint64 {
	return c.DetailViewBits >> bit & 1
}

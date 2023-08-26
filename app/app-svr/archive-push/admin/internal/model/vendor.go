package model

var DefaultVendors = []ArchivePushVendor{
	{
		ID:           1,
		Name:         "王者营地",
		UserBindable: true,
	},
	{
		ID:           2,
		Name:         "王者营地 - TGL",
		UserBindable: true,
	},
	{
		ID:           3,
		Name:         "暴雪",
		UserBindable: false,
	},
}

type ArchivePushVendor struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	UserBindable bool   `json:"userBindable"`
}

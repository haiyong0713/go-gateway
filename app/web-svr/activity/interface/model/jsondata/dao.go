package jsondata

// SummerGiftList ...
type SummerGiftList struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	ImgUrl string `json:"img_url"`
	Date   string `json:"date"`
	Order  int    `json:"order"`
}

// SummerGiftRes 夏令营夏日
type SummerGiftRes []*SummerGiftList

type SummerGift struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	ImgUrl string `json:"img_url"`
	Order  int    `json:"order"`
}

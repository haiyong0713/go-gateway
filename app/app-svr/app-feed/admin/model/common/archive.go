package common

type ArchiveResult struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    *struct {
		Archive *Archive
	} `json:"data"`
}

type Archive struct {
	State int    `json:"state"`
	Title string `json:"title"`
}

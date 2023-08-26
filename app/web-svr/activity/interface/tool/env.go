package tool

import "os"

func IsBnj2021LiveApplication() (b bool) {
	if os.Getenv("BNJ2021_LIVE") != "" {
		b = true
	}

	return
}

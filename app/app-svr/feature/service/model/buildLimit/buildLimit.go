package buildLimit

import "encoding/json"

type BuildLimit struct {
	KeyName    string          `json:"-"`
	Conditions json.RawMessage `json:"-"`
}

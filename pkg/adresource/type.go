package adresource

import (
	"strings"

	"github.com/pkg/errors"
)

type tScene string
type tResourceID int64

var (
	EmptyScene      = tScene("")
	EmptyResourceID = tResourceID(0)

	allSceneStore      = map[tScene]struct{}{}
	allResoueceIDStore = map[tResourceID]struct{}{}
)

func NewScene(s tScene) tScene {
	if _, ok := allSceneStore[s]; ok {
		panic(errors.Errorf("duplicated scene: %q", s))
	}
	allSceneStore[s] = struct{}{}
	return s
}

type SceneBuilder func(...string) tScene

func CurryingSceneBuilder(initialLabel ...string) SceneBuilder {
	return func(runtimeLabel ...string) tScene {
		allLabel := append(initialLabel, runtimeLabel...)
		return tScene(strings.Join(allLabel, ":"))
	}
}

func NewResoueceID(rawID int64) tResourceID {
	r := tResourceID(rawID)
	if _, ok := allResoueceIDStore[r]; ok {
		panic(errors.Errorf("duplicated resource id: %d", r))
	}
	allResoueceIDStore[r] = struct{}{}
	return r
}

package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetByAppID(t *testing.T) {
	res, ok := GetByAppID("main.dynamic.feed")
	assert.Equal(t, ok, true)
	assert.Equal(t, &App{Name: "main.dynamic.feed",
		Packages: map[string]*Package{
			"dynamic.service.feed.v1": {Name: "dynamic.service.feed.v1",
				Services: map[string]*Service{
					"Feed": {Name: "Feed",
						Methods: []string{
							"/dynamic.service.feed.v1.Feed/UpdateNum",
						},
					},
				},
			},
		},
	}, res)
	res, ok = GetByAppID("test")
	assert.Equal(t, false, ok)
}

func TestPrefixMatch(t *testing.T) {
	res, ok := PrefixMatch("main")
	assert.Equal(t, true, ok)
	assert.Equal(t, map[string]*App{
		"main.dynamic.feed": {Name: "main.dynamic.feed",
			Packages: map[string]*Package{
				"dynamic.service.feed.v1": {Name: "dynamic.service.feed.v1",
					Services: map[string]*Service{
						"Feed": {Name: "Feed",
							Methods: []string{
								"/dynamic.service.feed.v1.Feed/UpdateNum",
							},
						},
					},
				},
			},
		},
	}, res)
	res, ok = PrefixMatch("dynamic")
	assert.Equal(t, false, ok)
}

func TestFuzzyMatch(t *testing.T) {
	res, ok, err := FuzzyMatch("dynamic")
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
	assert.Equal(t, map[string]*App{
		"main.dynamic.feed": {Name: "main.dynamic.feed",
			Packages: map[string]*Package{
				"dynamic.service.feed.v1": {Name: "dynamic.service.feed.v1",
					Services: map[string]*Service{
						"Feed": {Name: "Feed",
							Methods: []string{
								"/dynamic.service.feed.v1.Feed/UpdateNum",
							},
						},
					},
				},
			},
		},
	}, res)
	_, ok, err = FuzzyMatch("fha**)#U*$)@")
	assert.NotNil(t, err)
	assert.Equal(t, false, ok)
}

func TestGetByAppIDs(t *testing.T) {
	res, ok := GetByAppIDs([]string{"main.dynamic.feed"})
	assert.Equal(t, true, ok)
	assert.Equal(t, map[string]*App{
		"main.dynamic.feed": {Name: "main.dynamic.feed",
			Packages: map[string]*Package{
				"dynamic.service.feed.v1": {Name: "dynamic.service.feed.v1",
					Services: map[string]*Service{
						"Feed": {Name: "Feed",
							Methods: []string{
								"/dynamic.service.feed.v1.Feed/UpdateNum",
							},
						},
					},
				},
			},
		},
	}, res)
}

func TestGetAllAppID(t *testing.T) {
	res := GetAllAppID()
	assert.Equal(t, []string{"main.dynamic.feed"}, res)
}

package tusmultipleversion

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionIncrease(t *testing.T) {
	tests := []struct {
		cvm            *ConfigVersionManager
		limit          []*BuildLimit
		tusValues      []string
		lateestVersion string
	}{
		{
			cvm: &ConfigVersionManager{
				Field: "field1",
				VersionInfos: []*VersionInfo{
					{
						ConfigVersion: FirstVersion,
						TusValues:     []string{"1", "2"},
					},
					{
						ConfigVersion: "v2.0",
						TusValues:     []string{"1", "2", "3"},
					},
				},
			},
			limit: []*BuildLimit{
				{
					Plat:     1,
					Operator: GT,
					Build:    1000,
				},
			},
			tusValues:      []string{"1", "2", "3", "4"},
			lateestVersion: "v3.0",
		},
		{
			cvm: &ConfigVersionManager{
				Field: "field2",
				VersionInfos: []*VersionInfo{
					{
						ConfigVersion: FirstVersion,
						TusValues:     []string{"a", "b"},
					},
					{
						ConfigVersion: "v3.0",
						TusValues:     []string{"a", "b", "c"},
					},
				},
			},
			limit: []*BuildLimit{
				{
					Plat:     0,
					Operator: LT,
					Build:    109999,
				},
			},
			tusValues:      []string{"a"},
			lateestVersion: "v4.0",
		},
	}
	for _, v := range tests {
		latestVersion, err := v.cvm.VersionIncrease(v.limit, v.tusValues)
		assert.NoError(t, err)
		assert.Equal(t, latestVersion, v.lateestVersion)
	}
}

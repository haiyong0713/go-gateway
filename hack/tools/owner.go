package tools

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

type DirOptions struct {
	NoParentOwners      bool `json:"no_parent_owners,omitempty"`
	UnitTestAll         bool `json:"unit_test_all,omitempty"`
	UnitTestRestrictive bool `json:"unit_test_restrictive,omitempty"`
}

// Config holds roles+usernames and labels for a directory considered as a unit of independent code
type Config struct {
	Approvers         []string `json:"approvers,omitempty"`
	Reviewers         []string `json:"reviewers,omitempty"`
	RequiredReviewers []string `json:"required_reviewers,omitempty"`
	Labels            []string `json:"labels,omitempty"`
}

// SimpleConfig holds options and Config applied to everything under the containing directory
type Owner struct {
	Options DirOptions `json:"options,omitempty"`
	Config  `json:",inline"`
}

func ReadOwner(path string) (*Owner, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("ERROR 文件路径: %s, yaml 文件格式错误: %v", path, err)
	}
	owner := &Owner{}
	if err = yaml.Unmarshal(data, owner); err != nil {
		return nil, fmt.Errorf("ERROR 文件路径: %s, OWNERS 格式错误: %v ", path, err)
	}
	return owner, nil
}

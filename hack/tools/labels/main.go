package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"go-gateway/hack/tools"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Approvers         []string    `json:"approvers,omitempty"`
	Reviewers         []string    `json:"reviewers,omitempty"`
	RequiredReviewers []string    `json:"required_reviewers,omitempty"`
	Labels            []string    `json:"labels,omitempty"`
	Options           *DirOptions `json:"options,omitempty"`
}

type Owner struct {
	Config `json:",inline"`
}

type DirOptions struct {
	NoParentOwners bool `json:"no_parent_owners,omitempty"`
}

// LabelTarget specifies the intent of the label (PR or issue)
type LabelTarget string

const (
	bothTarget = "both"
)

type Label struct {
	// Name is the current name of the label
	Name string `json:"name"`
	// Color is rrggbb or color
	Color string `json:"color"`
	// Description is brief text explaining its meaning, who can apply it
	Description string `json:"description"` // What does this label mean, who can apply it
	// Target specifies whether it targets PRs, issues or both
	Target LabelTarget `json:"target"`
	// ProwPlugin specifies which prow plugin add/removes this label
	ProwPlugin string `json:"prowPlugin,omitempty"`
	// AddedBy specifies whether human/munger/bot adds the label
	AddedBy string `json:"addedBy"`
	// Previously lists deprecated names for this label
	Previously []Label `json:"previously,omitempty"`
	// DeleteAfter specifies the label is retired and a safe date for deletion
	DeleteAfter *time.Time `json:"deleteAfter,omitempty"`
}

// Configuration is a list of Required Labels to sync in all kubernetes repos
type Configuration struct {
	Default RepoConfig `json:"default"`
}

func (c *Configuration) Output(labels map[string]Label) []byte {
	for _, label := range labels {
		c.Default.Labels = append(c.Default.Labels, label)
	}
	sort.Slice(c.Default.Labels, func(i, j int) bool {
		return strings.Compare(c.Default.Labels[i].Name, c.Default.Labels[j].Name) == -1
	})
	out, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("marshal configuration error: %v", err)
	}
	return out
}

func (c *Configuration) AddLabels() {

}

func (c *Configuration) AddNewProjectLabels(name string) {
	c.Default.Labels = append(c.Default.Labels)

}

// RepoConfig contains only labels for the moment
type RepoConfig struct {
	Labels []Label `json:"labels"`
}

func OwnerLabelsWalk(labels map[string]Label) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if _, ok := tools.PathBaseBlacklist[filepath.Base(path)]; ok {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}
		if !strings.HasSuffix(path, tools.OWNERSFileName) {
			return nil
		}
		bs, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		var owner Config
		if err := yaml.Unmarshal(bs, &owner); err != nil {
			return err
		}
		for _, label := range owner.Labels {
			color := "F0AD4E"
			if tools.IsAreaLabel(label) {
				labelLen := len(strings.Split(label, string(filepath.Separator)))
				if labelLen == tools.SubdivisionDirectoryLevel {
					color = "0033CC"
				} else {
					color = "428BCA"
				}
			}
			labels[label] = Label{
				Name:        label,
				Color:       color,
				Description: "Categorizes an issue or PR as relevant to " + label,
				Target:      bothTarget,
				AddedBy:     "anyone",
				ProwPlugin:  "label",
			}
		}
		return nil
	}
}

func DepartmentLabelsWalk(labels map[string]Label) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() || path == tools.RelativeProjectRootPath {
			return nil
		}
		paths := strings.Split(path, string(filepath.Separator))
		label := filepath.Join("new-project", paths[1])
		labels[label] = Label{
			Name:        label,
			Color:       "D9534F",
			Description: "Categorizes an issue or PR as relevant to " + label,
			Target:      bothTarget,
			AddedBy:     "anyone",
			ProwPlugin:  "label",
		}
		return filepath.SkipDir
	}
}

func yamlFileUnmarshal(path string, dest interface{}) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, dest)
}

func main() {
	log.SetOutput(os.Stdout)
	labelConfig := Configuration{}
	if err := yamlFileUnmarshal(filepath.Join(tools.ProwJobTemplatePath, "labels", "template.yaml"), &labelConfig); err != nil {
		log.Fatalf("read label template failed: %v", err)
	}
	labels := map[string]Label{}
	if err := filepath.Walk(".", OwnerLabelsWalk(labels)); err != nil {
		log.Fatalf("walk owner label error: %v", err)
	}
	if err := filepath.Walk(tools.RelativeProjectRootPath, DepartmentLabelsWalk(labels)); err != nil {
		log.Fatalf("walk department label error: %v", err)
	}
	origin, err := ioutil.ReadFile(tools.LabelsFilePath)
	if err != nil {
		log.Fatalf("read origin labels error: %v", err)
	}
	current := labelConfig.Output(labels)

	if 0 == bytes.Compare(origin, current) {
		return
	}
	if err := ioutil.WriteFile(tools.LabelsFilePath, current, 0644); err != nil {
		log.Fatalf("write label to %s error: %v", tools.LabelsFilePath, err)
	}
	os.Exit(1)
}

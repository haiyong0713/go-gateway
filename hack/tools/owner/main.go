package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go-gateway/hack/tools"

	"github.com/ghodss/yaml"
)

var Changed = false

func main() {
	log.SetOutput(os.Stdout)

	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
		}

		if !info.IsDir() || path == "." {
			return nil
		}

		if tools.InBlacklist(path) {
			if info.IsDir() {
				return filepath.SkipDir
			} else {
				return nil
			}
		}

		if !tools.IsProjectRootDirectory(path) {
			return filepath.SkipDir
		}

		paths := strings.Split(path, string(filepath.Separator))
		pathsLen := len(paths)

		if !(pathsLen == tools.SubdivisionDirectoryLevel || (pathsLen > tools.SubdivisionDirectoryLevel && tools.IsProject(path))) {
			return nil
		}

		var (
			owner         = &tools.Owner{}
			ownerPath     = filepath.Join(path, tools.OWNERSFileName)
			requireLabels = []string{tools.AreaLabel(paths[1:])}
			ownerSet      = make(map[string]struct{})
		)

		if _, err = os.Stat(ownerPath); !os.IsNotExist(err) {
			if owner, err = tools.ReadOwner(ownerPath); err != nil {
				return err
			}
		}

		if pathsLen > tools.SubdivisionDirectoryLevel {
			upperOwnerPath := strings.Join(paths[:2], "/") + tools.OWNERSFileName
			if owner.Options.NoParentOwners == true {
				requireLabels = append(requireLabels, tools.AreaLabel(paths[1:2]))
			} else if _, err = os.Stat(upperOwnerPath); !os.IsNotExist(err) {
				if upper, err := tools.ReadOwner(upperOwnerPath); err != nil {
					return err
				} else if upper.Options.NoParentOwners == true {
					requireLabels = append(requireLabels, tools.AreaLabel(paths[1:2]))
				}
			}
		}

		for _, label := range owner.Labels {
			ownerSet[label] = struct{}{}
		}

		skip := true
		for _, label := range requireLabels {
			if _, ok := ownerSet[label]; !ok {
				if skip {
					skip = false
				}
				ownerSet[label] = struct{}{}
			}
		}

		if !skip {
			var labels []string
			for label := range ownerSet {
				labels = append(labels, label)
			}
			sort.Strings(labels)
			owner.Labels = labels
			data, err := yaml.Marshal(owner)
			if err != nil {
				return fmt.Errorf("fail to Marshal %q: %v\n", path, err)
			}
			data = append([]byte("# See the OWNERS docs at https://go.k8s.io/owners\n\n"), data...)

			if err := ioutil.WriteFile(ownerPath, data, 0644); err != nil {
				return fmt.Errorf("fail to write yaml %q: %v\n", path, err)
			} else {
				log.Printf("wrote %q\n", ownerPath)
				Changed = true
			}
		}

		if pathsLen > tools.SubdivisionDirectoryLevel {
			return filepath.SkipDir
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if Changed {
		log.Fatal("Please git-add OWNERS file.")
	}
}

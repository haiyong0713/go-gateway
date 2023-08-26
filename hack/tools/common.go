package tools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	// OWNERSFileName is the OWNERS file name
	OWNERSFileName = "OWNERS"

	// SubdivisionDirectoryLevel is directory level of the subdivision
	SubdivisionDirectoryLevel = 2

	JenkinsServiceLevel = 4

	// GroupName is the group name of repository at gitlab
	GitlabGroupName = "platform"

	// RepositoryName is the name of repository
	GitlabRepositoryName = "go-gateway"

	// ProwConfigRootPath is root of prow job config
	ProwConfigRootPath = "hack/prow"

	// ProjectRootDirectoryName is the root of project
	ProjectRootDirectoryName = "app"

	LabelPrefix = "area"
)

var (
	// ProwJobTemplatePath is file path where all prow job templates are
	ProwJobTemplatePath = filepath.Join(ProwConfigRootPath, "template")

	//ProwJobFileName is the file name of prow jobs
	ProwJobFileName = strings.ReplaceAll(fmt.Sprintf("%s_jobs.yaml", GitlabRepositoryName), "-", "_")

	// ProwJobFilePath is file path where all prow jobs are
	ProwJobFilePath = filepath.Join(ProwConfigRootPath, ProwJobFileName)

	// ProwJobFilePath is file path where all prow jobs are
	LabelsFilePath = filepath.Join(ProwConfigRootPath, "labels.yaml")

	RelativeProjectRootPath = filepath.Join(".", ProjectRootDirectoryName)
)

func IsProject(str string) bool {

	f, err := os.Stat(filepath.Join(str, "cmd"))
	if err != nil {
		return false
	} else {
		return f.IsDir()
	}
}

func IsAreaLabel(str string) bool {
	return strings.HasPrefix(str, LabelPrefix)
}

func AreaLabel(sub []string) string {
	return filepath.Join(append([]string{"area"}, sub...)...)
}

func IsProjectRootDirectory(str string) bool {
	return strings.HasPrefix(str, ProjectRootDirectoryName)
}

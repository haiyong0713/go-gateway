package delay

import "time"

type DeployType string

const (
	NormalDeploy DeployType = "build-and-deploy"
	OnlyDeploy   DeployType = "only-deploy"

	//发布阶段
	DisplayGenCode          int8 = 1
	DisplayUploadBoss       int8 = 2
	DisplayBuild            int8 = 3
	DisplayUatSetConfig     int8 = 4
	DisplayUatCreatApp      int8 = 5
	DisplayUatDeploy        int8 = 6
	DisplayUatSuspend       int8 = 7
	DisplayPreCreatApp      int8 = 8
	DisplayPreDeploy        int8 = 9
	DisplayPreSuspend       int8 = 10
	DisplayProdSetConfig    int8 = 11
	DisplayProdCreatApp     int8 = 12
	DisplayProdDeployCanary int8 = 13
	DisplayProdSuspend      int8 = 14
	DisplayProdDeploy       int8 = 15

	DisplayNameGenCode          = "gen-code"
	DisplayNameUploadBoss       = "upload-boss"
	DisplayNameBuild            = "build"
	DisplayNameUatSetConfig     = "create-dynamic-config"
	DisplayNameUatCreatApp      = "create-uat-app"
	DisplayNameUatDeploy        = "uat-deploy"
	DisplayNameUatSuspend       = "uat-suspend"
	DisplayNamePreCreatApp      = "create-pre-app"
	DisplayNamePreDeploy        = "pre-deploy"
	DisplayNamePreSuspend       = "pre-suspend"
	DisplayNameProdSetConfig    = "set-default-canary-proportion"
	DisplayNameProdCreatApp     = "create-prod-app"
	DisplayNameProdDeployCanary = "prod-deploy-canary"
	DisplayNameProdSuspend      = "prod-suspend"
	DisplayNameProdDeploy       = "prod-deploy"

	//发布阶段的状态
	DisplayStateNormal    int8 = 0
	DisplayStateRunning   int8 = 1
	DisplayStateSucceeded int8 = 2
	DisplayStateFailed    int8 = 3

	//wf状态
	WFStateNormal     int8 = 0
	WFStateFinished   int8 = 1
	WFStateFailed     int8 = 2
	WFStateManualStop int8 = 3 //暂时没用

	//节点状态
	NodeStatusPending   = "Pending"
	NodeStatusRunning   = "Running"
	NodeStatusSucceeded = "Succeeded"
	NodeStatusSkipped   = "Skipped"
	NodeStatusFailed    = "Failed"
	NodeStatusError     = "Error"
	NodeStatusOmitted   = "Omitted"
)

var DisplayStatusMap = map[string]int8{
	NodeStatusRunning:   DisplayStateRunning,
	NodeStatusSucceeded: DisplayStateSucceeded,
	NodeStatusFailed:    DisplayStateFailed,
}

var DisplayNameMap = map[string]int8{
	DisplayNameGenCode:          DisplayGenCode,
	DisplayNameUploadBoss:       DisplayUploadBoss,
	DisplayNameBuild:            DisplayBuild,
	DisplayNameUatSetConfig:     DisplayUatSetConfig,
	DisplayNameUatCreatApp:      DisplayUatCreatApp,
	DisplayNameUatDeploy:        DisplayUatDeploy,
	DisplayNameUatSuspend:       DisplayUatSuspend,
	DisplayNamePreCreatApp:      DisplayPreCreatApp,
	DisplayNamePreDeploy:        DisplayPreDeploy,
	DisplayNamePreSuspend:       DisplayPreSuspend,
	DisplayNameProdSetConfig:    DisplayProdSetConfig,
	DisplayNameProdCreatApp:     DisplayProdCreatApp,
	DisplayNameProdDeployCanary: DisplayProdDeployCanary,
	DisplayNameProdSuspend:      DisplayProdSuspend,
	DisplayNameProdDeploy:       DisplayProdDeploy,
}

type CreateWorkflowParams struct {
	NameSpace    string   `json:"namespace"`
	ResourceName string   `json:"resource_name"`
	EntryPoint   string   `json:"entrypoint"`
	Parameters   []string `json:"parameters"`
}

type ResumeWorkflowParams struct {
	NameSpace   string `json:"namespace"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
}

type WFDetail struct {
	ID           int64     `json:"id"`
	ApiName      string    `json:"api_name"`
	Boss         string    `json:"boss"`
	Version      string    `json:"version"`
	WFName       string    `json:"wf_name"`
	DiscoveryID  string    `json:"discovery_id"`
	Image        string    `json:"image"`
	DisplayName  string    `json:"display_name"`
	DisplayState int8      `json:"display_state"`
	State        int8      `json:"state"`
	Log          string    `json:"log"`
	Mtime        time.Time `json:"mtime"`
	Ctime        time.Time `json:"ctime"`
}

type Node struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	DisplayName string   `json:"displayName"`
	Children    []string `json:"children"`
	Type        string   `json:"type"`
	Phase       string   `json:"phase"`
	Message     string   `json:"message"`
}

type OutputsData struct {
	Parameters []struct {
		Name      string `json:"name"`
		Value     string `json:"value"`
		ValueFrom struct {
			Path string `json:"path"`
		} `json:"valueFrom"`
	} `json:"parameters"`
	//Artifacts []struct {
	//	Name        string `json:"name"`
	//	ArchiveLogs bool   `json:"archiveLogs"`
	//	S3          struct {
	//		Endpoint        string `json:"endpoint"`
	//		Bucket          string `json:"bucket"`
	//		Insecure        string `json:"insecure"`
	//		AccessKeySecret struct {
	//			Name string `json:"name"`
	//			Key  string `json:"key"`
	//		} `json:"accessKeySecret"`
	//	} `json:"s3"`
	//	SecretKeySecret struct {
	//		Name string `json:"name"`
	//		Key  string `json:"key"`
	//	} `json:"secretKeySecret"`
	//	Key string `json:"key"`
	//} `json:"artifacts"`
	//ExitCode string `json:"exitCode"`
}

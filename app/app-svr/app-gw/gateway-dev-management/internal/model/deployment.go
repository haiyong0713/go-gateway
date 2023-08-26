package model

type GetAuthReply struct {
	ApiVersion string `json:"apiVersion"`
	Status     int64  `json:"status"`
	Message    string `json:"message"`
	Data       struct {
		Token      string `json:"token"`
		PlatformId string `json:"platform_id"`
		UserName   string `json:"user_name"`
		Secret     string `json:"secret"`
		Expired    int64  `json:"expired"`
		Admin      bool   `json:"admin"`
	} `json:"data"`
}

type DeploymentCluster struct {
	Cluster struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		Zone string `json:"zone"`
	} `json:"cluster"`
	ClusterDNS                 bool   `json:"cluster_dns"`
	ConfigurationPlatform      string `json:"configuration_platform"`
	ConfigurationPlatformBuild string `json:"configuration_platform_build"`
	ConfigurationPlatformEnv   string `json:"configuration_platform_env"`
	CPUPolicy                  string `json:"cpu_policy"`
	CurrentPodTemplate         struct {
		AdvancedConfig string `json:"advanced_config"`
		Cluster        struct {
			DefaultRegistry string `json:"default_registry"`
		} `json:"cluster"`
		ConfigurationPlatform        string `json:"configuration_platform"`
		ConfigurationPlatformBuild   string `json:"configuration_platform_build"`
		ConfigurationPlatformEnv     string `json:"configuration_platform_env"`
		ConfigurationPlatformVersion string `json:"configuration_platform_version"`
		ID                           int64  `json:"id"`
		Image                        string `json:"image"`
		Version                      string `json:"version"`
	} `json:"current_pod_template"`
	ID               int64  `json:"id"`
	OverCommitFactor int64  `json:"over_commit_factor"`
	Resource         string `json:"resource"`
	ResourceLimit    struct {
		ID int64 `json:"id"`
	} `json:"resource_limit"`
	ResourcePool struct {
		Name string `json:"name"`
	} `json:"resource_pool"`
}

type DeploymentDetail struct {
	ApplicationClusters []*DeploymentCluster `json:"application_clusters"`
	ID                  int64                `json:"id"`
	Name                string               `json:"name"`
}

type GetStatusReply struct {
	Data struct {
		Items []*DeploymentDetail `json:"items"`
	} `json:"data"`
	Message string `json:"message"`
	Status  int64  `json:"status"`
}

type CalcReq struct {
	ClusterId        int64  `json:"cluster_id"`
	ClusterName      string `json:"cluster_name"`
	Constraints      string `json:"constraints"`
	CpuPolicy        string `json:"cpu_policy"`
	ResourcePoolName string `json:"resource_pool_name"`
	OverCommitFactor int64  `json:"over_commit_factor"`
	ResourceLimitId  int64  `json:"resource_limit_id"`
	CpuReq           int64  `json:"cpu_req"`
	CpuLimit         int64  `json:"cpu_limit"`
	MemReq           int64  `json:"mem_req"`
	MemLimit         int64  `json:"mem_limit"`
	EstorageReq      int64  `json:"estorage_req"`
	EstorageLimit    int64  `json:"estorage_limit"`
}

type CalcReply struct {
	ApiVersion string `json:"apiVersion"`
	Status     int64  `json:"status"`
	Message    string `json:"message"`
	Data       int64  `json:"data"`
}

type CalcResource struct {
	CPULimit      int64 `json:"cpu_limit"`
	CPUReq        int64 `json:"cpu_req"`
	EstorageLimit int64 `json:"estorage_limit"`
	EstorageReq   int64 `json:"estorage_req"`
	GpuLimit      int64 `json:"gpu_limit"`
	GpuReq        int64 `json:"gpu_req"`
	MemLimit      int64 `json:"mem_limit"`
	MemReq        int64 `json:"mem_req"`
}

type AvailableResource struct {
	AvailableResource int64  `json:"available_resource"`
	ClusterID         int64  `json:"cluster_id"`
	ClusterName       string `json:"cluster_name"`
	Constraints       string `json:"constraints"`
	CPULimit          int64  `json:"cpu_limit"`
	CPUPolicy         string `json:"cpu_policy"`
	CPUReq            int64  `json:"cpu_req"`
	EstorageLimit     int64  `json:"estorage_limit"`
	EstorageReq       int64  `json:"estorage_req"`
	MemLimit          int64  `json:"mem_limit"`
	MemReq            int64  `json:"mem_req"`
	OverCommitFactor  int64  `json:"over_commit_factor"`
	ResourceLimitID   int64  `json:"resource_limit_id"`
	ResourcePoolName  string `json:"resource_pool_name"`
}

type DeploymentReq struct {
	App                             string             `json:"app"`
	AppName                         string             `json:"app_name"`
	ApplicationID                   int64              `json:"application_id"`
	AutoPausePoint                  bool               `json:"auto_pause_point"`
	AvailableResources              *AvailableResource `json:"available_resources"`
	BatchSize                       int64              `json:"batch_size"`
	BatchTimeout                    int64              `json:"batch_timeout"`
	Cluster                         int64              `json:"cluster"`
	ClusterID                       int64              `json:"cluster_id"`
	ConfigurationPlatform           string             `json:"configuration_platform"`
	ConfigurationPlatformBuild      string             `json:"configuration_platform_build"`
	ConfigurationPlatformBuildInput string             `json:"configuration_platform_build_input"`
	ConfigurationPlatformEnv        string             `json:"configuration_platform_env"`
	EnvInfo                         string             `json:"env_info"`
	Image                           string             `json:"image"`
	ImageBefore                     string             `json:"image_before"`
	ImageTagType                    string             `json:"image_tag_type"`
	ShowConfig                      bool               `json:"showConfig"`
	SideCarImagesInfo               struct{}           `json:"side_car_images_info"`
	Strategy                        string             `json:"strategy"`
	Type                            string             `json:"type"`
	Version                         string             `json:"version"`
	Versions                        string             `json:"versions"`
}

type DeploymentReply struct {
	Data struct {
		ID int64 `json:"id"`
	} `json:"data"`
	Message string `json:"message"`
	Status  int64  `json:"status"`
}

type StartReply struct {
	Data struct {
		ID int64 `json:"id"`
	} `json:"data"`
	Message string `json:"message"`
	Status  int64  `json:"status"`
}

type ResumeReply struct {
	Status  int64  `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Id int64 `json:"id"`
	} `json:"data"`
}

type Action struct {
	AutoPausePoint bool   `json:"auto_pause_point"`
	Ctime          string `json:"ctime"`
	Duration       int64  `json:"duration"`
	EndTime        string `json:"end_time"`
	ID             int64  `json:"id"`
	Message        string `json:"message"`
	Mtime          string `json:"mtime"`
	Name           string `json:"name"`
	NextActionID   int64  `json:"next_action_id"`
	Operator       string `json:"operator"`
	RestartTime    string `json:"restart_time"`
	StartTime      string `json:"start_time"`
	Status         string `json:"status"`
}

type GetDeployReply struct {
	APIVersion string `json:"apiVersion"`
	Data       struct {
		Actions     []*Action `json:"actions"`
		Application struct {
			ID   int64  `json:"id"`
			Name string `json:"name"`
		} `json:"application"`
		ID           int64 `json:"id"`
		LastRevision int64 `json:"last_revision"`
		PodTemplate  struct {
			AdvancedConfig string `json:"advanced_config"`
			Cluster        struct {
				ID   int64  `json:"id"`
				Zone string `json:"zone"`
			} `json:"cluster"`
			Instances int64  `json:"instances"`
			Version   string `json:"version"`
		} `json:"pod_template"`
		Revision int64  `json:"revision"`
		Status   string `json:"status"`
	} `json:"data"`
	Message string `json:"message"`
	Status  int64  `json:"status"`
}

type DeployStatus struct {
	Id          int64  `json:"id"`
	Service     string `json:"service"`
	Zone        string `json:"zone"`
	Version     string `json:"version"`
	Current     string `json:"current"`
	Action      string `json:"action"`
	Percent     string `json:"percent"`
	Start       bool   `json:"start"`
	Rollback    bool   `json:"rollback"`
	Next        bool   `json:"next"`
	Done        bool   `json:"done"`
	LastReplica string `json:"last_replica"`
	Replica     string `json:"replica"`
}

type GetRevisionReply struct {
	APIVersion string           `json:"apiVersion"`
	Data       *GetRevisionData `json:"data"`
	Message    string           `json:"message"`
	Status     int64            `json:"status"`
}

type GetRevisionData struct {
	Status struct {
		AvailableReplicas      int64  `json:"available_replicas"`
		ExpectRevisionReplicas int64  `json:"expect_revision_replicas"`
		ReadyReplicas          int64  `json:"ready_replicas"`
		Status                 string `json:"status"`
	} `json:"status"`
}

type GetRevisionReq struct {
	AppID     int64
	ClusterID int64
	Revision  int64
}

type CasterAuthReply struct {
	Status  int64  `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Token string `json:"token"`
	} `json:"data"`
}

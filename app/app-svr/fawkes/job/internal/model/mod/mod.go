package mod

type (
	Env          string
	VersionState string
)

var (
	EnvTest = Env("test")
	EnvProd = Env("prod")

	VersionProcessing = VersionState("processing")
	VersionSucceeded  = VersionState("succeeded")
	VersionDisable    = VersionState("disable")
)

type File struct {
	ID          int64  `json:"id"`
	VersionID   int64  `json:"version_id"`
	Name        string `json:"name"`
	ContentType string `json:"-"`
	Size        int64  `json:"size"`
	Md5         string `json:"md5"`
	URL         string `json:"url"`
	IsPatch     bool   `json:"is_patch"`
	FromVer     int64  `json:"from_ver"`
}

type Version struct {
	ID        int64        `json:"id"`
	ModuleID  int64        `json:"module_id"`
	Env       Env          `json:"env"`
	Version   int64        `json:"version"`
	FromVerID int64        `json:"from_ver_id"`
	State     VersionState `json:"state"`
}

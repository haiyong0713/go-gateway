package tusmultipleversion

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

type Operator string

func (o Operator) Valid() bool {
	return o == GT || o == LT || o == EQ || o == NEQ
}

const (
	GT  = Operator("gt")  // >
	LT  = Operator("lt")  // <
	EQ  = Operator("eq")  // ==
	NEQ = Operator("neq") // !=
)

const (
	PlatAndroid   = int64(0)
	PlatIPhone    = int64(1)
	PlatIPad      = int64(2)
	PlatWPhone    = int64(3)
	PlatAndroidG  = int64(4)
	PlatIPhoneI   = int64(5)
	PlatIPadI     = int64(6)
	PlatAndroidTV = int64(7)
	PlatAndroidI  = int64(8)
	PlatAndroidB  = int64(9)
	PlatIPhoneB   = int64(10)
	PlatH5        = int64(15)
	PlatIPadHD    = int64(20)
	PlatAndroidHD = int64(90)
)

func PlatConverter(mobiApp, device string) int64 {
	switch mobiApp {
	case "iphone":
		if device == "pad" {
			return PlatIPad
		}
		return PlatIPhone
	case "white":
		return PlatIPhone
	case "ipad":
		return PlatIPadHD
	case "android":
		return PlatAndroid
	case "win":
		return PlatWPhone
	case "android_G":
		return PlatAndroidG
	case "android_i":
		return PlatAndroidI
	case "android_b":
		return PlatAndroidB
	case "iphone_b":
		return PlatIPhoneB
	case "iphone_i":
		if device == "pad" {
			return PlatIPadI
		}
		return PlatIPhoneI
	case "ipad_i":
		return PlatIPadI
	case "android_tv":
		return PlatAndroidTV
	case "android_hd":
		return PlatAndroidHD
	case "h5":
		return PlatH5
	}
	return PlatIPhone
}

type VersionInfo struct {
	ConfigVersion string        `json:"config_version"`
	BuildLimit    []*BuildLimit `json:"build_limit"`
	CreateTime    int64         `json:"create_time"`
	TusValues     []string      `json:"tus_values"`
}

type BuildLimit struct {
	Plat     int64    `json:"plat"`
	Operator Operator `json:"operator"`
	Build    int64    `json:"build"`
}

type BuildLimits []*BuildLimit

func (b BuildLimits) Valid() bool {
	for _, v := range b {
		if v.Plat < 0 || v.Plat > 90 {
			return false
		}
		if !v.Operator.Valid() {
			return false
		}
	}
	return true
}

func (b BuildLimits) AllowDeviceToUse(plat int64, build int64) bool {
	for _, limit := range b {
		if limit.Plat != plat {
			continue
		}
		switch limit.Operator {
		case LT:
			if build < limit.Build {
				return true
			}
		case GT:
			if build > limit.Build {
				return true
			}
		case EQ:
			if build == limit.Build {
				return true
			}
		case NEQ:
			if build != limit.Build {
				return true
			}
		}
	}
	return false
}

type ConfigVersionManager struct {
	Field        string         `json:"field"`
	VersionInfos []*VersionInfo `json:"version_infos"`
}

const FirstVersion = "v1.0"

//VersionIncrease add version like v1.0
func (c *ConfigVersionManager) VersionIncrease(limit []*BuildLimit, tusValues []string) (string, error) {
	if len(c.VersionInfos) == 0 {
		return FirstVersion, nil
	}
	sort.Slice(c.VersionInfos, func(i, j int) bool {
		return c.VersionInfos[i].ConfigVersion < c.VersionInfos[j].ConfigVersion
	})
	latestVersion := c.VersionInfos[len(c.VersionInfos)-1].ConfigVersion
	latestVersionNumber, err := strconv.ParseInt(strings.ReplaceAll(strings.ReplaceAll(latestVersion, "v", ""), ".0", ""), 10, 64)
	if err != nil {
		return "", errors.Errorf("invalidate version number %d", latestVersionNumber)
	}
	latestVersionReplacer := fmt.Sprintf("v%d.0", latestVersionNumber+1)
	c.VersionInfos = append(c.VersionInfos, &VersionInfo{
		ConfigVersion: latestVersionReplacer,
		BuildLimit:    limit,
		TusValues:     tusValues,
		CreateTime:    time.Now().Unix(),
	})
	return latestVersionReplacer, nil
}

func NewTaishanKey(field string) string {
	return fmt.Sprintf("%s_version_manager", field)
}

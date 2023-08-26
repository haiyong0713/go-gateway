package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"regexp"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/appstatic/admin/model"
)

const nameFmt = `^[a-zA-Z0-9._-]+$`
const fileFmt = "Mod_%d-%s/%s/%s"

// validate required data
func validateRequired(reqInfo *model.RequestVer) (err error) {
	reg := regexp.MustCompile(nameFmt)
	if res := reg.MatchString(reqInfo.ModName); !res {
		err = fmt.Errorf("mod_name %s contains illegal character", reqInfo.ModName)
		return
	}
	if res := reg.MatchString(reqInfo.Department); !res {
		err = fmt.Errorf("department %s contains illegal character", reqInfo.Department)
		return
	}
	return
}

// transform []int to []string
func sliceString(is []int) (ss []string) {
	for _, v := range is {
		ss = append(ss, fmt.Sprintf("%d", v))
	}
	return
}

// check limit data and build the Limit Struct, error is json error here
func checkLimit(reqInfo *model.RequestVer) (res *model.Limit, err error) {
	getFormat := "GetLimit Param (%s), Value = (%s)"
	res = &model.Limit{}
	// mobi_app
	if len(reqInfo.MobiAPP) != 0 {
		res.MobiApp = reqInfo.MobiAPP
	}
	// device
	if len(reqInfo.Device) != 0 {
		res.Device = reqInfo.Device
	}
	// plat
	if len(reqInfo.Plat) != 0 {
		res.Plat = reqInfo.Plat
	}
	if reqInfo.IsWifi != 0 {
		res.IsWifi = reqInfo.IsWifi
	}
	// Scale & Arch & Level
	if len(reqInfo.Scale) != 0 {
		res.Scale = sliceString(reqInfo.Scale)
	}
	if len(reqInfo.Arch) != 0 {
		res.Arch = sliceString(reqInfo.Arch)
	}
	if reqInfo.Level != 0 {
		res.Level = sliceString([]int{reqInfo.Level}) // treat level as others ( []int )
	}
	// build_range
	if buildStr := reqInfo.BuildRange; buildStr != "" {
		log.Info(getFormat, "build_range", buildStr)
		var build = model.Build{}
		if err = json.Unmarshal([]byte(buildStr), &build); err != nil { // json err
			log.Error("buildStr (%s) json.Unmarshal error(%v)", buildStr, err)
			return
		}
		if isValid := build.CheckRange(); !isValid { // range not valid
			err = fmt.Errorf("build range (%s) not valid", buildStr)
			log.Error("buildStr CheckRange Error (%v)", err)
			return
		}
		res.Build = &build
	}
	// sysver
	if sysverStr := reqInfo.Sysver; sysverStr != "" {
		var build = model.Build{}
		if err = json.Unmarshal([]byte(sysverStr), &build); err != nil { // json err
			log.Error("buildStr (%s) json.Unmarshal error(%v)", sysverStr, err)
			return
		}
		if isValid := build.CheckRange(); !isValid { // range not valid
			err = fmt.Errorf("build range (%s) not valid", sysverStr)
			log.Error("sysverStr CheckRange Error (%v)", err)
			return
		}
		res.Sysver = &build
	}
	// time_range
	if timeStr := reqInfo.TimeRange; timeStr != "" {
		log.Info(getFormat, "time_range", timeStr)
		var tr = model.TimeRange{}
		if err = json.Unmarshal([]byte(timeStr), &tr); err != nil {
			log.Error("timeStr (%s) json.Unmarshal error(%v)", timeStr, err)
			return
		}
		if tr.Stime != 0 && tr.Etime != 0 && tr.Stime > tr.Etime {
			err = fmt.Errorf("Stime(%d) is bigger than Etime(%d)", tr.Stime, tr.Etime)
			log.Error("Time Range Error(%v)", err)
			return
		}
		res.TimeRange = &tr
	}
	return
}

func validateFile(content []byte, poolID int64, platForm, fileName string) (fInfo *model.FileInfo, err error) {
	// parse file, get type, size, md5
	fInfo, err = apsSvc.ParseFile(content)
	if err != nil {
		log.Error("validateFile error(%+v)", err)
		return
	}
	if !apsSvc.TypeCheck(fInfo.Type) {
		log.Error("validateFile file type(%v) not valid", fInfo.Type)
		err = fmt.Errorf("请上传指定类型文件")
		return
	}
	// regex checking
	reg := regexp.MustCompile(nameFmt)
	if res := reg.MatchString(fileName); !res {
		err = fmt.Errorf("fileName %s contains illegal character", fileName)
		return
	}
	// upload file to BFS
	fInfo.Name = fmt.Sprintf(fileFmt, poolID, fInfo.Md5, platForm, fileName) // rename with the MD5 and poolID
	return
}

// validate the file type, content and upload it to the BFS storage
func validateFileAndUpload(ctx *bm.Context, req *http.Request, pool *model.ResourcePool, upType int64) (fInfo *model.FileInfo, err error) {
	var (
		location string
		file     multipart.File
		header   *multipart.FileHeader
		content  []byte
	)
	// get the file
	if file, header, err = req.FormFile("file"); err != nil {
		log.Error("validateFileAndUpload FormFile error(%v)", err)
		return
	}
	defer file.Close()
	// read the file
	if content, err = ioutil.ReadAll(file); err != nil {
		log.Error("validateFileAndUpload ReadAll error(%v)", err)
		return
	}
	if fInfo, err = validateFile(content, pool.ID, pool.Platform, header.Filename); err != nil {
		log.Error("validateFileAndUpload validateFile error(%v)", err)
		return
	}
	if upType == model.UploadBfs {
		if len(content) >= model.BfsMaxSize {
			return nil, fmt.Errorf("bfs最大允许上传不超过20M的文件")
		}
		if location, err = apsSvc.UploadBfs(ctx, fInfo.Name, fInfo.Type, time.Now().Unix(), content); err != nil {
			log.Error("validateFileAndUpload uploadBFS error(%v)", err)
			return
		}
	} else if upType == model.UploadBoss {
		if len(content) >= model.BossMaxSize {
			return nil, fmt.Errorf("boss最大允许上传不超过100M的文件")
		}
		if location, err = apsSvc.UploadBigFile(ctx, content, fInfo); err != nil {
			log.Error("validateFileAndUpload UploadBigFile error(%v)", err)
			return
		}
	} else {
		return nil, fmt.Errorf("参数错误")
	}
	fInfo.URL = location
	return
}

// for other systems
func addVer(c *bm.Context) {
	res := map[string]interface{}{}
	tmpRes, err := addFile(c, model.UploadBoss)
	if err != nil {
		res["message"] = "上传失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(tmpRes, nil)
}

// addFile .
func addFile(c *bm.Context, upType int64) (respData *model.RespAdd, err error) {
	var (
		pool       = model.ResourcePool{}
		department = model.Department{}
		req        = c.Request
		limitData  *model.Limit
		fInfo      *model.FileInfo
		reqInfo    = model.RequestVer{}
	)
	respData = &model.RespAdd{}
	if err = c.Bind(&reqInfo); err != nil {
		return
	}
	// validate required data
	if err = validateRequired(&reqInfo); err != nil {
		log.Error("addVer ModName, ResName Error (%v)", err)
		return
	}
	// validate department
	if err = apsSvc.DB.Where("`name` = ?", reqInfo.Department).First(&department).Error; err != nil {
		log.Error("addVer First department Error (%v)", err)
		return
	}
	// validate mod Name
	// -2：限制，-1：冻结，0-未删除 1-已删除
	// 限制资源
	// 界面调整：在相关的mod名下红字提醒，该mod仅限于xxxx版本以下的应用版本
	if err = apsSvc.DB.Where("`name` = ? AND `department_id` = ? AND (`deleted` = 0 OR `deleted` = -2) AND `action` = 1", reqInfo.ModName, department.ID).First(&pool).Error; err != nil {
		log.Error("addVer First Pool Error (%v)", err)
		return
	}
	// check limit & config data
	if limitData, err = checkLimit(&reqInfo); err != nil {
		log.Error("addVer CheckLimit Error (%v)", err)
		return
	}
	// validate file data
	if fInfo, err = validateFileAndUpload(c, req, &pool, upType); err != nil {
		log.Error("addVer ValidateFile Error (%v)", err)
		return
	}
	// DB & storage operation
	if respData.ResID, respData.Version, err = apsSvc.GenerateVer(c, reqInfo.ResName, limitData, fInfo, &pool, reqInfo.DefaultPackage); err != nil {
		log.Error("addVer GenerateVer Error (%v)", err)
		return
	}
	return
}

// addBigVer .
func addBigVer(c *bm.Context) {
	res := map[string]interface{}{}
	tmpRes, err := addFile(c, model.UploadBoss)
	if err != nil {
		res["message"] = "上传失败 " + err.Error()
		c.JSONMap(res, ecode.RequestErr)
		return
	}
	c.JSON(tmpRes, nil)
}

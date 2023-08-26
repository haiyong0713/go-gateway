package http

import (
	"context"
	"encoding/csv"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/http/blademaster"
	bm "go-common/library/net/http/blademaster"
	"go-gateway/app/web-svr/activity/admin/model/college"
	"io/ioutil"
	"strconv"
	"strings"
)

func collegeImport(c *blademaster.Context) {
	var (
		err  error
		data []byte
	)
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		log.Errorc(c, "importDetailCSV upload err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	defer file.Close()
	data, err = ioutil.ReadAll(file)
	if err != nil {
		log.Errorc(c, "importDetailCSV ioutil.ReadAll err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	r := csv.NewReader(strings.NewReader(string(data)))
	records, err := r.ReadAll()
	if err != nil {
		log.Errorc(c, "importDetailCSV r.ReadAll() err(%v)", err)
		c.JSON(nil, ecode.RequestErr)
		return
	}
	var args []*college.College
Loop:
	for i, row := range records {
		// continue first row
		if i == 0 {
			continue
		}
		// import csv state online
		arg := &college.College{}
		for field, value := range row {
			value = strings.TrimSpace(value)
			switch field {
			case 0:
				if value == "" {
					log.Warn("importDetailCSV name provinceID(%s)", value)
					continue Loop
				}
				provinceID, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					// continue Loop
				}
				arg.ProvinceID = provinceID
			case 1:
				if value == "" {
					log.Warn("importDetailCSV  ProvinceInitial empty(%s)", value)
					// continue Loop
				}
				arg.ProvinceInitial = value
			case 2:
				if value == "" {
					log.Warn("importDetailCSV Province empty(%s)", value)
					// continue Loop
				}
				arg.Province = value
			case 3:
				if value == "" {
					log.Warn("importDetailCSV CollegeName empty(%s)", value)
					continue Loop
				}
				arg.CollegeName = value
			case 4:
				if value == "" {
					log.Warn("importDetailCSV Initial empty(%s)", value)
					// continue Loop
				}
				arg.Initial = value
			case 5:
				if value == "" {
					// continue Loop
					arg.Mid = 0
				} else {
					mid, err := strconv.ParseInt(value, 10, 64)
					if err != nil {
						continue Loop
					}
					arg.Mid = mid
				}
			case 6:
				var relationStr string
				if value == "" {
					log.Warn("importDetailCSV relationArr empty(%s)", value)
					relationStr = ""
				} else {
					relationArr := strings.Split(value, "_")
					if len(relationArr) > 0 {
						relationStr = strings.Join(relationArr, ",")
					}
					if err != nil {
						continue Loop
					}
				}

				arg.RelationMid = relationStr
			case 7:
				var whiteStr string

				if value == "" {
					// log.Warn("importDetailCSV right answer empty(%s)", value)
					// continue Loop
					whiteStr = ""
				} else {
					whiteArr := strings.Split(value, "_")
					if len(whiteArr) > 0 {
						whiteStr = strings.Join(whiteArr, ",")
					}
					if err != nil {
						continue Loop
					}
				}

				arg.White = whiteStr
			}

		}
		args = append(args, arg)
	}
	if len(args) == 0 {
		log.Errorc(c, "importDetailCSV args no after filter")
		c.JSON(nil, ecode.RequestErr)
		return
	}
	go func() {
		ctx := context.Background()
		collegeSrv.CollegeBatchInsert(ctx, args)
	}()
	c.JSON(nil, nil)
}

// collegeList 学校对象
func collegeList(c *bm.Context) {
	var (
		err   error
		count int
		reply *college.Reply
	)
	reply = &college.Reply{}
	reply.List = make([]*college.College, 0)
	v := new(struct {
		Page int    `form:"pn" default:"1"`
		Size int    `form:"ps" default:"20"`
		Name string `form:"college_name"`
	})
	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	db := collegeSrv.DB
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	if v.Name != "" {
		db = db.Where("college_name = ?", v.Name)
	}
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).
		Find(&reply.List).Error; err != nil {
		log.Error("CollegeList(%d,%d) error(%v)", v.Page, v.Size, err)
		c.JSON(nil, err)
		return
	}
	if err = db.Model(&college.College{}).Count(&count).Error; err != nil {
		log.Error("CollegeList count error(%v)", err)
		c.JSON(nil, err)
		return
	}
	reply.Page = map[string]interface{}{
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}

	data := map[string]interface{}{
		"data": reply,
	}
	c.JSONMap(data, nil)
}

// saveCollege 存储
func saveCollege(c *bm.Context) {
	var (
		request = &college.College{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(nil, collegeSrv.CollegeInsertOrUpdate(c, request))
}

// collegeSaveAID 存储
func collegeSaveAID(c *bm.Context) {
	var (
		request = &college.AIDList{}
	)
	if err := c.Bind(request); err != nil {
		return
	}
	c.JSON(nil, collegeSrv.CollegeAidInsertOrUpdate(c, request))
}

// collegeAIDList 稿件列表
func collegeAIDList(c *bm.Context) {
	var (
		err   error
		count int
		reply *college.AidReply
	)
	v := new(struct {
		ID   int64 `form:"aid"`
		Page int   `form:"pn" default:"1"`
		Size int   `form:"ps" default:"20"`
	})
	reply = &college.AidReply{}
	if err = c.Bind(v); err != nil {
		c.JSON(nil, ecode.RequestErr)
		return
	}
	db := collegeSrv.AidDB
	if v.Page == 0 {
		v.Page = 1
	}
	if v.Size == 0 {
		v.Size = 20
	}
	if v.ID != 0 {
		db = db.Where("aid = ?", v.ID)
	}
	if err = db.
		Offset((v.Page - 1) * v.Size).Limit(v.Size).
		Find(&reply.List).Error; err != nil {
		log.Error("CollegeList(%d,%d) error(%v)", v.Page, v.Size, err)
		c.JSON(nil, err)
		return
	}
	if err = db.Model(&college.AIDList{}).Count(&count).Error; err != nil {
		log.Error("CollegeList count error(%v)", err)
		c.JSON(nil, err)
		return
	}
	reply.Page = map[string]interface{}{
		"pn":    v.Page,
		"ps":    v.Size,
		"total": count,
	}

	data := map[string]interface{}{
		"data": reply,
	}
	c.JSONMap(data, nil)
}

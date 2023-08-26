package http

import (
	"encoding/json"
	"fmt"
	bm "go-common/library/net/http/blademaster"
	pb "go-gateway/app/app-svr/app-feed/admin/api/entry"
	"go-gateway/app/app-svr/app-feed/ecode"
	"io/ioutil"
	"net/http"
	"strconv"
)

func createEntry(c *bm.Context) {
	req := &pb.CreateEntryReq{}
	if err := ReadRequestBody(c.Request, req); err != nil {
		c.JSON(nil, err)
		return
	}
	req.CreatedBy = GetCurrentUsername(c)

	c.JSON(entrySvc.CreateEntry(c, req))
}

func deleteEntry(c *bm.Context) {
	req := &pb.DeleteEntryReq{}
	if err := ReadRequestBody(c.Request, req); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(entrySvc.DeleteEntry(c, req))
}

func editEntry(c *bm.Context) {
	req := &pb.EditEntryReq{}
	if err := ReadRequestBody(c.Request, req); err != nil {
		c.JSON(nil, err)
		return
	}

	if req.Id == 0 {
		c.JSON(nil, ecode.EntryParamsError)
		return
	}

	req.CreatedBy = GetCurrentUsername(c)

	c.JSON(entrySvc.EditEntry(c, req))
}

func toggleEntry(c *bm.Context) {
	req := &pb.ToggleEntryOnlineStatusReq{}
	if err := ReadRequestBody(c.Request, req); err != nil {
		c.JSON(nil, err)
		return
	}
	c.JSON(entrySvc.ToggleEntry(c, req))
}

func getEntryList(c *bm.Context) {
	var (
		pn  int64
		ps  int64
		err error
	)

	pageSize := c.Request.URL.Query().Get("ps")
	if ps, err = strconv.ParseInt(pageSize, 10, 32); pageSize == "" || err != nil || ps <= 0 {
		ps = 10
	}

	pageNum := c.Request.URL.Query().Get("pn")
	if pn, err = strconv.ParseInt(pageNum, 10, 32); pageNum == "" || err != nil || pn <= 0 {
		pn = 1
	}
	req := &pb.GetEntryListReq{
		PageSize: int32(ps),
		PageNum:  int32(pn),
	}
	c.JSON(entrySvc.GetEntryList(c, req))
}

func setNextTimeSettings(c *bm.Context) {
	req := &pb.SetNextStateReq{}
	if err := ReadRequestBody(c.Request, req); err != nil {
		c.JSON(nil, err)
		return
	}
	req.CreatedBy = GetCurrentUsername(c)

	c.JSON(entrySvc.SetNextState(c, req))
}

func getTimeSettingList(c *bm.Context) {
	req := &pb.GetTimeSettingListReq{}
	entryID := c.Request.URL.Query().Get("entry_id")
	if eid, err := strconv.ParseInt(entryID, 10, 32); entryID == "" || err != nil || eid <= 0 {
		c.JSON(nil, ecode.EntryParamsError)
		return
	} else {
		req.EntryID = int32(eid)
	}
	c.JSON(entrySvc.GetTimeSettingList(c, req))
}

//func getCurrentEntry(c *bm.Context) {
//	req := &pb.GetAppEntryStateReq{}
//	plat := c.Request.URL.Query().Get("plat")
//	if p, err := strconv.ParseInt(plat, 10, 32); plat == "" || err != nil || p <= 0 {
//		c.JSON(nil, ecode.EntryParamsError)
//		return
//	} else {
//		req.Plat = int32(p)
//	}
//	build := c.Request.URL.Query().Get("build")
//	if b, err := strconv.ParseInt(build, 10, 32); build == "" || err != nil || b <= 0 {
//		c.JSON(nil, ecode.EntryParamsError)
//		return
//	} else {
//		req.Build = int32(b)
//	}
//	c.JSON(entrySvc.GetAppEntryState(c, req))
//}

// -------------------------- common method -------------------------------
func ReadRequestBody(req *http.Request, target interface{}) (err error) {
	requestBody, readErr := ioutil.ReadAll(req.Body)

	if readErr != nil {
		fmt.Println("request body read error:", readErr)
		return ecode.EntryParamsError
	}
	if len(requestBody) == 0 {
		return nil
	}

	if parseErr := json.Unmarshal(requestBody, &target); parseErr != nil {
		fmt.Println("request parse error: ", parseErr)
		return ecode.EntryParamsError
	}

	return nil
}

func GetCurrentUsername(c *bm.Context) string {
	var username string
	if un, ok := c.Get("username"); ok {
		username = un.(string)
	}
	return username
}

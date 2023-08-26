package bubble

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-common/library/database/sql"
	"go-common/library/ecode"
	"go-common/library/log"
	xtime "go-common/library/time"

	bubblemdl "go-gateway/app/app-svr/app-feed/admin/model/bubble"
	"go-gateway/app/app-svr/app-feed/admin/model/common"
	"go-gateway/app/app-svr/app-feed/admin/util"
)

const (
	_logActionAdd       = "add"
	_logActionEdit      = "edit"
	_logActionEditState = "edit_state"
)

func (s *Service) List(c context.Context, pn, ps int) (res *bubblemdl.List, err error) {
	var bubbles []*bubblemdl.Bubble
	if bubbles, err = s.bubbleDao.List(c, pn, ps); err != nil {
		log.Error("%v", err)
		return
	}
	res = &bubblemdl.List{
		Page: &bubblemdl.Page{
			Total:    len(bubbles),
			PageNum:  pn,
			PageSize: ps,
		},
		Items: bubbles,
	}
	return
}

func (s *Service) PositionList(c context.Context) (res []*bubblemdl.Sidebar, err error) {
	var (
		sidebars      []*bubblemdl.Sidebar
		sids          []int64
		sidebarLimits map[int64][]*bubblemdl.SidebarLimit
	)
	if sidebars, err = s.bubbleDao.Siderbar(c); err != nil {
		log.Error("%v", err)
		return
	}
	for _, sidebar := range sidebars {
		if sidebar == nil {
			continue
		}
		if sidebar.ID == 0 {
			continue
		}
		sids = append(sids, sidebar.ID)
	}
	if len(sids) == 0 {
		return
	}
	if sidebarLimits, err = s.bubbleDao.SiderbarLimit(c, sids); err != nil {
		log.Error("%v", err)
		return
	}
	for _, sidebar := range sidebars {
		if sidebar == nil {
			continue
		}
		if sidebar.ID == 0 {
			continue
		}
		res = append(res, sidebar)
		// 聚合limit数据
		sidebarlimit, ok := sidebarLimits[sidebar.ID]
		if !ok || len(sidebarlimit) == 0 {
			continue
		}
		var b []byte
		if b, err = json.Marshal(sidebarlimit); err != nil {
			log.Error("%v", err)
			return
		}
		sidebar.Limit = string(b)
	}
	return
}

func (s *Service) Add(c context.Context, params *bubblemdl.Param, uid int64, username string, positions []*bubblemdl.ParamPostion) (err error) {
	var clashBubbles []*bubblemdl.Bubble
	if clashBubbles, err = s.bubbleDao.Clash(c, params.STime, params.ETime); err != nil {
		log.Error("%v", err)
		return
	}
	var isClash bool
	for _, cb := range clashBubbles {
		if cb == nil {
			continue
		}
		for _, clashPos := range cb.Position {
			if clashPos == nil {
				continue
			}
			for _, paramPos := range positions {
				if clashPos == nil {
					continue
				}
				if paramPos.Plat == clashPos.Plat {
					isClash = true
					break
				}
			}
		}
	}
	if isClash {
		err = ecode.Error(ecode.RequestErr, "投放位置和时间有冲突")
		return
	}
	var tx *sql.Tx
	if tx, err = s.bubbleDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	var bid int64
	if bid, err = s.bubbleDao.TxAddBubble(tx, params.Position, params.Icon, params.Desc, params.URL, params.STime, params.ETime, params.WhiteList, username, bubblemdl.StateOnline); err != nil {
		log.Error("%v", err)
		return
	}
	if err = s.cache.Do(c, func(ctx context.Context) {
		s.ReadWhiteListCSV(bid, params.WhiteList, params.ETime)
	}); err != nil {
		log.Error("d.ReadWhiteListCSV bid(%d) files(%v) etime(%v) error(%v)", bid, params.WhiteList, params.ETime, err)
		return
	}
	if err = util.AddLogs(common.LogBubble, username, uid, 0, _logActionAdd, params); err != nil {
		log.Error("AddLog error %v", err)
		err = nil
	}
	return
}

func (s *Service) Edit(c context.Context, params *bubblemdl.Param, uid int64, username string, positions []*bubblemdl.ParamPostion) (err error) {
	var clashBubbles []*bubblemdl.Bubble
	if clashBubbles, err = s.bubbleDao.Clash(c, params.STime, params.ETime); err != nil {
		log.Error("%v", err)
		return
	}
	var isClash bool
	for _, cb := range clashBubbles {
		if cb == nil || (params.ID == cb.ID) {
			continue
		}
		for _, clashPos := range cb.Position {
			if clashPos == nil {
				continue
			}
			for _, paramPos := range positions {
				if clashPos == nil {
					continue
				}
				if paramPos.Plat == clashPos.Plat {
					isClash = true
					break
				}
			}
		}
	}
	if isClash {
		err = ecode.Error(ecode.RequestErr, "投放位置和时间有冲突")
		return
	}
	var tx *sql.Tx
	if tx, err = s.bubbleDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.bubbleDao.TxUpdateBubble(tx, params.ID, params.Position, params.Icon, params.Desc, params.URL, params.STime, params.ETime, params.WhiteList, username); err != nil {
		log.Error("%v", err)
		return
	}
	if err = s.cache.Do(c, func(ctx context.Context) {
		s.ReadWhiteListCSV(params.ID, params.WhiteList, params.ETime)
	}); err != nil {
		log.Error("d.ReadWhiteListCSV bid(%d) files(%v) etime(%v) error(%v)", params.ID, params.WhiteList, params.ETime, err)
		return
	}
	if err = util.AddLogs(common.LogBubble, username, uid, params.ID, _logActionEdit, params); err != nil {
		log.Error("AddLog error %v", err)
		err = nil
	}
	return
}

func (s *Service) BuddleState(c context.Context, params *bubblemdl.Param, uid int64, username string) (err error) {
	var tx *sql.Tx
	if tx, err = s.bubbleDao.BeginTran(c); err != nil {
		log.Error("s.fkDao.BeginTran() error(%v)", err)
		return
	}
	defer func() {
		if r := recover(); r != nil {
			//nolint:errcheck
			tx.Rollback()
			log.Error("%v", r)
		}
		if err != nil {
			if err1 := tx.Rollback(); err1 != nil {
				log.Error("tx.Rollback() error(%v)", err1)
			}
			return
		}
		if err = tx.Commit(); err != nil {
			log.Error("tx.Commit() error(%v)", err)
		}
	}()
	if _, err = s.bubbleDao.TxUpdateBubbleState(tx, params.ID, params.State); err != nil {
		log.Error("%v", err)
		return
	}
	if err = util.AddLogs(common.LogBubble, username, uid, params.ID, _logActionEditState, params.State); err != nil {
		log.Error("AddLog error %v", err)
		err = nil
	}
	return
}

func (s *Service) ReadWhiteListCSV(bid int64, files string, etime xtime.Time) {
	if bid == 0 {
		return
	}
	// 拆文件名
	//nolint:gosimple
	var filenames []string
	filenames = strings.Split(files, ",")
	if len(filenames) == 0 {
		log.Error("ReadWhiteListCSV files error bid(%d) files(%v) etime(%v)", bid, files, etime.Time())
		return
	}
	// 计算过期时间
	expire := int32(etime.Time().Unix() - time.Now().Unix())
	if expire <= 0 {
		log.Error("ReadWhiteListCSV expire error bid(%d) filename(%v) etime(%v)", bid, files, etime.Time())
		return
	}
	// 聚合mid
	for _, filename := range filenames {
		//nolint:gosec
		resp, err := http.Get(filename)
		if err != nil {
			log.Error("%v", err)
			continue
		}
		defer resp.Body.Close()
		buf := new(bytes.Buffer)
		if _, err := io.Copy(buf, resp.Body); err != nil {
			log.Error("%v", err)
			continue
		}
		//nolint:gosimple
		r := csv.NewReader(strings.NewReader(string(buf.Bytes())))
		columns, err := r.ReadAll()
		if err != nil {
			log.Error("%v", err)
			continue
		}
		for _, col := range columns {
			if col[0] == "mid" {
				continue
			}
			mid, err := strconv.ParseInt(col[0], 10, 64)
			if err != nil || mid == 0 {
				log.Error("ReadWhiteListCSV mid error bid(%d) filename(%v) etime(%v)", bid, filename, etime.Time())
				continue
			}
			for i := 0; i < 5; i++ {
				if err := s.bubbleDao.SetBubbleConfig(context.Background(), bid, mid, bubblemdl.BubblePushing, expire); err != nil {
					log.Error("ReadWhiteListCSV set mc error bid(%d) mids(%v) expire(%v)", bid, mid, expire)
					time.Sleep(time.Microsecond * 200)
					continue
				}
				break
			}
		}
	}
}

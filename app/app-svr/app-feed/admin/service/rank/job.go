package rank

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	xtime "go-common/library/time"

	arcgrpc "git.bilibili.co/bapis/bapis-go/archive/service"

	"go-gateway/app/app-svr/app-feed/admin/dataplat"
	rankModel "go-gateway/app/app-svr/app-feed/admin/model/rank"
)

const (
	_jobFlagOngoing     = "ongoing"
	_jobFlagFailed      = "failed"
	_jobFlagSucceed     = "succeed"
	_jobFlagLifeTime    = 600
	_jobFlagLifeTimeDay = 86400
)

var ctx = context.Background()

// 定时任务，每天凌晨计算出所有给数据平台用的原始视频列表，写入 rank_dataplat_av_rank
func (s *Service) JobGenDataplatAvRank() {
	for {
		// 检查当前时间，如果超过了 03:00，就检查一下是否需要跑任务
		//nolint:gomnd
		if time.Now().Hour() >= 3 {

			// 生成今天的key
			runnerKey := "job_rank_dataplat_" + genLogDate()

			// 正在跑，成功了，那就忽略，等待下个周期
			// 没有过，失败了，就启动任务
			if exist, flag, err := s.dao.GetTaskFlag(ctx, runnerKey); err != nil || (exist && (flag == _jobFlagOngoing || flag == _jobFlagSucceed)) {
			} else {
				// 标记，正在跑
				if err := s.dao.SetTaskFlag(ctx, runnerKey, _jobFlagOngoing, _jobFlagLifeTime); err != nil {
					_ = s.dao.SetTaskFlag(ctx, runnerKey, _jobFlagFailed, _jobFlagLifeTime)
				} else {
					s.AtomGenDataplatAvRank(runnerKey)
					// 标记，成功了，24小时有效
					_ = s.dao.SetTaskFlag(ctx, runnerKey, _jobFlagSucceed, _jobFlagLifeTimeDay)
				}
			}
		}

		// 10分钟检查一次
		time.Sleep(10 * time.Minute)
	}
}

func (s *Service) AtomGenDataplatAvRank(runnerKey string) {
	// 数据平台有接口限频，需要控制速度，hive 同一个 key 40s 一次

	log.Info("______AtomGenDataplatAvRank start______")
	configList, count, err := s.dao.QueryRankConfigList(0, "", -1, 0, 1, 10000)
	if err != nil {
		log.Error("s.dao.QueryRankConfigList error(%v)", err)
		return
	}

	if count == 0 {
		return
	}

	// 取出所有状态为 1 的榜单配置，每个榜单单独操作
	for _, config := range configList {
		if config.State == 1 {
			var (
				avidList []int64
				avList   []*rankModel.DataPlatAvInfo
			)
			// 取出当前配置活动下的 avid，返回所有的稿件avid列表
			if avidList, err = s.QueryAvidInAct(runnerKey, config); err != nil {
				// 记录下来任务失败，等待下次重新开始
				_ = s.dao.SetTaskFlag(ctx, runnerKey, _jobFlagFailed, _jobFlagLifeTime)
				return
			}

			// 取出avid列表下面的，符合config条件的所有avid,mid
			if avList, err = s.QueryAvidInConfig(runnerKey, avidList, config); err != nil {
				// 记录下来任务失败，等待下次重新开始
				_ = s.dao.SetTaskFlag(ctx, runnerKey, _jobFlagFailed, _jobFlagLifeTime)
				return
			}

			// 将结果数据批量写入到 rank_dataplat_av_rank
			if err = s.InsertDataplatAvRankData(runnerKey, avList, config); err != nil {
				// 记录下来任务失败，等待下次重新开始
				_ = s.dao.SetTaskFlag(ctx, runnerKey, _jobFlagFailed, _jobFlagLifeTime)
				return
			}

			log.Info("rank job succeed rankId(%v)", config.ID)
		}
	}
	log.Info("______AtomGenDataplatAvRank end______")
}

// 取出当前配置活动下的 avid，返回所有的稿件avid列表
func (s *Service) QueryAvidInAct(runnerKey string, config *rankModel.RankConfig) (avidList []int64, err error) {
	log.Info("______QueryAvidInAct start______")
	var (
		_querySql = "select distinct(wid) as avid from b_dwd.dwd_cmpn_actvy_basic_lottery_act_likes_a_d where log_date=%v and sid in (%v)"
		logDate   = genLogDate()
	)

	querySql := fmt.Sprintf(_querySql, logDate, config.ActIds)

	jobStatusUrl := ""
	if err = s.dao.
		CallDataPlatHiveAPI(ctx, "http://berserker.bilibili.co/avenger/api/762/query", querySql, &jobStatusUrl); err != nil {
		log.Error("QueryAvidInAct s.dao.CallDataPlatHiveAPI error(%v)", err)
		return
	}

	var jobRes *dataplat.ResponseHive
	ok := false

	if ok, jobRes = s.PoolingJobStatusUrl(runnerKey, jobStatusUrl); !ok {
		log.Error("QueryAvidInAct s.PoolingJobStatusUrl failed key(%v)", runnerKey)
		return
	}

	// 开始下载
	for _, filePath := range jobRes.HdfsPath {
		target := ""
		if target, err = s.DownloadFileByUrl(filePath); err != nil {
			log.Error("QueryAvidInAct s.DownloadFileByUrl url(%v) error(%v)", filePath, err)
			return
		}
		// 每个文件读出来结果，写入到返回结果内
		var fileLines []string
		if fileLines, err = s.getFileLines(target); err != nil {
			log.Error("QueryAvidInAct s.getFileLines path(%v) error(%v)", target, err)
			return
		}
		for _, line := range fileLines {
			avid, err := strconv.ParseInt(line, 10, 64)
			if err != nil {
				log.Error("QueryAvidInAct strconv.ParseInt input(%v) error(%v)", line, err)
				continue
			}
			avidList = append(avidList, avid)
		}
	}

	log.Info("______QueryAvidInAct end______")

	return
}

// 取出avid列表下面的，符合config条件的所有avid,mid
func (s *Service) QueryAvidInConfig(runnerKey string, avidList []int64, config *rankModel.RankConfig) (avList []*rankModel.DataPlatAvInfo, err error) {
	log.Info("______QueryAvidInConfig start______")
	var (
		_querySql   = "select aid as avid, mid, attribute from b_dwd.dwd_ctnt_arch_basic_arch_info_a_d where log_date=%v and aid in (%v) and pubtime >= '%v' and pubtime <= '%v' and `state`=0"
		_widthTids  = " and b_typeid(typeid) in (%v)"
		logDate     = genLogDate()
		avidStrList []string
	)

	for _, avid := range avidList {
		avidStrList = append(avidStrList, strconv.FormatInt(avid, 10))
	}
	sqlAvidParam := strings.Join(avidStrList, ",")

	querySql := fmt.Sprintf(_querySql, logDate, sqlAvidParam, config.ArchiveStime.Time().Format("2006-01-02 15:04:05"), config.ArchiveEtime.Time().Format("2006-01-02 15:04:05"))

	if config.Tids != "" {
		querySql += fmt.Sprintf(_widthTids, config.Tids)
	}

	jobStatusUrl := ""
	if err = s.dao.
		CallDataPlatHiveAPI(ctx, "http://berserker.bilibili.co/avenger/api/763/query", querySql, &jobStatusUrl); err != nil {
		log.Error("QueryAvidInConfig s.dao.CallDataPlatHiveAPI error(%v)", err)
		return
	}

	var jobRes *dataplat.ResponseHive
	ok := false

	if ok, jobRes = s.PoolingJobStatusUrl(runnerKey, jobStatusUrl); !ok {
		log.Error("QueryAvidInConfig s.PoolingJobStatusUrl failed key(%v)", runnerKey)
		return
	}

	// 开始下载
	for _, filePath := range jobRes.HdfsPath {
		target := ""
		if target, err = s.DownloadFileByUrl(filePath); err != nil {
			log.Error("QueryAvidInConfig s.DownloadFileByUrl url(%v) error(%v)", filePath, err)
			return
		}
		// 每个文件读出来结果，写入到返回结果内
		var fileLines []string
		if fileLines, err = s.getFileLines(target); err != nil {
			log.Error("QueryAvidInConfig s.getFileLines path(%v) error(%v)", target, err)
			return
		}
		for _, line := range fileLines {
			var (
				avid      int64
				mid       int64
				attribute int64
			)
			for i, value := range hiveString2Int64Array(line) {
				if i == 0 {
					avid = value
				}
				if i == 1 {
					mid = value
				}
				//nolint:gomnd
				if i == 2 {
					attribute = value
				}
			}

			//nolint:gomnd
			if ((attribute>>12)&int64(1)) == 1 || ((attribute>>24)&int64(1)) == 1 {
				// 过滤掉私单和联合投稿
				continue
			}

			avList = append(avList, &rankModel.DataPlatAvInfo{
				Avid: avid,
				Mid:  mid,
			})
		}
	}

	avManuallyList := string2Int64Array(config.AvManuallyList)

	if len(avManuallyList) > 0 {
		var archives *arcgrpc.ArcsReply
		if archives, err = s.arcClient.Arcs(ctx, &arcgrpc.ArcsRequest{Aids: avManuallyList}); err != nil {
			log.Error("s.arcClient.Arcs error %v", err)
			err = nil
		}

		for _, avid := range avManuallyList {
			var mid int64
			if info, ok := archives.Arcs[avid]; ok {
				mid = info.Author.Mid
			}
			avList = append(avList, &rankModel.DataPlatAvInfo{
				Avid: avid,
				Mid:  mid,
			})
		}
	}

	log.Info("______QueryAvidInConfig end______")

	return
}

// 将结果数据批量写入到 rank_dataplat_av_rank
func (s *Service) InsertDataplatAvRankData(runnerKey string, avList []*rankModel.DataPlatAvInfo, config *rankModel.RankConfig) (err error) {
	log.Info("______InsertDataplatAvRankData start______")
	var (
		_insertSql  = "insert into rank_dataplat_av_rank (avid, av_author, rank_id, max_cnt_play, max_cnt_like, max_cnt_coin, max_cnt_share, base_play, base_like, base_coin, base_share, log_date) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) ON DUPLICATE KEY UPDATE av_author=?, max_cnt_play=?, max_cnt_like=?, max_cnt_coin=?, max_cnt_share=?, base_play=?, base_like=?, base_coin=?, base_share=?"
		maxCntPlay  = 1
		basePlay    = 1
		maxCntLike  = 1
		baseLike    = 1
		maxCntCoin  = 1
		baseCoin    = 1
		maxCntShare = 1
		baseShare   = 1
		logDate     = genLogDate()
	)

	var scoreConfig []*rankModel.ScoreConfig
	if err = json.Unmarshal([]byte(config.ScoreConfig), &scoreConfig); err != nil {
		log.Error("InsertDataplatAvRankData json.Unmarshal error(%v)", err)
		return
	}
	for _, v := range scoreConfig {
		if v.Action == "play" {
			maxCntPlay = v.CntPerDay
			if maxCntPlay < 1 {
				maxCntPlay = 1
			}
			basePlay = v.Base
			if basePlay < 1 {
				basePlay = 1
			}
		}
		if v.Action == "like" {
			maxCntLike = v.CntPerDay
			if maxCntLike < 1 {
				maxCntLike = 1
			}
			baseLike = v.Base
			if baseLike < 1 {
				baseLike = 1
			}
		}
		if v.Action == "coin" {
			maxCntCoin = v.CntPerDay
			if maxCntCoin < 1 {
				maxCntCoin = 2
			}
			baseCoin = v.Base
			if baseCoin < 1 {
				baseCoin = 1
			}
		}
		if v.Action == "share" {
			maxCntShare = v.CntPerDay
			if maxCntShare < 1 {
				maxCntShare = 1
			}
			baseShare = v.Base
			if baseShare < 1 {
				baseShare = 1
			}
		}
	}

	for i, info := range avList {
		// 每秒插入10条
		time.Sleep(100 * time.Millisecond)
		if err = s.dao.DB.Exec(_insertSql, info.Avid, info.Mid, config.ID, maxCntPlay, maxCntLike, maxCntCoin, maxCntShare, basePlay, baseLike, baseCoin, baseShare, logDate, info.Mid, maxCntPlay, maxCntLike, maxCntCoin, maxCntShare, basePlay, baseLike, baseCoin, baseShare).Error; err != nil {
			log.Error("InsertDataplatAvRankData s.dao.DB.Exec error(%v)", err)
			err = nil
			continue
		}

		// 每插入1000条，续期一次
		if i%1000 == 0 {
			if err = s.dao.SetTaskFlag(ctx, runnerKey, _jobFlagOngoing, _jobFlagLifeTime); err != nil {
				log.Error("PoolingJobStatusUrl s.dao.SetTaskFlag error(%v)", err)
				err = nil
			}
		}
	}

	log.Info("______InsertDataplatAvRankData end______")

	return
}

// 轮询结果
func (s *Service) PoolingJobStatusUrl(runnerKey string, jobStatusUrl string) (ok bool, jobRes *dataplat.ResponseHive) {
	const (
		_jobStatusSucceed = 1
		_jobStatusFailed  = 2
		_jobStatusOngoing = 3
		_jobStatusWaiting = 4
	)

	var (
		err error
	)

	jobRes = &dataplat.ResponseHive{}

	// 每分钟查一次结果
	for {
		time.Sleep(1 * time.Minute)
		if err = s.dao.SetTaskFlag(ctx, runnerKey, _jobFlagOngoing, _jobFlagLifeTime); err != nil {
			log.Error("PoolingJobStatusUrl s.dao.SetTaskFlag error(%v)", err)
			//nolint:ineffassign
			err = nil
		}

		if err = s.Client.Get(ctx, jobStatusUrl, "", url.Values{}, jobRes); err != nil {
			log.Error("PoolingJobStatusUrl s.Client.Get(%v) error(%v)", jobStatusUrl, err)
			//nolint:ineffassign
			err = nil
			continue
		}

		log.Info("______AtomGenDataplatAvRank url(%v) jobRes(%v)", jobStatusUrl, jobRes)

		if jobRes.JobStatusId == _jobStatusSucceed {
			// 成功了，结束轮询
			ok = true
			break
		}

		//nolint:staticcheck
		if jobRes.JobStatusId == _jobStatusOngoing || jobRes.JobStatusId == _jobStatusWaiting {
			// 等待
		}

		if jobRes.JobStatusId == _jobStatusFailed {
			// 失败了，放弃后面的操作
			ok = false
			return
		}
	}
	return
}

// 通过 URL 下载某个文件到本地
func (s *Service) DownloadFileByUrl(url string) (target string, err error) {
	var netResp *http.Response
	//nolint:gosec
	netResp, err = http.Get(url)
	if err != nil {
		log.Error("http.Get(%v) error %v", url, err)
		return
	}
	defer netResp.Body.Close()

	pwd := ""
	if pwd, err = os.Getwd(); err != nil {
		log.Error("os.Getwd() error(%v)", err)
		return
	}

	// 创建一个文件用于保存
	filePath := path.Join(pwd, "rank", genLogDate())
	if ext := path.Ext(filePath); ext == "" {
		err = os.MkdirAll(filePath, os.ModePerm)
		if err != nil {
			log.Error("os.Mkdir error %s", err)
			return
		}
	}

	fileName := fmt.Sprintf("%x", md5.Sum([]byte(url)))
	target = path.Join(filePath, fileName)

	var out *os.File
	out, err = os.Create(target)
	if err != nil {
		log.Error("os.Create %v error %v", target, err)
		return
	}
	defer out.Close()

	// 然后将响应流和文件流对接起来
	_, err = io.Copy(out, netResp.Body)
	if err != nil {
		log.Error("io.Copy error %v", err)
		return
	}

	log.Info("DownloadFileByUrl succeed url(%s) target(%s)", url, target)

	return
}

// 定时任务，每5s看一次当前所有榜单配置，是否要继续流转状态
func (s *Service) JobUpdateRankState() {
	for {
		s.AtomUpdateRankState()
		time.Sleep(5 * time.Second)
	}
}

func (s *Service) AtomUpdateRankState() {
	configList, count, err := s.dao.QueryRankConfigList(0, "", -1, 0, 1, 10000)
	if err != nil {
		log.Error("s.dao.QueryRankConfigList error(%v)", err)
		return
	}

	if count == 0 {
		return
	}

	for _, config := range configList {
		// 忽略掉的 state 有：2 已结束 3 已结榜
		if config.State == 2 || config.State == 3 {
			continue
		}
		currentTime := xtime.Time(time.Now().Unix())
		// 如果是 0 未开始，检查当前时间是否 >= 开始时间，如果是，变更为 1 进行中
		if config.State == 0 && currentTime >= config.STime {
			if err := s.dao.UpdateRankState(config.ID, 1); err != nil {
				log.Error("JobUpdateRankState s.dao.UpdateRankState error(%v)", err)
			}
		}
		// 如果是 1 进行中，检查当前时间是否 >= 结束时间，如果是，变更为 2 已结束
		if config.State == 1 && currentTime >= config.ETime {
			if err := s.dao.UpdateRankState(config.ID, 2); err != nil {
				log.Error("JobUpdateRankState s.dao.UpdateRankState error(%v)", err)
			}
		}
	}
}

// 读出文件的每一行
func (s *Service) getFileLines(filePath string) (fileLines []string, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		log.Error("os.Open() error %v", filePath)
		return
	}
	defer file.Close()

	fd := bufio.NewReader(file)
	for {
		line, err := fd.ReadString('\n')
		if err != nil {
			break
		}
		fileLines = append(fileLines, strings.Replace(line, "\n", "", 1))
	}

	return
}

package service

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"go-gateway/app/web-svr/esports/job/component"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	arcclient "git.bilibili.co/bapis/bapis-go/archive/service"
	tagmdl "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	"go-common/library/log"
	"go-common/library/queue/databus"
	errGroup "go-common/library/sync/errgroup.v2"
	"go-common/library/xstr"
	"go-gateway/app/web-svr/esports/job/model"
	"go-gateway/app/web-svr/esports/job/tool"
	"go-gateway/pkg/idsafe/bvid"

	"github.com/go-vgo/gt/file"
)

const (
	_insertAct     = "insert"
	_tableArchive  = "archive"
	_otherTitle    = "其他"
	_autoCheck     = 2
	_autoCheckPass = 4

	berserkerSql4ArchiveStats     = `select avid, play, reply, fav, coin, danmu, share, likes from archive.dws_archive_daily where log_date = "%v" and avid in (%v)`
	archiveScoreDownloadShell     = `/data/archive/./download.sh %v %v`
	archiveScoreTmpDir            = "/data/archive/%v"
	archiveScoreTmpDir4Shell      = "/data/archive"
	archiveScoreTmpFilePath       = "/data/archive/%v/%v"
	archiveScoreTmpFilePath4Shell = "/data/archive/download.sh"
	archiveStatsOriginFilename    = "archive_stats_origin_%v_%v.csv"
	archiveScoreBackupFilename    = "archive_score_%v.csv"
	archiveAutoFilePath           = "/data/%v"
	archiveAutoFilename           = "rule_auto_archive_%v.csv"

	shellContentOfDownload = `#!/bin/bash
wget -c "$1" -O "$2"
if [ $? -ne 0 ];then
    echo "download resource failed"
    exit 1
fi`

	//archiveScoreDownloadShell     = `/Users/leelei/data/archive/./download.sh %v %v`
	//archiveScoreTmpDir            = "/Users/leelei/data/archive/%v"
	//archiveScoreTmpDir4Shell      = "/Users/leelei/data/archive"
	//archiveScoreTmpFilePath       = "/Users/leelei/data/archive/%v/%v"
	//archiveScoreTmpFilePath4Shell = "/Users/leelei/data/archive/download.sh"
	//archiveStatsOriginFilename    = "archive_stats_origin_%v_%v.csv"
	//archiveScoreBackupFilename    = "archive_score_%v.csv"

	archiveScoreAlarmMsgTemplate = `赛事视频得分更新耗时：<font color=\"info\">%v</font>，请相关同事注意。\n
>预计更新赛事视频数:<font color=\"comment\">%v</font> \n
>成功更新赛事视频数:<font color=\"info\">%v</font> \n
>失败更新赛事视频数:<font color=\"warning\">%v</font> \n
>详情未匹配赛事视频数:<font color=\"warning\">%v</font> \n
>实时统计未匹配赛事视频数:<font color=\"warning\">%v</font> \n
>%v`
	archiveScoreAlarmMsgTemplate4InternalErr = `<font color=\"warning\">服务内部错误：%v, 请管理员及时查看</font>`

	maxLimit4ArchiveArcsQuery  = 50
	maxLimit4ArchiveStatsQuery = 100
	maxLimit4DBQuery           = 5000
)

// start archive score update worker(once a day)
// 1. query specified archives from db
// 2. fetch archives stats from data center and archive server
// 3. calculate score for every archive
// 4. update related archive in db, record scores into local files
func (s *Service) startArchiveScoreWorker() {
	if !tool.IsArchiveScoreBizEnabled() {
		return
	}

	if err := initShellScript(); err != nil {
		fmt.Println(fmt.Sprintf("ArchiveScoreBiz >>> startArchiveScoreWorker failed, err: %v", err))

		return
	}

	date := time.Now().Format("20060102")
	dir := fmt.Sprintf(archiveScoreTmpDir, date)
	if mkdirErr := os.MkdirAll(dir, os.ModePerm); mkdirErr != nil {
		if bs, err := tool.GenAlarmMsgDataByType(
			tool.AlarmMsgTypeOfText,
			fmt.Sprintf(archiveScoreAlarmMsgTemplate4InternalErr, mkdirErr)); err == nil {
			_ = tool.SendCorpWeChatRobotAlarm(bs)
		}

		return
	}

	var (
		archiveID            = int64(0)
		archives             = make([]*model.Arc, 0)
		ctx                  = context.Background()
		fetchArchiveErr      error
		fetchArchiveErrCount int
		loopIndex            = 0

		wg = new(sync.WaitGroup)

		expectedC        = make(chan int64, 1)
		succeedC         = make(chan int64, 1)
		failedC          = make(chan int64, 1)
		noArchiveDetailC = make(chan int64, 1)
		noArchiveStatsC  = make(chan int64, 1)

		expectedNum, succeedNum, failedNum, noArchiveDetailNum, noArchiveStatsNum int64
	)

	startTime := time.Now()
	defer func() {
		if fetchArchiveErr != nil {
			if bs, alarmErr := tool.GenAlarmMsgDataByType(
				tool.AlarmMsgTypeOfText,
				fmt.Sprintf(archiveScoreAlarmMsgTemplate4InternalErr, fetchArchiveErr)); alarmErr == nil {
				_ = tool.SendCorpWeChatRobotAlarm(bs)
			}

			return
		}

		bs, genAlarmMsgErr := genAlarmTemplate(
			expectedNum,
			succeedNum,
			failedNum,
			noArchiveDetailNum,
			noArchiveStatsNum,
			time.Since(startTime))
		if genAlarmMsgErr == nil {
			if err := tool.SendCorpWeChatRobotAlarm(bs); err != nil {
				fmt.Println(fmt.Sprintf("ArchiveScoreBiz >>> SendCorpWeChatRobotAlarm occur err: %v", err))
			}
		} else {
			fmt.Println(fmt.Sprintf("ArchiveScoreBiz >>> genAlarmTemplate occur err: %v", genAlarmMsgErr))
		}
	}()

	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		for {
			select {
			case v, ok := <-expectedC:
				if !ok {
					expectedC = nil
				} else {
					expectedNum = expectedNum + v
				}
			case v, ok := <-succeedC:
				if !ok {
					succeedC = nil
				} else {
					succeedNum = succeedNum + v
				}
			case v, ok := <-failedC:
				if !ok {
					failedC = nil
				} else {
					failedNum = failedNum + v
				}
			case v, ok := <-noArchiveDetailC:
				if !ok {
					noArchiveDetailC = nil
				} else {
					noArchiveDetailNum = noArchiveDetailNum + v
				}
			case v, ok := <-noArchiveStatsC:
				if !ok {
					noArchiveStatsC = nil
				} else {
					noArchiveStatsNum = noArchiveStatsNum + v
				}
			}

			if expectedC == nil && succeedC == nil && failedC == nil && noArchiveDetailC == nil && noArchiveStatsC == nil {
				return
			}
		}
	}()

	log.Infoc(ctx, "[Service][Cron][ArchivesScoreSync][Begin]")
	for {
		loopIndex++
		archives, archiveID, fetchArchiveErr = s.fetchArchives(ctx, archiveID)

		if fetchArchiveErr != nil {
			fetchArchiveErrCount++
		} else {
			fetchArchiveErrCount = 0

			expectedC <- int64(len(archives))
			s.updateArchivesScore(ctx, archives, loopIndex, succeedC, failedC, noArchiveDetailC, noArchiveStatsC)
			if isArchivesQueryDone(archives, archiveID) {
				break
			}
		}

		if fetchArchiveErrCount >= 3 {
			break
		}
	}
	log.Infoc(ctx, "[Service][Cron][ArchivesScoreSync][End]")

	{
		close(expectedC)
		close(succeedC)
		close(failedC)
		close(noArchiveDetailC)
		close(noArchiveStatsC)
	}

	wg.Wait()
}

func isArchivesQueryDone(archives []*model.Arc, lastArchiveID int64) bool {
	if lastArchiveID == 0 && len(archives) == 0 {
		return true
	}

	if lastArchiveID != 0 && len(archives) < maxLimit4DBQuery {
		return true
	}

	return false
}

func genAlarmTemplate(expectedNum, succeedNum, failedNum, noArchiveDetailNum, noArchiveStatsNum int64,
	duration time.Duration) (bs []byte, err error) {
	robot, robotErr := tool.Robot()
	if robotErr == nil {
		content := fmt.Sprintf(
			archiveScoreAlarmMsgTemplate,
			duration,
			expectedNum,
			succeedNum,
			failedNum,
			noArchiveDetailNum,
			noArchiveStatsNum,
			tool.MentionUserIDs(robot, tool.AlarmMsgTypeOfMarkdown))

		return tool.GenAlarmMsgDataByType(tool.AlarmMsgTypeOfMarkdown, content)
	}

	err = robotErr

	return
}

func initShellScript() error {
	if file.Exist(archiveScoreTmpFilePath4Shell) {
		return nil
	}

	if err := os.MkdirAll(archiveScoreTmpDir4Shell, os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(archiveScoreTmpFilePath4Shell)
	if err != nil {
		return err
	}

	defer func() {
		_ = f.Close()
	}()

	if _, err := f.Write([]byte(shellContentOfDownload)); err != nil {
		return err
	}

	return execShell(fmt.Sprintf("chmod +x %v", archiveScoreTmpFilePath4Shell))
}

func downloadRemoteFiles2Local(files []string, loopIndex int) []string {
	localFilePaths := make([]string, 0)

	if len(files) > 0 {
		for k, v := range files {
			localFilePath := fmt.Sprintf(
				archiveScoreTmpFilePath,
				time.Now().Format("20060102"),
				fmt.Sprintf(archiveStatsOriginFilename, loopIndex, k))

			cmd := fmt.Sprintf(archiveScoreDownloadShell, v, localFilePath)
			cmdErr := execShell(cmd)
			if cmdErr != nil {
				fmt.Println(fmt.Sprintf("downloadRemoteFiles2Local(cmd: %v) failed, err: %v", cmd, cmdErr))
			} else {
				localFilePaths = append(localFilePaths, localFilePath)
			}
		}
	}

	return localFilePaths
}

func execShell(cmd string) (err error) {
	_, err = exec.Command("/bin/bash", "-c", cmd).Output()

	return
}

func (s *Service) updateArchivesScore(ctx context.Context, archives []*model.Arc, loopIndex int,
	succeedC, failedC, noArchiveDetailC, noArchiveStatsC chan int64) {
	if len(archives) == 0 {
		return
	}

	archiveAIDs := make([]int64, 0)
	for _, v := range archives {
		if v.Aid != 0 {
			archiveAIDs = append(archiveAIDs, v.Aid)
		} else {
			failedC <- 1
		}
	}

	if len(archiveAIDs) == 0 {
		return
	}

	now := time.Now()
	archiveArcM, archiveStatM := s.fetchArchivesInfoAndStats(ctx, archiveAIDs)
	if d := len(archiveAIDs) - len(archiveArcM); d > 0 {
		noArchiveDetailList := make([]int64, 0)
		for _, aID := range archiveAIDs {
			if _, ok := archiveArcM[aID]; !ok {
				noArchiveDetailList = append(noArchiveDetailList, aID)
			}
		}
		if len(noArchiveDetailList) > 0 {
			fmt.Println("updateArchivesScore >>> noArchiveDetailList: ", loopIndex, noArchiveDetailList)
		}

		noArchiveDetailC <- int64(d)
	}

	if d := len(archiveAIDs) - len(archiveStatM); d > 0 {
		noArchiveStatsList := make([]int64, 0)
		for _, aID := range archiveAIDs {
			if _, ok := archiveStatM[aID]; !ok {
				noArchiveStatsList = append(noArchiveStatsList, aID)
			}
		}
		if len(noArchiveStatsList) > 0 {
			fmt.Println("updateArchivesScore >>> noArchiveStatsList: ", loopIndex, noArchiveStatsList)
		}

		noArchiveStatsC <- int64(d)
	}

	query := genBerserkerQuery(archiveAIDs, now)

	// build data center request 4 query
	remoteFiles, err := tool.FetchBerserkerJobFile(query)
	fmt.Println("updateArchivesScore >>> tool.FetchBerserkerJobFile: ", loopIndex, query, remoteFiles, err)
	if err != nil {
		fmt.Println(
			fmt.Sprintf("ArchiveScoreBiz >>> tool.FetchBerserkerJobFile failed, query: %v, err: %v", query, err))

		return
	}

	if localFilePaths := downloadRemoteFiles2Local(remoteFiles, loopIndex); len(localFilePaths) > 0 {
		localFilePath := localFilePaths[0]
		if archiveDWM, err := s.updateArchiveScoreByStatsFile(localFilePath); err == nil {
			scoreMap := genArchiveScoreMap(archiveArcM, archiveStatM, archiveDWM, succeedC, failedC)
			if len(scoreMap) > 0 {
				if tool.CanResetArchiveScoreInDB() {
					s.dao.UpdateScoreByArchiveMap(ctx, scoreMap)
				}
				log.Infoc(ctx, "[Service][Cron][ArchivesScoreSync][Handler][Running], len:%d", len(scoreMap))

				filename := fmt.Sprintf(archiveScoreBackupFilename, loopIndex)
				filePath := fmt.Sprintf(archiveScoreTmpFilePath, now.Format("20060102"), filename)
				recordArchiveScoreIntoLocalFile(filePath, scoreMap, archiveArcM)
			}
		}

		if !tool.KeepBackupFile() {
			_ = os.Remove(localFilePath)
		}
	}
}

func genArchiveScoreMap(archiveArcM map[int64]*arcclient.Arc, archiveStatM map[int64]*arcclient.Stat,
	archiveDWM map[int64]*model.ArchiveStats, succeedC, failedC chan int64) map[int64]int64 {
	scoreMap := make(map[int64]int64, 0)

	for archiveID, archiveArc := range archiveArcM {
		if statD, ok := archiveStatM[archiveID]; ok {
			if d, ok := archiveDWM[archiveID]; ok {
				d.Rebuild(statD)
				score := calculateScore(d, archiveArc.Ctime.Time())
				scoreMap[archiveID] = int64(score)
			} else {
				// if no data in data warehouse, use new ArchiveStats to calculate
				newArchiveStat := new(model.ArchiveStats)
				newArchiveStat.Rebuild(statD)
				score := calculateScore(newArchiveStat, archiveArc.Ctime.Time())
				scoreMap[archiveID] = int64(score)
			}

			succeedC <- 1
		} else {
			failedC <- 1
		}
	}

	return scoreMap
}

func recordArchiveScoreIntoLocalFile(filePath string, scoreMap map[int64]int64, archiveArcM map[int64]*arcclient.Arc) {
	f, err := os.Create(filePath)
	if err != nil {
		return
	}

	defer func() {
		_ = f.Close()
	}()

	w := csv.NewWriter(f)
	_ = w.Write([]string{"archive_aid", "archive_title", "archive_score"})
	for k, v := range scoreMap {
		title := ""
		if d, ok := archiveArcM[k]; ok {
			title = d.Title
		}

		_ = w.Write([]string{
			fmt.Sprintf("%v", k),
			fmt.Sprintf("%v", title),
			fmt.Sprintf("%v", v)})
	}
	w.Flush()
}

func (s *Service) fetchArchivesInfoAndStats(ctx context.Context, archiveAIDs []int64) (map[int64]*arcclient.Arc, map[int64]*arcclient.Stat) {
	var wg sync.WaitGroup

	archiveArcM := make(map[int64]*arcclient.Arc, 0)
	archiveStatM := make(map[int64]*arcclient.Stat, 0)

	wg.Add(2)
	go func() {
		defer func() {
			wg.Done()
		}()

		archivesLen := len(archiveAIDs)
		for i := 0; i < archivesLen; i = i + maxLimit4ArchiveArcsQuery {
			endIndex := calculateSliceEndIndex(archivesLen, i, maxLimit4ArchiveArcsQuery)
			if archivesRes, err := component.ArcClient.Arcs(
				ctx,
				&arcclient.ArcsRequest{
					Aids: archiveAIDs[i:endIndex],
				}); err == nil && archivesRes != nil {
				for k, v := range archivesRes.Arcs {
					archiveArcM[k] = v
				}
			}
		}
	}()

	go func() {
		defer func() {
			wg.Done()
		}()

		archivesLen := len(archiveAIDs)
		for i := 0; i < archivesLen; i = i + maxLimit4ArchiveStatsQuery {
			endIndex := calculateSliceEndIndex(archivesLen, i, maxLimit4ArchiveStatsQuery)
			if archivesRes, err := component.ArcClient.Stats(
				ctx,
				&arcclient.StatsRequest{
					Aids: archiveAIDs[i:endIndex],
				}); err == nil && archivesRes != nil {
				for k, v := range archivesRes.Stats {
					archiveStatM[k] = v
				}
			}
		}
	}()

	wg.Wait()

	return archiveArcM, archiveStatM
}

func calculateSliceEndIndex(totalLen, startIndex, needLen int) int {
	endIndex := startIndex + needLen - 1
	if d := startIndex + needLen - totalLen; d > 0 {
		endIndex = totalLen - 1
	}

	return endIndex
}

// Generate berserker query sql
func genBerserkerQuery(archiveAIDs []int64, now time.Time) string {
	timeBefore14Day := now.Add(-time.Hour * 24 * 13).Format("20060102")
	archiveAIDsStr := tool.Int64JoinStr(archiveAIDs, tool.DelimiterOfComma)

	return fmt.Sprintf(berserkerSql4ArchiveStats, timeBefore14Day, archiveAIDsStr)
}

func (s *Service) updateArchiveScoreByStatsFile(localFile string) (m map[int64]*model.ArchiveStats, err error) {
	m = make(map[int64]*model.ArchiveStats, 0)

	if !file.Exist(localFile) {
		err = fmt.Errorf("ArchiveScore: archive stats file(%v) is not existed", localFile)

		return
	}

	f, fileErr := os.Open(localFile)
	if fileErr != nil {
		err = fmt.Errorf("ArchiveScore: open archive stats file(%v) failed, err: %v", localFile, fileErr)

		return
	}

	defer func() {
		_ = f.Close()
	}()

	reader := bufio.NewReader(f)
	for {
		bs, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		raw := strings.Split(string(bs), "\u0001")
		stats, err := genArchiveStats(raw)
		if err != nil {
			continue
		}

		m[stats.AID] = stats
	}

	return
}

func genArchiveStats(raw []string) (stats *model.ArchiveStats, err error) {
	stats = new(model.ArchiveStats)

	if len(raw) == 8 {
		for k, v := range raw {
			bitSize := 32
			if k == 0 {
				bitSize = 64
			}

			vAfterParse, parseErr := strconv.ParseInt(v, 10, bitSize)
			if parseErr != nil {
				err = parseErr

				return
			}

			switch k {
			case 0:
				stats.AID = vAfterParse
			case 1:
				stats.ViewBefore14 = int32(vAfterParse)
			case 2:
				stats.ReplyBefore14 = int32(vAfterParse)
			case 3:
				stats.FavoriteBefore14 = int32(vAfterParse)
			case 4:
				stats.CoinBefore14 = int32(vAfterParse)
			case 5:
				stats.DanmakuBefore14 = int32(vAfterParse)
			case 6:
				stats.ShareBefore14 = int32(vAfterParse)
			case 7:
				stats.LikeBefore14 = int32(vAfterParse)
			}
		}

		return
	}

	err = errors.New("berserker: archive stats record is invalid")

	return
}

// calculate score biz
//  1. 14天前分数 =
//     (1-LN(MIN(IF((今天-视频投稿日期)<=1,1.1,(今天-视频投稿日期)),30))/LN(30))* 14天前 (硬币*0.4+收藏*0.3+弹幕*0.4+评论*0.4+播放*0.25+点赞*0.4+分享*0.6)
//  2. 最近14天内分数=
//     (LN(MIN(IF((今天-视频投稿日期)<=1,1.1,(今天-视频投稿日期)),30))/LN(30)) * 14天内 (硬币*0.4+收藏*0.3+弹幕*0.4+评论*0.4+播放*0.25+点赞*0.4+分享*0.6)
//  3. 最终得分 = 14天前分数+最近14天内分数
func calculateScore(stats *model.ArchiveStats, createTime time.Time) float64 {
	var scoreBefore14, scoreIn14 float64

	dateSubOfFloat := 1.1
	dateSub := tool.CalculateDateSub(time.Now(), createTime)
	if dateSub > 1 {
		dateSubOfFloat = float64(dateSub)
	}

	if dateSubOfFloat > 30 {
		dateSubOfFloat = 30
	}

	calculateRate := tool.LNX(dateSubOfFloat) / tool.LNX(30)
	scoreBefore14 = (1 - calculateRate) * stats.CalculateScore(model.ArchiveStatsTypeOfBefore14)
	scoreIn14 = calculateRate * stats.CalculateScore(model.ArchiveStatsTypeOfIn14)

	return scoreBefore14 + scoreIn14
}

func (s *Service) fetchArchives(ctx context.Context, archiveID int64) ([]*model.Arc, int64, error) {
	var (
		archives = make([]*model.Arc, 0)
		err      error
	)

	archives, err = s.dao.Arcs(ctx, archiveID, maxLimit4DBQuery)
	if err != nil {
		log.Error("ArcScore s.dao.Arcs ID(%d) Limit(%d) error(%v)", archiveID, maxLimit4DBQuery, err)

		return archives, archiveID, err
	}

	if len(archives) == 0 {
		return archives, 0, nil
	}

	return archives, archives[len(archives)-1].ID, nil
}

func (s *Service) fetchAutoArchives(ctx context.Context, archiveID, checkTp int64) ([]*model.Arc, int64, error) {
	archives, err := s.dao.ArcsAuto(ctx, archiveID, checkTp, maxLimit4DBQuery)
	if err != nil {
		log.Error("fetchAutoArchives s.dao.Arcs ID(%d) Limit(%d) error(%v)", archiveID, maxLimit4DBQuery, err)

		return archives, archiveID, err
	}
	if len(archives) == 0 {
		return archives, 0, nil
	}
	return archives, archives[len(archives)-1].ID, nil
}

// arcConsumeproc consumer archive
func (s *Service) arcConsumeproc(ctx context.Context) (err error) {
	defer s.waiter.Done()
	var (
		msg *databus.Message
		ok  bool
	)
	if s.archiveNotifySub == nil {
		return
	}
	msgs := s.archiveNotifySub.Messages()
	for {
		if msg, ok = <-msgs; !ok {
			log.Infoc(ctx, "arcConsumeproc databus Consumer exit")
			break
		}
		var ms = &model.ArcMsg{}
		if err = json.Unmarshal(msg.Value, ms); err != nil {
			msg.Commit()
			log.Errorc(ctx, "arcConsumeproc json.Unmarshal(%s) error(%v)", msg.Value, err)
			continue
		}
		switch ms.Table {
		case _tableArchive:
			s.archiveInsert(ms.Action, ms.New)
		}
		msg.Commit()
	}
	return
}

func (s *Service) Auto(msg string) (err error) {
	var ms = &model.ArcMsg{}
	if err = json.Unmarshal([]byte(msg), ms); err != nil {
		log.Error("Auto json.Unmarshal(%s) error(%v)", msg, err)
		return
	}
	switch ms.Table {
	case _tableArchive:
		s.archiveInsert(_insertAct, ms.New)
	}
	return
}

func (s *Service) NewAutoCheckPass() {
	var (
		archiveID            = int64(0)
		archives             []*model.Arc
		ctx                  = context.Background()
		fetchArchiveErr      error
		fetchArchiveErrCount int
	)
	for {
		archives, archiveID, fetchArchiveErr = s.fetchAutoArchives(ctx, archiveID, _autoCheck)
		if fetchArchiveErr != nil {
			fetchArchiveErrCount++
			log.Errorc(ctx, "NewAutoCheckPass s.fetchArchives(%d) error(%+v)", archiveID, fetchArchiveErr)
		} else {
			var archiveAIDs []int64
			for _, arc := range archives {
				archiveAIDs = append(archiveAIDs, arc.Aid)
			}
			archivesLen := len(archiveAIDs)
			for i := 0; i < archivesLen; i = i + maxLimit4ArchiveArcsQuery {
				endIndex := calculateSliceEndIndex(archivesLen, i, maxLimit4ArchiveArcsQuery)
				if archivesRes, err := component.ArcClient.Arcs(
					ctx,
					&arcclient.ArcsRequest{
						Aids: archiveAIDs[i:endIndex],
					}); err == nil && archivesRes != nil {
					for _, arcInfo := range archivesRes.Arcs {
						if arcInfo == nil {
							log.Errorc(ctx, "NewAutoCheckPass arcInfo nil")
							continue
						}
						if arcInfo.Stat.View < s.c.Rule.AutoPassView {
							log.Infoc(ctx, "NewAutoCheckPass arcInfo aid(%d) view(%d)", arcInfo.Aid, arcInfo.Stat.View)
							continue
						}
						if err := s.dao.AutoArcPass(ctx, arcInfo.Aid); err != nil {
							log.Errorc(ctx, "NewAutoCheckPass s.dao.AutoArcPass aid(%d) error(%+v)", arcInfo.Aid, err)
						}
					}
				}
			}
			if isArchivesQueryDone(archives, archiveID) {
				log.Infoc(ctx, "NewAutoCheckPass isArchivesQueryDone success")
				break
			}
		}
		if fetchArchiveErrCount >= 3 {
			log.Infoc(ctx, "NewAutoCheckPass fetchArchiveErrCount(%d)", fetchArchiveErrCount)
			break
		}
	}
}

func (s *Service) NewAutoCheck() (err error) {
	if !s.autoTagRun.LockWithCheck() {
		err = fmt.Errorf("执行速度过快")
	}
	return
}

func (s *Service) NewAutoTagHistoryArc() {
	var (
		archiveID            = int64(0)
		archives             []*model.Arc
		ctx                  = context.Background()
		fetchArchiveErr      error
		fetchArchiveErrCount int
	)
	defer func() {
		s.autoTagRun.Release()
	}()
	for {
		archives, archiveID, fetchArchiveErr = s.fetchAutoArchives(ctx, archiveID, _autoCheckPass)
		if fetchArchiveErr != nil {
			fetchArchiveErrCount++
			log.Errorc(ctx, "NewAutoHistoryArc s.fetchArchives(%d) error(%+v)", archiveID, fetchArchiveErr)
		} else {
			var archiveAIDs []int64
			for _, arc := range archives {
				archiveAIDs = append(archiveAIDs, arc.Aid)
			}
			archivesLen := len(archiveAIDs)
			for i := 0; i < archivesLen; i = i + maxLimit4ArchiveArcsQuery {
				endIndex := calculateSliceEndIndex(archivesLen, i, maxLimit4ArchiveArcsQuery)
				if archivesRes, err := component.ArcClient.Arcs(
					ctx,
					&arcclient.ArcsRequest{
						Aids: archiveAIDs[i:endIndex],
					}); err == nil && archivesRes != nil {
					for _, arcInfo := range archivesRes.Arcs {
						if arcInfo == nil {
							log.Errorc(ctx, "NewAutoTagHistoryArc arcInfo nil")
							continue
						}
						archive := &model.Archive{
							Aid:     arcInfo.Aid,
							Mid:     arcInfo.Author.Mid,
							TypeID:  int16(arcInfo.TypeID),
							Title:   arcInfo.Title,
							PubTime: arcInfo.PubDate.Time().Format("2006-01-02"),
						}
						if err := s.archiveAutoTag(_insertAct, archive, false); err != nil {
							log.Errorc(ctx, "NewAutoTagHistoryArc s.archiveAutoTag aid(%d) archive(%+v) error(%+v)", arcInfo.Aid, archive, err)
						}
					}
				}
			}
			if isArchivesQueryDone(archives, archiveID) {
				break
			}
		}
		if fetchArchiveErrCount >= 3 {
			break
		}
	}
}

func (s *Service) NewAutoOneArc(aid int64) (err error) {
	ctx := context.Background()
	archivesRes, e := component.ArcClient.Arc(
		ctx,
		&arcclient.ArcRequest{
			Aid: aid,
		})
	if e != nil || archivesRes == nil {
		log.Errorc(ctx, "NewAutoOneArc s.arcClient.Arc aid(%d) error(%+v)", aid, e)
		return
	}
	arcInfo := archivesRes.Arc
	archive := &model.Archive{
		Aid:     arcInfo.Aid,
		Mid:     arcInfo.Author.Mid,
		TypeID:  int16(arcInfo.TypeID),
		Title:   arcInfo.Title,
		PubTime: arcInfo.PubDate.Time().Format("2006-01-02"),
	}
	if err = s.archiveAutoTag(_insertAct, archive, false); err != nil {
		log.Errorc(ctx, "NewAutoOneArc s.archiveAutoTag aid(%d) archive(%+v) error(%+v)", arcInfo.Aid, archive, err)
		return
	}
	log.Infoc(ctx, "NewAutoOneArc success aid(%d)", aid)
	return
}

func (s *Service) NewAutoOnePass(aid int64) (err error) {
	ctx := context.Background()
	archivesRes, e := component.ArcClient.Arc(
		ctx,
		&arcclient.ArcRequest{
			Aid: aid,
		})
	if e != nil || archivesRes == nil {
		log.Errorc(ctx, "NewAutoOnePass s.arcClient.Arc aid(%d) error(%+v)", aid, e)
		return
	}
	arcInfo := archivesRes.Arc
	if arcInfo.Stat.View < s.c.Rule.AutoPassView {
		log.Infoc(ctx, "NewAutoCheckPass arcInfo aid(%d) view(%d)", arcInfo.Aid, arcInfo.Stat.View)
		return
	}
	if err = s.dao.AutoArcPass(ctx, arcInfo.Aid); err != nil {
		log.Errorc(ctx, "NewAutoOnePass s.dao.AutoArcPass aid(%d) error(%+v)", arcInfo.Aid, err)
		return
	}
	log.Infoc(ctx, "NewAutoOnePass success aid(%d)", aid)
	return
}

func (s *Service) archiveInsert(action string, newMsg []byte) {
	newArc := &model.Archive{}
	if err := json.Unmarshal(newMsg, newArc); err != nil {
		log.Error("archiveInsert json.Unmarshal(%s) error(%v)", newMsg, err)
		return
	}
	s.archiveAutoTag(action, newArc, true)
}

func (s *Service) archiveAutoTag(action string, newArc *model.Archive, isSub bool) (err error) {
	var (
		ctx                                       = context.Background()
		tagRs                                     []*tagmdl.Tag
		mid, haveID, intYear, officialTid         int64
		tags, keywords                            []string
		gameIDs, matchIDs, teamIDs                []int64
		tagMatchIDs, whiteMatchIDs, titleMatchIDs []int64
		tagGameIDs, whiteGameIDs, titleGameIDs    []int64
		gameTeams                                 []*model.Team
		BVID                                      string
		tagsMap                                   map[string]struct{}
	)
	if _, ok := s.gameTypeMap[int32(newArc.TypeID)]; !ok {
		return
	}
	switch action {
	case _insertAct:
		if isSub {
			if haveID, err = s.dao.AutoArc(ctx, newArc.Aid); err != nil {
				log.Error("archiveInsert s.dao.AutoArc aid(%d) error(%+v)", newArc.Aid, err)
				return
			}
			if haveID > 0 {
				log.Warn("archiveInsert s.dao.AutoArc aid(%d) archive table id(%d)", newArc.Aid, haveID)
				return
			}
		}
		log.Info("archiveInsert databus message aid(%d)", newArc.Aid)
		if white, ok := s.autoRules.Mids[newArc.Mid]; ok {
			mid = newArc.Mid
			log.Infoc(ctx, "archiveInsert 823 aid(%d) gameIDS(%+v) matchIDS(%+v) mid(%+v)", newArc.Aid, white.GameIDs, white.MatchIDs, newArc.Mid)
			if gs, e := xstr.SplitInts(white.GameIDs); e != nil {
				log.Error("archiveInsert white xstr.SplitInts gameids(%s) aid(%d) error(%v)", white.GameIDs, newArc.Aid, e)
			} else {
				gameIDs = append(gameIDs, gs...)
				whiteGameIDs = append(whiteGameIDs, gs...)
			}
			if ms, e := xstr.SplitInts(white.MatchIDs); e != nil {
				log.Error("archiveInsert white xstr.SplitInts matchids(%s) aid(%d) error(%v)", white.MatchIDs, newArc.Aid, e)
			} else {
				matchIDs = append(matchIDs, ms...)
				//whiteMatchIDs = append(whiteMatchIDs, ms...)
			}
		}
		reply, errG := component.TagClient.ArcTags(ctx, &tagmdl.ArcTagsReq{Aid: newArc.Aid})

		if errG != nil || reply == nil {
			err = errG
			log.Error("archiveInsert s.tag.ArcTags aid(%d) error(%v)", newArc.Aid, err)
		} else {
			tagRs = reply.Tags
			for _, t := range tagRs {
				if tag, ok := s.autoRules.Tags[strings.ToLower(t.Name)]; ok {
					tags = append(tags, strconv.FormatInt(tag.ID, 10))
					log.Infoc(ctx, "archiveInsert 843 aid(%d) gameIDS(%+v) matchIDS(%+v) mid(%+v)", newArc.Aid, tag.GameIDs, tag.MatchIDs, newArc.Mid)
					if gs, e := xstr.SplitInts(tag.GameIDs); e != nil {
						log.Error("archiveInsert tag  xstr.SplitInts gameids(%s) aid(%d) error(%v)", tag.GameIDs, newArc.Aid, e)
					} else {
						gameIDs = append(gameIDs, gs...)
						tagGameIDs = append(tagGameIDs, gs...)
					}
					if ms, e := xstr.SplitInts(tag.MatchIDs); e != nil {
						log.Error("archiveInsert tag xstr.SplitInts matchids(%s) aid(%d) error(%v)", tag.MatchIDs, newArc.Aid, e)
					} else {
						matchIDs = append(matchIDs, ms...)
						tagMatchIDs = append(tagMatchIDs, ms...)
					}
				}
			}
			if len(tagGameIDs) > 0 { // 有多个游戏时，再用tag匹配所有游戏，取交集
				var tmpGameIDs []int64
				for _, t := range tagRs {
					if gameInfo, ok := s.autoGames[strings.ToLower(t.Name)]; ok {
						tmpGameIDs = append(tmpGameIDs, gameInfo.ID)
					}
				}
				tagGameIDs = intersect(tagGameIDs, tmpGameIDs)
			}
		}
		for name, keyword := range s.autoRules.Keywords {
			if strings.Contains(strings.ToLower(newArc.Title), strings.ToLower(name)) {
				keywords = append(keywords, strconv.FormatInt(keyword.ID, 10))
				log.Infoc(ctx, "archiveInsert 871 aid(%d) gameIDS(%+v) matchIDS(%+v) mid(%+v)", newArc.Aid, keyword.GameIDs, keyword.MatchIDs, newArc.Mid)
				if gs, e := xstr.SplitInts(keyword.GameIDs); e != nil {
					log.Error("archiveInsert keyword  xstr.SplitInts gameids(%s) aid(%d) error(%v)", keyword.GameIDs, newArc.Aid, e)
				} else {
					gameIDs = append(gameIDs, gs...)
					titleGameIDs = append(titleGameIDs, gs...)
				}
				if ms, e := xstr.SplitInts(keyword.MatchIDs); e != nil {
					log.Error("archiveInsert tag xstr.SplitInts matchids(%s) aid(%d) error(%v)", keyword.MatchIDs, newArc.Aid, e)
				} else {
					matchIDs = append(matchIDs, ms...)
					titleMatchIDs = append(titleMatchIDs, ms...)
				}
			}
			if len(titleGameIDs) > 0 { // 有多个游戏时，再用标题匹配所有游戏，取交集
				var tmpGameIDs []int64
				for _, gameInfo := range s.autoGames {
					if strings.Contains(strings.ToLower(newArc.Title), strings.ToLower(gameInfo.Title)) {
						tmpGameIDs = append(tmpGameIDs, gameInfo.ID)
					}
				}
				titleGameIDs = intersect(titleGameIDs, tmpGameIDs)
			}
		}
		if mid == 0 && len(tags) == 0 && len(keywords) == 0 {
			return
		}
		log.Info("archiveInsert aid(%d) have rule mid(%d) tags(%+v) keywords(%+v)", newArc.Aid, mid, tags, keywords)
		strTag := strings.Join(tags, ",")
		strKeyword := strings.Join(keywords, ",")
		// 添加年份标签
		if len(newArc.PubTime) >= 4 {
			pubYear := newArc.PubTime[0:4]
			if intYear, err = strconv.ParseInt(pubYear, 10, 64); err != nil {
				log.Errorc(ctx, "archiveInsert pubYear(%s) pubtime(%s) aid(%d) error(%+v)", pubYear, newArc.PubTime, newArc.Aid, err)
			}
		}
		matchIDs = unique(matchIDs)
		gameIDs = unique(gameIDs)
		isOld := s.c.Rule.AutoFileSwitch == 1
		if isOld {
			if err = s.dao.AutoAdd(ctx, newArc.Aid, mid, officialTid, strTag, strKeyword, gameIDs, matchIDs, teamIDs, intYear, _autoCheckPass); err != nil {
				log.Error("archiveInsert s.dao.AutoAddArc aid(%d) error(%+v)", newArc.Aid, err)
				return
			}
		}
		log.Infoc(ctx, "archiveInsert 913 s.autoTeamsByGame aid(%d) tagMatchIDs(%+v) whiteMatchIDs(%+v) titleMatchIDs(%+v)", newArc.Aid, tagMatchIDs, whiteMatchIDs, titleMatchIDs)
		// 添加赛事标签
		matchCount := len(matchIDs)
		if matchCount > 1 {
			matchIDs = intersectMatch(tagMatchIDs, whiteMatchIDs, titleMatchIDs)
		}
		log.Infoc(ctx, "archiveInsert 918 s.autoTeamsByGame aid(%d) tagGameIDs(%+v) whiteGameIDs(%+v) titleGameIDs(%+v)", newArc.Aid, tagGameIDs, whiteGameIDs, titleGameIDs)
		// 添加游戏标签
		gameCount := len(gameIDs)
		if gameCount > 1 {
			gameIDs = intersectGame(tagGameIDs, whiteGameIDs, titleGameIDs)
		}
		log.Infoc(ctx, "archiveInsert 923 s.autoTeamsByGame gameID(%+v) aid(%d)", gameIDs, newArc.Aid)
		if len(gameIDs) == 1 {
			// 设置队伍(该游戏下的队伍)
			if gameTeams, err = s.autoTeamsByGame(ctx, gameIDs[0]); err != nil {
				log.Errorc(ctx, "archiveInsert s.autoTeamsByGame gameID(%d) aid(%d) error(%+v)", gameIDs[0], newArc.Aid, err)
			} else {
				tagsMap = make(map[string]struct{}, len(tags))
				for _, t := range tagRs {
					tagsMap[strings.ToLower(t.Name)] = struct{}{}
				}
				for _, team := range gameTeams {
					log.Infoc(ctx, "archiveInsert 940 s.autoTeamsByGame gameID(%+v) aid(%d) teamTitle(%s) gameTeams count(%d)", gameIDs[0], newArc.Aid, team.Title, len(gameTeams))
					if _, ok := tagsMap[strings.ToLower(team.Title)]; ok {
						log.Infoc(ctx, "archiveInsert 942 s.autoTeamsByGame gameID(%+v) aid(%d) team.Title(%s)", gameIDs[0], newArc.Aid, team.Title)
						teamIDs = append(teamIDs, team.ID)
					}
				}
			}
		} else {
			// 设置游戏为其他
			for _, gameInfo := range s.autoGames {
				if gameInfo.Title == _otherTitle {
					gameIDs = []int64{gameInfo.ID}
					break
				}
			}
			// 设置队伍为其他
			for _, teamInfo := range s.autoTeams {
				if teamInfo.Title == _otherTitle {
					teamIDs = []int64{teamInfo.ID}
					break
				}
			}
		}
		// 设置官方视频tag
		if mid > 0 {
			officialTid = s.c.Rule.AutoOfficialTid
		}
		// 写入文件
		matchIDs = unique(matchIDs)
		gameIDs = unique(gameIDs)
		teamIDs = unique(teamIDs)
		if isOld {
			filename := fmt.Sprintf(archiveAutoFilename, time.Now().Format("20060102"))
			filePath := fmt.Sprintf(archiveAutoFilePath, filename)
			BVID, err = bvid.AvToBv(newArc.Aid)
			if err != nil {
				log.Errorc(ctx, "archiveInsert bvid.AvToBv aid(%d) error(%+v)", newArc.Aid, err)
				return
			}
			log.Info("archiveInsert aid(%d) new rule gameIDs(%+v) matchIDs(%+v) teamIDs(%+v)", newArc.Aid, gameIDs, matchIDs, teamIDs)
			s.NewRuleArchiveLabelIntoLocalFile(filePath, BVID, gameIDs, matchIDs, teamIDs, intYear, mid)
		} else {
			if isSub {
				if err = s.dao.AutoAdd(ctx, newArc.Aid, mid, officialTid, strTag, strKeyword, gameIDs, matchIDs, teamIDs, intYear, _autoCheck); err != nil {
					log.Errorc(ctx, "archiveInsert s.dao.AutoAddArc aid(%d) error(%+v)", newArc.Aid, err)
				}
			} else { // 历史老数据修复
				if haveID, err = s.dao.AutoArc(ctx, newArc.Aid); err != nil {
					log.Error("archiveInsert history s.dao.AutoArc aid(%d) error(%+v)", newArc.Aid, err)
					return
				}
				if err = s.dao.AutoUpdate(ctx, newArc.Aid, mid, officialTid, strTag, strKeyword, gameIDs, matchIDs, teamIDs, intYear, _autoCheckPass, haveID); err != nil {
					log.Errorc(ctx, "archiveInsert s.dao.AutoUpdate aid(%d) error(%+v)", newArc.Aid, err)
				}
			}
		}
	}
	return
}

func (s *Service) NewRuleArchiveLabelIntoLocalFile(filePath, BVID string, gameIDs, matchIDs, teamIDs []int64, intYear, mid int64) {
	var (
		games, matchs, teams []string
		w                    *csv.Writer
		f                    *os.File
		err                  error
		fileNotExists        bool
	)
	if _, err = os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			fileNotExists = true
		}
	}
	if f, err = os.OpenFile(filePath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644); err != nil {
		return
	}
	w = csv.NewWriter(f)
	if fileNotExists {
		_ = w.Write([]string{"BVID", "游戏", "赛事", "战队", "年份", "白名单Mid"})
	}
	defer func() {
		_ = f.Close()
	}()
	f.Seek(0, io.SeekEnd)
	for _, gameID := range gameIDs {
		for _, gameInfo := range s.autoGames {
			if gameID == gameInfo.ID {
				games = append(games, gameInfo.Title)
				break
			}
		}
	}
	for _, matchID := range matchIDs {
		if match, ok := s.autoMatchs[matchID]; ok {
			matchs = append(matchs, match.Title)
		}
	}
	for _, teamID := range teamIDs {
		for _, teamInfo := range s.autoTeams {
			if teamID == teamInfo.ID {
				teams = append(teams, teamInfo.Title)
				break
			}
		}
	}
	_ = w.Write([]string{
		fmt.Sprintf("%v", BVID),
		fmt.Sprintf("%v", strings.Join(games, ",")),
		fmt.Sprintf("%v", strings.Join(matchs, ",")),
		fmt.Sprintf("%v", strings.Join(teams, ",")),
		fmt.Sprintf("%v", intYear),
		fmt.Sprintf("%v", mid)})
	w.Flush()
}

func intersectMatch(tagMatchIDs, whiteMatchIDs, titleMatchIDs []int64) (res []int64) {
	var tagMatchCount, whiteMatchCount, titleMatchCount int
	tagMatchCount = len(tagMatchIDs)
	whiteMatchCount = len(whiteMatchIDs)
	titleMatchCount = len(titleMatchIDs)
	if tagMatchCount > 0 && whiteMatchCount > 0 && titleMatchCount > 0 {
		res = intersect(tagMatchIDs, whiteMatchIDs)
		if len(res) == 0 {
			res = []int64{}
			return
		}
		res = intersect(res, titleMatchIDs)
	} else if tagMatchCount > 0 && whiteMatchCount > 0 {
		res = intersect(tagMatchIDs, whiteMatchIDs)
	} else if tagMatchCount > 0 && titleMatchCount > 0 {
		res = intersect(tagMatchIDs, titleMatchIDs)
	} else if whiteMatchCount > 0 && titleMatchCount > 0 {
		res = intersect(whiteMatchIDs, titleMatchIDs)
	} else { // 命中一种的情况
		res = append(res, tagMatchIDs...)
		res = append(res, whiteMatchIDs...)
		res = append(res, titleMatchIDs...)
	}
	return
}

func intersectGame(tagGameIDs, whiteGameIDs, titleGameIDs []int64) (res []int64) {
	var tagGameCount, whiteGameCount, titleGameCount int
	tagGameCount = len(tagGameIDs)
	whiteGameCount = len(whiteGameIDs)
	titleGameCount = len(titleGameIDs)
	if tagGameCount > 0 && whiteGameCount > 0 && titleGameCount > 0 {
		res = intersect(tagGameIDs, whiteGameIDs)
		if len(res) == 0 {
			res = []int64{}
			return
		}
		res = intersect(res, titleGameIDs)
	} else if tagGameCount > 0 && whiteGameCount > 0 {
		res = intersect(tagGameIDs, whiteGameIDs)
	} else if tagGameCount > 0 && titleGameCount > 0 {
		res = intersect(tagGameIDs, titleGameIDs)
	} else if whiteGameCount > 0 && titleGameCount > 0 {
		res = intersect(whiteGameIDs, titleGameIDs)
	} else { // 命中一种的情况
		res = append(res, tagGameIDs...)
		res = append(res, whiteGameIDs...)
		res = append(res, titleGameIDs...)
	}
	return
}

func intersect(slice1, slice2 []int64) []int64 {
	m := make(map[int64]struct{})
	nn := make([]int64, 0)
	if len(slice1) == 0 {
		return slice2
	}
	if len(slice2) == 0 {
		return slice1
	}
	for _, v := range slice1 {
		m[v] = struct{}{}
	}
	for _, v := range slice2 {
		if _, ok := m[v]; ok {
			nn = append(nn, v)
		}
	}
	return nn
}

func (s *Service) autoRule() {
	var (
		mids           map[int64]*model.RuleRs
		tags, keywords map[string]*model.RuleRs
		c              = context.Background()
		err            error
	)
	group := errGroup.WithContext(c)
	group.Go(func(ctx context.Context) error {
		if mids, err = s.dao.RuleWhite(ctx); err != nil {
			log.Error("autoRule s.dao.RuleWhite error(%+v)", err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if tags, err = s.dao.RuleTag(ctx); err != nil {
			log.Error("autoRule s.dao.RuleTag error(%+v)", err)
			return err
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		if keywords, err = s.dao.RuleKeyword(ctx); err != nil {
			log.Error("autoRule s.dao.RuleKeyword error(%+v)", err)
			return err
		}
		return nil
	})
	err = group.Wait()
	s.autoRules = &model.AutoRule{
		Mids:     mids,
		Tags:     tags,
		Keywords: keywords,
	}
}

func (s *Service) autoGame() {
	ctx := context.Background()
	games, err := s.dao.AllGames(ctx)
	if err != nil {
		log.Errorc(ctx, "autoGame s.dao.AllGames error(%+v)", err)
		return
	}
	s.autoGames = games
}

func (s *Service) autoMatch() {
	ctx := context.Background()
	matchs, err := s.dao.AllMatchs(ctx)
	if err != nil {
		log.Errorc(ctx, "autoMatch s.dao.AllGames error(%+v)", err)
		return
	}
	s.autoMatchs = matchs
}

func (s *Service) autoTeam() {
	ctx := context.Background()
	teams, err := s.dao.AllTeams(ctx)
	if err != nil {
		log.Errorc(ctx, "autoTeam s.dao.AllTeams error(%+v)", err)
		return
	}
	s.autoTeams = teams
}

func (s *Service) autoTeamsByGame(ctx context.Context, gameID int64) (res []*model.Team, err error) {
	var teamIDs []int64
	if teamIDs, err = s.dao.TeamsByGame(ctx, gameID); err != nil {
		log.Errorc(ctx, "autoTeamsByGame s.dao.TeamsByGame gid(%d) error(%d)", gameID, err)
		return
	}
	log.Infoc(ctx, "archiveInsert 1191 teamIDs count(%+v),s.autoTeams count(%+v)", len(teamIDs), len(s.autoTeams))
	for _, teamID := range teamIDs {
		if team, ok := s.autoTeams[teamID]; ok {
			res = append(res, team)
		}
	}
	log.Infoc(ctx, "archiveInsert 1206 res count(%+v)", len(res))
	return
}

func unique(ids []int64) (outs []int64) {
	idMap := make(map[int64]int64, len(ids))
	for _, v := range ids {
		if _, ok := idMap[v]; ok {
			continue
		} else {
			idMap[v] = v
		}
		outs = append(outs, v)
	}
	return
}

func (s *Service) gameTypeID() {
	var (
		err      error
		typesRes *arcclient.TypesReply
		game     = make(map[int32]int32)
	)
	if typesRes, err = component.ArcClient.Types(context.Background(), &arcclient.NoArgRequest{}); err != nil {
		log.Error("[regionTypeID] s.arcClient.Types error(%v)", err)
		return
	}
	if typesRes == nil || len(typesRes.Types) == 0 {
		log.Error("[regionTypeID] s.arcClient.Types return nil")
		return
	}
	for _, v := range typesRes.Types {
		if (v.Pid == 0 && v.ID == _gamePid) || (v.Pid == _gamePid) {
			game[v.ID] = v.ID
		}
	}
	s.gameTypeMap = game
}

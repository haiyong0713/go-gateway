package service

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/stat/prom"
	"go-gateway/app/web-svr/esports/job/model"

	errGroup "go-common/library/sync/errgroup.v2"
)

const (
	_codeSuccess          = "200"
	_living               = 0
	_tpOffLineHero        = 1
	_tpOffLineSkill       = 2
	_tpOffLineDevice      = 3
	_tpOffLineTeam        = 4
	_tpOffLinePlayer      = 5
	_tpOffLinePlayerThumb = 6
	_tpOffLineRepair      = 10
	_tpBattleInfo         = 1
	_thumbImg             = "_100x100"
)

func (s *Service) StoreLiveOffLineImageMap(m map[string]bool) {
	s.liveOffLineImageMap.Store(m)
}

func (s *Service) LoadLiveOffLineImageMap() map[string]bool {
	return s.liveOffLineImageMap.Load().(map[string]bool)
}

func (s *Service) ScoreLivePage() {
	ctx := context.Background()
	ticker := time.NewTicker(time.Second * 1)
	defer func() {
		ticker.Stop()
	}()
	for {
		select {
		case <-ticker.C:
			matchOne, err := s.matchOne(ctx)
			if err != nil || len(matchOne) == 0 {
				s.NoMatch(ctx)
				continue
			}
			liveOffLineImage := s.LoadLiveOffLineImageMap()
			for _, match := range matchOne {
				battleString, err := s.BattleListTwo(ctx, match.MatchID)
				if err != nil {
					log.Errorc(ctx, "ScoreLivePage s.BattleListTwo matchID(%s) error(%+v)", match.MatchID, err)
					continue
				}
				s.BattleInfoThree(ctx, battleString, liveOffLineImage)
			}
		}
	}
}

func (s *Service) NoMatch(ctx context.Context) {
	matchOne, e := s.dao.MatchOne(ctx)
	if e != nil || len(matchOne) == 0 {
		return
	}
	for _, match := range matchOne {
		if _, err := s.BattleListTwo(ctx, match.MatchID); err != nil {
			log.Errorc(ctx, "noMatch s.BattleListTwo matchID(%s) error(%+v)", match.MatchID, err)
			continue
		}
	}
	// 删除第一个接口正在进行的比赛
	s.dao.DelCacheMatch(ctx)
}

func (s *Service) matchOne(ctx context.Context) (res []*model.ScoreMatch, err error) {
	var (
		bs      []byte
		matchRs struct {
			Code string
			Data struct {
				List []*model.ScoreMatch
			} `json:"data"`
		}
	)
	if bs, err = s.score(&model.ParamScore{}, _scoreLiveMatchList); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &matchRs); err != nil {
		log.Error("matchOne tp(%s) json.Unmarshal rs(%s)  error(%+v)", _scoreLiveMatchList, string(bs), err)
		return
	}
	if matchRs.Code != _codeSuccess {
		err = ecode.NothingFound
		return
	}
	res = matchRs.Data.List
	s.dao.AddCacheMatch(ctx, res)
	return
}

func (s *Service) BattleListTwo(ctx context.Context, matchID string) (res string, err error) {
	var (
		bs       []byte
		battleRs struct {
			Code string
			Data *model.BattleList `json:"data"`
		}
	)
	intMatchID, err := strconv.ParseInt(matchID, 10, 64)
	if err != nil {
		log.Error("battleListTwo strconv.ParseInt matchID(%s) error(%+v)", matchID, err)
		return
	}
	if bs, err = s.score(&model.ParamScore{MatchID: intMatchID}, _scoreLiveBattleList); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &battleRs); err != nil {
		log.Error("battleListTwo tp(%s) json.Unmarshal rs(%s)  error(%+v)", _scoreLiveBattleList, string(bs), err)
		return
	}
	if battleRs.Code != _codeSuccess {
		err = ecode.NothingFound
		return
	}
	for _, battle := range battleRs.Data.List {
		if battle.Status == _living {
			// 取最后一个进行中的battleString
			res = battle.BattleString
		}
	}
	if battleRs.Data == nil {
		return
	}
	battleRs.Data.LastTime = time.Now().Unix()
	s.dao.AddCacheBattleList(ctx, matchID, battleRs.Data)
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.AddBattleList(ctx, matchID, string(bs))
	})
	return
}

func (s *Service) BattleInfoThree(ctx context.Context, battleString string, liveOffLineImage map[string]bool) {
	var (
		err      error
		bs       []byte
		battleRs struct {
			Code string
			Data *model.BattleInfo `json:"data"`
		}
	)
	if battleString == "" {
		return
	}
	if bs, err = s.score(&model.ParamScore{BattleString: battleString}, _scoreLiveBattleInfo); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &battleRs); err != nil {
		log.Error("battleInfoThree tp(%s) json.Unmarshal rs(%s)  error(%+v)", _scoreLiveBattleList, string(bs), err)
		return
	}
	if battleRs.Code != _codeSuccess {
		log.Errorc(ctx, "error(%+v) code(%s)", err, battleRs.Code)
		return
	}
	if battleRs.Data == nil {
		return
	}
	s.rebuildBattleInfo(ctx, battleRs.Data, liveOffLineImage)
	battleRs.Data.LastTime = time.Now().Unix()
	s.dao.AddCacheBattleInfo(ctx, battleString, battleRs.Data)
	s.cache.Do(ctx, func(ctx context.Context) {
		s.dao.AddBattleInfo(ctx, _tpBattleInfo, battleString, string(bs))
	})
	return
}

func (s *Service) rebuildBattleInfo(ctx context.Context, data *model.BattleInfo, liveOffLineImage map[string]bool) {
	for _, pick := range data.PickList {
		pick.HeroImage = s.replaceImg(ctx, pick.HeroImage, "PickList", liveOffLineImage)
	}
	for _, ban := range data.BanList {
		ban.HeroImage = s.replaceImg(ctx, ban.HeroImage, "BanList", liveOffLineImage)
	}
	data.Teama.TeamImageThumbA = s.replaceImg(ctx, data.Teama.TeamImageThumbA, "TeamaImageThumbA", liveOffLineImage)
	data.Teama.TeamImageThumb = s.replaceImg(ctx, data.Teama.TeamImageThumb, "TeamaImageThumb", liveOffLineImage)
	for _, teamaDragon := range data.Teama.Dragons {
		teamaDragon.DragonImage = s.replaceImg(ctx, teamaDragon.DragonImage, "TeamaDragonImage", liveOffLineImage)
	}
	for _, teamaPlayer := range data.Teama.Players {
		teamaPlayer.HeroImage = s.replaceImg(ctx, teamaPlayer.HeroImage, "TeamaHeroImage", liveOffLineImage)
		teamaPlayer.PlayerImage = s.replaceImg(ctx, teamaPlayer.PlayerImage, "TeamaPlayerImage", liveOffLineImage)
	}
	data.Teamb.TeamImageThumb = s.replaceImg(ctx, data.Teamb.TeamImageThumb, "TeambImageThumb", liveOffLineImage)
	data.Teamb.TeamImageThumbB = s.replaceImg(ctx, data.Teamb.TeamImageThumbB, "TeambImageThumbB", liveOffLineImage)
	for _, teambDragon := range data.Teamb.Dragons {
		teambDragon.DragonImage = s.replaceImg(ctx, teambDragon.DragonImage, "TeambDragonImage", liveOffLineImage)
	}
	for _, teambPlayer := range data.Teamb.Players {
		teambPlayer.HeroImage = s.replaceImg(ctx, teambPlayer.HeroImage, "TeambHeroImage", liveOffLineImage)
		teambPlayer.PlayerImage = s.replaceImg(ctx, teambPlayer.PlayerImage, "TeambPlayerImage", liveOffLineImage)
	}
}

func (s *Service) replaceImg(ctx context.Context, scoreImg, name string, liveOffLineImage map[string]bool) string {
	if scoreImg == "" {
		return ""
	}
	img := strings.Replace(scoreImg, _scoreImgage, _bfsImage, 1)
	if strings.Index(img, _scoreDomain) >= 0 { // 没有替换成功不是https或域名不对原因
		prom.BusinessErrCount.Incr("score:ReplaceLiveImageError" + name)
		log.Error("replaceImg  scoreImg(%s) error", scoreImg)
		return s.c.Score.LiveBackupImg
	}
	if _, ok := liveOffLineImage[scoreImg]; ok {
		return img
	}
	// 不存在图先返回默认图
	prom.BusinessErrCount.Incr("score:ReplaceLiveImageErrorRepair")
	log.Infoc(ctx, "BattleInfoThree replaceImg repair scoreImg(%s) img(%s)", scoreImg, img)
	// 异步处理离线包没有的图片
	s.cache.Do(ctx, func(ctx context.Context) {
		missImage := make(map[string]string, 1)
		itemID := s.imageName(scoreImg)
		log.Infoc(ctx, "BattleInfoThree replaceImg repair itemID(%s) scoreImg(%s) img(%s)", itemID, scoreImg, img)
		if itemID != "" {
			missImage[itemID] = scoreImg
			s.liveImageCh <- missImage
		}
	})

	// 不存在返回默认图
	return s.c.Score.LiveBackupImg
}

func (s *Service) imageName(imgURL string) string {
	start := strings.LastIndex(imgURL, "/")
	end := strings.LastIndex(imgURL, ".")
	if start <= 0 || end <= 0 || start > end {
		return ""
	}
	return imgURL[start+1 : end]
}

func (s *Service) setLiveMissImage(ctx context.Context) (err error) {
	for missImage := range s.liveImageCh {
		for itemID, scoreImg := range missImage {
			liveOffLineImage := s.LoadLiveOffLineImageMap()
			if _, ok := liveOffLineImage[scoreImg]; ok {
				continue
			}
			bfsImage := s.BfsProxy(ctx, scoreImg)
			if bfsImage == "" {
				log.Errorc(ctx, "BattleInfoThree replaceImg repair scoreImg(%s) bfsImage empty", scoreImg)
				continue
			}
			s.dao.AddOffLineImage(ctx, []*model.OffLineImage{{
				ItemType:   _tpOffLineRepair,
				ItemId:     itemID,
				ScoreImage: scoreImg,
				BfsImage:   bfsImage,
			}})
		}
	}
	return
}

func (s *Service) RefreshOffLineImageLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.OffLineImage(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) OffLineImage(ctx context.Context) {
	var heroCount, skillCount, deviceCount int
	group := errGroup.WithContext(ctx)
	group.Go(func(ctx context.Context) error {
		heros, heroErr := s.offLineHero()
		if heroErr != nil {
			log.Errorc(ctx, "s.offLineHero error(%+v)", heroErr)
			return heroErr
		}
		heroCount = len(heros)
		if heroErr = s.SetOffLineHero(ctx, heros); heroErr != nil {
			log.Errorc(ctx, "s.SetOffLineHero error(%+v)", heroErr)
			return heroErr
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		skills, skillErr := s.offLineSkill()
		if skillErr != nil {
			log.Errorc(ctx, "s.offLineSkill error(%+v)", skillErr)
			return skillErr
		}
		skillCount = len(skills)
		if skillErr = s.SetOffLineSkill(ctx, skills); skillErr != nil {
			log.Errorc(ctx, "s.offLineSkill error(%+v)", skillErr)
			return skillErr
		}
		return nil
	})
	group.Go(func(ctx context.Context) error {
		devices, deviceErr := s.offLineDevice()
		if deviceErr != nil {
			log.Errorc(ctx, "s.offLineDevice error(%+v)", deviceErr)
			return deviceErr
		}
		deviceCount = len(devices)
		if deviceErr = s.SetOffLineDevice(ctx, devices); deviceErr != nil {
			log.Errorc(ctx, "s.SetOffLineDevice error(%+v)", deviceErr)
			return deviceErr
		}
		return nil
	})
	group.Wait()
	log.Info("OffLineImage success image count(%d)", heroCount+skillCount+deviceCount)
	// 设置战队与队员离线图片
	s.setOffLineTeamAndPlayer(ctx)
	// 更新内存中的图片
	s.StoreOffLineImage(ctx)
}

func (s *Service) setOffLineTeamAndPlayer(ctx context.Context) {
	// 查询进行中的赛季
	goingSeasons := loadComponentGoingSeasons()
	if len(goingSeasons) == 0 {
		log.Warnc(ctx, "OffLineImage setOffLineTeamAndPlayer watchGoingSeasonContests goingSeasons empty")
		return
	}
	for _, season := range goingSeasons {
		if season.LeidaSid == 0 {
			continue
		}
		teams, teamErr := s.getOffLineTeamAndPlayer(ctx, season.LeidaSid)
		if teamErr != nil {
			log.Errorc(ctx, "OffLineImage setOffLineTeamAndPlayer s.getOffLineTeamAndPlayer() LeidaSid(%d) error(%+v)", season.LeidaSid, teamErr)
			continue
		}
		if teamErr = s.SetOffLineTeam(ctx, teams); teamErr != nil {
			log.Errorc(ctx, "OffLineImage setOffLineTeamAndPlayer s.SetOffLineTeam() LeidaSid(%d) error(%+v)", season.LeidaSid, teamErr)
			continue
		}
		log.Infoc(ctx, "OffLineImage s.SetOffLineTeam() LeidaSid(%d) team count(%d)", season.LeidaSid, len(teams))
	}
}

func (s *Service) LoadLiveImageMap() {
	// 更新内存中的图片
	s.StoreOffLineImage(context.Background())
}

func (s *Service) StoreOffLineImage(ctx context.Context) {
	liveOffLineImage, err := s.dao.OffLineImage(ctx)
	if err != nil {
		log.Errorc(ctx, "StoreOffLineImage s.dao.OffLineImage error(%+v)", err)
		return
	}
	s.StoreLiveOffLineImageMap(liveOffLineImage)
}

func (s *Service) offLineHero() (res []*model.OffLineHero, err error) {
	var (
		bs        []byte
		offLineRs struct {
			Code string
			Data struct {
				List []*model.OffLineHero
			} `json:"data"`
		}
	)
	if bs, err = s.score(&model.ParamScore{}, _scoreOfflineHero); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &offLineRs); err != nil {
		log.Error("offLineHero tp(%s) json.Unmarshal rs(%s)  error(%+v)", _scoreOfflineHero, string(bs), err)
		return
	}
	if offLineRs.Code != _codeSuccess {
		err = ecode.NothingFound
		return
	}
	res = offLineRs.Data.List
	return
}

func (s *Service) offLineSkill() (res []*model.OffLineJsZb, err error) {
	var (
		bs        []byte
		offLineRs struct {
			Code string
			Data struct {
				List []*model.OffLineJsZb
			} `json:"data"`
		}
	)
	if bs, err = s.score(&model.ParamScore{}, _scoreOfflineSkill); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &offLineRs); err != nil {
		log.Error("offLineSkill tp(%s) json.Unmarshal rs(%s)  error(%+v)", _scoreOfflineSkill, string(bs), err)
		return
	}
	if offLineRs.Code != _codeSuccess {
		err = ecode.NothingFound
		return
	}
	res = offLineRs.Data.List
	return
}

func (s *Service) offLineDevice() (res []*model.OffLineJsZb, err error) {
	var (
		bs        []byte
		offLineRs struct {
			Code string
			Data struct {
				List []*model.OffLineJsZb
			} `json:"data"`
		}
	)
	if bs, err = s.score(&model.ParamScore{}, _scoreOfflineDevice); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &offLineRs); err != nil {
		log.Error("offLineDevice tp(%s) json.Unmarshal rs(%s)  error(%+v)", _scoreOfflineDevice, string(bs), err)
		return
	}
	if offLineRs.Code != _codeSuccess {
		err = ecode.NothingFound
		return
	}
	res = offLineRs.Data.List
	return
}

func (s *Service) offLineTeam() (res []*model.OffLineTeam, err error) {
	var (
		bs        []byte
		offLineRs struct {
			Code string
			Data struct {
				List []*model.OffLineTeam
			} `json:"data"`
		}
	)
	if bs, err = s.score(&model.ParamScore{}, _scoreOfflineTeam); err != nil {
		return
	}
	if err = json.Unmarshal(bs, &offLineRs); err != nil {
		log.Error("offLineDevice tp(%s) json.Unmarshal rs(%s)  error(%+v)", _scoreOfflineTeam, string(bs), err)
		return
	}
	if offLineRs.Code != _codeSuccess {
		err = ecode.NothingFound
		return
	}
	res = offLineRs.Data.List
	return
}

func (s *Service) SetOffLineHero(ctx context.Context, data []*model.OffLineHero) (err error) {
	count := len(data)
	if count == 0 {
		return
	}
	log.Error("SetOffLineHero  offlineimage hero count(%d)", count)
	var res []*model.OffLineImage
	for _, hero := range data {
		bfsImg := s.BfsProxy(ctx, hero.Image)
		log.Warn("SetOffLineHero  scoreImg(%s) bfsImg(%s)", hero.Image, bfsImg)
		if bfsImg == "" {
			log.Errorc(ctx, "SetOffLineHero s.BfsProxy hero image empty")
			continue
		}
		heroImg := &model.OffLineImage{
			ItemType:   _tpOffLineHero,
			ItemId:     hero.HeroID,
			ItemName:   hero.Name,
			NickName:   hero.Nickname,
			ScoreImage: hero.Image,
			BfsImage:   bfsImg,
		}
		res = append(res, heroImg)
	}
	// 写入DB
	s.dao.AddOffLineImage(ctx, res)
	return
}

func (s *Service) SetOffLineSkill(ctx context.Context, data []*model.OffLineJsZb) (err error) {
	count := len(data)
	if count == 0 {
		return
	}
	log.Error("SetOffLineSkill  offlineimage skill count(%d)", count)
	var res []*model.OffLineImage
	for _, skill := range data {
		bfsImg := s.BfsProxy(ctx, skill.Image)
		log.Warn("SetOffLineSkill  scoreImg(%s) bfsImg(%s)", skill.Image, bfsImg)
		if bfsImg == "" {
			log.Errorc(ctx, "SetOffLineSkill s.BfsProxy skill image empty")
			continue
		}
		skillImg := &model.OffLineImage{
			ItemType:   _tpOffLineSkill,
			ItemId:     skill.ID,
			ItemName:   skill.Name,
			NickName:   skill.NameEn,
			ScoreImage: skill.Image,
			BfsImage:   bfsImg,
		}
		res = append(res, skillImg)
	}
	// 写入DB
	s.dao.AddOffLineImage(ctx, res)
	return
}

func (s *Service) SetOffLineDevice(ctx context.Context, data []*model.OffLineJsZb) (err error) {
	count := len(data)
	if count == 0 {
		return
	}
	log.Error("SetOffLineDevice  offlineimage device count(%d)", count)
	var res []*model.OffLineImage
	for _, device := range data {
		bfsImg := s.BfsProxy(ctx, device.Image)
		log.Warn("SetOffLineDevice  scoreImg(%s) bfsImg(%s)", device.Image, bfsImg)
		if bfsImg == "" {
			log.Errorc(ctx, "SetOffLineDevice s.BfsProxy device image empty")
			continue
		}
		deviceImg := &model.OffLineImage{
			ItemType:   _tpOffLineDevice,
			ItemId:     device.ID,
			ItemName:   device.Name,
			NickName:   device.NameEn,
			ScoreImage: device.Image,
			BfsImage:   bfsImg,
		}
		res = append(res, deviceImg)
	}
	// 写入DB
	s.dao.AddOffLineImage(ctx, res)
	return
}

func (s *Service) SetOffLineTeam(ctx context.Context, data []*model.OffLineTeam) (err error) {
	count := len(data)
	if count == 0 {
		return
	}
	log.Error("SetOffLineTeam  offlineimage team count(%d)", count)
	var (
		res []*model.OffLineImage
	)
	for _, team := range data {
		bfsImg := s.BfsProxy(ctx, team.TeamImage)
		log.Warn("SetOffLineTeam team  scoreImg(%s) bfsImg(%s)", team.TeamImage, bfsImg)
		if bfsImg == "" {
			log.Errorc(ctx, "SetOffLineTeam team s.BfsProxy team(%+v) image empty", team)
			continue
		}
		teamImg := &model.OffLineImage{
			ItemType:   _tpOffLineTeam,
			ItemId:     team.TeamID,
			ItemName:   team.TeamName,
			NickName:   team.TeamShortName,
			ScoreImage: team.TeamImage,
			BfsImage:   bfsImg,
		}
		res = append(res, teamImg)
		// 添加用户图片
		var (
			playerRes []*model.OffLineImage
		)
		for _, player := range team.PlayerList {
			bfsPlayerImg := s.BfsProxy(ctx, player.Image)
			log.Warn("SetOffLineTeam player  scoreImg(%s) bfsImg(%s)", player.Image, bfsPlayerImg)
			if bfsPlayerImg == "" {
				log.Errorc(ctx, "SetOffLineTeam player s.BfsProxy empty")
				continue
			}
			playerImg := &model.OffLineImage{
				ItemType:   _tpOffLinePlayer,
				ItemId:     player.PlayerID,
				ItemName:   player.Name,
				NickName:   player.Nickname,
				ScoreImage: player.Image,
				BfsImage:   bfsPlayerImg,
			}
			playerRes = append(playerRes, playerImg)
		}
		// player写入DB
		s.dao.AddOffLineImage(ctx, playerRes)
		// 添加用户缩略图
		var (
			thumbRes []*model.OffLineImage
		)
		for _, playerThumb := range team.PlayerList {
			bfsThumbImg := s.BfsProxy(ctx, playerThumb.ImageThumb)
			log.Warn("SetOffLineTeam playerThumb t scoreImg(%s) bfsImg(%s)", playerThumb.ImageThumb, bfsThumbImg)
			if bfsThumbImg == "" {
				log.Errorc(ctx, "SetOffLineTeam playerThumb s.BfsProxy empty")
				continue
			}
			thumbImg := &model.OffLineImage{
				ItemType:   _tpOffLinePlayerThumb,
				ItemId:     playerThumb.PlayerID + _thumbImg,
				ItemName:   playerThumb.Name,
				NickName:   playerThumb.Nickname,
				ScoreImage: playerThumb.ImageThumb,
				BfsImage:   bfsThumbImg,
			}
			thumbRes = append(thumbRes, thumbImg)
		}
		// playerThumb写入DB
		s.dao.AddOffLineImage(ctx, thumbRes)
	}
	// team写入DB
	s.dao.AddOffLineImage(ctx, res)
	return
}

func (s *Service) BattleUpImage(ctx context.Context, url string) (res string, err error) {
	return s.BfsProxy(ctx, url), nil
}

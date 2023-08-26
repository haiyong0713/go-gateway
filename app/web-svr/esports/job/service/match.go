package service

import (
	"context"
	"encoding/json"
	"go-gateway/app/web-svr/esports/job/component"
	"net/url"
	"strconv"
	"strings"
	"time"

	"go-common/library/log"
	"go-gateway/app/web-svr/esports/job/model"

	tunnelmdl "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"
)

const _platform = 3

func (s *Service) matchGameCron() {
	var (
		contestDatas []*model.Contest
		err          error
	)
	if contestDatas, err = s.dao.ContestDatas(context.Background()); err != nil {
		log.Error("matchGame  s.dao.ContestDatas error(%v)", err)
		time.Sleep(time.Second)
	}
	for _, data := range contestDatas {
		tmp := data
		log.Warn("matchGameCron success matchID(%d) contestID(%d)", tmp.MatchID, tmp.ID)
		go s.writeGames(tmp)
		time.Sleep(time.Second)
	}
	log.Info("matchGameCron success time(%d)", time.Now().Unix())
}

func (s *Service) FixMatchGame(matchID int64) {
	var (
		contestDatas []*model.Contest
		err          error
	)
	if contestDatas, err = s.dao.ContestDatas(context.Background()); err != nil {
		log.Error("matchGame  s.dao.ContestDatas error(%v)", err)
		time.Sleep(time.Second)
	}
	for _, data := range contestDatas {
		if data.MatchID != matchID {
			continue
		}
		s.writeFixGames(data)
		log.Info("FixMatchGame contestID(%d) matchID(%d) tp(%d) success time(%d)", data.ID, data.MatchID, data.DataType, time.Now().Unix())
		return
	}
}

func (s *Service) writeGames(data *model.Contest) {
	var (
		err           error
		params        url.Values
		rs            json.RawMessage
		dotaGames     []*model.DotaGame
		owGames       []*model.OwLdGame
		endTime       time.Time
		scoreLolGames struct {
			Data struct {
				List []*model.LolGame
			}
		}
	)
	params = url.Values{}
	params.Set("match_id", strconv.FormatInt(data.MatchID, 10))
	if data.Etime > 0 {
		endTime = time.Unix(data.Etime, 0).Add(time.Duration(s.c.Leidata.After.GameEnd))
		if time.Now().Unix() > endTime.Unix() {
			return
		}
	}
	if data.Stime > 0 && time.Now().Unix() < data.Stime {
		return
	}
	switch data.DataType {
	case _lolType:
		// use score.
		if rs, err = s.score(&model.ParamScore{MatchID: data.MatchID}, _scoreLolGame); err != nil || len(rs) == 0 {
			return
		}
		if err = json.Unmarshal(rs, &scoreLolGames); err != nil {
			log.Error("writeGames matchID(%d) lol json.Unmarshal rs(%+v)  error(%+v)", data.MatchID, rs, err)
			return
		}
		log.Error("writeGames matchID(%d) lol json.Unmarshal rs(%+v)  error(%+v)", data.MatchID, string(rs), err)
		s.writeLOL(data.MatchID, scoreLolGames.Data.List)
	case _dotaType:
		if rs, _, err = s.leida(params, _dotaGame); err != nil || len(rs) == 0 {
			return
		}
		if err = json.Unmarshal(rs, &dotaGames); err != nil {
			log.Error("writeGames  matchID(%d) dota json.Unmarshal rs(%+v)  error(%+v)", data.MatchID, rs, err)
			return
		}
		s.writeDota(data.MatchID, dotaGames)

	case _owType:
		if rs, _, err = s.leida(params, _owGame); err != nil || len(rs) == 0 {
			return
		}
		if err = json.Unmarshal(rs, &owGames); err != nil {
			log.Error("writeGames  matchID(%d) ow json.Unmarshal  rs(%+v)  error(%+v)", data.MatchID, rs, err)
			return
		}
		s.writeOw(data.MatchID, owGames)

	}
	log.Warn("writeGames contestID(%d) matchID(%d) tp(%d) success", data.ID, data.MatchID, data.DataType)
}

func (s *Service) writeFixGames(data *model.Contest) {
	var (
		err           error
		params        url.Values
		rs            json.RawMessage
		dotaGames     []*model.DotaGame
		owGames       []*model.OwLdGame
		scoreLolGames struct {
			Data struct {
				List []*model.LolGame
			}
		}
	)
	params = url.Values{}
	params.Set("match_id", strconv.FormatInt(data.MatchID, 10))
	switch data.DataType {
	case _lolType:
		if rs, err = s.score(&model.ParamScore{MatchID: data.MatchID}, _scoreLolGame); err != nil || len(rs) == 0 {
			return
		}
		if err = json.Unmarshal(rs, &scoreLolGames); err != nil {
			log.Error("writeGames matchID(%d) lol json.Unmarshal rs(%+v)  error(%+v)", data.MatchID, rs, err)
			return
		}
		s.writeLOL(data.MatchID, scoreLolGames.Data.List)
	case _dotaType:
		if rs, _, err = s.leida(params, _dotaGame); err != nil || len(rs) == 0 {
			return
		}
		if err = json.Unmarshal(rs, &dotaGames); err != nil {
			log.Error("writeFixGames  matchID(%d) dota json.Unmarshal rs(%+v)  error(%+v)", data.MatchID, rs, err)
			return
		}
		s.writeDota(data.MatchID, dotaGames)
	case _owType:
		if rs, _, err = s.leida(params, _owGame); err != nil || len(rs) == 0 {
			return
		}
		if err = json.Unmarshal(rs, &owGames); err != nil {
			log.Error("writeFixGames  matchID(%d) ow json.Unmarshal  rs(%+v)  error(%+v)", data.MatchID, rs, err)
			return
		}
		s.writeOw(data.MatchID, owGames)
	}
	log.Info("writeFixGames contestID(%d) matchID(%d) tp(%d) success", data.ID, data.MatchID, data.DataType)
}

func (s *Service) writeLOL(matchID int64, ldGames []*model.LolGame) {
	var (
		err      error
		ids      []*model.Oid
		idMap    map[int64]struct{}
		c        = context.Background()
		addGames []*model.LolGame
	)
	if ids, err = s.dao.LolGames(c, matchID); err != nil {
		log.Error("writeLOL s.dao.LolGames MatchID(%d) error(%+v)", matchID, err)
		return
	}
	idMap = make(map[int64]struct{}, len(ids))
	for _, p := range ids {
		idMap[p.ID] = struct{}{}
	}
	for _, game := range ldGames {
		for _, team := range game.Teams {
			team.Team.ImageURL = s.BfsProxy(c, team.Team.ImageURL)
		}
		for _, player := range game.Players {
			player.Player.ImageURL = s.BfsProxy(c, player.Player.ImageURL)
			player.Champion.ImageURL = s.BfsProxy(c, player.Champion.ImageURL)
			player.Team.ImageURL = s.BfsProxy(c, player.Team.ImageURL)
			for _, spell := range player.Spells {
				spell.ImageURL = s.BfsProxy(c, spell.ImageURL)
			}
			for _, item := range player.Items {
				item.ImageURL = s.BfsProxy(c, item.ImageURL)
			}
		}
		if _, ok := idMap[game.ID]; ok {
			if err = s.dao.UpLolGame(c, game); err != nil {
				log.Error("writeLOL s.dao.UpLOLGame MatchID(%d) gameID(%d) error(%+v)", matchID, game.ID, err)
			}
			time.Sleep(time.Millisecond * 10)
		} else {
			addGames = append(addGames, game)
		}
	}
	if len(addGames) > 0 {
		if err = s.dao.AddLolGame(c, addGames); err != nil {
			log.Error("writeLOL s.dao.AddLolGame MatchID(%d) error(%+v)", matchID, err)
		}
	}
}

func (s *Service) writeDota(matchID int64, ldGames []*model.DotaGame) {
	var (
		err      error
		ids      []*model.Oid
		idMap    map[int64]struct{}
		c        = context.Background()
		addGames []*model.DotaGame
	)
	if ids, err = s.dao.DotaGames(c, matchID); err != nil {
		log.Error("writeDota s.dao.DotaGames MatchID(%d) error(%+v)", matchID, err)
		return
	}
	idMap = make(map[int64]struct{}, len(ids))
	for _, p := range ids {
		idMap[p.ID] = struct{}{}
	}
	for _, game := range ldGames {
		for _, team := range game.Teams {
			team.Team.ImageURL = s.BfsProxy(c, team.Team.ImageURL)
		}
		for _, player := range game.Players {
			player.Player.ImageURL = s.BfsProxy(c, player.Player.ImageURL)
			player.Hero.ImageURL = s.BfsProxy(c, player.Hero.ImageURL)
			player.Team.ImageURL = s.BfsProxy(c, player.Team.ImageURL)
			for _, ability := range player.Abilities {
				ability.ImageURL = s.BfsProxy(c, ability.ImageURL)
			}
			for _, item := range player.Items {
				item.ImageURL = s.BfsProxy(c, item.ImageURL)
			}
		}
		if _, ok := idMap[game.ID]; ok {
			if err = s.dao.UpDotaGame(c, game); err != nil {
				log.Error("writeDota s.dao.UpDotaGame MatchID(%d) gameID(%d) error(%+v)", matchID, game.ID, err)
			}
			time.Sleep(time.Millisecond * 10)
		} else {
			addGames = append(addGames, game)
		}
	}
	if len(addGames) > 0 {
		if err = s.dao.AddDotaGame(c, addGames); err != nil {
			log.Error("writeDota s.dao.AddDotaGame MatchID(%d) error(%+v)", matchID, err)
		}
	}
}

func (s *Service) writeOw(matchID int64, ldGames []*model.OwLdGame) {
	var (
		err      error
		ids      []*model.Oid
		idMap    map[int64]struct{}
		c        = context.Background()
		addGame  *model.OwGame
		addGames []*model.OwGame
		players  map[int64]model.OwPlayerStats
	)
	if ids, err = s.dao.OwGames(c, matchID); err != nil {
		log.Error("writeOw s.dao.OwGames MatchID(%d) error(%+v)", matchID, err)
		return
	}
	idMap = make(map[int64]struct{}, len(ids))
	for _, p := range ids {
		idMap[p.ID] = struct{}{}
	}
	for _, game := range ldGames {
		addGame = &model.OwGame{}
		if len(game.Rounds) == 0 {
			addGame = &model.OwGame{WinTeam: game.Winner.ID, Position: game.Position, MatchID: game.MatchID, ID: game.ID, Finished: game.Finished, EndAt: game.EndAt, BeginAt: game.BeginAt}
		} else {
			for roundIndex, round := range game.Rounds {
				if roundIndex == _firstOwGame {
					for teamIndex, team := range round.Teams {
						if teamIndex == _firstOwGame {
							pCount := len(team.Players)
							players = make(map[int64]model.OwPlayerStats, pCount)
						}
						team.Team.ImageURL = s.BfsProxy(c, team.Team.ImageURL)
						for _, player := range team.Players {
							player.Player.ImageURL = s.BfsProxy(c, player.Player.ImageURL)
							if _, ok := players[player.PlayerID]; ok {
								ultimate := players[player.PlayerID].Ultimate + player.Ultimate
								resurrections := players[player.PlayerID].Resurrections + player.Resurrections
								kills := players[player.PlayerID].Kills + player.Kills
								destructions := players[player.PlayerID].Destructions + player.Destructions
								deaths := players[player.PlayerID].Deaths + player.Deaths
								players[player.PlayerID] = model.OwPlayerStats{
									Ultimate:      ultimate,
									Resurrections: resurrections,
									Kills:         kills,
									Destructions:  destructions,
									Deaths:        deaths,
								}
							} else {
								players[player.PlayerID] = model.OwPlayerStats{
									Ultimate:      player.Ultimate,
									Resurrections: player.Resurrections,
									Kills:         player.Kills,
									Destructions:  player.Destructions,
									Deaths:        player.Deaths,
								}
							}
						}
					}
					addGame = &model.OwGame{WinTeam: game.Winner.ID, Teams: round.Teams, Position: game.Position, MatchID: game.MatchID, ID: game.ID, Finished: game.Finished, EndAt: game.EndAt, BeginAt: game.BeginAt}
				} else {
					for _, team := range round.Teams {
						for _, player := range team.Players {
							if _, ok := players[player.PlayerID]; ok {
								ultimate := players[player.PlayerID].Ultimate + player.Ultimate
								resurrections := players[player.PlayerID].Resurrections + player.Resurrections
								kills := players[player.PlayerID].Kills + player.Kills
								destructions := players[player.PlayerID].Destructions + player.Destructions
								deaths := players[player.PlayerID].Deaths + player.Deaths
								players[player.PlayerID] = model.OwPlayerStats{
									Ultimate:      ultimate,
									Resurrections: resurrections,
									Kills:         kills,
									Destructions:  destructions,
									Deaths:        deaths,
								}
							}
						}
					}
				}
			}
			for _, team := range addGame.Teams {
				for _, player := range team.Players {
					if addPlayer, ok := players[player.PlayerID]; ok {
						player.Ultimate = addPlayer.Ultimate
						player.Resurrections = addPlayer.Resurrections
						player.Kills = addPlayer.Kills
						player.Destructions = addPlayer.Destructions
						player.Deaths = addPlayer.Deaths
					}
				}
			}
		}
		if game.Map != nil && game.Map.ThumbnailURL != "" {
			game.Map.ThumbnailURL = s.BfsProxy(c, game.Map.ThumbnailURL)
			addGame.Map = game.Map
		}
		if _, ok := idMap[addGame.ID]; ok {
			if err = s.dao.UpOwGame(c, addGame); err != nil {
				log.Error("writeOw s.dao.UpOwGame MatchID(%d) gameID(%d) error(%+v)", matchID, game.ID, err)
			}
			time.Sleep(time.Millisecond * 10)
		} else {
			addGames = append(addGames, addGame)
		}
	}
	if len(addGames) > 0 {
		if err = s.dao.AddOwGame(c, addGames); err != nil {
			log.Error("writeOw s.dao.AddOwGame MatchID(%d) error(%+v)", matchID, err)
		}
	}
}

// writeInfo lol champion;dota hero;overwatch hero.
func (s *Service) writeInfo(tp string, rs []byte) (err error) {
	var (
		ldInfos, verInfos []*model.LdInfo
		c                 = context.Background()
		verID             map[int]struct{}
	)
	if err = json.Unmarshal(rs, &ldInfos); err != nil {
		log.Error("writeInfo json.Unmarshal tp(%v) rs(%+v) error(%v)", tp, string(rs), err)
		return
	}
	if tp == _lolVerChampions {
		verCount := len(s.c.Leidata.Hero.IDs)
		if verCount > 0 {
			verID = make(map[int]struct{}, verCount)
			for _, v := range s.c.Leidata.Hero.IDs {
				verID[v] = struct{}{}
			}
			for _, info := range ldInfos {
				if _, ok := verID[info.ID]; ok {
					info.ImageURL = s.BfsProxy(c, info.ImageURL)
					info.Name = strings.Replace(info.Name, "'", "\\'", -1)
					verInfos = append(verInfos, info)
				}
			}
		}
	} else {
		for _, info := range ldInfos {
			info.ImageURL = s.BfsProxy(c, info.ImageURL)
			info.Name = strings.Replace(info.Name, "'", "\\'", -1)
		}
	}
	switch tp {
	case _lolChampions:
		if err = s.dao.AddLolCham(c, ldInfos); err != nil {
			log.Error("writeInfo lolChampions  s.dao.AddLolCham error(%+v)", err)
		}
	case _lolVerChampions:
		if err = s.dao.AddLolCham(c, verInfos); err != nil {
			log.Error("writeInfo  lolVersionsChampions  s.dao.AddLolCham  error(%+v)", err)
		}
	case _dotaHeroes:
		if err = s.dao.AddDotaHero(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddDotaHero error(%+v)", err)
		}
	case _owHeroes:
		if err = s.dao.AddOwHero(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddOwHero error(%+v)", err)
		}
	case _lolItems:
		if err = s.dao.AddLolItem(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddLolItem error(%+v)", err)
		}
	case _dotaItems:
		if err = s.dao.AddDotaItem(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddDotaItem error(%+v)", err)
		}
	case _owMaps:
		if err = s.dao.AddOwMap(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddOwMap error(%+v)", err)
		}
	case _lolPlayers:
		if err = s.dao.AddLolMatchPlayer(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddLolMatchPlayer error(%+v)", err)
		}
	case _dotaPlayers:
		if err = s.dao.AddDotaMatchPlayer(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddDotaMatchPlayer error(%+v)", err)
		}
	case _owPlayers:
		if err = s.dao.AddOwMatchPlayer(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddOwMatchPlayer error(%+v)", err)
		}
	case _lolSpells:
		if err = s.dao.AddLolAbility(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddLolAbility error(%+v)", err)
		}
	case _dotaAbilities:
		if err = s.dao.AddDotaAbility(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddDotaAbility error(%+v)", err)
		}
	case _lolTeams:
		if err = s.dao.AddLolTeams(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddLolTeams error(%+v)", err)
		}
	case _dotaTeams:
		if err = s.dao.AddDotateams(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddDotateams error(%+v)", err)
		}
	case _owTeams:
		if err = s.dao.AddOwTeams(c, ldInfos); err != nil {
			log.Error("writeInfo  s.dao.AddOwTeams error(%+v)", err)
		}
	}
	return
}

func (s *Service) tunnelPushActiveEvent() {
	var (
		err      error
		contests []*model.Contest
		ctx      = context.Background()
	)
	for {
		stime := time.Now()
		etime := stime.Add(time.Minute)
		if contests, err = s.getPushContests(ctx, stime.Unix(), etime.Unix()); err != nil {
			log.Error("contestsPush contests stime(%d) etime(%d) error(%+v)", stime.Unix(), etime.Unix(), err)
		} else {
			for _, contest := range contests {
				s.activeEvent(ctx, contest)
			}
		}
		time.Sleep(time.Minute)
	}
}

func (s *Service) activeEvent(ctx context.Context, contest *model.Contest) (err error) {
	if contest.Stime == 0 || contest.Etime == 0 {
		log.Errorc(ctx, "activeEvent contest  Stime(%d)  Etime(%d)", contest.Stime, contest.Etime)
		return
	}
	// 激活事件
	activeArg := &tunnelmdl.ActiveEventReq{
		BizId:     s.c.Rule.TunnelBizID,
		UniqueId:  contest.ID,
		Platform:  _platform,
		StartTime: time.Unix(contest.Stime, 0).Format("2006-01-02 15:04:05"),
		EndTime:   time.Unix(contest.Etime, 0).Format("2006-01-02 15:04:05"),
	}
	for i := 0; i < _retry; i++ {
		if _, err = component.TunnelClient.ActiveEvent(ctx, activeArg); err == nil {
			break
		}
	}
	if err != nil {
		log.Errorc(ctx, "activeEvent s.tunnelClient.ActiveEvent contestID(%d) error(%v)", contest.ID, err)
		return err
	}
	if err = s.dao.UpContestPush(ctx, contest.ID); err != nil {
		log.Errorc(ctx, "activeEvent s.dao.UpContestPush contestID(%d) error(%v)", contest.ID, err)
		return err
	}
	log.Infoc(ctx, "activeEvent success contestID(%d)", contest.ID)
	return
}

func (s *Service) getPushContests(c context.Context, stime, etime int64) (res []*model.Contest, err error) {
	for i := 0; i < _tryTimes; i++ {
		if res, err = s.dao.ContestsPush(c, stime, etime); err == nil {
			break
		}
		time.Sleep(time.Millisecond * 200)
	}
	if err != nil {
		log.Error("s.dao.ContestsPush error(%v)", err)
	}
	return
}

func (s *Service) TunnelPush(ctx context.Context, contestID, mid int64) (err error) {
	var (
		contest *model.Contest
	)
	// 发送 databus
	if err = s.dao.AsyncSendTunnelDatabus(ctx, _platform, mid, contestID); err != nil {
		log.Errorc(ctx, "TunnelPush s.dao.AsyncSendTunnelDatabus contestID(%d) error(%+v)", contestID, err)
		return
	}
	if contest, err = s.dao.PushHandUse(ctx, contestID); err != nil {
		log.Errorc(ctx, "TunnelPush s.dao.PushHandUse contestID(%d) error(%+v)", contestID, err)
		return
	}
	// 激活事件
	if err = s.activeEvent(ctx, contest); err != nil {
		log.Errorc(ctx, "TunnelPush s.activeEvent contestID(%d) error(%+v)", contestID, err)
	}
	return
}

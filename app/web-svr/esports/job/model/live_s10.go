package model

type ScoreMatch struct {
	MatchID   string `json:"matchID"`
	Teama     string `json:"teama"`
	Teamb     string `json:"teamb"`
	StartTime string `json:"start_time"`
}

type BattleList struct {
	MatchInfo struct {
		MatchID string `json:"matchID"`
		Status  string `json:"status"`
	} `json:"match_info"`
	List []*struct {
		Status       int64       `json:"status"`
		RedClanName  string      `json:"red_clan_name"`
		WinClanName  string      `json:"win_clan_name"`
		BlueClanName string      `json:"blue_clan_name"`
		WinTeamID    interface{} `json:"win_teamID"`
		BlueTeamID   interface{} `json:"blue_teamID"`
		RedTeamID    interface{} `json:"red_teamID"`
		WinClanColor string      `json:"win_clan_color"`
		BattleString string      `json:"battle_string"`
	} `json:"list"`
	LastTime int64 `json:"last_time"`
}

type BattleInfo struct {
	PickList []*struct {
		HeroName     string      `json:"hero_name"`
		HeroID       interface{} `json:"heroID"`
		HeroNickname string      `json:"hero_nickname"`
		GroupID      interface{} `json:"group_id"`
		HeroImage    string      `json:"hero_image"`
	} `json:"pick_list"`
	EcoList []interface{} `json:"eco_list"`
	WinRate struct {
		BlueWinRate []struct {
			Timestamp interface{} `json:"timestamp"`
			WinRate   interface{} `json:"win_rate"`
		} `json:"blue_win_rate"`
	} `json:"win_rate"`
	Game struct {
		Status       int64       `json:"status"`
		RedClanName  string      `json:"red_clan_name"`
		BlueClanName string      `json:"blue_clan_name"`
		MatchID      interface{} `json:"match_id"`
		NumberTxt    string      `json:"number_txt"`
		BlueTeamID   interface{} `json:"blue_teamID"`
		StartTime    interface{} `json:"start_time"`
		GameTime     interface{} `json:"game_time"`
		Number       interface{} `json:"number"`
		WinTeamID    interface{} `json:"win_teamID"`
		EndTime      interface{} `json:"end_time"`
		GameTimeTxt  string      `json:"game_time_txt"`
		RedTeamID    interface{} `json:"red_teamID"`
	} `json:"game"`
	Timeline []struct {
		GameTime interface{} `json:"game_time"`
		Devices  []struct {
			DeviceID interface{} `json:"device_id"`
		} `json:"devices,omitempty"`
		GameTimeTxt     string        `json:"game_time_txt"`
		Group_ID        interface{}   `json:"group_id"`
		Type            string        `json:"type"`
		GroupID         interface{}   `json:"groupId"`
		HeroID          interface{}   `json:"hero_id,omitempty"`
		AssistantCount  interface{}   `json:"assistant_count,omitempty"`
		AssistantIDList []interface{} `json:"assistant_id_list,omitempty"`
		KillerID        interface{}   `json:"killer_id,omitempty"`
		DeadID          interface{}   `json:"dead_id,omitempty"`
		AxisX           interface{}   `json:"axis_x,omitempty"`
		AxisY           interface{}   `json:"axis_y,omitempty"`
		ID              interface{}   `json:"id,omitempty"`
		DragonType      interface{}   `json:"dragon_type,omitempty"`
		TowerID         interface{}   `json:"tower_id,omitempty"`
		DestroyType     interface{}   `json:"destroy_type,omitempty"`
		TowerName       string        `json:"tower_name,omitempty"`
	} `json:"timeline"`
	BanList []*struct {
		HeroName     string      `json:"hero_name"`
		HeroID       interface{} `json:"heroID"`
		HeroNickname string      `json:"hero_nickname"`
		GroupID      interface{} `json:"group_id"`
		HeroImage    string      `json:"hero_image"`
	} `json:"ban_list"`
	Teama *struct {
		WardsKilled     interface{} `json:"wardsKilled"`
		First10Kill     interface{} `json:"first10Kill"`
		Deaths          interface{} `json:"deaths"`
		WardsPlaced     interface{} `json:"wardsPlaced"`
		TeamImageThumb  string      `json:"team_image_thumb"`
		TeamImageThumbA string      `json:"team_image_thumb_a"`
		FirstTowerKill  interface{} `json:"firstTowerKill"`
		First5Kill      interface{} `json:"first5Kill"`
		ClanID          interface{} `json:"clan_id"`
		Damages         string      `json:"damages"`
		FirstBloodKill  interface{} `json:"firstBloodKill"`
		Kills           interface{} `json:"kills"`
		Towers          interface{} `json:"towers"`
		TeamShortName   string      `json:"team_short_name"`
		Assists         int         `json:"assists"`
		Dragons         []*struct {
			DragonImage string      `json:"dragon_image"`
			DragonType  interface{} `json:"dragon_type"`
			GameTime    interface{} `json:"game_time"`
			GameTimeTxt string      `json:"game_time_txt"`
			GroupID     interface{} `json:"groupId"`
			Group_ID    interface{} `json:"group_id"`
			ID          interface{} `json:"id"`
			Type        string      `json:"type"`
		} `json:"dragons"`
		GroupID   interface{} `json:"groupId"`
		TeamIDA   interface{} `json:"teamID_a"`
		Economics string      `json:"economics"`
		Players   []*struct {
			Sort          string        `json:"sort"`
			WardsKilled   interface{}   `json:"wardsKilled"`
			AssisNum      interface{}   `json:"assis_num"`
			PlayerID      interface{}   `json:"player_id"`
			HeroID        interface{}   `json:"heroID"`
			Spells        []interface{} `json:"spells"`
			HeroName      string        `json:"hero_name"`
			PlayerName    string        `json:"player_name"`
			LasthitNum    interface{}   `json:"lasthit_num"`
			WardsPlaced   interface{}   `json:"wardsPlaced"`
			KillNum       interface{}   `json:"kill_num"`
			DeadNum       interface{}   `json:"dead_num"`
			ChampLevel    interface{}   `json:"champLevel"`
			Hero_ID       interface{}   `json:"hero_id"`
			PositionID    interface{}   `json:"positionID"`
			HeroDamage    interface{}   `json:"hero_damage"`
			HeroWound     interface{}   `json:"hero_wound"`
			Economics     interface{}   `json:"economics"`
			Devices       []interface{} `json:"devices"`
			Perks         []interface{} `json:"perks"`
			EconomicsRate interface{}   `json:"economics_rate"`
			HeroImage     string        `json:"hero_image"`
			PlayerImage   string        `json:"player_image"`
		} `json:"players"`
		FirstDragonKill interface{} `json:"firstDragonKill"`
		Group_ID        interface{} `json:"group_id"`
		FirstBaronKill  interface{} `json:"firstBaronKill"`
		TeamShortNameA  string      `json:"team_short_name_a"`
	} `json:"teama"`
	Teamb *struct {
		WardsKilled interface{} `json:"wardsKilled"`
		First10Kill interface{} `json:"first10Kill"`
		Deaths      interface{} `json:"deaths"`
		Dragons     []*struct {
			DragonImage string      `json:"dragon_image"`
			Group_ID    interface{} `json:"group_id"`
			GameTimeTxt string      `json:"game_time_txt"`
			DragonType  interface{} `json:"dragon_type"`
			GameTime    interface{} `json:"game_time"`
			Type        string      `json:"type"`
			ID          interface{} `json:"id"`
			GroupID     interface{} `json:"groupId"`
		} `json:"dragons"`
		TeamImageThumb  string      `json:"team_image_thumb"`
		FirstTowerKill  interface{} `json:"firstTowerKill"`
		WardsPlaced     interface{} `json:"wardsPlaced"`
		ClanID          interface{} `json:"clan_id"`
		Damages         string      `json:"damages"`
		FirstBloodKill  interface{} `json:"firstBloodKill"`
		TeamImageThumbB string      `json:"team_image_thumb_b"`
		Kills           interface{} `json:"kills"`
		Towers          interface{} `json:"towers"`
		TeamShortName   string      `json:"team_short_name"`
		Assists         interface{} `json:"assists"`
		First5Kill      interface{} `json:"first5Kill"`
		GroupID         interface{} `json:"groupId"`
		TeamIDB         interface{} `json:"teamID_b"`
		TeamShortNameB  string      `json:"team_short_name_b"`
		Economics       string      `json:"economics"`
		Players         []*struct {
			Sort          string        `json:"sort"`
			WardsKilled   interface{}   `json:"wardsKilled"`
			AssisNum      interface{}   `json:"assis_num"`
			PlayerID      interface{}   `json:"player_id"`
			HeroID        interface{}   `json:"heroID"`
			Spells        []interface{} `json:"spells"`
			HeroName      string        `json:"hero_name"`
			PlayerName    string        `json:"player_name"`
			LasthitNum    interface{}   `json:"lasthit_num"`
			WardsPlaced   interface{}   `json:"wardsPlaced"`
			KillNum       interface{}   `json:"kill_num"`
			DeadNum       interface{}   `json:"dead_num"`
			ChampLevel    interface{}   `json:"champLevel"`
			Hero_ID       interface{}   `json:"hero_id"`
			PositionID    interface{}   `json:"positionID"`
			HeroDamage    interface{}   `json:"hero_damage"`
			HeroWound     interface{}   `json:"hero_wound"`
			Economics     interface{}   `json:"economics"`
			Devices       []interface{} `json:"devices"`
			Perks         []interface{} `json:"perks"`
			EconomicsRate interface{}   `json:"economics_rate"`
			HeroImage     string        `json:"hero_image"`
			PlayerImage   string        `json:"player_image"`
		} `json:"players"`
		FirstDragonKill interface{} `json:"firstDragonKill"`
		Group_ID        interface{} `json:"group_id"`
		FirstBaronKill  interface{} `json:"firstBaronKill"`
	} `json:"teamb"`
	LastTime int64 `json:"last_time"`
}

type OffLineImage struct {
	ItemType   int64  `json:"item_type"`
	ItemId     string `json:"item_id"`
	ItemName   string `json:"item_name"`
	NickName   string `json:"nick_name"`
	ScoreImage string `json:"score_image"`
	BfsImage   string `json:"bfs_image"`
}

type OffLineHero struct {
	HeroID   string `json:"heroID"`
	Name     string `json:"name"`
	Nickname string `json:"nickname"`
	Image    string `json:"image"`
}

type OffLineJsZb struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	NameEn string `json:"name_en"`
	Image  string `json:"image"`
}

type OffLineTeam struct {
	TeamID        string `json:"teamID"`
	TeamName      string `json:"team_name"`
	TeamShortName string `json:"team_short_name"`
	TeamImage     string `json:"team_image"`
	PlayerList    []struct {
		PlayerID   string `json:"playerID"`
		Nickname   string `json:"nickname"`
		Name       string `json:"name"`
		Image      string `json:"image"`
		TeamID     string `json:"teamID"`
		PositionID string `json:"positionID"`
		StatusID   string `json:"statusID"`
		ImageThumb string `json:"image_thumb"`
	} `json:"player_list"`
}

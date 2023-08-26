package model

// LolGame lol game.
type LolGame struct {
	Teams []*struct {
		TowerKills int64 `json:"tower_kills"`
		Team       struct {
			Slug     string `json:"slug"`
			Name     string `json:"name"`
			ImageURL string `json:"image_url"`
			ID       int64  `json:"id"`
			Acronym  string `json:"acronym"`
		} `json:"team"`
		PlayerIds      []int64 `json:"player_ids"`
		InhibitorKills int64   `json:"inhibitor_kills"`
		GoldEarned     int64   `json:"gold_earned"`
		FirstTower     bool    `json:"first_tower"`
		FirstInhibitor bool    `json:"first_inhibitor"`
		FirstDragon    bool    `json:"first_dragon"`
		FirstBlood     bool    `json:"first_blood"`
		DragonKills    int64   `json:"dragon_kills"`
		Color          string  `json:"color"`
		BaronKills     int64   `json:"baron_kills"`
		Bans           []int64 `json:"bans"`
	} `json:"teams"`
	Position int64 `json:"position"`
	Players  []*struct {
		TotalDamage struct {
			Taken            int64 `json:"taken"`
			DealtToChampions int64 `json:"dealt_to_champions"`
			Dealt            int64 `json:"dealt"`
		} `json:"total_damage"`
		PlayerID    int64  `json:"player_id"`
		Assists     int64  `json:"assists"`
		GoldEarned  int64  `json:"gold_earned"`
		WardsPlaced int64  `json:"wards_placed"`
		GameID      int64  `json:"game_id"`
		Role        string `json:"role"`
		Kills       int64  `json:"kills"`
		Spells      []*struct {
			Name     string `json:"name"`
			ImageURL string `json:"image_url"`
			ID       int64  `json:"id"`
		} `json:"spells"`
		Player struct {
			Slug      string `json:"slug"`
			Role      string `json:"role"`
			Name      string `json:"name"`
			LastName  string `json:"last_name"`
			ImageURL  string `json:"image_url"`
			ID        int64  `json:"id"`
			Hometown  string `json:"hometown"`
			FirstName string `json:"first_name"`
		} `json:"player"`
		Champion *struct {
			Name     string `json:"name"`
			ImageURL string `json:"image_url"`
			ID       int64  `json:"id"`
		} `json:"champion"`
		Deaths int64 `json:"deaths"`
		Items  []*struct {
			Name      string `json:"name"`
			IsTrinket bool   `json:"is_trinket"`
			ImageURL  string `json:"image_url"`
			ID        int64  `json:"id"`
		} `json:"items"`
		Level int64 `json:"level"`
		Team  struct {
			Slug     string `json:"slug"`
			Name     string `json:"name"`
			ImageURL string `json:"image_url"`
			ID       int64  `json:"id"`
			Acronym  string `json:"acronym"`
		} `json:"team"`
		MinionsKilled int64 `json:"minions_killed"`
	} `json:"players"`
	MatchID  int64  `json:"match_id"`
	ID       int64  `json:"id"`
	Finished bool   `json:"finished"`
	EndAt    string `json:"end_at"`
	BeginAt  string `json:"begin_at"`
}

// DotaGame dota game.
type DotaGame struct {
	Teams []*struct {
		Team struct {
			Slug     string `json:"slug"`
			Name     string `json:"name"`
			ImageURL string `json:"image_url"`
			ID       int64  `json:"id"`
			Acronym  string `json:"acronym"`
		} `json:"team"`
		TowerStatus struct {
			TopTier3      bool `json:"top_tier_3"`
			TopTier2      bool `json:"top_tier_2"`
			TopTier1      bool `json:"top_tier_1"`
			MiddleTier3   bool `json:"middle_tier_3"`
			MiddleTier2   bool `json:"middle_tier_2"`
			MiddleTier1   bool `json:"middle_tier_1"`
			BottomTier3   bool `json:"bottom_tier_3"`
			BottomTier2   bool `json:"bottom_tier_2"`
			BottomTier1   bool `json:"bottom_tier_1"`
			AncientTop    bool `json:"ancient_top"`
			AncientBottom bool `json:"ancient_bottom"`
		} `json:"tower_status"`
		BarracksStatus struct {
			TopRanged    bool `json:"top_ranged"`
			TopMelee     bool `json:"top_melee"`
			MiddleRanged bool `json:"middle_ranged"`
			MiddleMelee  bool `json:"middle_melee"`
			BottomRanged bool `json:"bottom_ranged"`
			BottomMelee  bool `json:"bottom_melee"`
		} `json:"barracks_status"`
		Score      int64   `json:"score"`
		PlayerIds  []int64 `json:"player_ids"`
		Picks      []int64 `json:"picks"`
		FirstBlood bool    `json:"first_blood"`
		Faction    string  `json:"faction"`
		Bans       []int64 `json:"bans"`
	} `json:"teams"`
	Position int64 `json:"position"`
	Players  []*struct {
		Hero struct {
			Name          string `json:"name"`
			LocalizedName string `json:"localized_name"`
			ImageURL      string `json:"image_url"`
			ID            int64  `json:"id"`
		} `json:"hero"`
		GoldPerMin int64  `json:"gold_per_min"`
		Faction    string `json:"faction"`
		Assists    int64  `json:"assists"`
		LaneCreep  int64  `json:"lane_creep"`
		HeroDamage int64  `json:"hero_damage"`
		Abilities  []*struct {
			Name     string `json:"name"`
			Level    int64  `json:"level"`
			ImageURL string `json:"image_url"`
			ID       int64  `json:"id"`
		} `json:"abilities"`
		GameID               int64 `json:"game_id"`
		DamageTaken          int64 `json:"damage_taken"`
		SentryUsed           int64 `json:"sentry_used"`
		TeamID               int64 `json:"team_id"`
		Denies               int64 `json:"denies"`
		SentryWardsDestroyed int64 `json:"sentry_wards_destroyed"`
		Kills                int64 `json:"kills"`
		GoldSpent            int64 `json:"gold_spent"`
		TowerKills           int64 `json:"tower_kills"`
		Player               struct {
			Slug      string `json:"slug"`
			Role      string `json:"role"`
			Name      string `json:"name"`
			LastName  string `json:"last_name"`
			ImageURL  string `json:"image_url"`
			ID        int64  `json:"id"`
			Hometown  string `json:"hometown"`
			FirstName string `json:"first_name"`
		} `json:"player"`
		GoldRemaining int64 `json:"gold_remaining"`
		Deaths        int64 `json:"deaths"`
		HeroLevel     int64 `json:"hero_level"`
		Team          struct {
			Slug     string `json:"slug"`
			Name     string `json:"name"`
			ImageURL string `json:"image_url"`
			ID       int64  `json:"id"`
			Acronym  string `json:"acronym"`
		} `json:"team"`
		Items []*struct {
			Name     string `json:"name"`
			ImageURL string `json:"image_url"`
			ID       int64  `json:"id"`
		} `json:"items"`
		Heal                   int64 `json:"heal"`
		XpPerMin               int64 `json:"xp_per_min"`
		ObserverWardsDestroyed int64 `json:"observer_wards_destroyed"`
		CampsStacked           int64 `json:"camps_stacked"`
		LastHits               int64 `json:"last_hits"`
		TowerDamage            int64 `json:"tower_damage"`
		ObserverUsed           int64 `json:"observer_used"`
		SentryWardsPurchased   int64 `json:"sentry_wards_purchased"`
	} `json:"players"`
	MatchID  int64  `json:"match_id"`
	ID       int64  `json:"id"`
	Finished bool   `json:"finished"`
	EndAt    string `json:"end_at"`
	BeginAt  string `json:"begin_at"`
}

// OwLdGame overwatch leida resource game.
type OwLdGame struct {
	Winner struct {
		ID int64 `json:"id"`
	} `json:"winner"`
	Rounds []*struct {
		Teams []*struct {
			Team struct {
				Slug     string `json:"slug"`
				Name     string `json:"name"`
				ImageURL string `json:"image_url"`
				ID       int64  `json:"id"`
				Acronym  string `json:"acronym"`
			} `json:"team"`
			Players []*struct {
				Ultimate      int64 `json:"ultimate"`
				Resurrections int64 `json:"resurrections"`
				PlayerID      int64 `json:"player_id"`
				Player        struct {
					Role     string `json:"role"`
					Name     string `json:"name"`
					ImageURL string `json:"image_url"`
					ID       int64  `json:"id"`
				} `json:"player"`
				Kills        int64 `json:"kills"`
				Destructions int64 `json:"destructions"`
				Deaths       int64 `json:"deaths"`
			} `json:"players"`
		} `json:"teams"`
		Round int64 `json:"round"`
	} `json:"rounds"`
	Position int64 `json:"position"`
	MatchID  int64 `json:"match_id"`
	Map      *struct {
		ThumbnailURL string `json:"thumbnail_url"`
		Slug         string `json:"slug"`
		Name         string `json:"name"`
		ID           int64  `json:"id"`
		GameMode     string `json:"game_mode"`
	} `json:"map"`
	ID       int64  `json:"id"`
	Finished bool   `json:"finished"`
	EndAt    string `json:"end_at"`
	BeginAt  string `json:"begin_at"`
}

// OwGame overwatch sum round game.
type OwGame struct {
	WinTeam int64 `json:"win_team"`
	Teams   []*struct {
		Team struct {
			Slug     string `json:"slug"`
			Name     string `json:"name"`
			ImageURL string `json:"image_url"`
			ID       int64  `json:"id"`
			Acronym  string `json:"acronym"`
		} `json:"team"`
		Players []*struct {
			Ultimate      int64 `json:"ultimate"`
			Resurrections int64 `json:"resurrections"`
			PlayerID      int64 `json:"player_id"`
			Player        struct {
				Role     string `json:"role"`
				Name     string `json:"name"`
				ImageURL string `json:"image_url"`
				ID       int64  `json:"id"`
			} `json:"player"`
			Kills        int64 `json:"kills"`
			Destructions int64 `json:"destructions"`
			Deaths       int64 `json:"deaths"`
		} `json:"players"`
	} `json:"teams"`
	Position int64 `json:"position"`
	MatchID  int64 `json:"match_id"`
	Map      *struct {
		ThumbnailURL string `json:"thumbnail_url"`
		Slug         string `json:"slug"`
		Name         string `json:"name"`
		ID           int64  `json:"id"`
		GameMode     string `json:"game_mode"`
	} `json:"map"`
	ID       int64  `json:"id"`
	Finished bool   `json:"finished"`
	EndAt    string `json:"end_at"`
	BeginAt  string `json:"begin_at"`
}

// OwPlayerStats overwatch player stats.
type OwPlayerStats struct {
	Ultimate      int64 `json:"ultimate"`
	Resurrections int64 `json:"resurrections"`
	Kills         int64 `json:"kills"`
	Destructions  int64 `json:"destructions"`
	Deaths        int64 `json:"deaths"`
}

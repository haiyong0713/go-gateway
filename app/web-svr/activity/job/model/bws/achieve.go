package bws

type Achieve struct {
	ID           int64
	AchievePoint int64
}

type UserAchieve struct {
	Aid int64
	Key string
}

type AchieveRank struct {
	Mid        int64
	TotalPoint int64
}

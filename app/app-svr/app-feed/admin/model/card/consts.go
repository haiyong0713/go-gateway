package card

const (
	CardTypeContent    = "content"    // 内容聚合卡
	CardTypeNavigation = "navigation" // 导航卡
)

// Constants here are same with model/search/recommend.go
// Aiming at integrating all "cardTypes" defined in various places.
const (
	// CardContent same with CardUnion
	CardContent = 7
	// CardNavigation baike navigation card
	CardNavigation = 20
)

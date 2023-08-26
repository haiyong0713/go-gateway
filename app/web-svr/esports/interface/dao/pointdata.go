package dao

const (
	_lolGame  = "1"
	_dotaGame = "2"
)

var (
	lolRole  = map[string]string{"1": "上路", "2": "中路", "3": "下路", "4": "打野", "5": "辅助"}
	dotaRole = map[string]string{"1": "1号位", "2": "2号位", "3": "3号位", "4": "4号位", "5": "5号位"}
)

// Roles get lol dota roles.
func (d *Dao) Roles(tp string) (rs map[string]string) {
	if tp == _lolGame {
		rs = lolRole
	} else if tp == _dotaGame {
		rs = dotaRole
	} else {
		rs = make(map[string]string)
	}
	return
}

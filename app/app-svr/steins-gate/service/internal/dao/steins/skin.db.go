package steins

import (
	"context"

	"go-common/library/log"
	"go-gateway/app/app-svr/steins-gate/service/internal/model"
)

const _skinListSQL = "SELECT id,`name`,cover,image,title_text_color,title_shadow_color,title_shadow_offset_x,title_shadow_offset_y,title_shadow_radius,progress_bar_color,progress_shadow_color FROM skin WHERE state=1 AND is_deleted=0 ORDER BY rank ASC,id DESC"

// RawSkinList raw skin list.
func (d *Dao) RawSkinList(c context.Context) (list []*model.Skin, err error) {
	rows, err := d.db.Query(c, _skinListSQL)
	if err != nil {
		log.Error("RawSkinList db.Query error(%v)", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var offsetX, offsetY, radius int64
		skin := new(model.Skin)
		if err = rows.Scan(&skin.ID, &skin.Name, &skin.Cover, &skin.Image, &skin.TitleTextColor, &skin.TitleShadowColor, &offsetX, &offsetY, &radius, &skin.ProgressBarColor, &skin.ProgressShadowColor); err != nil {
			log.Error("RawSkinList rows.Scan error(%v)", err)
			return
		}
		//nolint:gomnd
		skin.TitleShadowOffsetX = float32(offsetX) / 100
		//nolint:gomnd
		skin.TitleShadowOffsetY = float32(offsetY) / 100
		//nolint:gomnd
		skin.TitleShadowRadius = float32(radius) / 100
		list = append(list, skin)
	}
	return

}

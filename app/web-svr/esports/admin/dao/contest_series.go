package dao

import (
	"context"
	"encoding/json"
	"go-common/library/log"
	"go-gateway/app/web-svr/esports/admin/model"
)

// 吃鸡类赛季阶段的积分配置
func (d *Dao) ScoreRuleConfigUpdate(ctx context.Context, seriesId int64, scoreRules *model.PUBGContestSeriesScoreRule) (err error) {
	var configStr string
	if scoreRules != nil {
		configBytes, err := json.Marshal(scoreRules)
		configStr = string(configBytes)
		if err != nil {
			log.Errorc(ctx, "[DB][ScoreRuleConfigUpdate][Marshal][Error], err:(%+v)", err)
		}
	} else {
		configStr = ""
	}

	if err = d.DB.Model(model.ContestSeriesByScoreRule{}).Where("id = ?", seriesId).Updates(map[string]interface{}{"score_rule_config": configStr}).Error; err != nil {
		log.Errorc(ctx, "[DB][ExtraConfigUpdate][DoUpdate][Error], err:(%+v)", err)
	}
	return
}

func (d *Dao) GetScoreRuleConfigBySeriesId(ctx context.Context, seriesId int64) (contestSeries *model.ContestSeriesByScoreRule, err error) {
	contestSeries = new(model.ContestSeriesByScoreRule)
	err = d.DB.Model(model.ContestSeriesByScoreRule{}).Where("id = ? AND is_deleted = ?", seriesId, model.Identity4NotDeleted).Find(&contestSeries).Error
	if err != nil {
		log.Errorc(ctx, "[DB][ScoreRuleConfigGet][Error], err:(%+v)", err)
		return
	}
	return
}

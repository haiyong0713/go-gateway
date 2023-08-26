package api

import (
	"strconv"
)

func (player *DanmuPlayerConfig) Default() {
	if player != nil {
		player.PlayerDanmakuSwitch = true
		player.PlayerDanmakuSwitchSave = false
		player.PlayerDanmakuUseDefaultConfig = false
		player.PlayerDanmakuAiRecommendedSwitch = true
		player.PlayerDanmakuAiRecommendedLevel = 3
		player.PlayerDanmakuBlocktop = false
		player.PlayerDanmakuBlockscroll = false
		player.PlayerDanmakuBlockbottom = false
		player.PlayerDanmakuBlockcolorful = false
		player.PlayerDanmakuBlockrepeat = false
		player.PlayerDanmakuBlockspecial = false
		player.PlayerDanmakuOpacity = 0.8
		player.PlayerDanmakuScalingfactor = 1
		player.PlayerDanmakuDomain = 1
		player.PlayerDanmakuSpeed = 30
		player.PlayerDanmakuEnableblocklist = true
	}
	return
}

func (player *DanmuPlayerConfig) ToPlayerSha1() (res *DanmuPlayerConfigSha1) {
	res = &DanmuPlayerConfigSha1{
		PlayerDanmakuSwitch:              player.GetPlayerDanmakuSwitch(),
		PlayerDanmakuSwitchSave:          player.GetPlayerDanmakuSwitchSave(),
		PlayerDanmakuUseDefaultConfig:    player.GetPlayerDanmakuUseDefaultConfig(),
		PlayerDanmakuAiRecommendedSwitch: player.GetPlayerDanmakuAiRecommendedSwitch(),
		PlayerDanmakuAiRecommendedLevel:  player.GetPlayerDanmakuAiRecommendedLevel(),
		PlayerDanmakuBlocktop:            player.GetPlayerDanmakuBlocktop(),
		PlayerDanmakuBlockscroll:         player.GetPlayerDanmakuBlockscroll(),
		PlayerDanmakuBlockbottom:         player.GetPlayerDanmakuBlockbottom(),
		PlayerDanmakuBlockcolorful:       player.GetPlayerDanmakuBlockcolorful(),
		PlayerDanmakuBlockrepeat:         player.GetPlayerDanmakuBlockrepeat(),
		PlayerDanmakuBlockspecial:        player.GetPlayerDanmakuBlockspecial(),
		PlayerDanmakuOpacity:             player.GetPlayerDanmakuOpacity(),
		PlayerDanmakuScalingfactor:       player.GetPlayerDanmakuScalingfactor(),
		PlayerDanmakuDomain:              player.GetPlayerDanmakuDomain(),
		PlayerDanmakuSpeed:               player.GetPlayerDanmakuSpeed(),
		PlayerDanmakuEnableblocklist:     player.GetPlayerDanmakuEnableblocklist(),
	}

	return
}

func (player *DanmuPlayerConfig) UseDefaultConfig() {
	if player != nil {
		player.PlayerDanmakuAiRecommendedSwitch = true
		player.PlayerDanmakuAiRecommendedLevel = 5
		player.PlayerDanmakuBlocktop = false
		player.PlayerDanmakuBlockscroll = false
		player.PlayerDanmakuBlockbottom = false
		player.PlayerDanmakuBlockcolorful = false
		player.PlayerDanmakuBlockrepeat = false
		player.PlayerDanmakuBlockspecial = false
		player.PlayerDanmakuOpacity = 0.8
		player.PlayerDanmakuScalingfactor = 1
		player.PlayerDanmakuDomain = 1
		player.PlayerDanmakuSpeed = 30
	}
	return
}

func (req *DmPlayerConfigReq) ToMap() (res map[string]string) {
	if req != nil {
		res = make(map[string]string, 20)
		if req.GetSwitch() != nil {
			res["switch"] = strconv.FormatBool(req.GetSwitch().GetValue())
		}
		if req.GetSwitchSave() != nil {
			res["switch_save"] = strconv.FormatBool(req.GetSwitchSave().GetValue())
		}
		if req.GetUseDefaultConfig() != nil {
			res["use_default_config"] = strconv.FormatBool(req.GetUseDefaultConfig().GetValue())
		}
		if req.GetAiRecommendedSwitch() != nil {
			res["ai_recommended_switch"] = strconv.FormatBool(req.GetAiRecommendedSwitch().GetValue())
		}
		if req.GetAiRecommendedLevel() != nil {
			res["ai_recommended_level"] = strconv.FormatInt(int64(req.GetAiRecommendedLevel().GetValue()), 10)
		}
		if req.GetBlocktop() != nil {
			res["blocktop"] = strconv.FormatBool(req.GetBlocktop().GetValue())
		}
		if req.GetBlockscroll() != nil {
			res["blockscroll"] = strconv.FormatBool(req.GetBlockscroll().GetValue())
		}
		if req.GetBlockbottom() != nil {
			res["blockbottom"] = strconv.FormatBool(req.GetBlockbottom().GetValue())
		}
		if req.GetBlockcolorful() != nil {
			res["blockcolorful"] = strconv.FormatBool(req.GetBlockcolorful().GetValue())
		}
		if req.GetBlockrepeat() != nil {
			res["blockrepeat"] = strconv.FormatBool(req.GetBlockrepeat().GetValue())
		}
		if req.GetBlockspecial() != nil {
			res["blockspecial"] = strconv.FormatBool(req.GetBlockspecial().GetValue())
		}
		if req.GetOpacity() != nil {
			res["opacity"] = strconv.FormatFloat(float64(req.GetOpacity().GetValue()), 'g', 5, 64)
		}
		if req.GetScalingfactor() != nil {
			res["scalingfactor"] = strconv.FormatFloat(float64(req.GetScalingfactor().GetValue()), 'g', 5, 64)
		}
		if req.GetDomain() != nil {
			res["domain"] = strconv.FormatFloat(float64(req.GetDomain().GetValue()), 'g', 5, 64)
		}
		if req.GetSpeed() != nil {
			res["speed"] = strconv.FormatInt(int64(req.GetSpeed().GetValue()), 10)
		}
		if req.GetEnableblocklist() != nil {
			res["enableblocklist"] = strconv.FormatBool(req.GetEnableblocklist().GetValue())
		}
	}
	return
}

func (player *DanmuPlayerConfig) Diff(dst *DanmuPlayerConfig, platForm string) (msg string, res bool) {
	if platForm != "android" || dst.PlayerDanmakuSwitchSave {
		if player.GetPlayerDanmakuSwitch() != dst.GetPlayerDanmakuSwitch() {
			res = true
			msg += "switch x "
		}
	}
	if player.GetPlayerDanmakuSwitchSave() != dst.GetPlayerDanmakuSwitchSave() {
		res = true
		msg += "switch_save x "
	}
	if player.GetPlayerDanmakuUseDefaultConfig() != dst.GetPlayerDanmakuUseDefaultConfig() {
		res = true
		msg += "use_default_config x "
	}
	if !dst.GetPlayerDanmakuUseDefaultConfig() {
		if player.GetPlayerDanmakuAiRecommendedSwitch() != dst.GetPlayerDanmakuAiRecommendedSwitch() {
			res = true
			msg += "ai_recommended_switch x "
		}
		if player.GetPlayerDanmakuAiRecommendedLevel() != dst.GetPlayerDanmakuAiRecommendedLevel() {
			res = true
			msg += "ai_level x "
		}
		if player.GetPlayerDanmakuBlocktop() != dst.GetPlayerDanmakuBlocktop() {
			res = true
			msg += "top x "
		}
		if player.GetPlayerDanmakuBlockscroll() != dst.GetPlayerDanmakuBlockscroll() {
			res = true
			msg += "scroll x "
		}
		if player.GetPlayerDanmakuBlockbottom() != dst.GetPlayerDanmakuBlockbottom() {
			res = true
			msg += "bottom x "
		}
		if player.GetPlayerDanmakuBlockcolorful() != dst.GetPlayerDanmakuBlockcolorful() {
			res = true
			msg += "colorful x "
		}
		if player.GetPlayerDanmakuBlockrepeat() != dst.GetPlayerDanmakuBlockrepeat() {
			res = true
			msg += "repeat x "
		}
		if player.GetPlayerDanmakuBlockspecial() != dst.GetPlayerDanmakuBlockspecial() {
			res = true
			msg += "special x "
		}
		opacity, _ := strconv.ParseFloat(strconv.FormatFloat(float64(dst.GetPlayerDanmakuOpacity()), 'g', 5, 64), 64)
		dstopacity, _ := strconv.ParseFloat(strconv.FormatFloat(float64(player.GetPlayerDanmakuOpacity()), 'g', 5, 64), 64)
		if float32(dstopacity) != float32(opacity) {
			res = true
			msg += "Opacity x "
		}
		scalingfactor, _ := strconv.ParseFloat(strconv.FormatFloat(float64(dst.GetPlayerDanmakuScalingfactor()), 'g', 5, 64), 64)
		dstscalingfactor, _ := strconv.ParseFloat(strconv.FormatFloat(float64(player.GetPlayerDanmakuScalingfactor()), 'g', 5, 64), 64)
		if float32(dstscalingfactor) != float32(scalingfactor) {
			res = true
			msg += "Scalingfactor x "
		}
		domain, _ := strconv.ParseFloat(strconv.FormatFloat(float64(dst.GetPlayerDanmakuDomain()), 'g', 5, 64), 64)
		dstdomain, _ := strconv.ParseFloat(strconv.FormatFloat(float64(player.GetPlayerDanmakuDomain()), 'g', 5, 64), 64)
		if float32(dstdomain) != float32(domain) {
			res = true
			msg += "Domain x "
		}
		if player.GetPlayerDanmakuSpeed() != dst.GetPlayerDanmakuSpeed() {
			res = true
			msg += "Speed x "
		}
		if player.GetPlayerDanmakuEnableblocklist() != dst.GetPlayerDanmakuEnableblocklist() {
			res = true
			msg += "Enableblocklist x "
		}
	}
	return
}

func (player *DanmuPlayerConfig) Change(dst ConfigModify) (modify bool) {
	var (
		req *DmPlayerConfigReq
		ok  bool
	)
	if req, ok = dst.(*DmPlayerConfigReq); !ok {
		return
	}
	if req.GetSwitch() != nil {
		if req.GetSwitch().GetValue() != player.GetPlayerDanmakuSwitch() {
			player.PlayerDanmakuSwitch = req.GetSwitch().GetValue()
			modify = true
		}
	}
	if req.GetSwitchSave() != nil {
		if req.GetSwitchSave().GetValue() != player.GetPlayerDanmakuSwitchSave() {
			player.PlayerDanmakuSwitchSave = req.GetSwitchSave().GetValue()
			modify = true
		}
	}
	if req.GetUseDefaultConfig() != nil {
		if req.GetUseDefaultConfig().GetValue() != player.GetPlayerDanmakuUseDefaultConfig() {
			player.PlayerDanmakuUseDefaultConfig = req.GetUseDefaultConfig().GetValue()
			modify = true
		}
	}
	if req.GetAiRecommendedSwitch() != nil {
		if req.GetAiRecommendedSwitch().GetValue() != player.GetPlayerDanmakuAiRecommendedSwitch() {
			player.PlayerDanmakuAiRecommendedSwitch = req.GetAiRecommendedSwitch().GetValue()
			modify = true
		}
	}
	if req.GetAiRecommendedLevel() != nil {
		if req.GetAiRecommendedLevel().GetValue() != player.GetPlayerDanmakuAiRecommendedLevel() {
			player.PlayerDanmakuAiRecommendedLevel = req.GetAiRecommendedLevel().GetValue()
			modify = true
		}
	}
	if req.GetBlocktop() != nil {
		if req.GetBlocktop().GetValue() != player.GetPlayerDanmakuBlocktop() {
			player.PlayerDanmakuBlocktop = req.GetBlocktop().GetValue()
			modify = true
		}
	}
	if req.GetBlockscroll() != nil {
		if req.GetBlockscroll().GetValue() != player.GetPlayerDanmakuBlockscroll() {
			player.PlayerDanmakuBlockscroll = req.GetBlockscroll().GetValue()
			modify = true
		}
	}
	if req.GetBlockbottom() != nil {
		if req.GetBlockbottom().GetValue() != player.GetPlayerDanmakuBlockbottom() {
			player.PlayerDanmakuBlockbottom = req.GetBlockbottom().GetValue()
			modify = true
		}
	}
	if req.GetBlockcolorful() != nil {
		if req.GetBlockcolorful().GetValue() != player.GetPlayerDanmakuBlockcolorful() {
			player.PlayerDanmakuBlockcolorful = req.GetBlockcolorful().GetValue()
			modify = true
		}
	}
	if req.GetBlockrepeat() != nil {
		if req.GetBlockrepeat().GetValue() != player.GetPlayerDanmakuBlockrepeat() {
			player.PlayerDanmakuBlockrepeat = req.GetBlockrepeat().GetValue()
			modify = true
		}
	}
	if req.GetBlockspecial() != nil {
		if req.GetBlockspecial().GetValue() != player.GetPlayerDanmakuBlockspecial() {
			player.PlayerDanmakuBlockspecial = req.GetBlockspecial().GetValue()
			modify = true
		}
	}
	if req.GetOpacity() != nil {
		if req.GetOpacity().GetValue() != player.GetPlayerDanmakuOpacity() {
			player.PlayerDanmakuOpacity = req.GetOpacity().GetValue()
			modify = true
		}
	}
	if req.GetScalingfactor() != nil {
		if req.GetScalingfactor().GetValue() != player.GetPlayerDanmakuScalingfactor() {
			player.PlayerDanmakuScalingfactor = req.GetScalingfactor().GetValue()
			modify = true
		}
	}
	if req.GetDomain() != nil {
		if req.GetDomain().GetValue() != player.GetPlayerDanmakuDomain() {
			player.PlayerDanmakuDomain = req.GetDomain().GetValue()
			modify = true
		}
	}
	if req.GetSpeed() != nil {
		if req.GetSpeed().GetValue() != player.GetPlayerDanmakuSpeed() {
			player.PlayerDanmakuSpeed = req.GetSpeed().GetValue()
			modify = true
		}
	}
	if req.GetEnableblocklist() != nil {
		if req.GetEnableblocklist().GetValue() != player.GetPlayerDanmakuEnableblocklist() {
			player.PlayerDanmakuEnableblocklist = req.GetEnableblocklist().GetValue()
			modify = true
		}
	}
	return
}

func (dst *DmPlayerConfigReq) Merge(req interface{}) {
	var (
		src *DmPlayerConfigReq
		ok  bool
	)
	if src, ok = req.(*DmPlayerConfigReq); !ok {
		return
	}
	if src.GetSwitch() != nil {
		dst.Switch = src.GetSwitch()
	}
	if src.GetSwitchSave() != nil {
		dst.SwitchSave = src.GetSwitchSave()
	}
	if src.GetUseDefaultConfig() != nil {
		dst.UseDefaultConfig = src.GetUseDefaultConfig()
	}
	if src.GetAiRecommendedSwitch() != nil {
		dst.AiRecommendedSwitch = src.GetAiRecommendedSwitch()
	}
	if src.GetAiRecommendedLevel() != nil {
		dst.AiRecommendedLevel = src.GetAiRecommendedLevel()
	}
	if src.GetBlocktop() != nil {
		dst.Blocktop = src.GetBlocktop()
	}
	if src.GetBlockscroll() != nil {
		dst.Blockscroll = src.GetBlockscroll()
	}
	if src.GetBlockbottom() != nil {
		dst.Blockbottom = src.GetBlockbottom()
	}
	if src.GetBlockcolorful() != nil {
		dst.Blockcolorful = src.GetBlockcolorful()
	}
	if src.GetBlockrepeat() != nil {
		dst.Blockrepeat = src.GetBlockrepeat()
	}
	if src.GetBlockspecial() != nil {
		dst.Blockspecial = src.GetBlockspecial()
	}
	if src.GetOpacity() != nil {
		dst.Opacity = src.GetOpacity()
	}
	if src.GetScalingfactor() != nil {
		dst.Scalingfactor = src.GetScalingfactor()
	}
	if src.GetDomain() != nil {
		dst.Domain = src.GetDomain()
	}
	if src.GetSpeed() != nil {
		dst.Speed = src.GetSpeed()
	}
	if src.GetEnableblocklist() != nil {
		dst.Enableblocklist = src.GetEnableblocklist()
	}
	return
}

func (player *DanmuPlayerConfig) Merge(ucDoc map[string]string) {
	if val, ok := ucDoc["switch"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuSwitch = boolean
		}
	}
	if val, ok := ucDoc["switch_save"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuSwitchSave = boolean
		}
	}
	if val, ok := ucDoc["use_default_config"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuUseDefaultConfig = boolean
		}
	}
	if val, ok := ucDoc["ai_recommended_switch"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuAiRecommendedSwitch = boolean
		}
	}
	if val, ok := ucDoc["ai_recommended_level"]; ok {
		if intval, nerr := strconv.ParseInt(val, 10, 64); nerr == nil {
			player.PlayerDanmakuAiRecommendedLevel = int32(intval)
		}
	}
	if val, ok := ucDoc["blocktop"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuBlocktop = boolean
		}
	}
	if val, ok := ucDoc["blockscroll"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuBlockscroll = boolean
		}
	}
	if val, ok := ucDoc["blockbottom"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuBlockbottom = boolean
		}
	}
	if val, ok := ucDoc["blockcolorful"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuBlockcolorful = boolean
		}
	}
	if val, ok := ucDoc["blockrepeat"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuBlockrepeat = boolean
		}
	}
	if val, ok := ucDoc["blockspecial"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuBlockspecial = boolean
		}
	}
	if val, ok := ucDoc["opacity"]; ok {
		if floatval, nerr := strconv.ParseFloat(val, 64); nerr == nil {
			player.PlayerDanmakuOpacity = float32(floatval)
		}
	}
	if val, ok := ucDoc["scalingfactor"]; ok {
		if floatval, nerr := strconv.ParseFloat(val, 64); nerr == nil {
			player.PlayerDanmakuScalingfactor = float32(floatval)
		}
	}
	if val, ok := ucDoc["domain"]; ok {
		if floatval, nerr := strconv.ParseFloat(val, 64); nerr == nil {
			player.PlayerDanmakuDomain = float32(floatval)
		}
	}
	if val, ok := ucDoc["speed"]; ok {
		if intval, nerr := strconv.ParseInt(val, 10, 64); nerr == nil {
			player.PlayerDanmakuSpeed = int32(intval)
		}
	}
	if val, ok := ucDoc["enableblocklist"]; ok {
		if boolean, nerr := strconv.ParseBool(val); nerr == nil {
			player.PlayerDanmakuEnableblocklist = boolean
		}
	}
	if player.PlayerDanmakuUseDefaultConfig {
		player.UseDefaultConfig()
	}
	return
}

func NewDmCfg(data interface{}) Config {
	if data != nil {
		if res, ok := data.(*DanmuPlayerConfig); ok {
			return res
		}
		return nil
	}
	return &DanmuPlayerConfig{}
}

func NewDmCfgModify() ConfigModify {
	return &DmPlayerConfigReq{}
}

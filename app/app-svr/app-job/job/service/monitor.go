package service

import (
	"context"
	"fmt"

	"go-common/library/conf/env"
	"go-common/library/log"
	resmdl "go-gateway/app/app-svr/resource/service/model"
)

// watchSideBar is
func (s *Service) watchSideBar() {
	var (
		platToString = map[int8]string{
			0:  "Android",
			1:  "iPhone",
			2:  "iPad",
			8:  "Android 国际版",
			9:  "Android 蓝",
			10: "iPhone 蓝",
		}
		moduleToString = map[int]string{
			1:  "安卓侧边栏",
			2:  "安卓首页顶部",
			3:  "iphone我的页",
			4:  "iphone首页顶部",
			5:  "ipad首页顶部",
			6:  "iphone个人中心",
			7:  "iphone我的服务",
			8:  "首页顶部tab",
			9:  "首页底部tab",
			10: "首页顶部icon",
			11: "iphone创作中心",
			12: "ipad个人中心",
			13: "ipad创作中心",
			14: "安卓个人中心",
			15: "安卓创作中心",
			16: "安卓我的服务",
		}
		c           = context.Background()
		tmpSideBars = make(map[int8]map[int][]*resmdl.SideBar)
		sideBars    *resmdl.SideBars
		err         error
	)
	if sideBars, err = s.resourceRPC.SideBars(c); err != nil {
		log.Error("s.resourceRPC.SideBars error(%v)", err)
		return
	}
	for _, v := range sideBars.SideBar {
		if _, ok := tmpSideBars[v.Plat]; !ok {
			tmpSideBars[v.Plat] = make(map[int][]*resmdl.SideBar)
		}
		tmpSideBars[v.Plat][v.Module] = append(tmpSideBars[v.Plat][v.Module], v)
	}
	if s.sideBars == nil {
		s.sideBars = tmpSideBars
	} else {
		for plat, moudles := range s.sideBars {
			newModule, ok := tmpSideBars[plat]
			if !ok {
				_ = s.monitorDao.Send(c, fmt.Sprintf("env:%s ,%s 模块没了！！！", env.DeployEnv, platToString[plat]))
				continue
			}
			for module, sidebar := range moudles {
				if len(newModule[module]) != len(sidebar) {
					message := fmt.Sprintf("env:%s %s module:%s 模块发生变更！！！上一次有[%d]个配置项，本次有[%d]个配置项\n", env.DeployEnv, platToString[plat], moduleToString[module], len(sidebar), len(newModule[module]))
					message += "老的:\n"
					for _, v := range sidebar {
						message += fmt.Sprintf("name:%s\n", v.Name)
					}
					message += "\n新的:\n"
					for _, v := range newModule[module] {
						message += fmt.Sprintf("name:%s\n", v.Name)
					}
					_ = s.monitorDao.Send(c, message)
					continue
				}
				log.Info("sidebar %s module(%d) not change", platToString[plat], module)
			}
		}
		s.sideBars = tmpSideBars
	}
}

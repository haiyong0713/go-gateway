### esports

#### Version 1.23.2
##### Features
> 1. 奥运赛程配置

#### Version 1.23.1
##### Features
> 1. web 专题页api
> 2. fix 大数据页修复数据失败
> 3. 一键订阅过滤已过期赛程

#### Version 1.23.0
##### Features
> 1. fix mid check

#### Version 1.23.0
##### Features
> 1. 下掉s10接口

#### Version 1.22.3
##### Features
> 1. 增加赛事专题页接口

#### Version 1.22.2
##### Features
> 1. 吃鸡类比赛-fix排序规则

#### Version 1.22.1
##### Features
> 1. 吃鸡类比赛-信息流赛程组件

#### Version 1.22.0
##### Features
> 1. dev config

#### Version 1.21.5
##### Features
> 1. 赛季下的战队信息加uid字段

#### Version 1.21.4
##### Features
> 1. 树状图功能

#### Version 1.21.3
##### Features
> 1. /x/esports/season/teams/info忽略待定队伍

#### Version 1.21.2
##### Features
> 1. add SubContestUsersV2 grpc 方法
> 2. 直播同学联系：@溜哥 @魅影

#### Version 1.21.1
##### Features
> 1. 阶段支持积分表

#### Version 1.21.1
##### Features
> 1. 增加赛事下的赛季信息接口
> 2. 增加赛季下战队信息接口

#### Version 1.21.0
##### Features
> 1. fix fold api bug

#### Version 1.20.3
##### Features
> 1. 赛季预测实例切换

#### Version 1.20.2
##### Features
> 1. fix WLNum back

#### Version 1.20.1
##### Features
> 1. fix ldSeasonGame panic

#### Version 1.20.0
##### Features
> 1. 删除重复metric定义,增加回源相关metric
> 2. 修复不适当的错误处理

#### Version 1.19.0
##### Features
> 1. 战队详情页优化

#### Version 1.18.372
> 1. 修复可能的链接泄露问题
#### Version 1.18.371
> 1. 修复查询赛季下战队回源DB后接口返回500的问题
> 2. 查询不到正在进行中的赛季时, 后台同步协程保持运行
#### Version 1.18.37
> 1. 增加赛季下战队的查看功能
> 2. 正在进行的赛季战队信息会被缓存到内存.
#### Version 1.18.364
> 1. 结束赛程按照结束时间降序排列

#### Version 1.18.363
> 1. 赛程热点数据增加最近七天的赛程信息

#### Version 1.18.362
> 1. 缓存prometheus调整

#### Version 1.18.361
> 1. prometheus重复注册

#### Version 1.18.36
> 1. 新缓存方案优化：live Contest

#### Version 1.18.35
> 触发缓存更新时,不使用http context.(http context会被cancel)
> 缓存更新增加重试逻辑, 并新增对应metrics
#### Version 1.18.34
##### Features
* 去除infov2的强制登陆逻辑.
* 赛程没有配置竞猜时, 不请求竞猜列表
#### Version 1.18.33
> 合并`/guess`和`/guess/match/record`, 作为新的api提供给前端.
>  * 变更原因: 当前前端渲染一个页面需要同时调用两个api, 合并后可节省前端成本
>  * 方案: 原有api实现不做代码变更, 避免扩大影响面  

> 新api所涉及的model优化缓存使用逻辑
>  * update db时主动更新redis
>  * 适当的调整缓存过期时间
>  * 原有api和新api的缓存对象隔离,互相不影响(不同的key)
#### Version 1.18.32
##### Features
> 1. 直播更多赛程固定取未来四天数据

#### Version 1.18.31
##### Features
> 1. 修复nil变量问题

#### Version 1.18.3
##### Features
> 1. live contest增加内存缓存
> 2. 添加hit/miss计数，便于track回源状态

#### Version 1.18.20
##### Features
> 1. 动态附加卡jump_url

#### Version 1.18.19
##### Features
> 1. 赛程定位再次调整
> 2. score统计增加第二/第三权重(比赛场次/关联id)

#### Version 1.18.18
##### Features
> 1. 修复赛事阶段定位bug

#### Version 1.18.17
##### Features
> 1. 战绩排名积分字段更新

#### Version 1.18.16
##### Features
> 1. live fix

>#### Version 1.18.15
##### Features
> 1. 直播web积分商城图区分下发

#### Version 1.18.14
##### Features
> 1. 收藏服务增加path级别熔断
> 2. 增加grpc限流

#### Version 1.18.13
##### Features
> 1. 修复live more默认定位问题

#### Version 1.18.12
##### Features
> 1. 修复live more问题

#### Version 1.18.11
##### Features
> 1. 修复TasksAndPoints转化

#### Version 1.18.10
##### Features
> 1. 战绩排名干预字段

#### Version 1.18.9
##### Features
> 1. 修复TasksAndPoints错误返回

#### Version 1.18.9
##### Features
> 1. 赛程天马订阅卡

#### Version 1.18.8
##### Features
> 1. 添加trace

#### Version 1.18.7
##### Features
> 1. 修复按钮引用错误

#### Version 1.18.6
##### Features
> 1. 忽略已收藏错误

#### Version 1.18.5
##### Features
> 1. 增加主库配置

#### Version 1.18.4
##### Features
> 1. 增加直播web banner

#### Version 1.18.3
##### Features
> 1. 修复批量收藏状态查询 + 编辑战队bugs

#### Version 1.18.2
##### Features
> 1. 直播统计字段全部设置为string给前端
> 2. 优化S10代码

#### Version 1.18.1
##### Features
> 1. 一键订阅

#### Version 1.17.12
##### Features
> 1. empty contest invalid range

#### Version 1.17.1
##### Features
> 1. 战绩排名增加干预信息

#### Version 1.16.9
##### Features
> 1. fix err code return about points and taskprogress

#### Version 1.16.8
##### Features
> 1. 赛程定位，补未来赛程卡bug fix

#### Version 1.16.7
##### Features
> 1. 新增直播标题

#### Version 1.16.6
##### Features
> 1. invalid index range

#### Version 1.16.5
##### Features
> 1. score分析排序产品业务调整

#### Version 1.16.5
##### Features
> 1.fix the login problem about points interface

#### Version 1.16.4
##### Features
> 1.fix analysis and poster guess bug

#### Version 1.16.3
##### Features
> 1. batch add fav biz

#### Version 1.16.2
##### Features
> 1. add guess list check biz

#### Version 1.16.1
##### Features
> 1. s10 live add last time

#### Version 1.16.0
##### Features
> 1. add s10 live api
> 2. add S10 biz

#### Version 1.15.3
##### Features
> 1. add guess identifier and link for OTT

#### Version 1.15.2
##### Features
> 1.ContestList grpc add game_stage1, 2

#### Version 1.15.1
##### Features
> 1.add game season list

#### Version 1.15.0
##### Features
> 1.add ott contests grpc no cache

#### Version 1.14.3
##### Features
> 1.fix match info bvid regexp

#### Version 1.14.2
##### Features
> 1.del archive goprc model

#### Version 1.14.1
##### Features
> 1.contest list improve 

#### Version 1.14.0
##### Features
> 1.ott add game map grpc

#### Version 1.13.1
##### Features
> 1.fix leida game

#### Version 1.13.0
##### Features
> 1.del fav contest 15 day limit
> 2.add es time out config

#### Version 1.12.2
##### Features
> 1.fix contest list stime etime

#### Version 1.12.1
##### Features
> 1.add contest list gprc

#### Version 1.12.0
##### Features
> 1.tv OTT add games grpc

#### Version 1.11.1
##### Features
> 1.use favorite service ecode

#### Version 1.11.0
##### Features
> 1.修改ecode

#### Version 1.10.0
##### Features
> 1.bvid

#### Version 1.9.1
##### Features
> 1.season add full logo

#### Version 1.9.0
##### Features
> 1.add search card intervene

#### Version 1.8.0
##### Features
> 1.add s9 guess

#### Version 1.7.0
##### Features
> 1.matchs list add s9 cache

#### Version 1.6.3
##### Features
> 1.add contest guess

#### Version 1.6.2
##### Features
> 1.use go-main ecode

#### Version 1.6.1
##### Features
> 1.rpc SubContestUsers struct

#### Version 1.6.0
##### Features
> 1.del fav go-common

#### Version 1.5.4
##### Features
> 1.special player api

#### Version 1.5.3
##### Features
> 1.special teams sort
> 2.team add reply id
> 3.fix grpc 404 

#### Version 1.5.2
##### Features
> 1.fix is fav

#### Version 1.5.1
##### Features
> 1.add live push 
> 2.add lol dota 战队专题页接口

#### Version 1.4.1
##### Features
> 1.add overwatch 比赛数据页接口
> 2.lol,dota2,overwatch from db

#### Version 1.4.0
##### Features
> 1.直播对接电竞

#### Version 1.3.7
##### Features
> 1.大数据页接口

#### Version 1.3.6
##### Features
> 1.fix search act sid 

#### Version 1.3.5
##### Features
> 1.fix search sid 

#### Version 1.3.4
##### Features
> 1.fix point live
> 2.live add cache

#### Version 1.3.3
##### Features
> 1.活动页优化

#### Version 1.3.2
##### Features
> 1.修复第三方API英雄版本缺失问题

#### Version 1.3.1
##### Features
> 1.赛程info接口game status

#### Version 1.3.0
##### Features
> 1.接雷达积分数据

#### Version 1.2.9
##### Features
> 1.bug fix fav list

#### Version 1.2.8
##### Features
> 1.修复积分赛API倒序排序错误

#### Version 1.2.7
##### Features
> 1.电竞赛事库1.2 增加H5配置

#### Version 1.2.6
##### Features
> 1.活动赛事顶部赛程接口 fix time 

#### Version 1.2.4
##### Features
> 1.电竞赛事库1.2

#### Version 1.2.3
##### Features
> 1.订阅赛选接口fix null

#### Version 1.2.2
##### Features
> 1.添加APP订阅列表接口

#### Version 1.2.1
##### Features
> 1.add internal

#### Version 1.2.0
##### Features
> 1.添加订阅和取消订阅
> 2.添加app赛程、赛季接口

#### Version 1.1.2
##### Features
> 1.修改remoteip方法

#### Version 1.1.1
##### Features
> 1.赛程加比赛中状态

#### Version 1.1.0
##### Features
> 1.筛选联动

#### Version 1.0.1
##### Features
> 1.筛选联动

#### Version 1.0.0
##### Features
> 1.初始化项目  
> 2.新增赛事库相关接口

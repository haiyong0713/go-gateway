#### App channel 移动端运营类接口
#### Version 4.1.23
>1. 图文说明字段修复

#### Version 4.1.22
>1.【频道-知识树】知识点增加图文说明及埋点优化

#### Version 4.1.21
1.修复频道详情页panic

#### Version 4.1.20
1.频道新增电影频道页跳转

#### Version 4.1.19
1.【百科】百科详情页顶部tab调整

#### Version 4.1.18
1.【百科】【频道】百科详情页

#### Version 4.1.16
1.【稿件服务】app-channel对接稿件禁止项迁移

#### Version 4.1.15
1. 频道mid int64过滤

#### Version 4.1.14
1. 去掉proxy_handler

#### Version 4.1.13
1. 频道话题中心添加缩略图

#### Version 4.1.12
1. mine接口版本判断采用libbdevice替换

#### Version 4.1.11
1. 频道话题中心--话题上线版本

#### Version 4.1.10
1. 注释掉话题未上线部分

#### Version 4.1.9
1. 新话题频道修改版本变更为6.47

#### Version 4.1.8
1. /x/v2/channel/home、/x/v2/channel/tab3、/x/v2/channel/detail 订阅 =》 收藏

#### Version 4.1.7
> 1. 频道话题中心接口增加我的话题入口

#### Version 4.1.6
1. /v2/channel/detail 订阅 =》 收藏
2. 待新话题一起上线

#### Version 4.1.5
> 1. 错误码转 status code

#### Version 4.1.4
> 1. 频道新主页细节修改

#### Version 4.1.3
> 1. 简体转繁体v2

#### Version 4.1.2
> 1. 修复issure

#### Version 4.1.1
> 1. 频道话题合并主页

#### Version 4.1.0
> 1. 版本控制迁移

#### Version 4.0.0
> 1. 全量issues修改

##### Version 3.0.2
> 1. 下线go-main关系链client

##### Version 3.0.1
> 1. 使用通用降级方法
> 2. 接入standby

##### Version 3.0.0
> 1. 新秒开接口

##### Version 2.12.1
> 1. lint bugs修复

##### Version 2.12.0
> 1. 调活动接口增加id分组逻辑

##### Version 2.11.0
> 1. 频道广场页增加重试机制

##### Version 2.10.7
> 1.话题活动grpc接口迁移

##### Version 2.10.6
> 1.tag接口迁移至gRPC

##### Version 2.10.5
> 1.重新encode的时候空格变成了+号问题修复

##### Version 2.10.4
> 1.自动翻译繁体功能
> 2.翻译功能版本控制

##### Version 2.10.3
> 1.频道列表秒开修复

##### Version 2.10.2
> 1.拜年祭跳转详情页地址改造

##### Version 2.10.1
> 1.删除所有非middleware处理的秒开参数
> 2.删除明确的未被使用的变量

##### Version 2.10.0
> 1.使用archive.middleware传递秒开参数

##### Version 2.9.19
> 1.删除无用代码

##### Version 2.9.18
> 1.历史进度seek

##### Version 2.9.17
> 1.特殊分区tab限制

##### Version 2.9.16
> 1.国际版出秒开

##### Version 2.9.15
> 1.审核模式分区增加ipad hd数据

##### Version 2.9.14
> 1.分区ipad和ipad hd数据拆分

##### Version 2.9.13
> 1.频道分区ios国际版审核模式

##### Version 2.9.12
> 1.频道mine接口增加国际版版本判断

##### Version 2.9.11
> 1.广场页和我的关注页活动tag更换图标和跳转逻辑

##### Version 2.9.10
> 1.文案删除实验逻辑 

##### Version 2.9.9
> 1.频道广场页和红点接口全量切新策略接口 

##### Version 2.9.8
> 1.国际版频道详不出话题tab   

##### Version 2.9.7
> 1.秒开免流下发

##### Version 2.9.6
> 1.订阅频道页优化 
> 2.全部频道label文案修改
> 3.我的订阅和全部频道下发顶部title

##### Version 2.9.5
> 1.频道排序接口指向sh001   

##### Version 2.9.4
> 1.删除archive-service的GoRPC依赖 

##### Version 2.9.3
> 1.详情页顶部增加OGV开关  

##### Version 2.9.2
> 1.广场页换mychannels接口  
> 2.红点接口更新  

##### Version 2.9.1
> 1.旧频道默认图更换  

##### Version 2.9.0
> 1.定时任务迁移cron 

##### Version 2.8.8
> 1.频道详情OGV卡增加服务端开关 

##### Version 2.8.7
> 1.bugfix排行榜卡无详情的问题  

##### Version 2.8.6
> 1.account gorpc 多余代码删除

##### Version 2.8.5
> 1.风控-秒开逻辑修改  

##### Version 2.8.4
> 1.新增客户端埋点上报字段  

##### Version 2.8.3
> 1.广场页文案修改：添加频道->添加订阅  

##### Version 2.8.2
> 1.分享接口文案修改  

##### Version 2.8.1
> 1.广场页接口增加版本控制  
> 2.预防panic  

##### Version 2.8.0
> 1.全部频道tab增加我的订阅  
> 2.我订阅的频道增加最近浏览过  
> 3.广场页我订阅的频道重构  
> 4.广场页我订阅的更新增加频道管理  
> 5.广场页新增我看过的频道  
> 6.广场页新增热门频道模块(热门频道+频道动态)  
> 7.详情页接口返回夜间模式色值  

##### Version 2.7.5
> 1.角标逻辑重写

##### Version 2.7.4
> 1.我看过的频道、推荐频道修改跳转地址  

##### Version 2.7.3
> 1.频道页的分区入口屏蔽判断

##### Version 2.7.2
> 1.广场页不出推荐频道  

##### Version 2.7.1
> 1.分区列表审核模式从黑名单改成白名单

##### Version 2.7.0
> 1.我订阅的更新文案修改  
> 2.增加广场页补充接口：最近浏览过的频道+历史+推荐频道  
> 3.增加推荐频道接口  
> 4.广场页上报增加pos  
> 5.详情页增加PGC卡  

##### Version 2.6.0
> 1.旧频道详情页相关tag增加新旧频道判断逻辑 

##### Version 2.5.0
> 1.频道分享接口  

##### Version 2.4.1
> 1.修复排行榜卡上报类型字段  

##### Version 2.4.0
> 1.android端xx版本，全部频道改为第一条数据*2  

##### Version 2.3.11
> 1.archive gorpc迁移grpc 

##### Version 2.3.10
> 1.优化广场页我订阅的频道文案  

##### Version 2.3.9
> 1.分区粉和蓝拆分

##### Version 2.3.8
> 1.排行榜卡  
> 2.详情页头图增加透明度  

##### Version 2.3.7
> 1.location grpc 删除go-main引用

##### Version 2.3.6
> 1.去掉广场页数据校验  
> 2.已投币增加参数校验

##### Version 2.3.5
> 1.location grpc

##### Version 2.3.4
> 1.精选增加按年份筛选  
> 2.自定义频道增加查看更多  
> 3.增加已投币、已收藏逻辑  
> 4.广场页我的订阅增加红圈更新提醒  
> 5.频道详情增加父子频道逻辑  
> 6.接入秒开(版本限制5.50)
> 7.优化全部频道和频道历史下发文案间距  

##### Version 2.3.3
> 1.频道精选角标换底图  

##### Version 2.3.2
> 1.新频道上报增加auto_refresh字段  
> 2.未登录状态，广场页不请求我订阅的更新和频道历史  
> 3.综合-播放最多、最新投稿屏蔽角标 

##### Version 2.3.1
> 1.新频道一期不下发秒开

##### Version 2.3.0
> 1.增加新频道逻辑  

##### Version 2.2.8

> 1.切grpc

##### Version 2.2.7

> 1.分区修改

##### Version 2.2.6

> 1.审核模式下去掉vlog

##### Version 2.2.5

1.ipad hd三点面板版本兼容

##### Version 2.2.4

> 1.分区去重修复

##### Version 2.2.3

> 1.青少年模式

##### Version 2.2.2

> 1. 迁移点赞、硬币grpc

##### Version 2.2.1

> 1.推荐获取房间信息接口切换

##### Version 2.2.0

> 1.pgc卡片样式调整
> 2.卡片事件新增event_v2

##### Version 2.1.22

> 1.国际版秒开版本修改

##### Version 2.1.21

> 1.account grpc
> 2.relation grpc

##### Version 2.1.20

> 1.稿件秒开层级修改

##### Version 2.1.19

> 1.频道广场页增加infoc上报
> 2.频道详情页上报增加from_spmid
> 3.频道详情页上报增加from_page

##### Version 2.1.18

> 1.单双列卡片增加点赞开关修改

##### Version 2.1.17

> 1.单双列卡片增加点赞开关修改

##### Version 2.1.16

> 1.单双列卡片增加点赞

##### Version 2.1.15

> 1.Archive3 改成 Arc

##### Version 2.1.14

> 1.拜年祭秒开版本限制

##### Version 2.1.13

> 1.蓝强转粉逻辑修改

##### Version 2.1.11

> 1.蓝强转粉逻辑修改

##### Version 2.1.9

> 1.频道蓝强转粉

##### Version 2.1.8

> 1.审核版本屏蔽原创排行榜入口

##### Version 2.1.7

> 1.新接口多机房走云立方

##### Version 2.1.6

> 1.频道tablist

##### Version 2.1.4

> 1.PGC卡片统一epid

##### Version 2.1.3

> 1.ecode init

##### Version 2.1.2

> 1.广场频道数量配置

##### Version 2.1.1

> 1.频道运营卡片buvid缓存

##### Version 2.1.0

> 1.去掉zlimit，接入location 

##### Version 2.0.21

> 1.频道分类国际版
> 2.分区海外默认繁体

##### Version 2.0.20

> 1.频道广场推荐频道改为4个

##### Version 2.0.19

> 1.频道广场推荐频道改为3个

##### Version 2.0.18

> 1.审核模式屏蔽排行榜入口

##### Version 2.0.17

> 1.频道海外版

##### Version 2.0.16

> 1.GRPC Panic

##### Version 2.0.15

> 1.UGC付费卡片

##### Version 2.0.14

> 1.PGC卡片展示后台配置的标题

##### Version 2.0.13

> 1.广场页接口拆除我的订阅，卡片新增from_type字段

##### Version 2.0.12

> 1.修复infoc上报问题

##### Version 2.0.11

> 1.上报增加字段

##### Version 2.0.10

> 1.新接口上报修改

##### Version 2.0.9

> 1.老接口频道单推卡片URI问题

##### Version 2.0.8

> 1.新增PGC卡片、单推UP卡片
> 2.新增头图卡片

##### Version 2.0.7

> 1.频道广场页加更新类型参数

##### Version 2.0.6

> 1.频道广场页改版，我的订阅换新接口

##### Version 2.0.5

> 1.稿件UP头像和用户名数据源修改

##### Version 2.0.4

> 1.老接口频道id大于0，频道name置为空

##### Version 2.0.3

> 1.频道id大于0，频道name置为空

##### Version 2.0.2

> 1.专栏分区schema换成新地址

##### Version 2.0.1

> 1.频道list游戏中心写死from

##### Version 2.0.0

> 1.频道卡片重构

##### Version 1.2.7

> 1.频道详情页返回内容修改

##### Version 1.2.6

> 1.频道详情页返回code

##### Version 1.2.5

> 1.修复build限制问题

##### Version 1.2.4

> 1.分区build bug

##### Version 1.2.3

> 1.替换c.remoteip获取

##### Version 1.2.2

> 1.autoplay

##### Version 1.2.1

> 1.返回error修复

##### Version 1.2.0

> 1.使用grpc auth

##### Version 1.1.12

> 1.频道卡片增加cid

##### Version 1.1.11

> 1.直播卡片横竖屏

##### Version 1.1.10

> 1.分区去重修改

##### Version 1.1.9

> 1.fix avHandler

##### Version 1.1.8

> 1.av卡片展示desc为弹幕数

##### Version 1.1.7

> 1.UP三连推BUG修复

##### Version 1.1.6

> 1.频道稿件卡片返回竖屏URI信息
> 2.频道详情页关注三连卡片增加关注状态判断

##### Version 1.1.5

> 1.卡片和AI视频去重

##### Version 1.1.4

> 1.infoc上报修改

##### Version 1.1.3

> 1.分区多区间限制

##### Version 1.1.2

> 1.卡片增加字段
> 2.频道列表接口修改

##### Version 1.1.1

> 1.增加template

### Version 1.1.0

> 1. update infoc sdk

##### Version 1.0.1

> 1.分区去重处理

##### Version 1.0.0

> 1.项目初始化 

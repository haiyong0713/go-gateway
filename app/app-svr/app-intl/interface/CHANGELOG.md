### app-intl version
# v1.1.50
> 1. app-intl对接稿件禁止项迁移增加分片调用

# v1.1.49
> 1.【稿件服务】app-intl对接稿件禁止项迁移

# v1.1.48
> 1.logstream上报-000025, 000027

# v1.1.46
> 1.下线logstream上报-000026

# v1.1.45
> 1.繁体版详情页和天马logstream迁移

# v1.1.43
> 1.数平logstream迁移

# v1.1.42
> 1.相关推荐增加zone_id传参

# v1.1.41
> 1.隐藏稿件page接口 attribute

# v1.1.40
> 1.隐藏稿件 attribute

# v1.1.39
> 1.app-intl内网SLB核心调用迁移到Discovery

# v1.1.38
> 1.history move to bapis

# v1.1.37
> 1.修复异味

# v1.1.36
> 1.删除gorpc

# v1.1.35
> 1.delete unused creative grpc

# v1.1.34
> 1.app-intl接chinese.v2-v1.1

# v1.1.33
> 1.将go-main的dm迁移到bapis

# v1.1.32
> 1.app-intl接chinese.v2

# v1.1.31
> 1.下线go-main关系链client

# v1.1.30
> 1.多p支持秒开

# v1.1.29
> 1.lint bugs修复

# v1.1.28
> 1.秒开参数优化

# v1.1.27
> 1.tag goRPC切gRPC

# v1.1.26
> 1.热门tag请求迁移至热门服务化

# v1.1.25
> 1.下线泰国包天马infoc双写

# v1.1.24
> 1.天马请求ai加timeout参数

# v1.1.23
> 1.活动gorpc切换到grpc

# v1.1.22
> 1.去掉无用的上报代码

# v1.1.21
> 1.天马展示上报双写
>
# v1.1.20
> 1.app-intl fix android loss banner_v5

# v1.1.19
> 1.intl feed/index/tab代码intl自己处理

# v1.1.18
> 1.intl 增加下发ogv_play_url字段

# v1.1.17
> 1.详情页展示剧集
> 2.同步粉版秒开
> 3.详情页同步粉dislikeV2

# v1.1.16
> 1.国际版搜索修改zoneid逻辑

# v1.1.15
> 1.天马列表添加bangumi和pgc卡片

# v1.1.14
> 1.vip确认数据源2年内无数据，删除vip"/internal/v1/notice/active"调用

# v1.1.13
> 1.优化繁体判断逻辑

# v1.1.12
> 1.isHant判断修改

# v1.1.11
> 1.isHant判断修改

# v1.1.10
> 1.搜索稿件标题、描述、推荐理由、用户认证信息简转繁

# v1.1.9
> 1.审核模式去掉banner

# v1.1.8
> 1.天马国际版增加审核模式

# v1.1.7
> 1.国际版搜索PGC逻辑同步粉版本逻辑

# v1.1.6
> 1.国际版支持互动视频

# v1.1.5
> 1.国际版同步粉版详情页tag逻辑

# v1.1.4
> 1.国际版增加Fawkes逻辑

# v1.1.3
> 1.国际版版本号限制修改

# v1.1.2
> 1.国际版入口屏蔽

# v1.1.1
> 1.修复不雅之词

# v1.1.0
> 1.删除稿件二级缓存的memcache
> 2.删除废弃代码

# v1.0.52
> 1.弹幕接口迁移grpc

# v1.0.51
> 1.综合搜索和垂搜增加回粉  

# v1.0.50
> 1.删除arc gorpc相关代码

# v1.0.49
> 1.删除displayID

# v1.0.48
> 1.联合投稿商业样式
> 2.风控-秒开逻辑修改 

# v1.0.47
> 1.修复潜在的panic

# v1.0.46
> 1.dao层ut补全

# v1.0.45
> 1.修改投币model

# v1.0.44
> 1.取消联合投稿和ugc剧集互斥逻辑

# v1.0.43
> 1.推荐池禁止稿件屏蔽相关推荐列表

# v1.0.42
> 1.location grpc 删除go-main引用

# v1.0.41
> 1.搜索和详情页支持bvid

# v1.0.40
> 1.ugc-season接口从archive切为走season-ugc-service服务

# v1.0.39
> 1.playurl切grpc

# v1.0.38
> 1.删除playurl无用代码

# v1.0.37
> location gorpc to grpc

# v1.0.36
> 1.history gorpc to grpc

# v1.0.35
> steins-gate-service的proto调用去掉archive引用 

# v1.0.34
> 1.详情页限制pugv视频返回404

# v1.0.33
> 1.离线下载默认https

# v1.0.32
> 1.修复panic

# v1.0.31
> 1.pgc内容去掉对isMovie的判断

# v1.0.30
> 1.热门标签加跳转

# v1.0.29
> 1.清理不需要的cache

# v1.0.28
> 1.国际版需求同步到545

# v1.0.27
> 1.uat host conf

# v1.0.26
> 1.切grpc

# v1.0.25
> 1.三点面板下发老样式

# v1.0.24
> 1.coin路径修改

# v1.0.23
> 1. 同步playurl is_sp会员鉴权

# v1.0.22
> 1. 迁移点赞、硬币grpc

# v1.0.21
> 1.搜索文章suggest

# v1.0.20
> 1.卡片事件新增event_v2

# v1.0.19
> 1.秒开增加4k

# v1.0.18
> 1.国际版去掉秒开

# v1.0.17
> 1.国际版秒开版本修改

# v1.0.16
> 1.稿件秒开层级修改

# v1.0.15
> 1.account gorpc迁移到grpc

# v1.0.14
> 1.fix playicon content 

# v1.0.11
> 1.灾备缓存优化

# v1.0.10
> 1.fix panic

# v1.0.9
> 1.接入新的playicon接口  

# v1.0.9
> 1.国际版Banner

# v1.0.8
> 1.播放器控件增加临时增加初音逻辑  

# v1.0.7
> 1.删除无用的配置

# v1.0.6
> 1.详情页去掉投稿数
> 2.首页详情页支持繁体
> 3.清晰度支持大会员

# v1.0.5
> 1.搜索展示pgc卡片

# v1.0.4
> 1.去掉zlimit，接入location

# v1.0.3
> 1.修复搜索卡片的uri

# v1.0.2
> 1.国际版feed暂时不秒开

# v1.0.1
> 1.去掉详情页的开屏广告


# v1.0.0
> 1.上线国际版功能
git

### app-job

#### Version 2.5.39
> 1.app空间下掉相册功能

#### Version 2.5.38
> 1.修复闭包中引用循环变量的bug

#### Version 2.5.37
> 1.首页校园灾备

#### Version 2.5.36
> 1.热门分品类增加人群包设置
> 2.删除app-job中无用的mogul代码

#### Version 2.5.35
> 1.删除每周必看生成相关逻辑

#### Version 2.5.34
> 1.增加运动分区

#### Version 2.5.33
> 1.刷119期每周必看的荣誉榜单

#### Version 2.5.32
> 1.修复异味

#### Version 2.5.31
> 1.增加大佬白名单上报过滤

#### Version 2.5.30
> 1.迁移老投稿接口

#### Version 2.5.29
> 1.定时迁移更改数据结构

#### Version 2.5.28
> 1.一级分区汽车排行榜

#### Version 2.5.27
> 1.修复异味

#### Version 2.5.26
> 空间小视频接口下线

#### Version 2.5.25
> 修改解析job数据库解析错误

#### Version 2.5.24
> app-show定时获取db数据迁移到app-job中

#### Version 2.5.23
> 迁移up-archive接口

#### Version 2.5.22
> 1.lint bugs修复

#### Version 2.5.21
> 排行版新增动物圈一级分区

#### Version 2.5.20
> 1.每周必看推送和红点任务拆分到feed-admin

#### Version 2.5.19
> 1.tag接口迁移至gRPC

#### Version 2.5.18
> 1.每周必看推送加重试

#### Version 2.5.17
> 1.热门tag请求迁移至热门服务化

#### Version 2.5.16
> 1.增加每周必看消息日志

#### Version 2.5.15
> 1.修复tag proto重复定义的问题

#### Version 2.5.14
> 1.每周必看推送na_url至稿件荣誉

#### Version 2.5.13
> 1.每周必看补救时从redis中取数据

#### Version 2.5.12
> 1.每周必看推送新老接口加开关

#### Version 2.5.11
> 1.每周必看获取推送用户迁移至TagSub接口

#### Version 2.5.10
> 1.热门车载单独的redis key

#### Version 2.5.9
> 1.排行榜增加美食区

#### Version 2.5.8
> 1.新增每周必看刷新缓存日志告警

#### Version 2.5.7
> 1.榜单数据获取与定时缓存

#### Version 2.5.6
> 1.排行榜数据统一数据源

#### Version 2.5.5
> 1.更新push model

#### Version 2.5.4
> 1.s10订阅后台状态切换推送消息，发送broadcast更新首页入口

#### Version 2.5.3
> 1.fawkes laser的mobi_app换字段

#### Version 2.5.2
> 1.fawkes laser增加mobi_app

#### Version 2.5.1
> 1.同步热门module_id=hot-topic的ai数据

#### Version 2.5.0
> 1.删除所有稿件二级缓存

#### Version 2.4.14
> 1.fix mogul return

#### Version 2.4.13
> 1.fix ecode equal

#### Version 2.4.12
> 1.每周必看发送消息给ott

#### Version 2.4.11
> 1.静默推送mid和buvid二选一且buvid优先    

#### Version 2.4.10
> 1.静默推送参数过滤  

#### Version 2.4.9
> 1.fawkes的laser增加静默推送任务线  

#### Version 2.4.8
> 1.fix水印图本地缓存空的问题   

#### Version 2.4.7
> 1.删除archive-service的GoRPC依赖

#### Version 2.4.6
> 1.更新bfs SDK  

#### Version 2.4.5
> 1.修复databus关闭的时候只对云立方关闭

#### Version 2.4.4
> 1.修复contributeSub databus关闭的时候只对云立方关闭

#### Version 2.4.3
> 1.app-show排行版缓存放入redis

#### Version 2.4.2
> 1. cron

#### Version 2.4.1
> 1. 修复异常尺寸热门稿件分享水印图  

#### Version 2.4.0
> 1. 增加load热门模块ai数据

#### Version 2.3.10
> 1. 热门稿件增加右上角和底部横幅两种水印图  

#### Version 2.3.8
> 1. mogul add duration field

#### Version 2.3.7
> 1. 修复潜在的panic

#### Version 2.3.6
> 1. 每周必看push文案支持配置

#### Version 2.3.5
> 1. 修改热点后台databus消费  

#### Version 2.3.4
> 1.分区定时上下线更新数据库

#### Version 2.3.3
> 1. mogul 消费

#### Version 2.3.2
> 1. 每周必看有更新自动前置

#### Version 2.3.1
> 1. 修复 ecode.Cause(err).Equal(ecode.NothingFound) 判断  

#### Version 2.3.0
> 1. 站外热点刷新增加过滤  

#### Version 2.2.3
> 1. 分品类热门AI数据缓存增加推荐理由

#### Version 2.2.2
> 1. 接入刷新站外热点的逻辑  

#### Version 2.2.1
> 1. 分品类热门AI数据写入缓存

#### Version 2.2.0
> 1. 每周必看链接区分推送链接和成就链接

#### Version 2.1.18
> 1. 入站必刷增加订阅binlog消息逻辑，刷新播单+发送变更消息给成就系统
> 2. 每周必看发送变更消息给成就系统

#### Version 2.1.18
> 1. 入站必刷增加订阅binlog消息逻辑，刷新播单+发送变更消息给成就系统
> 2. 每周必看发送变更消息给成就系统

#### Version 2.1.17
> 1. 每周必看发布流程失败则企业微信报警

#### Version 2.1.16
> 1. 限制contribute的更新频率

#### Version 2.1.15
> 1. 修复wechat报警参数传错

#### Version 2.1.14
> 1. contribute加日志区分来源

#### Version 2.1.13
> 1. 每周必看获取订阅列表改为游标获取

#### Version 2.1.12
> 1. 扩大flush并发数

#### Version 2.1.11
> 1. contribute cache local build and batch update

#### Version 2.1.10
> 1. 去掉未使用的redis

#### Version 2.1.9
> 1. 修复闭包问题

#### Version 2.1.8
> 1.bap hosts

#### Version 2.1.7
> 1. laser修改逻辑运算符  

#### Version 2.1.6
> 1. 热门稿件分享封面水印参数变更  

#### Version 2.1.5
> 1. 修复每周必看更新时module_id错误的问题

#### Version 2.1.4
> 1. 每周必看发布时，调接口自动更新红点 

#### Version 2.1.3
> 1. 热门稿件分享链接封面打水印  

#### Version 2.1.2
> 1. fawkes的laser增加expire和ack  

#### Version 1.9.1
> 1. 调整databus日志

#### Version 1.9.0
> 1. 热点后台AI同步databus信息

#### Version 1.8.9
> 1. 空间投稿增加漫画逻辑  

#### Version 1.8.8
> 1. 增加laser任务逻辑  

#### Version 1.8.7
> 1. 增加回源views接口失败的retry逻辑

#### Version 1.8.6
> 1. service关闭时增加逻辑关闭的判断，避免send on closed channel

#### Version 1.8.5
> 1. 更新view的二级缓存

#### Version 1.8.4
> 1. 消费河童子数据的databus在嘉定不初始化

#### Version 1.8.3
> 1. 双机房都在等到播单更新后再刷缓存&推送（云立方)

#### Version 1.8.2
> 1.热门每周必看删缓存操作改为调grpc刷新缓存
> 2.新增发布时生产播单和修改审核通过时修改播单的逻辑

#### Version 1.8.1
> 1.模块变更告警+env

#### Version 1.8.0
> 1. 新增热门每周精选企业微信警告和数据更新操作
> 2. 去掉service层的

#### Version 1.7.7
> 1.增加ai聚合卡片消费

#### Version 1.7.6
> 1. 新增app-cache的mc缓存配置，适配app-feed/intl读取的mc

#### Version 1.7.5
> 1. app-feed灾备缓存从打http接口改为存memcache
> 2. 补充dao层ut

#### Version 1.7.4

> 1.根据aid顺序消费，避免同时收到多条消息时执行顺序不一致

#### Version 1.7.3

> 1.个人空间多机房


#### Version 1.7.2

> 1.个人空间多机房回源

#### Version 1.7.1

> 1.使用account grpc

#### Version 1.7.0
> 1.增加监控sidebar

#### Version 1.6.19

> 1.账号昵称头像变更更新稿件优化

#### Version 1.6.18

> 1.账号昵称头像变更更新稿件

#### Version 1.6.17

> 1.历史排行榜计数更新

#### Version 1.6.6

> 1.历史排行榜计数更新

#### Version 1.6.5

> 1.flush使用批量接口初始化

#### Version 1.6.4

> 1.cpu num

#### Version 1.6.3

> 1.新增稿件缓存刷新功能

#### Version 1.6.2

> 1.稿件计数绝对值更新

#### Version 1.6.1

> 1.http bm

#### Version 1.6.0

> 1.stat切新topic

#### Version 1.5.11

> 1.空间全部投稿缓存换key

#### Version 1.5.10

> 1.fix空间投稿小视频

#### Version 1.5.9

> 1.空间投稿脏mid过滤

#### Version 1.5.8

> 1.补充单元测试

#### Version 1.5.7

> 1.job参数过滤
> 2.fix bug

#### Version 1.5.6

> 1.视频详情页用户投稿job

#### Version 1.5.5

> 1.fix retry

#### Version 1.5.4

> 1.空间投稿异常稿件剔除列表
> 2.error wrap

#### Version 1.5.3

> 1.stat切新databus

#### Version 1.5.2

> 1.去掉旧逻辑代码

#### Version 1.5.1

> 1.去除掉空间投稿中的异常稿件

#### Version 1.5.0

> 1.job自己处理稿件缓存

#### Version 1.4.2

> 1.修复全部投稿panic

#### Version 1.4.1

> 1.修复音频分页bug 

#### Version 1.4.0

> 1.空间全部投稿job重构

#### Version 1.3.8

> 1.修复投稿时间

#### Version 1.3.7

> 1.修复死循环

#### Version 1.3.6

> 1.忽略clip接口错误

#### Version 1.3.5

> 1.空间全量投稿忽略稿件-404返回码修复

#### Version 1.3.4

> 1.空间全量投稿忽略稿件-404返回码

#### Version 1.3.3

> 1.空间全量投稿
> 2.切新的httpclient

#### Version 1.3.2

> 1.切新的httpclient

#### Version 1.3.1

> 1.去掉新旧稿件判断，无脑更新

#### Version 1.3.0

> 1.稿件计数并发消费

#### Version 1.2.11

> 1.初始化稿件id缓存设置过期时间

#### Version 1.2.10

> 1.稿件缓存切新

#### Version 1.2.9

> 1.monitor ping

#### Version 1.2.8

> 1.热门推荐接口切新

#### Version 1.2.7

> 1.修复重试队列的退出

#### Version 1.2.6

> 1.view archive 初始化化功能

#### Version 1.2.1

> 1.archive notify 队列

#### Version 1.2.0

> 1.app-feed灾备缓存
> 2.app-show数据库定时同步

#### Version 1.1.2

> 1.增加sleep

#### Version 1.1.1

> 1.增加定时更新数据库操作

#### Version 1.1.0

> 1.改为同步更新

#### Version 1.0.5

> 1.增加state

#### Version 1.0.4

> 1.修复config和消费问题

#### Version 1.0.2

> 1.增加http monitor ping

#### Version 1.0.0

> 1.基于databus消费更新app-interface缓存

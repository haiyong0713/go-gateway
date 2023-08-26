#### Web show 网页端运营及广告类接口

#### Version 2.21.35
>1.商业接口上报稿件tag id

#### Version 2.21.34
>1.播放页新增对直播、活动、广告位、C位卡的屏蔽功能

#### Version 2.21.33
>1.【PC首页焦点图】首帧插帧逻辑修复

#### Version 2.21.32
> 1.商业接口请求结果打印

#### Version 2.21.31
> 1.脏逻辑，用黑白名单控制详情页特殊稿件的播放页不出投放
> 2.兼容aid请求详情页

#### Version 2.21.30
> 1.【埋点需求】投放数据治理（pc端）

#### Version 2.21.29
> 1.脏逻辑，用黑白名单控制详情页特殊稿件的播放页不出投放

#### Version 2.21.28
> 1.相关推荐支持投放稿件

#### Version 2.21.27
> 1.脏逻辑，详情页特殊稿件的播放页不出投放

#### Version 2.21.26
> 1.脏逻辑，详情页特殊稿件的播放页不出投放

#### Version 2.21.25
> 1.【PC首页改版】banner焦点图接入Inline能力

#### Version 2.21.24
> 1.透传前端from_spm_id给商业广告引擎

#### Version 2.21.23
> 1.移除NothingFound错误的降级处理

#### Version 2.21.22
> 1.网关透传字段ads_control给前端

#### Version 2.21.21
> 1.【【PC首页改版】Banner焦点图 后台改造】

#### Version 2.21.20
> 1.【PC-WEB】主站首页版头配置后台优化

#### Version 2.21.19
> 1.【PC-WEB】主站首页版头配置后台优化

#### Version 2.21.10
> 1.PC端请求广告接口增加ua字段

#### Version 2.21.9
> 1.把videoup迁移到bapis

#### Version 2.21.8
> 1.消除异味

#### Version 2.21.7
> 1.lint bugs修复

#### Version 2.21.6
> 1.http status 500，slb跨机房重试

#### Version 2.21.5
> 1.焦点图增加空帧标识

#### Version 2.21.4
> 1.banner 去掉空帧逻辑 回滚

#### Version 2.21.3
> 1.banner 去掉空帧逻辑

#### Version 2.21.2
> 1.banner id 配置化

#### Version 2.21.1
> 1.广告接口，播放页增加参数，防止特殊稿件出广告

#### Version 2.21.0
> 1.go-common 更新

##### Version 2.20.9
> 1.loc，locs接口接口新增广告主adver_name字段

##### Version 2.20.8
> 1.增加版接口

##### Version 2.20.7
> 1.loc，locs接口新增ad_desc字段

##### Version 2.20.6
> 1.稿件隐藏attribute相关字段

##### Version 2.20.5
> 1.修复固定投放历史遗留老坑(指针传递)

##### Version 2.20.4
> 1.版头增加美食区

##### Version 2.20.3
> 1.use cron

##### Version 2.20.2
> 1.loc和locs增加字段判断是否不请求广告   

##### Version 2.20.1
> 1.增加副标题    

##### Version 2.20.0
> 1.实现强运营帧逻辑    

##### Version 2.19.1
> 1.删除空中课堂测试日志    

##### Version 2.19.0
> 1.上海地区版头空中课堂    

##### Version 2.18.0
> 1.acccount grpc 迁移  

##### Version 2.17.1
> 1.banner部分增加防panic判断  

##### Version 2.17.0
> 1.请求广告用discovery

##### Version 2.16.0
> 1.archive/relation 接口接入开关管控能力

##### Version 2.15.0
> 1.loc接入bv开关

##### Version 2.14.0
> 1.迁移archive grpc    

##### Version 2.13.1
> 1.现有支持投放直播的位置都增加拉去直播间信息的逻辑  

##### Version 2.13.0
> 1.location切grpc接口  

##### Version 2.12.11
> 1.资源位支持直播间数据获取 

##### Version 2.12.10
> 1.内容运营加锁，防止稿件和付费同时投放导致并发读写  

##### Version 2.12.9
> 1.内容运营增加OGV付费跳转类型  

##### Version 2.12.8

> 1.修复硬编码  

##### Version 2.12.7

> 1.修复编译错误  

##### Version 2.12.6

> 1.删除ping接口  

##### Version 2.12.5

> 1.运营位根据活动时间聚合活动状态  

##### Version 2.12.4

> 1.res增加数码分区版头位置ID  

##### Version 2.12.3

> 1.接入location RPC IP查询接口  

##### Version 2.12.2

> 1.修复移动端内容运营位无法获取稿件信息的问题  

##### Version 2.12.1

> 1.移动端内容运营位增加获取稿件信息逻辑  

##### Version 2.12.0

> 1.auth grpc
> 2.bm 启动修改
> 1.ip 传入修改

##### Version 2.11.0

> 1.添加mid

##### Version 2.10.2

> 1.添加相关  creative_type

##### Version 2.10.1

> 1.素材添加 stime

##### Version 2.10.0

> 1.接入话题

##### Version 2.9.2

> 1.使用discovery接入Resource

##### Version 2.9.1

> 1.使用discovery接入archive

##### Version 2.9.0

> 1.使用discovery接入account

##### Version 2.8.0

> 1.move to main

##### Version 2.7.3

> 1.调整固定投放、推荐池、广告优先级的

##### Version 2.7.2

> 1.使用account-service v7

##### Version 2.7.1

> 1.删除 statsd

##### Version 2.7.0

> 1.http参数修改为binging

##### Version 2.6.0

> 1.广告改为插入逻辑

##### Version 2.5.0

> 1.http切换为bm

##### Version 2.4.1

1.add new banner
2.fix pinc

##### Version 2.4.0

> 1.广告接口，透传buvid
> 2.相关视频接入大数据接口
> 3.archive3接入

##### Version 2.3.6

> 1.熔断报警 http

##### Version 2.3.5

> 1.全局版头

##### Version 2.3.4

> 1.prom 监控添加

##### Version 2.3.2

> 1.cpm接口改为GET接口，dao层

##### Version 2.3.1

> 1.添加广告上报字段

##### Version 2.3.0

> 1.接入新版配置中心

##### Version 2.2.0

> 1.为ad 到添加ping方法检测
> 2.支持根据平台查询后台url配置列表监控

##### Version 2.1.0

> 1.新增接口返回全量运营广告的url

##### Version 1.10.0

> 1.更新vendor
> 2.接入docker

##### Version 1.9.4

> 1.更新vendor
> 1.增加国创区默认版头配置

##### Version 1.9.3 

> 1.修改资源位https

##### Version 1.9.2

> 1.更新govendor

##### Version 1.9.1

> 1.更新govendor

##### Version 1.9.0

> 1.接入熔断
> 2.更新vendor依赖

##### Version 1.8.1

> 1.修改静态降级参数
> 2.兼容推荐视频

##### Version 1.8.0

> 1.相关视频推荐接入cpm
> 2.接入引擎cpm&cpt聚合
> 3.升级vendor使用identify
> 4.增加内容运营为agency字段

##### Version 1.7.1

> 1.修复video ad

##### Version 1.7.0

> 1.更新go-common依赖

##### Version 1.6.8

> 1.修复跨域问题

##### Version 1.6.7

> 1.更新govendor依赖

##### Version 1.6.6

> 1.更新govendor 依赖
> 2.删除多余错误日志

##### Version 1.6.5

> 1.修复manage运营内容位置乱序

##### Version 1.6.4

> 1.添加cpm全局开关

##### Version 1.6.3

> 1.修复缓存被修改bug

##### Version 1.6.2

> 1.CPT&CRM 表结构优化
> 2.番剧贴片广告对大会员不展示

##### Version 1.6.1

> 1.判断所有请求area
> 2.设置manager指定内容为广告

##### Version 1.6.0

> 1.增加cpm广告

##### Version 1.5.0

> 1.增加govendor支持
> 2.兼容crm广告数据

##### Version 1.4.0

> 1.公共资源添加文字链及版头

##### Version 1.3.2

> 1.PGC移动端广告数据处理

##### Version 1.3.1

> 1.增加广告id唯一标识

##### Version 1.3.0

> 1.接入trace2
> 2.接入新的router
> 3.批量查询公共资源接口

##### Version 1.2.1

> 1.修复缓存bug

##### Version 1.2.0

> 1.统一公共资源管理

##### Version 1.1.0

> 1.添加广告接口

##### Version 1.0.1

> 1.优化配置
> 2.添加服务发现

##### Version 1.0.0

> 1.初始化完成招聘信息查询接口
> 2.初始化完成通告信息查询接口

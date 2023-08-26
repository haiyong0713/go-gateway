#### 稿件字段同步
#### Version 2.15.50
> 1.查询首映和付费稿件信息dbfix

#### Version 2.15.49
> 1.sql接入api

#### Version 2.15.48
> 稿件投稿属地转化

#### Version 2.15.47
> 1.paladin v2迁移
> 新增稿件投稿属地

#### Version 2.15.46
> 1.稿件缩略图处理日志

#### Version 2.15.45
> 1.付费稿件

#### Version 2.15.44
> 1.竖版缩略图

#### Version 2.15.43
> 1.首映风控文案

#### Version 2.15.42
> 1.稿件禁止重试队列告警增加数量限制
 
#### Version 2.15.41
> 1.稿件禁止项目过滤oid小于等于0

#### Version 2.15.40
> 1.稿件禁止项目过滤oid为0

#### Version 2.15.39
> 1.首映稿件缓存写入fix

#### Version 2.15.38
> 1.首映稿件消息
> 2.首映稿件缓存更新

#### Version 2.15.37
> 1.获取稿件禁止项sign加密fix

#### Version 2.15.36
> 1.同步稿件存量禁止项数据job删除

#### Version 2.15.35
> 1.同步稿件存量禁止项数据

#### Version 2.15.34
> 1.同步稿件inner禁止项数据

#### Version 2.15.33
> 1.合集消息逻辑优化

#### Version 2.15.32
> 1.psb_{aid}_{cid}缓存增加过期时间

#### Version 2.15.31
> 1.psb_{aid}缓存增加过期时间
> 2.videoshot老key=vst_{cid}写入删除
> 3.desc老key=desc_{aid}写入删除

#### Version 2.15.30
> 1.支持高清缩略图

#### Version 2.15.29
> 1.a3p_{aid}缓存增加过期时间

#### Version 2.15.25
> 1.ff消息需要更新view_cache

#### Version 2.15.24
> 1.修复bug

#### Version 2.15.23
> 1.video_up接rail_gun

#### Version 2.15.22
> 1.稿件首帧封面逻辑

#### Version 2.15.21
> 1.接rail_gun

#### Version 2.15.20
> 1.databus降级接口

#### Version 2.15.19
> 1.修复异味

#### Version 2.15.18
> 1.修复short_link字段和up_from字段

#### Version 2.15.17
> 1.新增投稿来源up_from同步

#### Version 2.15.16
> 1.引入quota保护db

#### Version 2.15.15
> 1.删除 archive-job 依赖评论接口

#### Version 2.15.15
> 1.fix del video cache 

#### Version 2.15.14
> 1.fix send notify retry msg

#### Version 2.15.13
> 1.删除弹幕计数逻辑

#### Version 2.15.12
> 1.年报视频独立chanel处理

#### Version 2.15.11
> 1.修改非正常稿件告警条件

#### Version 2.15.10
> 1.删除历史遗留的废弃代码

#### Version 2.15.9
> 1.分区名称不写入缓存

#### Version 2.15.8
> 1.fix simple arc cache 

#### Version 2.15.7
> 1.初始化分P的taishan数据

#### Version 2.15.6
> 1.所有稿件缓存的key使用统一model

#### Version 2.15.5
> 1.增加batchDel泰山的逻辑

#### Version 2.15.4
> 1.加速泰山初始化的速度

#### Version 2.15.3
> 1.增加初始化泰山逻辑

#### Version 2.15.2
> 1.加快初始化泰山的速度
> 2.videoshot初始化不再往redis里塞，改为串行执行

#### Version 2.15.1
> 1.指定泰山为本机房

#### Version 2.15.0
> 1.使用taishan双写稿件缓存数据

#### Version 2.14.10
> 1.区分videoshot和arc的retry list告警阈值

#### Version 2.14.9
> 1.重新校验videoshot合法性

#### Version 2.14.8
> 1.setcache忽略pages=nil

#### Version 2.14.7
> 1.增加SimpleArc缓存逻辑

#### Version 2.14.6
> 1.增加up投稿列表过滤规则

#### Version 2.14.5
> 1.删除账号系统databus的消费代码
> 2.去除灰度配置由job闭环更新稿件缓存的代码
> 3.写缓存时不再写入账号昵称与头像

#### Version 2.14.4
> 1.修改打日志时mtime的逻辑 取databus的msg

#### Version 2.14.3
> 1.修改打日志时mtime的逻辑

#### Version 2.14.2
> 1.优化error日志
> 2.细化耗时打点

#### Version 2.14.1
> 1.迁移小仓

#### Version 2.14.0
> 1.将稿件缓存更新的逻辑全部收敛到job中
> 1.将up主投稿列表简化为一套redis
> 2.删除报警接口的调用，统一使用bilimoni平台自带的报警功能

#### Version 2.13.51
> 1.修复addit表同步问题

#### Version 2.13.50
> 1.初始化时如果非长简介则不入库

#### Version 2.13.49
> 1.同简介的不做插入

#### Version 2.13.48
> 1.同步B端稿件分区表到archive_result结果库

#### Version 2.13.47
> 1.从B端同步archive_addit表的长简介信息，使 service 层不再需要使用B端的数据库从库
> 2.增加初始化archive_addit表的功能，并支持断点续传

#### Version 2.13.46
> 1.将videoshot独立到单独redis重试队列

#### Version 2.13.45
> 1.增加消费B端videoshot_changed的逻辑，促使缩略图实时更新

#### Version 2.13.44
> 1.弹幕接口迁移grpc

#### Version 2.13.43
> 1.通过errgroup增加初始化速度

#### Version 2.13.42
> 1.增加初始化videoshot的速度

#### Version 2.13.41
> 1.优化普罗米修斯，避免打爆监控

#### Version 2.13.40
> 1.videoshot逻辑放到job中进行操作

#### Version 2.13.39
> 1.video cache逻辑收敛在job处理

#### Version 2.13.38
> 1.删除archive-service的GoRPC依赖

#### Version 2.13.37
> 1.notify增加attribute_v2

##### Version 2.13.36
> 1.联合投稿商业样式

##### Version 2.13.35
> 1.accountNotify无脑重试

##### Version 2.13.34
> 1.accountNotify增加重试

##### Version 2.13.33
> 1.使用cron处理mid缓存

##### Version 2.13.32
> 1.增加nil判断防止panic

##### Version 2.13.31
> 1.删除历史遗留的灰度代码
> 2.查询审核库时补齐rows.Err,排查分P为0的情况

##### Version 2.13.30
> 1.增加同步attribute_v2字段的逻辑

##### Version 2.13.29
> 1.删除ugc-season相关

##### Version 2.13.28
> 1.修复ecode

##### Version 2.13.27
> 1.删除调用修改分区的接口

##### Version 2.13.26
> 1.稿件审核库由使用id切到aid

##### Version 2.13.25
> 1.增加日志

##### Version 2.13.24
> 1.构建最新go-main发版

##### Version 2.13.23
> 1.修复aid在不同season时更新season_id的问题

##### Version 2.13.22
> 1.archive接口全部迁移grpc

##### Version 2.13.21
> 1.修改go-common的tag为0.1.20，修复grpc问题

##### Version 2.13.20
> 1.引包切换为go-main

##### Version 2.13.19
> 1. 兼容非过审稿件加入剧集的情况

##### Version 2.13.18
> 1. 基础库grpc加日志同步

##### Version 2.13.17
> 1. arcUpdate的重试优化，部分自带重试的逻辑不返回可重试的error

##### Version 2.13.16
> 1. 移除ugc剧集相关的代码
> 2. 新增处理互动视频剧情图根节点逻辑，更新pages的index_order以及稿件的firstCid

##### Version 2.13.15
> 1.同步状态-20互动视频的稿件

##### Version 2.13.14
> 1.优化日志

##### Version 2.13.13
> 1.通过databus消息更新archive的seasonid字段

##### Version 2.13.12
> 1.更新稿件表的seasonID操作独立出来，走arcUpdate发archive notify

##### Version 2.13.11
> 1.修改剧集redis retry逻辑

##### Version 2.13.10
> 1.ugc剧集数据同步

##### Version 2.13.9
> 1.联合投稿排序走index_order

##### Version 2.13.8
> 1.联合投稿按照审核库staff id 排序

##### Version 2.13.7
> 1.去除track表迁移first_pass表

##### Version 2.13.6
> 1.同步联合创作人archive_staff

##### Version 2.13.5
> 1.修复err后重复retry tranResult & -->chan

##### Version 2.13.4
> 1.走account的grpc

##### Version 2.13.3
> 1.扩展archive-service的配置

##### Version 2.13.2
> 1.账号信息空判断

##### Version 2.13.1
> 1.账号databus有重复消息

##### Version 2.13.0
> 1.账号notify走databus，记录日志

##### Version 2.12.13
> 1.账号notify走databus

##### Version 2.12.12
> 1.增加databus消费，重新生成用户昵称和头像缓存

##### Version 2.12.11
> 1.修改第一P的判断逻辑

##### Version 2.12.10
> 1.同步稿件时判断是否为vupload

##### Version 2.12.9
> 1.更新archive表的视频id，分辨率字段

##### Version 2.12.8
> 1.新表字段修改

##### Version 2.12.7
> 1.增加分辨率字段

##### Version 2.12.6
> 1.全量新表

##### Version 2.12.5
> 1.灰度新表

##### Version 2.12.4
> 1.切换BM

##### Version 2.12.3
> 1.计算稿件总时长

##### Version 2.12.2
> 1.聚合所有topic消费

##### Version 2.12.1
> 1.发送statDm-T消息

##### Version 2.12.0
> 1.迁移到主站目录下

##### Version 2.11.1
> 1.弹幕计数做聚合

##### Version 2.11.0
> 1.弹幕计数

##### Version 2.10.4
> 1.删除所有插入track的代码

##### Version 2.10.3
> 1.配置offset，停止track记录

##### Version 2.10.2
> 1.补充单元测试

##### Version 2.10.2
> 1.改进报警文案

##### Version 2.10.1
> 1.pgc异步

##### Version 2.10.0
> 1.增加监控
> 2.pgc与ugc分开

##### Version 2.9.4
> 1.增加无限塞回重试逻辑

##### Version 2.9.3
> 1.增加force_sync消息

##### Version 2.9.2
> 1.处理first_round消息

##### Version 2.9.1
> 1.全量通过databus更新稿件缓存

##### Version 2.9.0
> 1.通过databus消息更新稿件缓存
> 2.灰度30%

##### Version 2.8.1
> 1.调整 passed 然后又改了回来 = =

##### Version 2.8.0
> 1.archive_track 迁移至 hbase
> 2.调整 passed 逻辑改查 archive_oper 表

##### Version 2.7.1
> 1.databus通知吐出dynamic字段

##### Version 2.7.0
> 1.删除hzxs缓存清楚

##### Version 2.6.1
> 1.archive增加dynamic字段

##### Version 2.6.0
> 1.去除pgc的databus订阅

##### Version 2.5.2
> 1.回滚，video relation表数据有误

##### Version 2.5.1
> 1.result库支持video分表逻辑(灰度10%的稿件)

##### Version 2.5.0
> 1.删除cdn代码
> 2.完善更新db的日志

##### Version 2.4.1
> 1.切新的httpclient

##### Version 2.4.0
> 1.同步result库判断微调

##### Version 2.3.0
> 1.删除purge cdn逻辑

##### Version 2.2.7
> 1.routine数量配置化&优化报警逻辑

##### Version 2.2.6
> 1.多routine更新数据库&缓存&每分钟报警

##### Version 2.2.5
> 1.cid为0不同步result

##### Version 2.2.4
> 1.add error log

##### Version 2.2.3
> 1.fix cids 0

##### Version 2.2.1
> 1.修复pgc同步result逻辑

##### Version 2.2.0
> 1.video缓存更新逻辑优化

##### Version 2.1.2
> 1.pgc第二次过审不同步result库

##### Version 2.1.1
> 1.修复videos插入失败的bug

##### Version 2.1.0
> 1.模块分层

##### Version 2.0.0
> 1.大仓库版本，依赖新的go-common
> 2.archive_result库的archive_video表有数据更新会调archive-service接口更新分P详情信息

##### Version 1.14.0
> 1.databus send error 时无限重试

##### Version 1.13.0
> 1.暂时去掉HK节点的通知
> 2.增加notify的databus

##### Version 1.12.0
> 1.增加https页面的cdn purge

##### Version 1.11.0
> 1.增加result库的partition消费监控

##### Version 1.10.0
> 1.update field增加group2

##### Version 1.9.1
> 1.archive的channal根据aid取余，每个channal只有一个goroutine消费，避免消费乱序

##### Version 1.9.0
> 1.升级go-common&go-business
> 2.兼容manager后台修改稿件归属的mid缓存

##### Version 1.8.0
> 1.所有缓存清理走archive_result库

##### Version 1.7.15
> 1.迁移archive-service的group2缓存到result库

##### Version 1.7.14
> 1.增加monitor/ping

##### Version 1.7.13
> 1.group1使用archive_result库更新缓存

##### Version 1.7.12
> 1.bugfix,任务分发:稿件修改分区，写redis逻辑错误

##### Version 1.7.11
> 1.移除所有group1的调用

##### Version 1.7.10
> 1.bugfix,修复track的redis

##### Version 1.7.9
> 1.bugfix,修复track的redis

##### Version 1.7.8
> 1.bugfix,修复sql语句

##### Version 1.7.7
> 1.增加香港节点的配置

##### Version 1.7.6
> 1.已删除的稿件不记录track

##### Version 1.7.5
> 1.dede_arctype替换为archive_type并修改相关逻辑
> 2.添加杭州group

##### Version 1.7.5
> 1.视频track变更增加标题简介

##### Version 1.7.4
> 1.去掉feed push

##### Version 1.7.3
> 1.升级go-common和go-business
> 2.接入新版配置中心

##### Version 1.7.2
> 1.增加proc处理
> 2.修改track video待审重复计数bug

##### Version 1.7.1
> 1.增加track video

##### Version 1.7.0
> 1.增加track信息记录
> 2.计数更新硬币数

##### Version 1.6.24
> 1.增加后台报表统计
> 2.track表备注增加活动id
> 3.增加upcache错误重试

##### Version 1.6.23
> 1.增加接口失败重试

##### Version 1.6.22
> 1.稿件变更无脑add/del
> 2.基于无脑变更，去掉发布时间和属性变化通知

##### Version 1.6.21
> 1.升级vendor

##### Version 1.6.20
> 1.增加archive-service group2 缓存增量更新

##### Version 1.6.19
> 1.修复日志错误

##### Version 1.6.18
> 1.移除stat的databus消费

##### Version 1.6.17
> 1.移除评论注册

##### Version 1.6.16
> 1.archive缓存走RPC

##### Version 1.6.15
> 1.根据state状态记录access变更

##### Version 1.6.14
> 1.track去除没必要的记录

##### Version 1.6.13
> 1.内部接口地址规范化

##### Version 1.6.12
> 1.track兼容round改动2

##### Version 1.6.11
> 1.track兼容round的改动

##### Version 1.6.10
> 1.track增加attr和access

##### Version 1.6.9
> 1.增加attr不在列表输出的判断

##### Version 1.6.8
> 1.增加纪录片的表同步

##### Version 1.6.7
> 1.up更改后通知评论替换subject的mid

##### Version 1.6.6
> 1.修复track时间错误

##### Version 1.6.5
> 1.修复xcode判断错误

##### Version 1.6.4
> 1.增加PGC表同步逻辑

##### Version 1.6.3
> 1.临时去除attr发邮件判断

##### Version 1.6.2
> 1.增加track追踪

##### Version 1.6.1
> 1.修复insert时的json解析错误

##### Version 1.6.0

> 1.稿件依赖databus

##### Version 1.5.7

> 1.kafka to databus

##### Version 1.5.6

> 1.增加insert事件 评论subject注册

##### Version 1.5.5

> 1.增加monitor监控

##### Version 1.5.4

> 1.fix sub databus bug

##### Version 1.5.3

> 1.刷新cdn加条件

##### Version 1.5.2

> 1.稿件计数消费databus

##### Version 1.5.1

> 1.更新stat cache的Rpc为2

##### Version 1.5.0

> 1.番剧和电影状态变更需要发送邮件

##### Version 1.4.4

> 1.[consumer]去除archive binlog调用tag change接口

##### Version 1.4.3

> 1.[consumer]稿件推送动态改为过审就推

##### Version 1.4.2

> 1.[consumer]修复tag过审同步问题

##### Version 1.4.1

> 1.[consumer]增加sleep控制消费速度

##### Version 1.4.0

> 1.[consumer]增加稿件变动purge

##### Version 1.3.0

> 1.[consumer]增加archive的事件字段变更，新老字段

##### Version 1.2.0

> 1.[consumer]修改archive的事件注册，分通过、不通过、无差别删除cache

##### Version 1.1.5

> 1.[consumer]修改archive的cache更新接口

##### Version 1.1.4

> 1.[consumer]修复没过审视频被推tag动态问题

##### Version 1.1.3

> 1.[consumer]修正通过tag服务批量获取tag信息签名错误
> 2.[consumer]修改调用tag change 接口mid错误

##### Version 1.1.2

> 1.[consumer]feed调用历史数量判断修改，防止feed不成功

##### Version 1.1.1

> 1.[consumer]feed调用加稿件审核历史判断

##### Version 1.1.0

> 1.[consumer]增加分区视频列表cache更新和删除
> 2.[consumer]修复改tag的更新

##### Version 1.0.0

> 1.[producer]binlog同步
> 2.[consumer]信息消费更新缓存等

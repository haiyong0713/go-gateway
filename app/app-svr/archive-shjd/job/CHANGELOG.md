# 稿件字段同步
#### Version 1.6.24
> 1.查询首映和付费稿件信息dbfix

#### Version 1.6.23
> 1.sql del

#### Version 1.6.22
> 1.paladin v2迁移
> 新增稿件投稿属地

#### Version 1.6.21
> 1.付费稿件

#### Version 1.6.20
> 1.竖版缩略图

#### Version 1.6.19
> 1.多活改造, 消费点赞多机房消息

## Version 1.6.18
> 1.首映稿件缓存更新

## Version 1.6.17
> 1.稿件inner禁止项写入缓存

## Version 1.6.16
> 1.代码优化

## Version 1.6.15
> 1.psb_{aid}_{cid}缓存增加过期时间

## Version 1.6.14
> 1.psb_{aid}缓存增加过期时间
> 2.videoshot老key=vst_{cid}写入删除
> 3.desc老key=desc_{aid}写入删除

## Version 1.6.13
> 1.支持高清缩略图

## Version 1.6.12
> 1.a3p_{aid}缓存增加过期时间

## Version 1.6.11
> 1.修复jd-job首帧单独更新时未更新psb_aid缓存

## Version 1.6.10
> 1.支持稿件首帧提取

## Version 1.6.9
> 1.修复异味

## Version 1.6.8
> 1.修复short_link字段和up_from字段

## Version 1.6.7
> 1.add up_from 

## Version 1.6.6
> 1.lint bugs修复

## Version 1.6.5
> 1.分区名称不写入缓存

## Version 1.6.4
> 1.fix simple arc cache 

## Version 1.6.3
> 1.所有稿件缓存改为统一model

## Version 1.6.2
> 1.修复db返回值错误的问题

## Version 1.6.1
> 1.增加初始化taishan的速度

## Version 1.6.0
> 1.消费稿件时增加泰山&redis的双写逻辑
> 2.增加全量初始化泰山数据的逻辑

## Version 1.5.5
> 1.日志输出过大继续删除一些不必要的info log

## Version 1.5.4
> 1.删除一些不必要的info log

## Version 1.5.3
> 1.setcache忽略pages=nil

## Version 1.5.2
> 1.增加SimpleArc缓存逻辑

## Version 1.5.1
> 1.日志告警全部修改为Error级别

## Version 1.5.0
> 1.全面删除账号databus的依赖
> 2.删除archive-service的依赖，闭环在job内部的缓存或db处理
> 3.老代码中打印%v err的地方改为%+v

## Version 1.4.4
> 1.增加up投稿列表过滤规则

## Version 1.4.3
> 1.使用metrics记录UpdateCache的耗时

## Version 1.4.2
> 1.使用缓存更新投稿列表总数

## Version 1.4.1
> 1. 新增follow相关逻辑

## Version 1.4.0
> 1.稿件缓存的更新收敛在job中执行
> 2.upper主的投稿列表因为访问量的降低，使用一套redis集群维护即可，降低缓存的操作成本

## Version 1.2.23
> 1.videoshot逻辑收敛在Job进行删除操作

## Version 1.2.22
> 1.video cache逻辑收敛在job处理

## Version 1.2.21
> 1.删除archive-service的GoRPC依赖

## Version 1.2.20
> 1.notify增加attribute_v2

## Version 1.2.19
> 1.删掉MC相关逻辑

## Version 1.2.18
> 1.同步caster版本

## Version 1.2.17
> 1.修改accountNotify重试规则

## Version 1.2.16
> 1.accountNotify增加重试

## Version 1.2.15
> 1.使用cron代替timeSleep，防止panic

## Version 1.2.14
> 1.增加nil判断防止panic

## Version 1.2.13
> 1.重构archive-shjd-job 依赖redis

## Version 1.2.12
> 1.删除ugc-season相关

## Version 1.2.11
> 1.构建master发版

## Version 1.2.10
> 1.通过databus直接更新click缓存

## Version 1.2.9
> 1.archive接口全部迁移grpc

## Version 1.2.8
> 1.up信息为空时，通过消息更新up主信息

## Version 1.2.7
> 1.修改go-common的tag为0.1.20，修复grpc问题

## Version 1.2.6
> 1.fix

## Version 1.2.5
> 1.fix

## Version 1.2.4
> 1.ugc剧集嘉定缓存处理

## Version 1.2.3
> 1.迁移account grpc

## Version 1.2.2
> 1.修改消费canal时的数据处理

## Version 1.2.1
> 1.nw-->new

## Version 1.2.1
> 1.消费嘉定机房的accountNotify更新用户缓存

## Version 1.2.0
> 1.增加notify的databus pub，闭环在一个机房

## Version 1.1.0
> 1.增加stat消费

## Version 1.0.2
> 1.databus bugfix

## Version 1.0.1

> 1.修改配置文件名

## Version 1.0.0

> 1.初始化项目

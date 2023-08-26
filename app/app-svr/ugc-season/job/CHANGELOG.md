# UGC剧集job模块
## Version 1.1.12
> 1.稿件分库分表DB下线

## Version 1.1.11
> 1.稿件分库分表接口迁移

## Version 1.1.10
> 1.job计数优化

## Version 1.1.9
> 1.job启动优化

## Version 1.1.8
> 1.新增error log

## Version 1.1.7
> 1.启动初始化优化删除不一致检查

## Version 1.1.6
> 1.启动初始化优化
> 
## Version 1.1.5
> 1.付费合集
> 
## Version 1.1.4
> 1.coin接rail_gun

## Version 1.1.3
> 1.fix concurrent map

## Version 1.1.2
> 1.解决lint

## Version 1.1.1
> 1.删除多余client配置 & 增加pub重试

## Version 1.1.0
> 1.删除memcache依赖

## Version 1.0.16
> 1.stat增加redis双写逻辑

## Version 1.0.15
> 1.修复锁的错误使用

## Version 1.0.14
> 1.删除更新stat缓存和给jd发消息

## Version 1.0.13
> 1.删除archive关联及双写逻辑

## Version 1.0.12
> 1.稿件审核库由使用id切到aid

## Version 1.0.11
> 1.构建master发版

## Version 1.0.10
> 1.初始化upperList时忽略404错误

## Version 1.0.9
> 1.请求archive-service时 stats切片请求

## Version 1.0.8
> 1.stat-job 计算season的相关逻辑迁移到ugc-season-job中
> 2.原有的ugc-season-job发送databus给到stat-job更新结构的逻辑取消（以及相关重试逻辑取消），改为直接走channel
> 3.stat相关逻辑如更新season-aid结构体、发送databus消息给嘉定失败时，增加重试

## Version 1.0.7
> 1.更新maxptime只关心season的show

## Version 1.0.6

> 1.修复sql单引号拼接&稿件被换到删除小节未更新问题

## Version 1.0.5

> 1.兼容换源情况更新archive的season_id

## Version 1.0.4

> 1.更新空间合集列表缓存了

## Version 1.0.3

> 1.修改go-common的tag为0.1.20，修复grpc问题

## Version 1.0.2

> 1.job调用service更新缓存

## Version 1.0.1

> 1.修改同步逻辑

## Version 1.0.0

> 1.ugc season job 初始化

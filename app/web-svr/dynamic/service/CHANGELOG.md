### dynamic-service

#### Version 3.8.6
##### Features
> 1.迁移配置平台

#### Version 3.8.5
##### Features
> 1.稿件属性位禁止项批量接口查询

#### Version 3.8.4
##### Features
> 1.网关侧稿件属性位禁止项查询由解析attribute迁移为接口查询

#### Version 3.8.3
##### Features
> 1.车载小程序过滤分 P 稿件

#### Version 3.8.2
##### Features
> 1.新增落地页：虚拟主播

#### Version 3.8.1
##### Features
> 1.消除异味
> 2.fix casttype

#### Version 3.8.0
##### Features
> 1.fix sub msg panic

#### Version 3.7.8
##### Features
> 1.删除调用gorpc代码

#### Version 3.7.7
##### Features
> 1.调用Arcs方法的aid数量, 不超过100

#### Version 3.7.6
##### Features
> 1.aid超过100进行切片

#### Version 3.7.2
##### Features
> 1.迁移archive-service gorpc方法RankTopArcs3和RankAllArcs3

#### Version 3.7.1
##### Features
> 1.archive-service gorpc切grpc

#### Version 3.7.0
##### Features
> 1.优化二级分区databus消费逻辑
> 2.切回archive-service方法
> 3.添加一级分区当天投稿总数接口
> 4.对比archive-和dynamic二级分区接口吐出数据的一致性

#### Version 3.5.9
##### Features
> 1.修复二级分区数据在redis同一个key的问题

#### Version 3.5.8
##### Features
> 1.提供分区当天投稿总数

#### Version 3.5.7
##### Features
> 1.提供分区grpc方法
> 1.优化读取数据逻辑

#### Version 3.5.6
##### Features
> 1.初始化分区依赖archive-service方法

#### Version 3.5.5
##### Features
> 1.优化初始化数据库逻辑

#### Version 3.5.4
##### Features
> 1.修复重启时，关闭chan导致的panic

#### Version 3.5.3
##### Features
> 1.迁移archive-service下二级分区稿件

#### Version 3.5.2
##### Features
> 1.提供grpc方法

#### Version 3.5.1
##### Features
> 1.dao ut

#### Version 3.5.0
##### Features
> 1.identify使用grpc
> 2.改用metadata.RemoteIP方法

#### Version 3.4.1
##### Features
> 1.common conf

#### Version 3.3.1
##### Features
> 1.HTTPServer
> 2.del default conf

#### Version 3.2.2
##### Features
> 1.live 替换新的API

#### Version 3.2.1
##### Features
> 1.使用discovery接入archive

#### Version 3.1.1
##### Features
> 1.test to archive3

#### Version 3.1.0
##### Features
> 1.use bm
> 2.move main

#### Version 3.0.0
##### Features
> 1.使用common mc ，去掉redis

#### Version 2.5.5
##### Features
> 1.add archives3 log

#### Version 2.5.4
##### Features
> 1.rm config dentify-mc

#### Version 2.5.3
##### Features
> 1.rm archive2
> 2.config xlog to log

#### Version 2.5.2
##### Features
> 1.统一prom使用方法

#### Version 2.5.1
##### Features
> 1.切archive3 rid to int32

#### Version 2.5.0
##### Features
> 1.http切archive3
> 2.新增archive3 rpc方法

#### Version 2.4.0
##### Features
> 1.trace、httpClient修改

#### Version 2.3.0
##### Features
> 1.去掉DB依赖

#### Version 2.2.1
##### Bug Fixes
> 1.删除client下文件

#### Version 2.2.0
##### Features
> 1.合并大仓库

#### Version 2.1.8
##### Bug Fixes
> 1.fix 动态数空指针

#### Version 2.1.7
##### Bug Fixes
> 1.fix 批量动态空指针

#### Version 2.1.6
##### Features
> 1.remote cache 优化    
> 2.返回结果加调用最新投稿     
> 3.加单元测试    

#### Version 2.1.5
##### Bug Fixes
> 1.添加禁止动态过滤

#### Version 2.1.4
##### Features
> 1.remote cache接入prometheus
> 2.错误码QPS接入prometheus

#### Version 2.1.3
##### Bug Fixes
> 1.db,redis接入prometheus

#### Version 2.1.2
##### Features
> 1.接入prometheus

#### Version 2.1.0
##### Features
> 1.添加remote cache

#### Version 2.0.0
##### Features
> 1.接入新的配制中心

#### Version 1.9.1
##### Features
> 1.动态加Archive IsNormal判断 and 升级ZKOff

#### Version 1.9.0
##### Features
> 1.增加分区动态总数rpc方法

#### Version 1.8.0
##### Features
> 1.更新最新的net/rpc  
> 2.tag和分区最新动态轮训时间区分  

#### Version 1.7.0
##### Features
> 1.分区动态总数,增加直播动态总数  

#### Version 1.6.0
##### Features
> 1.批量获取分区动态接口  

#### Version 1.5.0
##### Bug Fixes
> 1.fix 大数据接口容错的bug  

#### Version 1.4.2
##### Bug Fixes
> 1.增加rpc ping方法  

#### Version 1.4.1
##### Features
> 1.更新vendor  

#### Version 1.4.0
##### Features
> 1.更新vendor，支持rpcx  

#### Version 1.3.2
##### Features
> 1.升级vendor  

#### Version 1.3.1
##### Features
> 1.分区动态总数一次性返回所有分区  

#### Version 1.3.0
##### Features
> 1.增加分区动态总数  

#### Version 1.2.1
##### Features
> 1.初始化rpc  

#### Version 1.2.0
##### Features
> 1.接入配置中心  
> 2.升级vendor  

#### Version 1.1.0
##### Features
> 1.支持一级分区动态  
> 2.fix map 同时读写bug  
> 3.更新vendor  

#### Version 1.0.0
##### Features
> 1.支持分区和tag的最新视频以及最新动态  

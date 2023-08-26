### 计数更新job

#### Version 1.7.12
> 1.【降本增效】删除一些日志

#### Version 1.7.11
> 1.多活改造, 消费点赞多机房消息

#### Version 1.7.10
> 1.修复自增id增长过快
> 2.删去消费dislike逻辑

#### Version 1.7.9
> 1.删除archive_click库的依赖
> 2.删除已经不再使用的配置
> 3.使用日志告警平台，去除早期的手动报警

#### Version 1.7.8
> 1.更新archive_service redis集群加重试

#### Version 1.7.7
> 1.报警增加类型判断

#### Version 1.7.6
> 1.增加播放数回源数据比消息数据要大的报警

#### Version 1.7.5
> 1.暂停follow的监控

#### Version 1.7.4
> 1.新增follow相关消费的落地逻辑

#### Version 1.7.3
> 1.删除MC相关逻辑

#### Version 1.7.2
> 1.删除clickdb的配置
> 2.凌晨落库时的间隔改为配置化

#### Version 1.7.1
> 1.删除灰度逻辑
> 2.添加并发更新
> 3.去掉smaller部分
> 4.修改启动的go routine数目

#### Version 1.7.0
> 1.删除废弃代码
> 2.调整代码顺序

#### Version 1.6.12
> 1.重构，使用redis取代内存中的sm进行av数据缓存

#### Version 1.6.11
> 1.修复test文件

#### Version 1.6.10
> 1.重启时通过读取service接口初始化数据，以防数据出现回退

#### Version 1.6.9
> 1.各平台的播放数由databus进行更新

#### Version 1.6.8
> 1.season计数相关逻辑迁移到ugc-season-job中

#### Version 1.6.7
> 1.配合aid无序，去掉maxAid的校验

#### Version 1.6.6
> 1.stat-job中剧集的逻辑双写ugc season缓存
> 2.memcache.Pool迁移到memcache.Memcache

#### Version 1.6.5
> 1.修复panic

#### Version 1.6.4
> 1.修复判断aid增减的逻辑

#### Version 1.6.3
> 1.修复关闭时send on closed channel导致panic
> 2.season被删时候不删除season_stat的mc和db

#### Version 1.6.2
> 1.新增剧集计数

#### Version 1.6.1
> 1.dislike强制0

#### Version 1.6.0
> 1.监控增加env

#### Version 1.5.1
> 1.修改conn close时机

#### Version 1.5.0
> 1.job直接更新缓存

#### Version 1.4.8
> 1.迁移BM

#### Version 1.4.7
> 1.增加dislike

#### Version 1.4.6
> 1.调整日志  

#### Version 1.4.5
> 1.增加企业微信报警  

#### Version 1.4.4
> 1.打开更新数据库的逻辑  

#### Version 1.4.3
> 1.重新消费伪代码  

#### Version 1.4.2
> 1.修复所有计数无法从1变成0的bug  

#### Version 1.4.1
> 1.like > 0

#### Version 1.4.0
> 1.全部消费新的databus  

#### Version 1.3.1
> 1.fix databus老数据  

#### Version 1.3.0
> 1.迁移到main目录  

#### Version 1.2.3
> 1.maxAid限制  

#### Version 1.2.2
> 1.点赞消费修复bug  

#### Version 1.2.1
> 1.点赞消费  

#### Version 1.2.0
> 1.升级go-common,go-business  
> 2.实现点赞数更新  

#### Version 1.1.2
> 1.恢复group2的调用  

#### Version 1.1.1
> 1.删除group1的调用  

#### Version 1.1.0
> 1.重写更新逻辑  

#### Version 1.0.9
> 1.增加 archiveRPC2

#### Version 1.0.0
> 1.从databus订阅stat修改消息，存入MySQL和更新Cache

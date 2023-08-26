## up-archive-job
## Version 1.2.26
1.B端分库分表DB下线

## Version 1.2.25
1.B端分库分表接口迁移

## Version 1.2.24
1.接入paladin.v2

## Version 1.2.23
1.过滤付费合集

### v1.2.13
1.修复cache为0

### v1.2.12
1.限制并发数

### v1.2.11
1.禁止空间&动态

### v1.2.10
1. 修复非联合投稿列表出现联合投稿

### v1.2.9
1. archive-job的异常重试，会导致insert变为update，所以考虑无脑needAdd

### v1.2.8
1. fix railgun close panic
2. railgun 配置化

### v1.2.6
1. 修复过审稿件更换up主，导致未更新投稿列表

### v1.2.5
1. 空缓存设置过期时间

### v1.2.4
1. fix del story log

### v1.2.3
1. 日志优化，便于查找问题

### v1.2.2
1. 联合投稿去除attribute判断

### v1.2.1
1. 联合投稿状态变更特殊处理

### v1.2.0
1. story投稿列表  
2. 使用railgun

### v1.0.6
1. 单片数据写入不用延时 

### v1.0.5
1. 空缓存写入fix

### v1.0.4
1. 空缓存判断fix

### v1.0.3
1. fix 分片数量错误

### v1.0.2
1. score 值修改

### v1.0.1
1. 新稿件过滤attr

### v1.0.0
1. 初始化list缓存，监听databus  

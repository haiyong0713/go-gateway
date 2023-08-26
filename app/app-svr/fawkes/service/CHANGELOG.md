#### Fawkes 各种管理~

##### Version 1.18.42
##### feature
> 1. 修复appInfo查询异常

##### Version 1.18.41
##### feature
> 1. 修复异常 sample_rate 无法置为0

##### Version 1.18.40
##### feature
> 1. mod资源同步到pcdn

##### Version 1.18.39
##### feature
> 1. 添加 - 高峰采样配置

##### Version 1.18.38
##### feature
> 1. 添加活动高峰字段

##### Version 1.18.37
##### bugfix
> 1. 模块体积配置查询异常修复

##### Version 1.18.36
##### bugfix
> 1. apm event clickhouse建表去掉datacenter_table_id

##### Version 1.18.35
##### feature
> 1. apm event 添加字段 lowest_sample_rate

##### Version 1.18.34
##### feature
> 1. CI编译日志上报到hive

##### Version 1.18.33
##### feature
> 1. crash聚合hash字段切换

##### Version 1.18.32
##### feature
> 1. crash info增加manufacturer和domestic_rom_ver字段

##### Version 1.18.31
##### bugfix
> 1. app申请时排除datacenterAppid为0的情况

##### Version 1.18.30
##### feature
> 1. /auth/user/name/list接口返回值增加user_id

##### Version 1.18.29
##### feature
> 1. 新增openApi 数据平台采样率设置/发布

##### Version 1.18.28
##### bugfix
> 1. 流作业map表生成逻辑修复

##### Version 1.18.27
##### bugfix
> 1. 修复laser mid/buvid共存的触达失败异常

##### Version 1.18.26
##### feature
> 1. 告警根因配置增加描述和自定义查询类型

##### Version 1.18.25
##### bugfix
> 1. fix 创建关联 MR 时重复写子仓 MR 的问题

##### Version 1.18.24
##### feature
> 1. 企微应用通知增加发送图片的功能
> 2. 增加技术埋点监测通知配置功能

##### Version 1.18.23
##### bugfix
> 1. fix 创建关联 MR 后丢失关键字的问题

##### Version 1.18.22
##### bugfix
> 1. ci删除逻辑配置reload

##### Version 1.18.21
##### feature
> 1. XCrash 聚合数据兼容

##### Version 1.18.20
##### feature
> 1. 添加接口 add pcdn file api
> 2. 添加接口 sre config list

##### Version 1.18.19
##### feature
> 1. 适配 gitlab 升级 14，Draft 功能替代 WIP 功能
> 2. 新增 create mr 接口，bbgit 访问 gitlab api 相关功能迁移到后端，不再在客户端保存 key

##### Version 1.18.18
##### feature
> 1. MOD-patch包大于原包则过滤不返回

##### Version 1.18.17
##### feature
> 1. webhook trigger error不再影响接口返回

##### Version 1.18.16
##### feature
> 1. openApi添加新接口：OOM index List

##### Version 1.18.15
##### feature
> 1. cdn添加contant-disposition 逻辑优化

##### Version 1.18.14
##### feature
> 1. grouup为空. 不再进行分组

##### Version 1.18.13
##### feature
> 1. 自动分组 && cdn添加contant-disposition

##### Version 1.18.12
##### feature
> 1. 根据cdnurl，获取渠道包下载详情。 添加sole_cdn_url字段

##### Version 1.18.11
##### feature
> 1. 新增接口 - Fawkes Config Key 发布历史列表

##### Version 1.18.10
##### feature
> 1. 堆栈解析增加crash_version字段

##### Version 1.18.9
##### feature
> 1. 带宽预警样式调整

##### Version 1.18.8
##### feature
> 1. 修改埋点字段名同步到数据平台的逻辑

##### Version 1.18.7
##### feature
> 1. 增加埋点监测数据的聚合统计功能

##### Version 1.18.6
##### feature
> 1. NAS盘渠道包清理去掉"已推CDN"限制

##### Version 1.18.5
##### feature
> 1. MOD带宽预警
> 2. fix no rows 

##### Version 1.18.4
##### feature
> 1. iOS上传bugly操作逻辑调整

##### Version 1.18.3
##### bugfix
> 1. 修复编译耗时异常

##### Version 1.18.2
##### feature
> 1. 告警指标增加回滚功能

##### Version 1.18.1
##### feature
> 1. ci编译规则
> 2. ci common add接口添加返回值

##### Version 1.17.48
##### feature
> 1. 配置刷新

##### Version 1.17.47
##### feature
> 1. PC渠道包

##### Version 1.17.46
##### bugfix
> 1. ios上传bugly文件逻辑修复

##### Version 1.17.45
##### feature
> 1. 提交pipeline增加日志

##### Version 1.17.44
##### feature
> 1. windows OSS发布上传优化

##### Version 1.17.43
##### bugfix
> 1. mobiapp和appId校验逻辑修复

##### Version 1.17.42
##### feature
> 1. ci消息推送增加日志

##### Version 1.17.41
##### feature
> 1. apm告警增加根因配置功能

##### Version 1.17.40
##### feature
> 1. ios上传bugly逻辑调整

##### Version 1.17.39
##### bugfix
> 1. openapi接口增加兜底username

##### Version 1.17.38
##### feature
> 1. 告警指标临时模板生成

##### Version 1.17.37
##### feature
> 1. ci common add 兼容operator

##### Version 1.17.36
##### feature
> 1. paladin v2

##### Version 1.17.35
##### feature
> 1. 技术埋点告警配置增加datacenterAppId字段

##### Version 1.17.34
##### feature
> 1. ff增加mid黑名单过滤功能

##### Version 1.17.33
##### feature
> 1. apm增加对chid的筛选

##### Version 1.17.32
##### feature
> 1. app申请和修改对mobi_app以及datacenter_app_id的校验

##### Version 1.17.31
##### bugfix
> 1. hotfix推送到正式env_vars字段同步的修复

##### Version 1.17.30
##### bugfix
> 1. command sql拼接错误修复

##### Version 1.17.29
##### feature
> 1. cd发布支持根据设备品牌过滤

##### Version 1.17.28
##### bugfix
> 1. 空指针fix

##### Version 1.17.27
##### feature
> 1. 网络和图片增加country筛选
> 2. pie增加orderby和limit

##### Version 1.17.26
##### bugfix
> 1. mobile-ep同名文件上传冲突fix

##### Version 1.17.25
##### feature
> 1. 告警指标增加生效状态配置
> 2. 告警指标和告警规则的联动

##### Version 1.17.24
##### feature
> 1. 修复APM网络端到端 - 业务状态码占比异常

##### Version 1.17.23
##### feature
> 1. 渠道包上传CDN使用队列限速 

##### Version 1.17.22
##### feature
> 1. 埋点上报采样率配置 

##### Version 1.17.21
##### feature
> 1. clickhouse旧集群配置删除

##### Version 1.17.20
##### feature
> 1. 应用组权限共享

##### Version 1.17.19
##### bugfix
> 1. 修复calculate读旧集群逻辑

##### Version 1.17.18
##### feature
> 1. 技术埋点增加忽略日志平台同步的参数

##### Version 1.17.17
##### feature
> 1. cdn固定地址文件名由接口入参指定

##### Version 1.17.16
##### feature
> 1. apm告警规则增加微调规则的逻辑

##### Version 1.17.15
##### feature
> 1. apm告警规则增加微调规则的逻辑
> 2. 告警规则筛选逻辑调整

##### Version 1.17.14
##### feature
> 1. 删除已推送过CDN的渠道包

##### Version 1.17.13
##### feature
> 1. laser webhook方式拉取日志

##### Version 1.17.12
##### feature
> 1. 增加apm告警功能
> 2. 告警映射关系修改
> 3. apm告警优化以及增加根因分析
> 4. 搜索时间改为start_time

##### Version 1.17.11
##### feature
> 1. ck源表切换至信息群

##### Version 1.17.10
##### feature
> 1. 各类安装包提供固定下载地址

##### Version 1.17.9
##### bugfix
> 1. trackt统计功能name查询修改

##### Version 1.17.8
##### bugfix
> 1. 埋点字段同步到数据平台逻辑修复

##### Version 1.17.7
##### feature
> 1. 未注册的技术埋点数据分析统计

##### Version 1.17.6
##### bugfix
> 1. workflow代码修复

##### Version 1.17.5
##### feature
> 1. 增加技术埋点数据统计看板

##### Version 1.17.4
##### bugfix
> 1. ci构建消息取消、失败模板修改

##### Version 1.17.3
##### feature
> 1. web容器白名单

##### Version 1.17.2
##### bugfix
> 1. ci构建消息默认通知操作人
> 2. 通知群组添加默认接收者

##### Version 1.17.1
##### feature
> 1. ci代码重构

##### Version 1.17.0
##### feature
> 1. fflist返回值增加mtime

##### Version 1.16.99
##### bugfix
> 1. 修复tribe跨版本兼容问题

##### Version 1.16.98
##### bugfix
> 1. 技术埋点字段类型映射修复

##### Version 1.16.97
##### feature
> 1. 优化堆栈解析状态变更接口

##### Version 1.16.96
##### bugfix
> 1. clickhouse Web端换表逻辑修复

##### Version 1.16.95
##### feature
> 1. 添加service ping接口

##### Version 1.16.94
##### feature
> 1. Job 单机执行

##### Version 1.16.93
##### feature
> 1. 技术埋点资源占用逻辑更改

##### Version 1.16.92
##### bugfix
> 1. mod下发逻辑fix

##### Version 1.16.91
##### feature
> 1. 状态变更接口增加通知及日志写入功能，增加指派人字段
> 2. 增加日志列表

##### Version 1.16.90
##### bugfix
> 1. mod页面patch展示

##### Version 1.16.89
##### feature
> 1. app管理员申请增加workflow

##### Version 1.16.88
##### feature
> 1. 打包失败通知 

##### Version 1.16.87
##### feature
> 1. 增加查询索引状态接口
> 2. 定时处理任务接口
> 3. 堆栈列表开放至openapi

##### Version 1.16.86
##### feature
> 1. apm告警规则

##### Version 1.16.85
##### feature
> 1. MOD下发逻辑优化

##### Version 1.16.84
##### feature
> 1. app权限申请审批人不显示已离职的人员

##### Version 1.16.83
##### feature
> 1. 修改网络正确率/网络降级正确率逻辑

##### Version 1.16.82
##### feature
> 1. 技术埋点数据补全定时任务

##### Version 1.16.81
##### feature
> 1. laser log

##### Version 1.16.80
##### feature
> 1. job 统计新增字段

##### Version 1.16.79
##### feature
> 1. 技术埋点属性增加示例字段

##### Version 1.16.78
##### feature
> 1. setup 增加 ff 和config version

##### Version 1.16.77
##### feature
> 1. 技术埋点属性增加名称字段

##### Version 1.16.76
##### feature
> 1. 技术埋点统一以appId维度管理

##### Version 1.16.75
##### feature
> 1. 技术埋点资源占用统计数据源改为hive

##### Version 1.16.74
##### feature
> 1. flink配置文件逻辑更改

##### Version 1.16.73
##### feature
> 1. app内部技术埋点增加数仓分流

##### Version 1.16.72
##### feature
> 1. 业务组增加数据仓库分流表配置

##### Version 1.16.71
##### feature
> 1. 技术埋点增加数据仓库分流表

##### Version 1.16.70
##### feature
> 1. 编译统计增加build_source字段

##### Version 1.16.69
##### feature
> 1. 猫耳游戏CD测试环境推正式增加EP通知

##### Version 1.16.68
##### feature
> 1. 002312、011130、自定义埋点集群迁移

##### Version 1.16.67
##### feature
> 1. 增加app权限审批列表用户当前权限显示和忽略接口

##### Version 1.16.66
##### bugfix
> 1. 修复app_key为空时的broadcast切换逻辑

##### Version 1.16.65
##### feature
> 1. 添加入口查询fawkes用户总权限

##### Version 1.16.64
##### feature
> 1. 增加企微机器人列表OpenApi

##### Version 1.16.63
##### feature
> 1. clickhouse修改quantile拼接语句

##### Version 1.16.62
##### feature
> 1. 技术埋点字段类型增加UInt64和Map

##### Version 1.16.61
##### feature
> 1. 技术埋点自动建表扩展字段排序去除

##### Version 1.16.60
##### feature
> 1. 猫耳游戏资源包推送到正式环境

##### Version 1.16.59
##### bugfix
> 1. fix tapd jon

##### Version 1.16.58
##### feature
> 1. 增加tapd所有的字段

##### Version 1.16.57
##### bugfix
> 1. 告警规则增加字段agg_percentile

##### Version 1.16.56
##### bugfix
> 1. 修正CI定时出包参数异常

##### Version 1.16.55
##### feature
> 1. 增加技术埋点自动建表接口

##### Version 1.16.54
##### bugfix
> 1. ci编译统计增加字段after_sync_task

##### Version 1.16.53
##### bugfix
> 1. 数据平台基础字段忽略
> 2. 告警规则更新接口参数修改

##### Version 1.16.52
##### feature
> 1. 申请应用接入workflow

##### Version 1.16.51
##### feature
> 1. 增加git trigger token缓存

##### Version 1.16.50
##### bugfix
> 1. 技术埋点字段返回修复

##### Version 1.16.49
##### feature
> 1. Fawkes AppInfo 剔除

##### Version 1.16.48
##### feature
> 1. 技术埋点归属业务组精确查询

##### Version 1.16.47
##### feature
> 1. 技术埋点字段同步到数据平台

##### Version 1.16.46
##### feature
> 1. 权限校验逻辑上线

##### Version 1.16.45
##### feature
> 1. 技术埋点增加告警规则

##### Version 1.16.44
##### feature
> 1. 技术埋点增加字段

##### Version 1.16.43
##### feature
> 1. 技术埋点文件结构调整

##### Version 1.16.42
##### feature
> 1. tribe迭代 文档更新

##### Version 1.16.41
##### bugfix
> 1. 修复技术埋点监控的返回值

##### Version 1.16.40
##### bugfix
> 1. appKey不存在panic修复

##### Version 1.16.39
##### feature
> 1. 客服反馈-状态流转相关 

##### Version 1.16.39
##### bugfix
> 1. fix error sql

##### Version 1.16.38
##### feature
> 1. laser查询增加md5返回字段

##### Version 1.16.37
##### feature
> 1. 增加日志

##### Version 1.16.36
##### feature
> 1. cd upgrade 支持排除版本 

##### Version 1.16.35
##### feature
> 1. tribe relation接口支持只按照appkey拉取

##### Version 1.16.34
##### feature
> 1. openapi /ci/pack/report/info

##### Version 1.16.33
##### feature
> 1. 国际版本 - 海外版迁移

##### Version 1.16.32
##### feature
> 1. android oom

##### Version 1.16.31
##### feature
> 1. tribe feature

##### Version 1.16.30
##### bugfix
> 1. fix mod/add is_wifi

##### Version 1.16.29
##### feature
> 1. 权限树 

##### Version 1.16.28
##### feature
> 1. 告警指标删除撤销

##### Version 1.16.27
##### feature
> 1. 开启访问中间件

##### Version 1.16.26
##### feature
> 1. 重构机器人&邮件配置
> 2. feedback查询条件修改

##### Version 1.16.25
##### feature
> 1. openapi header支持fawkes-user

##### Version 1.16.24
##### bugfix
> 1. sql语句拼接错误fix

##### Version 1.16.23
##### feature
> 1. app内部技术埋点增加基础字段组

##### Version 1.16.22
##### bugfix
> 1. 补充技术埋点查询接口废弃标识

##### Version 1.16.21
##### feature
> 1. 字段过滤中间件

##### Version 1.16.20
##### feature
> 1. OpenAPI增加 /ci/common/add

##### Version 1.16.19
##### feature
> 1. 技术埋点增加是否废弃状态标识

##### Version 1.16.18
##### feature
> 1. 新增主仓子仓切分支服务

##### Version 1.16.17
##### feature
> 1. OpenAPI增加 /app/laser /apm/moni/calculate

##### Version 1.16.16
##### feature
> 1. 技术埋点生成sql读取途径修改成field_file表

##### Version 1.16.15
##### feature
> 1. 技术埋点字段增加es标识
> 2. 同步es字段类型

##### Version 1.16.14
##### feature
> 1. feedback迁移OpenAPI

##### Version 1.16.13
##### feature
> 1. 添加异常访问的uri日志

##### Version 1.16.12
##### feature
> 1. feedback增加业务线标识

##### Version 1.16.11
##### feature
> 1. bus/feedback/list下线

##### Version 1.16.10
##### feature
> 1. 技术埋点和app内部关联同步归属应用

##### Version 1.16.9
##### feature
> 1. 技术埋点字段审核态

##### Version 1.16.8
##### bugfix
> 1. feedback 消息通知

##### Version 1.16.7
##### bugfix
> 1. ios灰度包增加日志与修复

##### Version 1.16.6
##### feature
> 1. 灰度包的推送增加日志

##### Version 1.16.5
##### bugfix
> 1. eventbus fix

##### Version 1.16.4
##### feature
> 1. tribe增加API解析功能
> 2. ci/info返回bbr_url

##### Version 1.16.3
##### feature
> 1. ios灰度包推送逻辑调整

##### Version 1.16.2
##### feature
> 1. tapd 增加token

##### Version 1.16.2
##### feature
> 1. eventHub结构调整
> 2. 自动构建渠道包时接口超时fix

##### Version 1.16.1
##### feature
> 1. 已切openapi的接口下线
> 2. 优化提示tribe提示

##### Version 1.16.0
##### feature
> 1. 渠道包自动删除任务 过滤已删除的包

##### Version 1.15.99
##### bugfix
> 1. 灰度包查询接口空数组判断

##### Version 1.15.98
##### feature
> 1. APP编辑接口微调逻辑

##### Version 1.15.97
##### feature
> 1. 灰度包查询接口条件和返回参数增加

##### Version 1.15.96
##### feature
> 1. 渠道组增加构建优先级

##### Version 1.15.95
##### feature
> 1. 技术埋点监控增加
> 2. 技术埋点字段增加以及app内部技术埋点存储监控的增加

##### Version 1.15.94
##### feature
> 1. tribe增加是否内置属性

##### Version 1.15.93
##### feature
> 1. 技术埋点字段增加类型text
> 2. 增加server_zone字段

##### Version 1.15.92
##### bugfix
> 1. ios灰度包判断逻辑更改
> 2. 增加灰度包查询接口

##### Version 1.15.91
##### feature
> 1. merge告警逻辑优化

##### Version 1.15.90
##### feature
> 1. patch包 定时清理

##### Version 1.15.89
##### bugfix
> 1. 渠道包自动上传逻辑fix

##### Version 1.15.88
##### bugfix
> 1. flink 文件生成逻辑更改

##### Version 1.15.87
##### feature
> 1. 全量拉取分支信息

##### Version 1.15.86
##### fix
> 1. 自动构建渠道包逻辑优化

##### Version 1.15.85
##### bugfix
> 1. crash index uptade fix

##### Version 1.15.84
##### bugfix
> 1. mod cache fix

##### Version 1.15.83
##### feature
> 1. 增加xcrash解析相关查询接口

##### Version 1.15.82
##### bugfix
> 1. mod cache 回退

##### Version 1.15.81
##### bugfix
> 1. 添加clickhouse sql日志

##### Version 1.15.80
##### feature
> 1. ci包重签名增加自定义分支

##### Version 1.15.79
##### feature
> 1. 移除无用配置引用

##### Version 1.15.78
##### feature
> 1. 添加kibana配置申请通知

##### Version 1.15.77
##### bugfix
> 1. clickhouse的聚合函数判空逻辑增加

##### Version 1.15.76
##### feature
> 1. mod/appkey/file/list md5逻辑放入缓存fix

##### Version 1.15.75
##### feature
> 1. 增加查询指定构建号的可用tribe的接口

##### Version 1.15.74
##### feature
> 1. 用户追踪增加时间范围筛选
> 2. crash info增加返回字段

##### Version 1.15.73
##### feature
> 1. 增加event flink relation的返回数据

##### Version 1.15.72
##### feature
> 1. 增加指定logId的埋点进入数据平台逻辑

##### Version 1.15.71
##### bugfix
> 1. event flink 空历史文件判断逻辑增加

##### Version 1.15.70
##### bugfix
> 1. clickhouse statistics rate的sql逻辑更改

##### Version 1.15.69
##### feature
> 1. openapi config/add接口

##### Version 1.15.68
##### feature
> 1. app platform 增加win

##### Version 1.15.67
##### feature
> 1. 增加日志平台event field mapping接口
> 2. event field set接口代码逻辑重构

##### Version 1.15.66
##### feature
> 1. tribe修改构建产物保存路径

##### Version 1.15.65
##### feature
> 1. 添加event_fields基础字段设置条件

##### Version 1.15.64
##### feature
> 1. openapi 接口字典序排序

##### Version 1.15.63
##### feature
> 1. 追加openapi /apm/event

##### Version 1.15.62
##### feature
> 1. pipeline build 手动触发接口改为 GET，方便 webhook 调用，联动多仓库

##### Version 1.15.61
##### feature
> 1. clickhouse增加tribe_bundles的查询条件

##### Version 1.15.60
##### bugfix
> 1. tribe_pack_version fix operator

##### Version 1.15.59
##### feature
> 1. 下线business/laser/all和business/laser/all/silence接口

##### Version 1.15.58
##### feature
> 1. nas盘清理统计 
> 2. fix文件大小计算方式

##### Version 1.15.57
##### bugfix
> 1. ff 空指针修复

##### Version 1.15.56
##### feature
> 1. nas clean收口到fawkes-admin

##### Version 1.15.55
##### feature
> 1. 热修复支持环境变量

##### Version 1.15.54
##### feature
> 1. 增加ip解析外部接口

##### Version 1.15.53
##### feature
> 1. 增加openapi接口
> 2. log caller fix

##### Version 1.15.52
##### bugfix
> 1. clickhouse增加查询quantile、avg为nan的处理判断

##### Version 1.15.51
##### feature
> 1. fawkes task收口

##### Version 1.15.50
##### feature
> 1. tribe cd列表增加最后修改人和最后修改时间

##### Version 1.15.49
##### bugfix
> 1. laser msgId越界修复

##### Version 1.15.48
##### feature
> 1. 灰度包信息datacenterAppId更改

##### Version 1.15.47
##### feature
> 1. 卡顿详情接口增加返回route字段

##### Version 1.15.46
##### feature
> 1. 增加灰度包信息记录以及推送databus

##### Version 1.15.45
##### feature
> 1. fix laser msgID

##### Version 1.15.44
##### feature
> 1. business_api to openapi

##### Version 1.15.43
##### bugfix
> 1. tapd 同步新增 端 字段， 用于测试统计区分hd和粉板

##### Version 1.15.42
##### feature
> 1. laser/user 接口增加broadcast推送逻辑

##### Version 1.15.41
##### feature
> 1. business_api to openapi

##### Version 1.15.40
##### bugfix
> 1. 修正ios在只推送appstore正式版时无法正确复制模块体积配置的问题

##### Version 1.15.39
##### feature
> 1. laser business接口取消下线

##### Version 1.15.38
##### bugfix
> 1. 修复重复log

##### Version 1.15.37
##### feature
> 1. feedback增加BV号

##### Version 1.15.36
##### bugfix
> 1. merge超时提醒监控 fix

##### Version 1.15.35
##### feature
> 1. config/ff/cd增加数分调用的接口
> 2. app内部埋点查询增加归属应用的筛选条件

##### Version 1.15.34
##### bugfix
> 1. event count查询fix

##### Version 1.15.33
##### feature
> 1. laser add msgId

##### Version 1.15.32
##### feature
> 1. tribe增加调用git trigger的log 

##### Version 1.15.31
##### feature
> 1. openapi set username
> 2. bug fix

##### Version 1.15.30
##### feature
> 1. openapi 添加部分新接口

##### Version 1.15.29
##### feature
> 1. openapi 添加部分新接口

##### Version 1.15.28
##### bugfix
> 1. 渠道包自动构建分组逻辑fix

##### Version 1.15.27
##### feature
> 1. openapi 添加部分新接口

##### Version 1.15.26
##### bugfix
> 1. laser升级到broadcast消息离线推送版本

##### Version 1.15.25
##### bugfix
> 1. patch 规则改为 5个稳定版本 + 15个最新版本

##### Version 1.15.24
##### bugfix
> 1. 单元测试修复

##### Version 1.15.23
##### bugfix
> 1. 空指针修复

##### Version 1.15.22
##### feature
> 1. 稳定版本渠道包自动构建&&状态通知

##### Version 1.15.21
##### feature
> 1. 获取视频云资源url方式调整

##### Version 1.15.20
##### bugfix
> 1. metrics 发布状态判定忽略sql空格

##### Version 1.15.19
##### bugfix
> 1. hotfix 版本拉取修复

##### Version 1.15.18
##### feature
> 1. log日志优化 && 添加检测中间件

##### Version 1.15.17
##### bugfix
> 1. 修复windows appinstaller文件内容

##### Version 1.15.16
##### bugfix
> 1. open api 权限表插入错误fix

##### Version 1.15.15
##### feature
> 1. 单独CI支持webhook

##### Version 1.15.14
##### bugfix
> 1. 删除路由时同步删除权限&&接口描述信息更新
> 2. 中间件代码位置调整

##### Version 1.15.13
##### feature
> 1. 路由整理

##### Version 1.15.12
##### feature
> 1. windows上传服务

##### Version 1.15.11
##### feature
> 1. 接口接入openApi

##### Version 1.15.10
##### feature
> 1. app打包可选择同时打包tribe

##### Version 1.15.9
##### bugfix
> 1. 修复tribe context value

##### Version 1.15.8
##### feature
> 1. 修复ci排队告警sql

##### Version 1.15.7
##### feature
> 1. open api

##### Version 1.15.6
##### bugfix
> 1. tribe包oss上传地址修改

##### Version 1.15.3
##### bugfix
> 1. ci排队告警规则修改

##### Version 1.15.2
##### bugfix
> 1. 定时出包取消重试逻辑

##### Version 1.15.1
##### feature
> 1. 卡顿耗时改为分位

##### Version 1.14.117
##### feature
> 1. app内部技术埋点的实现

##### Version 1.14.116
##### bugfix
> 1. crash&jank count逻辑更改

##### Version 1.14.115
##### bugfix
> 1. moni query key分位逻辑修改

##### Version 1.14.114
##### feature
> 1. crash&jank info分页逻辑增加

##### Version 1.14.113
##### bugfix
> 1. event空查询逻辑增加

##### Version 1.14.112
##### bugfix
> 1. 去掉bilibili.com后缀

##### Version 1.14.111
##### feature
> 1. crash index和jank index 增加分页
> 2. 增加clickhouse聚合函数查询方法

##### Version 1.14.110
##### feature
> 1. 技术埋点增加企业微信通知

##### Version 1.14.109
##### feature
> 1. tribe cd list 根据version_name 聚合

##### Version 1.14.108
##### feature
> 1. business/cd/list 接口开放

##### Version 1.14.107
##### feature
> 1. app_attribute增加app_symbolso_name字段标识Android符号表

##### Version 1.14.106
##### feature
> 1. 新增apm接口：单条指标 数据的加载返回

##### Version 1.14.105
##### bugfix
> 1. crash info配置laser逻辑更改

##### Version 1.14.104
##### feature
> 1. mod/module/update 增加文件类型的修改功能
> 2. app/ci/record 接口operator如果有邮箱后缀则去掉后缀

##### Version 1.14.103
##### feature
> 1. laser新增md5和uposuri

##### Version 1.14.102
##### feature
> 1. modules/size/groupversion 开放biz接口

##### Version 1.14.102
##### bugfix
> 1. tribe cd 排序

##### Version 1.14.101
##### feature
> 1. business接口：feedback && app ci add

##### Version 1.14.100
##### feature
> 1. tribe增加version_name
> 2. 增加version接口

##### Version 1.14.99
##### bugfix
> 1. crash info前端入参判定修改

##### Version 1.14.98
##### feature
> 1. crash info增加laser日志拉取功能

##### Version 1.14.97
##### feature
> 1. 补充crash/jank info列表字段lifetime crash_type build_id

##### Version 1.14.96
##### feature
> 1. crash index接口增加解析后堆栈
> 2. crash/jank info列表根据时间倒序

##### Version 1.14.95
##### feature
> 1. 添加根据mid、buvid查询apm数据

##### Version 1.14.94
##### bugfix
> 1. 补全 MRHook 有部分情况触发的 Pipeline FAWKES_USER 值为空的逻辑

##### Version 1.14.93
##### feature
> 1. tribe2.0
> 2. fix:version_code以ci的version_code为准
> 3. fix:compatible_version

##### Version 1.14.92
##### bugfix
> 1. 业务组添加修复

##### Version 1.14.91
##### bugfix
> 1. 日志平台重复添加监控事件的限制解除

##### Version 1.14.90
##### bugfix
> 1. 监控事件生成sql字段完善

##### Version 1.14.89
##### feature
> 1. app打包新增bbr

##### Version 1.14.88
##### feature
> 1. apm查询添加"analyse_jank_stack"模糊匹配

##### Version 1.14.87
##### feature
> 1. 监控事件自动生成sql
> 2. 消除metric生成文件中sql首部与末尾的空格

##### Version 1.14.86
##### bugfix
> 1. 修正sql拼接OR的逻辑

##### Version 1.14.85
##### bugfix
> 1. 修正count接口filter逻辑

##### Version 1.14.84
##### feature
> 1. 重构moni count逻辑，增加filter筛选

##### Version 1.14.83
##### feature
> 1. 增加监控事件高级配置接口

##### Version 1.14.82
##### feature
> 1. crashRule 增加按照id查询

##### Version 1.14.81
##### bugfix
> 1. crashRule list的pn、ps增加默认值

##### Version 1.14.80
##### bugfix
> 1. 日志平台代码重构
> 2. 修复Flink生成文件格式
> 3. 修改告警指标发布的修改数的判定

##### Version 1.14.79
##### bugfix
> 1. 修正新注册 AppStore 只会更新一台机器的问题

##### Version 1.14.78
##### feature
> 1. APM读取数据Table字段切换为DistributedTableName

##### Version 1.14.77
##### feature
> 1. 添加分布式表增改查

##### Version 1.14.76
##### feature
> 1. 添加cdlist 外部接口

##### Version 1.14.75
##### feature
> 1. ctime mtime展示调整

##### Version 1.14.74
##### feature
> 1. 字段名称调整修正

##### Version 1.14.73
##### feature
> 1. 增加通过hash列表获取崩溃/卡顿index信息的接口
> 2. 补充crash/jank info字段IsHarmony LifeTime
> 3. 补充crash/jank index字段ctime mtime

##### Version 1.14.72
##### feature
> 1. 追加新接口. 通过appkey + commit 查找子仓信息

##### Version 1.14.71
##### bugfix
> 1. 修复field增加、查询逻辑
> 2. 修复metric生成的字段

##### Version 1.14.70
##### bugfix
> 1. 修复bugly符号表上传异常

##### Version 1.14.69
##### feature
> 1. event的字段增加是否进入clickhouse标识

##### Version 1.14.68
##### feature
> 1. 临时全路径搜搜java

##### Version 1.14.67
##### bugfix
> 1. flink publish接口按时间生成文件

##### Version 1.14.66
##### bugfix
> 1. scanRow接口补充
> 2. 修复卡顿对战index列表查询bug

##### Version 1.14.65
##### feature
> 1. 监控事件和数据平台与日志平台同步
> 2. 日志平台扩展字段key前缀统一,增加基础字段

##### Version 1.14.64
##### bugfix
> 1. 修复定时出包问题

##### Version 1.14.63
##### feature
> 1. 弹幕统计增加通知人

##### Version 1.14.62
##### feature
> 1. Apm下增加堆栈匹配规则

##### Version 1.14.61
##### feature
> 1. 修复 git_prj_id 字段类型异常

##### Version 1.14.60
##### feature
> 1. app表追加字段"app_dsym_name"
> 2. app表添加 "project_id" 的修改

##### Version 1.14.59
##### feature
> 1. bug fix

##### Version 1.14.58
##### bugfix
> 1. ci构建错误处理修复

##### Version 1.14.57
##### feature
> 1. 抽取ci构建逻辑， 定时出包复用

##### Version 1.14.56
##### fix
> 1. cron job的并发问题fix

##### Version 1.14.55
##### feature
> 1. 弹幕机器人修改

##### Version 1.14.54
##### feature
> 1. 弹幕关键词修改

##### Version 1.14.53
##### feature
> 1. 弹幕问题统计

##### Version 1.14.52
##### feature
> 1. 渠道包按照渠道配置自动上传cdn 

##### Version 1.14.51
##### feature
> 1. bus表追加 dc_bus_key

##### Version 1.14.50
##### bugfix
> 1. event追加数据中心的event_id进行同步

##### Version 1.14.49
##### feature
> 1. app表追加 dc_app_id

##### Version 1.14.48
##### feature
> 1. mrhook 追加 "target_branch, source_branch"

##### Version 1.14.47
##### feature
> 1. ci列表支持id搜索
> 2. ci重签名优化

##### Version 1.14.46
##### feature
> 1. 增加崩溃/卡顿堆栈解决人、解决状态等
> 2. moni数量接口增加pn ps

##### Version 1.14.45
##### feature
> 1. ci 支持commit搜索

##### Version 1.14.44
##### fix
> 1. 修复flink任务发布的文件的逻辑

##### Version 1.14.43
##### feature
> 1. 增加配置接口，返回配置平台上的前端配置

##### Version 1.14.42
##### feature
> 1. 监控长时间未完成的git merge && fix
> 2. mr_start_time逻辑调整
> 3. 加上时间范围限制的配置

##### Version 1.14.41
##### feature
> 1. config多选发布
> 2. ff搜索备注

##### Version 1.14.40
##### feature
> 1. 增加监控指标接口并修改原有接口代码逻辑
> 2. 增加监控指标接口按业务组划分的功能
> 3. 修复流处理任务关联相关逻辑

##### Version 1.14.39
##### feature
> 1. 去除apm部分接口appkey限制

##### Version 1.14.38
##### feature
> 1. 修改flink job生成Json文件的功能

##### Version 1.14.37
##### feature
> 1. iOS包构建新增企业包重签逻辑

##### Version 1.14.36
##### feature
> 1. feedback 同步 tapd

##### Version 1.14.35
##### feature
> 1. 监控事件和拓展字段方法分离
> 2. 监控事件增加字段并重构方法

##### Version 1.14.34
##### feature
> 1. dashboard 版本拉去数量 动态化

##### Version 1.14.33
##### feature
> 1. 增加监控事件的字段和查询字段

##### Version 1.14.32
##### bugfix
> 1. laser修正panic 空指针异常

##### Version 1.14.31
##### feature
> 1. 增加flink任务的接口
> 2. 增加flink和event的关联

##### Version 1.14.30
##### feature
> 1. 增加apm下prometheus接口
> 2. 修改publish和diff的格式
 
##### Version 1.14.29
##### feature
> 1. 卡顿堆栈查询

##### Version 1.14.28
##### bugfix
> 1. web端 v2 表查询兼容

##### Version 1.14.27
##### bugfix
> 1. 修复laser active 数据库查询性能异常

##### Version 1.14.26
##### bugfix
> 1. hotfix sql错误

##### Version 1.14.25
##### feature
> 1. 修复laser active 数据库查询性能异常

##### Version 1.14.24
##### feature
> 1. 增加UT
> 2. 修复重复发送的问题

##### Version 1.14.23
##### feature
> 1. nas盘清理逻辑整理
> 2. 根据ctime清理测试包

##### Version 1.14.22
##### feature
> 1. command查询优化

##### Version 1.14.21
##### feature
> 1. 优化APM查询逻辑

##### Version 1.14.20
##### feature
> 1. 修复卡片链接地址打不开

##### Version 1.14.19
##### feature
> 1. 修正apm无法过滤hash
> 2. 修正操作记录ios丢失

##### Version 1.14.18
##### feature
> 1. tf包上传增加消息通知

##### Version 1.14.17
##### feature
> 1. nas盘清理地址bug fix && 设置context为background

##### Version 1.14.16
##### feature
> 1. Laser 拉取成功回调, 追加消息通知

##### Version 1.14.15
##### feature
> 1. 修复apm堆栈解析查询条件

##### Version 1.14.14
##### feature
> 1. 优化堆栈解析query sql

##### Version 1.14.13
##### feature
> 1. 优化项目异味

##### Version 1.14.12
##### feature
> 1. 优化项目异味

##### Version 1.14.11
##### feature
> 1. 体积参数计算增加xcassets字段

##### Version 1.14.10
##### feature
> 1. ci排队告警规则改为2小时内

##### Version 1.14.9
##### feature
> 1. 增加nasci接口删除功能包类型的判断

##### Version 1.14.8
##### feature
> 1. 优化项目异味

##### Version 1.14.7
##### feature
> 1. 修复nasci接口参数异常

##### Version 1.14.6
##### feature
> 1. 用户权限表. 联表获取用户昵称

##### Version 1.14.5
##### feature
> 1. 追加堆栈模糊匹配

##### Version 1.14.4
##### feature
> 1. 定时出包修复pipline传参丢失问题

##### Version 1.14.3
##### feature
> 1. 增加读取CrashIndex和CrashInfo接口

##### Version 1.14.2
##### feature
> 1. 修复异味

##### Version 1.14.1
##### feature
> 1. 添加appConnect Command 追加日志

##### Version 1.13.125
##### feature
> 1. 客诉反馈增加昵称、编辑人、按照ctime范围检索功能

##### Version 1.13.124
##### feature
> 1. 增加查询某段时间的CI包流程
> 2. 增加更新CI包过期字段功能
> 3. 增加删除指定包的功能

##### Version 1.13.123
##### feature
> 1. patch接口修复版本索引

##### Version 1.13.122
##### feature
> 1. batch sql

##### Version 1.13.121
##### feature
> 1. 客服反馈流程根据描述内容模糊查询
> 2. update可以更新零值

##### Version 1.13.120
##### feature
> 1. 增加客服反馈流程

##### Version 1.13.119
##### bugfix
> 1. 优化sql查询效率

##### Version 1.13.118
##### bugfix
> 1. 修复数据信息

##### Version 1.13.117
##### bugfix
> 1. query 查询添加联表速率

##### Version 1.13.116
##### bugfix
> 1. 渠道包发布的时候. 记录generate变更的历史

##### Version 1.13.115
##### feature
> 1. apm route 增加memory统计

##### Version 1.13.114
##### bugfix
> 1. request url 若存在参数. 截断后面的参数；

##### Version 1.13.113
##### bugfix
> 1. 修正定时出包异常

##### Version 1.13.112
##### feature
> 1. mod open 新接口

##### Version 1.13.111
##### feature

> 1. 启动数据 - 添加version字段

##### Version 1.13.110
##### bugfix
> 1. 修正定时出包异常

##### Version 1.13.109
##### feature
> 1. moni支持internet_protocol_version查询

##### Version 1.13.108
##### feature
> 1. 修正 mod MD5

##### Version 1.13.107
##### feature
> 1. ci自动出包添加环境变量

##### Version 1.13.106
##### feature
> 1. LIKE 模糊规则改为 "%v%" -> "v%"

##### Version 1.13.105
##### feature
> 1. 慢查询优化

##### Version 1.13.104
##### feature
> 1. 慢查询优化

##### Version 1.13.103
##### bugfix
> 1. 修复写入多app_key异常

##### Version 1.13.102
##### feature
> 1. 添加config business查询接口

##### Version 1.13.101
##### feature
> 1. 添加config business写入接口

##### Version 1.13.100
##### feature
> 1. 添加用户登录列表数据

##### Version 1.13.99
##### bugfix
> 1. 修复hotfix异常

##### Version 1.13.98
##### feature
> 1. laser上报并发优化

##### Version 1.13.97
##### feature
> 1. cd 上传符号表使用channel

##### Version 1.13.96
##### feature
> 1. ci info 增加gl_job_url

##### Version 1.13.95
##### feature
> 1. laser broasdcast 离线消息对外接口

##### Version 1.13.94
##### feature
> 1. 迁移databus消费端

##### Version 1.13.93
##### bugfix
> 1. 消息推送查询build_pack备注字段遗漏修正

##### Version 1.13.92
##### feature
> 1. 增加ios推送姬邮件/企业微信推送模板
> 2. build_pack增加备注字段，新增构建时可填

##### Version 1.13.91
##### feature
> 1. laser cmd 企业微信通知

##### Version 1.13.90
##### bugfix
> 1. 模块自动分组超时问题优化

##### Version 1.13.89
##### feature
> 1. 重构发送邮件通知，增加版本姬

##### Version 1.13.88
##### bugfix
> 1. 修正patch/all4取最近版本的问题

##### Version 1.13.87
##### bugfix
> 1. 修复databus消息类型转换异常

##### Version 1.13.86
##### feature
> 1. mod下载校验

##### Version 1.13.85
##### feature
> 1. 开启databus消费

##### Version 1.13.84
##### feature
> 1. 权限管理用户名查询接口限制100条
> 2. 模块体积查询接口增加get_newest参数，为true时若当前版本不存在则取最新配置，用于未推线上的新版本查看是否超体积

##### Version 1.13.83
##### feature
> 1. 增加packAll待替换接口

##### Version 1.13.82
##### feature
> 1. versionAll upgradeAll packAll hotfixAll 增加内存cache

##### Version 1.13.81
##### feature
> 1. User-Agent和Referer校验路由去除business

##### Version 1.13.80
##### feature
> 1. mtc-marcross 三台机器没有挂载redis伴生容器. 停用databus消费

##### Version 1.13.79
##### feature
> 1. 模块自动分组功能

##### Version 1.13.78
##### feature
> 1. packAll bizApkListAll接口增加内存cache（30s）
> 2. User-Agent和Referer校验路由改为所有接口

##### Version 1.13.77
##### bugfix
> 1. 修正broadcast消息推送异常

##### Version 1.13.76
##### feature
> 1. 移除用户上线日志

##### Version 1.13.75
##### feature
> 1. Broadcast Laser v1

##### Version 1.13.74
##### feature
> 1. 增加Fawkes请求的User-Agent和Referer校验

##### Version 1.13.73
##### feature
> 1. 增加模块配置接口 增加 dashboard 总体积计算参数

##### Version 1.13.72
##### feature
> 1. 新增cd 发版消息通知接口

##### Version 1.13.71
##### feature
> 1. 机器人通知部分日志补充build_id

##### Version 1.13.70
##### bugfix
> 1. java 环境变量优化绝对路径

##### Version 1.13.69
##### bugfix
> 1. tribe business 接口新增版本字段

##### Version 1.13.68
##### bugfix
> 1. 新增tribe下发的 系统版本过滤

##### Version 1.13.67
##### bugfix
> 1. 修复企业微信通知模板折行问题

##### Version 1.13.66
##### feature
> 1. 新增anr大盘查询接口

##### Version 1.13.65
##### bugfix
> 1. 修复审核推送信息错误

##### Version 1.13.64
##### bugfix
> 1. 修复 sys_ver app_ver 参数校验

##### Version 1.13.63
##### feature
> 1. 重构ci定时出包代码

##### Version 1.13.62
##### feature
> 1. ci构建的bundle包不上传cdn

##### Version 1.13.61
##### feature
> 1. 新增laser command 上报接口

##### Version 1.13.60
##### feature
> 1. mod manager 四期

##### Version 1.13.59
##### feature
> 1. bundle包新增环境变量

##### Version 1.13.58
##### feature
> 1. laser指令扩展功能

##### Version 1.13.57
##### feature
> 1. tribe新增 pkg_type参数

##### Version 1.13.56
##### feature
> 1. database.Query()跳过lint扫描

##### Version 1.13.55
##### feature
> 1. lint bugs修复

##### Version 1.13.54
##### feature
> 1. iPhone 配置升级

##### Version 1.13.53
##### bugfix
> 1. 去除定时任务中triggerPipeline失效后十分钟重新循环构建的逻辑

##### Version 1.13.52
##### feature
> 1. 定时出包webhook添加新的透传参数 gl_job_id
> 2. 添加 laser pending 信息接口

##### Version 1.13.51
##### bugfix
> 1. 修复ci定时任务终止和删除失效的问题

##### Version 1.13.50
##### feature
> 1. 新增supervisor接口. 反馈用户基础信息

##### Version 1.13.49
##### feature
> 1. 新增编译统计字段

##### Version 1.13.48
##### bugfix
> 1. 删除 tester 频率改为每分钟删除200个

##### Version 1.13.48
##### bugfix
> 1. 修复删除 Tester 时，一个 appstore 信息错误直接 return 导致其他 App 删除失败的问题

##### Version 1.13.47
##### feature
> 1. cd 增加更新策略

##### Version 1.13.46
##### feature
> 1. ci packtype 补充

##### Version 1.13.45
##### feature
> 1. testflight 千分比最大值由 config 控制

##### Version 1.13.44
##### bugfix
> 1. 修复user_info写入接口字段顺序错误逻辑

##### Version 1.13.43
##### feature
> 1. 权限管理补充用户昵称
> 2. 查询单个包的信息查询条件由 pipline_id 改为 build_id或pipline_id

##### Version 1.13.42
##### feature
> 1. 增加企业微信图片推送接口

##### Version 1.13.41
##### bugfix
> 1. 编辑用户搜索

##### Version 1.13.40
##### bugfix
> 1. 修复渠道包状态更新

##### Version 1.13.39
##### bugfix
> 1. 修复 bizapk upload 重复上传 fix

##### Version 1.13.38
##### bugfix
> 1. 修复 bizapk upload 重复上传

##### Version 1.13.37
##### feature
> 1. app 修改信息接口，添加update icon逻辑

##### Version 1.13.36
##### bugfix
> 1. 修复 mod open config add

##### Version 1.13.35
##### bugfix
> 1. 修复 tribe 组件定时刷新任务的逻辑

##### Version 1.13.34
##### bugfix
> 1. 修复ci排队 异常告警

##### Version 1.13.33
##### bugfix
> 1. debug laserall 返回空数组

##### Version 1.13.32
##### feature
> 1. 修复subrepo Commit写入异常

##### Version 1.13.31
##### feature
> 1. 渠道包重构 -修复状态更新

##### Version 1.13.30
##### feature
> 1. 渠道包重构

##### Version 1.13.29
##### feature
> 1. 子仓信息 model 修改

##### Version 1.13.28
##### feature
> 1. mod白名单可以通过提供的三方api发布中优先级的资源

##### Version 1.13.27
##### bugfix
> 1. tribe上传的 名称大小写敏感

##### Version 1.13.26
##### feature
> 1. 修复manager 增量修改异常

##### Version 1.13.25
##### feature
> 1. 修复manager 增量修改异常

##### Version 1.13.24
##### feature
> 1. CI Upload接口添加可选参数"subrepo_commits"

##### Version 1.13.23
##### feature
> 1. 新增cienv 删除当前app的接口

##### Version 1.13.22
##### bugfix
> 1. config历史接口新增

##### Version 1.13.21
##### bugfix
> 1. 修复config新增同名同值的bug

##### Version 1.13.20
##### bugfix
> 1. 修复正式版也会推送tf组的问题

##### Version 1.13.19
##### bugfix
> 1. 修复ci等待状态上报的sql

##### Version 1.13.18
##### bugfix
> 1. 从 BetaGroup 删除 Tester 的逻辑改为从 App 删除

##### Version 1.13.17
##### feature
> 1. 添加查询日志

##### Version 1.13.16
##### bugfix
> 1. 修复删除 Testers 的逻辑问题

##### Version 1.13.15
##### feature
> 1. ci upload 添加日志

##### Version 1.13.14
##### feature
> 1. 增加prometheus上报模块，新增ci 打包队列中的数量上报

##### Version 1.13.13
##### feature
> 1. 业务分享渠道对外设置Config接口. 开放历史查询接口

##### Version 1.13.12
##### feature
> 1. 业务分享渠道对外设置Config接口

##### Version 1.13.11
##### feature
> 1. APM Config/FF SQL优化

##### Version 1.13.10
##### feature
> 1. bizapk 新增built_in字段的上报和查询

##### Version 1.13.9
##### bugfix
> 1. ci job统计增加joburl, 修复compile统计错误

##### Version 1.13.8
##### feature
> 1. ci编辑统计上报接口优化

##### Version 1.13.7
##### feature
> 1. ci编辑统计上报接口

##### Version 1.13.6
##### bugfix
> 1. 相同mod名不可重复创建提示

##### Version 1.13.5
##### feature
> 1. 分发人数更新频率变为10分钟
> 2. 统计人数超过分发上线逻辑区分环境

##### Version 1.13.4
##### bugfix
> 1. 修复stage筛选

##### Version 1.13.2
##### feature
> 1. fawkes新增job 数据统计 和laser数据统计

##### Version 1.13.1
##### bugfix
> 1. 新增全量mod拆分接口

##### Version 1.13.0
##### bugfix
> 1. 修复 testflight 推送 CD 的问题

##### Version 1.12.99
##### bugfix
> 1. 修复 err = rows.Err()

##### Version 1.12.98
##### bugfix
> 1. 优化fawkes的全量数据接口返回

##### Version 1.12.97
##### bugfix
> 1. 优化fawkes统计通用接口

##### Version 1.12.96
##### feature
> 1. Fawkes 聚合数据通用接口

##### Version 1.12.95
##### bugfix
> 1. tf 超时问题修复

##### Version 1.12.94
##### bugfix
> 1. 临时兼容 PATH 找不到 java 的问题

##### Version 1.12.93
##### bugfix
> 1. mod全量接口增加mtime字段

##### Version 1.12.92
##### bugfix
> 1. 停止分发增加测试环境逻辑
> 2. tf 定时刷新任务限定只在一台机器执行
> 3. 恢复分发人数功能

##### Version 1.12.91
##### bugfix
> 1. cimdl 重复引用

##### Version 1.12.90
##### feature
> 1. ci job 记录接口

##### Version 1.12.89
##### feature
> 1. 配合升级 gitlab

##### Version 1.12.88
##### feature
> 1. 消息推送. 用户去重

##### Version 1.12.87
##### feature
> 1. 添加修改昵称接口
> 2. 用户列表ps上限去除
> 3. 用户列表查询添加昵称字段

##### Version 1.12.86
##### feature
> 1. testflight 测试环境分发修复

##### Version 1.12.85
##### feature
> 1. ci统计优化

##### Version 1.12.84
##### feature
> 1. Laser日志静默上报串消逻辑修正

##### Version 1.12.83
##### feature
> 1. CI mobile-ep 文件上传接口优化

##### Version 1.12.82
##### bugfix
> 1. 修复lastpack 逻辑

##### Version 1.12.81
##### feature
> 1. CI mobile-ep 文件上传接口

##### Version 1.12.80
##### feature
> 1. ci环境变量新增是否可推送cd开关

##### Version 1.12.79
##### bugfix
> 1. testflight 接口修复

##### Version 1.12.78
##### feature
> 1. fawkes 企业应用 - 消息推送接口

##### Version 1.12.77
##### feature
> 1. TestFlight 增加黑白名单

##### Version 1.12.76
##### feature
> 1. ci模型修改

##### Version 1.12.75
##### feature
> 1. ci构建统计

##### Version 1.12.74
##### feature
> 1. 异步 load mod 缓存

##### Version 1.12.73
##### feature
> 1. APM事件配置数据接口查询

##### Version 1.12.72
##### feature
> 1. TestFlight 包可推正式环境，测试环境和正式环境数据隔离
> 2. 可设置测试环境和正式环境的外部公开链接
> 3. CI 推 CD 时就上传符号表

##### Version 1.12.71
##### feature
> 1. Fawkes CI 统计接口

##### Version 1.12.70
##### feature
> 1. 添加laser_id查询主动列表逻辑

##### Version 1.12.69
##### bugfix
> 1. 上传 dsym 前增加 JAVA_HOME 环境遍量
> 2. 去除 apk 双写 BOSS 的逻辑
> 3. 机器人微信消息只显示归档路径第一层的文件，不再遍历

##### Version 1.12.68
##### feature
> 1. 应用申请添加管理员通知

##### Version 1.12.67
##### feature
> 1. 新增laser解析状态修改的接口

##### Version 1.12.66
##### fix
> 1. 修改接口权限

##### Version 1.12.65
##### fix
> 1. 修改接口文案

##### Version 1.12.64
##### feature
> 1. monkey接口优化

##### Version 1.12.63
##### feature
> 1. 添加打包消息推送日志信息

##### Version 1.12.62
##### bugfix
> 1. 修复 mod 权限

##### Version 1.12.61
##### bugfix
> 1. 修复 mod 权限

##### Version 1.12.60
##### bugfix
> 1. 修复 mod 版本号校验

##### Version 1.12.59
##### bugfix
> 1. 修复 mod 推动正式

##### Version 1.12.58
##### Feature
> 1. app表添加 debug_url 参数

##### Version 1.12.57
##### Feature
> 1. mod 权限管理

##### Version 1.12.56
##### bugfix
> 1. ci耗时接口变更

##### Version 1.12.54
##### bugfix
> 1. 时间格式错误修正

##### Version 1.12.53
##### Feature
> 1. EP MonkeyTest相关功能接口服务

##### Version 1.12.52
##### Feature
> 1. ci 构建耗时

##### Version 1.12.51
##### Feature
> 1. mod 测试角色允许推送正式

##### Version 1.12.50
##### Feature
> 1. mod 增加角色接口

##### Version 1.12.49
##### Feature
> 1. patch打包git分之改为keep/build_patch_pipeline

##### Version 1.12.48
##### Feature
> 1. patch打包git分之改为master

##### Version 1.12.47
##### Feature
> 1. ci环境变量

##### Version 1.12.46
##### Feature
> 1. APM. 添加GroupKey索引条件

##### Version 1.12.45
##### Feature
> 1. CI接口删除冗余代码

##### Version 1.12.44
##### Bugfix
> 1. 请求方法修改为GET

##### Version 1.12.43
##### Feature
> 1. 添加APM自定义运算查询接口
> 2. 开放CIList对外查询

##### Version 1.12.42
##### Bugfix
> 1. mod fix

##### Version 1.12.41
##### Feature
> 1. 自定义埋点. 添加新的字段

##### Version 1.12.40
##### Feature
> 1. patchAll数据迁移至缓存
> 2. laser上报添加"task_id"字段

##### Version 1.12.39
##### Bugfix
> 1. app审核接口中appPass无法查询app信息的bug

##### Version 1.12.38
##### Feature
> 1. hotpatch. 添加返回字段 cdn_url

##### Version 1.12.37
##### Feature
> 1. 添加通过id获取config历史版本接口

##### Version 1.12.36
##### Feature
> 1. 添加通过id获取ff历史版本接口
> 2. APM添加新的筛选字段

##### Version 1.12.35
##### Bugfix
> 1. 修复修改 AppStore 配置后只触发一台机器更新内存中的 AppStore 配置的问题

##### Version 1.12.34
##### Feature
> 1. cd列表新增gl_job_url

##### Version 1.12.33
##### Feature
> 1. patch默认android

##### Version 1.12.32
##### Feature
> 1. patch包上传限制

##### Version 1.12.31
##### Feature
> 1. 应用激活逻辑修正

##### Version 1.12.30
##### Feature
> 1. 添加应用停用状态

##### Version 1.12.29
##### Feature
> 1. 云视听渠道包同步manager对外接口调整

##### Version 1.12.28
##### Feature
> 1. patch打包流程重构

##### Version 1.12.27
##### Feature
> 1. 云视听业务. 对外查询上传到cdn的包信息

##### Version 1.12.26
##### Feature
> 1. 修改引导参与 testflight 内测的默认文案

##### Version 1.12.25
##### Feature
> 1. 修复patch status默认值问题

##### Version 1.12.24
##### Feature
> 1. Config. FF 历史页面新增分页

##### Version 1.12.23
##### Feature
> 1. mod 支持多种CDN

##### Version 1.12.23
##### Feature
> 1. Web前端Config兼容支持

##### Version 1.12.22
##### Bugfix
> 1. patch表结构修改，patchALl sql增加状态条件

##### Version 1.12.21
##### Bugfix
> 1. 修复 CD 自动推正式环境后 testflight info 没有复制到正式环境的问题
> 2. 增加轮询任务，如果外部测试组发生变化可以及时更新

##### Version 1.12.20
##### Feature
> 1. 限制类型资源禁止修改操作

##### Version 1.12.19
##### Bugfix
> 1. 修正条件查询异常

##### Version 1.12.18
##### Feature
> 1. 新增 testflight 对外数据接口

##### Version 1.12.17
##### Feature
> 1. appPass 增加当前角色的字段

##### Version 1.12.16
##### Feature
> 1. 支持web端apm查询. 添加clickhouse双通道配置

##### Version 1.12.15
##### Feature
> 1. 临时下线定时打包接口

##### Version 1.12.14
##### Bugfix
> 1. 修复 mod 已知的问题

##### Version 1.12.13
##### Feature
> 1. 线上包过审后自动设置 tag, 符号表上传bugly，推正式环境
> 2. 增加设置 TF 包引导升级文案，提醒升级文案，强制升级文案功能

##### Version 1.12.12
##### Bugfix
> 1. 修复 mod 已知的问题

##### Version 1.12.11
##### Bugfix
> 1. appInfo缺省id则查appPass

##### Version 1.12.10
##### Bugfix
> 1. 修复 appInfo 查询问题

##### Version 1.12.9
##### Bugfix
> 1. 修复 mod 已知的问题

##### Version 1.12.8
##### Bugfix
> 1. 关闭前端埋点接口

##### Version 1.12.7
##### Feature
> 1. mod 功能发布

##### Version 1.12.6
##### Feature
> 1. 获取线上版本资料
> 2. App store 设置新增 tag prefix, bugly app id, bugly app key

##### Version 1.12.5
##### Feature
> 1. 新增机器人管理
> 2. ci推送可选

##### Version 1.12.4
##### Bugfix
> 1. 修复 size type sql 过慢的问题
> 2. 修复新增 app store 配置不去 appstoreconnect 里注册的问题

##### Version 1.12.3
##### Feature
> 1. 自定义埋点. 添加85 90分位

##### Version 1.12.2
##### Feature
> 1. saga 通知接口失败不影响通知机器人

##### Version 1.12.1
##### Feature
> 1. APM 崩溃率聚合数据接口. 调整支持全版本数据直接查询
> 2. 新增 Monkey自动化测试服务

##### Version 1.12.0
##### Bugfix
> 1. 优化体积接口速度，修复超时

##### Version 1.11.99
##### Feature
> 1. 添加 fawkes-fe 前端埋点接口. 用于监控前端用户体验报错异常&&日活等使用情况

##### Version 1.11.98
##### Feature
> 1. testflight 分发人数统计逻辑变更，由大表查询(只支持一周)更改为独立表查询(时间跨度可以更长)

##### Version 1.11.97
##### Bugfix
> 1. 修复事件组列表筛选功能

##### Version 1.11.96
##### Feature
> 1. 新增接口获取还在分发中的 testflight 包

##### Version 1.11.95
##### Feature
> 1. 增加 moni 的 jar 备份上传接口

##### Version 1.11.94
##### Feature
> 1. 聚合表查询timestamp索引timeRange修正

##### Version 1.11.93
##### Bugfix
> 1. 修复 testflight 分发人数

##### Version 1.11.92
##### Feature
> 1. 渠道包对外查询接口

##### Version 1.11.91
##### Bugfix
> 1. 修复 appstore connect 出口协议证明选择否以后 fawkes 无法更新 beta 状态的问题

##### Version 1.11.90
##### Bugfix
> 1. 解决 appstore connect 包修改接口报错的问题
> 2. 解决 testflight 功能中某些轮询任务超时的问题

##### Version 1.11.89
##### Bugfix
> 1. TestFlight 后端接口部分问题修正

##### Version 1.11.88
##### Feature
> 1. 修复OTT拉取android laser日志异常

##### Version 1.11.87
##### Bugfix
> 1. 修正 cd list 接口
> 2. 分发千分比添加千分之十的上限

##### Version 1.11.86
##### Feature
> 1. TestFlight 后端接口完成

##### Version 1.11.85
##### Feature
> 1. App 新增 TestFlight 配置

##### Version 1.11.84
##### Bugfix
> 1. 修复webhook启动数据异常

##### Version 1.11.83
##### Bugfix
> 1. 修正新的网络成功率/失败率sql

##### Version 1.11.82
##### Feature
> 1. 网络成功失败率. 添加新的配置项查询新的降级成功率。 原成功率key还原保持不变

##### Version 1.11.81
##### Bugfix
> 1. 修复 bizapk 上传 cdn 偶尔超时的问题

##### Version 1.11.80
##### Feature
> 1. apm网络成功率算法优化

##### Version 1.11.79
##### Feature
> 1. apm网络成功率算法优化

##### Version 1.11.78
##### Feature
> 1. 新增渠道分页接口

##### Version 1.11.77
##### Bugfix
> 1. 渠道包-500异常修复

##### Version 1.11.76
##### Feature
> 1. 网络追溯信息 - 添加小时/天 打点数据查询

##### Version 1.11.75
##### Bugfix
> 1. fawkes功能函数读取应用属性函数. AppInfo改为AppPass

##### Version 1.11.74
##### Feature
> 1. laser active接口提速； 秒级提升至毫秒级

##### Version 1.11.73
##### Feature
> 1. laser任务logdate必选改为可选

##### Version 1.11.72
##### Bugfix
> 1. bizapk 的 meta 和 mapping 参数改为非必传

##### Version 1.11.71
##### Bugfix
> 1. 修复渠道默认参数问题
> 2. 增加批量操作渠道分组

##### Version 1.11.70
##### Feature
> 1. 新增 渠道分组功能

##### Version 1.11.69
##### Feature
> 1. 新增 网络聚合表查询
> 2. 优化 CI列表 job_path 转化逻辑
> 3. 优化 Apm 成功率/失败率 算法

##### Version 1.11.68
##### Bugfix
> 1. 修复APM WHERE查询条件
> 2. 新增laser report2接口

##### Version 1.11.67
##### Feature
> 1. 接入苹果后台 API

##### Version 1.11.66
##### Bugfix
> 1. 修复Webhook消息推送异常

##### Version 1.11.65
##### Feature
> 1. FF/Config 添加消息推送至管理员

##### Version 1.11.64
##### Feature
> 1. 添加崩溃率webhook接口

##### Version 1.11.63
##### Feature
> 1. laser接口添加新字段 desc,mobi_app,build

##### Version 1.11.62
##### Feature
> 1. 审核列表&&关注列表 添加mobi_app字段

##### Version 1.11.61
##### Feature
> 1. app申请增加mobi字段
> 2. 审核通过同步设置为管理员
> 3. 审核增加可编辑字段

##### Version 1.11.60
##### Feature
> 1. 修复申请消息发送错误

##### Version 1.11.59
##### Feature
> 1. 权限申请列表增加state搜索

##### Version 1.11.58
##### Feature
> 1. oss 双写到 Boss

##### Version 1.11.57
##### Bugfix
> 1. FF消息通知url修正

##### Version 1.11.56
##### Bugfix
> 1. 修正权限操作回执逻辑

##### Version 1.11.55
##### Feature
> 1. 接入comet进行对企业微信，邮箱，电话的通知告警功能
> 2. 替换原markdown消息推送

##### Version 1.11.54
##### Bugfix
> 1. flowmap查询 增加 real_name_from

##### Version 1.11.53
##### Bugfix
> 1. apm 增加路由别名管理

##### Version 1.11.52
##### Bugfix
> 1. 未生效默认值使用占位符

##### Version 1.11.51
##### Bugfix
> 1. 未生效默认值为-1； 将设置1调整为-1

##### Version 1.11.50
##### Bugfix
> 1. 移除权限申请.req异常的错误捕获

##### Version 1.11.49
##### Feature
> 1. 新增网关版本列表接口

##### Version 1.11.48
##### Bugfix
> 1. 修复 Apm CommandList' resp total数据缺失
> 2. 优化 Apm Commands查询条件剔除 bus_id查询限制

##### Version 1.11.47
##### Bugfix
> 1. 修复未推 CD 的包无法查询全模块大小分布的问题

##### Version 1.11.46
##### Bugfix
> 1. 优化聚合表查询逻辑

##### Version 1.11.45
##### Bugfix
> 1. 分组删除从 DELETE 改为 POST

##### Version 1.11.44
##### Feature
> 1. 删除触发 size 上传的接口（已无调用）
> 2. 增加修改和删除 group 的接口
> 3. 增加单个版本全 group size 分布接口

##### Version 1.11.43
##### Feature
> 1. 添加laser对外业务查询接口

##### Version 1.11.42
##### Feature
> 1. 机器人文件上传接口

##### Version 1.11.41
##### Bugfix
> 1. ci推送cd cd推送正式默认不生效

##### Version 1.11.40
##### Feature
> 1. CI 新建构建的时候返回 pipeline 信息
> 2. tribe 组件列表返回 pipeline 和 job 的 URL

##### Version 1.11.39
##### Bugfix
> 1. 修复激活按钮失效

##### Version 1.11.38
##### Bugfix
> 1. 流量配置增加 pack_build_id 参数

##### Version 1.11.37
##### Bugfix
> 1. 修复 CI 推 CD 时 cdn 只传一个包的问题

##### Version 1.11.36
##### Feature
> 1. tribe 组件接口

##### Version 1.11.35
##### Bugfix
> 1. 修复Apm查询异常(添加Where匹配条件)
> 2. 还原1.11.34的修改内容. 通过修改源机器的host解决根本问题

##### Version 1.11.34
##### Bugfix
> 1. 修正数据报告req异常情况 - x509: certificate is valid

##### Version 1.11.33
##### Feature
> 1. 新增 fawkes 消息推送至管理员接口

##### Version 1.11.32
##### Feature
> 1. 新增 fawkes 公告管理列表，新增，编辑接口

##### Version 1.11.31
##### Feature
> 1. FF、Config发布通知添加管理员通知信息

##### Version 1.11.30
##### Feature
> 1. 提供fawkes的两个查询crash和setup的聚合数据的接口

##### Version 1.11.29
##### Bugfix
> 1. 修正最近稳定版的查询条件

##### Version 1.11.28
##### Feature
> 1. 提供fawkes的auth supervisor接口用来判断用户是否是超级管理员

##### Version 1.11.27
##### Feature
> 1. apm 查询逻辑优化

##### Version 1.11.26
##### Feature
> 1. generate/list 列表分页 , 排序, 筛选
> 2. /pack/latestStable 增加 version_code 筛选

##### Version 1.11.25
##### Feature
> 1. CI 新增 文件上传接口
> 2. CI 新增 构建报告查询

##### Version 1.11.24
##### Feature
> 1. 通知机器人整合到通知群组接口，CI前端不再单独调用

##### Version 1.11.23
##### Feature
> 1. 添加自定义渠道添加权限（对外服务. 仅支持动态渠道, 不支持静态渠道）

##### Version 1.11.22
##### Feature
> 1. 体积统计 sum 逻辑变更，原本所有类型相加变为只加 code 和 res 类型（android 有 apk 和 arr 两种不同的计算体积的方式，都有上传，但全部相加会导致 sum 大小变为约 2倍）

##### Version 1.11.21
##### Bugfix
> 1. config 修改配置项描述信息异常

##### Version 1.11.20
##### Feature
> 1. apm 自定义事件；事件组Apis && 自定义事件查询

##### Version 1.11.19
##### Feature
> 1. ctime参数在转到pack dao时需要改成ptime接收已修复

##### Version 1.11.18
##### Feature
> 1. business/pack/latestStable接口数据展示不全，增加更多的数据

##### Version 1.11.17
##### Feature
> 1. fawkes-admin增加business/pack/latestStable接口做最近一次稳定版本的查询

##### Version 1.11.16
##### Feature
> 1. 增加移动端 release 分支开出时的 webhook，辅助发布分支锁定 bapis commit 点
> 2. 精简 CI 模块一些稳定功能的 log

##### Version 1.11.16
##### Bugfix
> 1. 修复laser查询接口报错  

##### Version 1.11.15
##### Feature
> 1. 增加laser静默推送回调接口  
> 2. 增加静默推送uri和推送状态  
> 3. 增加获取静默推送任务的逻辑

##### Version 1.11.14
##### Bugfix
> 1. 尝试修复patch包生成异常的问题  

##### Version 1.11.14
##### Bugfix
> 1. 修复 CI 列表按构建号搜索时返回数量不对的问题

##### Version 1.11.13
##### Bugfix
> 1. 修复 BUILD_ID 为 0

##### Version 1.11.12
##### Feature
> 1. CI构建接口增加自定义环境变量参数
> 2. CI列表接口返回增加自定义环境变量
> 3. CI列表接口增加根据构建号查询
> 4. 模块体积上报 context 超时处理

##### Version 1.11.11
##### Bugfix
> 1. CI定时任务修复状态判断逻辑  

##### Version 1.11.10
##### Feature
> 1. 添加CD列表筛选接口

##### Version 1.11.9
##### Feature
> 1. 构建任务sleep时间改为1秒  

##### Version 1.11.8
##### Feature
> 1. 优化定时构建任务队列 

##### Version 1.11.7
##### Feature
> 1. 排序方式从创建时间变更为 version code 排序

##### Version 1.11.6
##### Feature
> 1. 新增单版本单模块组体积分布接口

##### Version 1.11.5
##### Feature
> 1. ci/info 接口去除身份校验
> 2. ci自动出包回执，添加app_key字段

##### Version 1.11.4
##### Feature
> 1. 模块列表接口增加空模块组（无下属模块）的数据
> 2. 增加组列表接口和体积类型接口

##### Version 1.11.3
##### Bugfix
> 1. 修复一个新的库上传多类型 size 时重复 insert 导致的报错

##### Version 1.11.2
##### Feature
> 1. 定时出包文案修改
> 2. apm http异常条件优化

##### Version 1.11.1
##### Feature
> 1. 新增模块大小统计接口

##### Version 1.11.0
##### Feature
> 1. 定时任务单机器运行方案调整  
> 2. 增加删除接口  

##### Version 1.10.2
##### Feature
> 1. 定时任务增加状态回写和单机器执行  

##### Version 1.10.1
##### Feature
> 1. 定时任务增加结果通知  

##### Version 1.10.0
##### Feature
> 1. 定时任务迁移fawkes  

##### Version 1.9.4
##### Feature
> 1.增加对外的频道数据查询接口  
> 2.增加定时任务管理接口和定时逻辑  

##### Version 1.9.3
##### Feature
> 1. apm 网络错误率sql 过滤ios端取消数据

##### Version 1.9.2
##### Feature
> 1. laser用户上报日志超时优化  

##### Version 1.9.1
##### Feature
> 1. 添加apm大版本接口
> 2. 版本list添加过滤分页项

##### Version 1.8.21
##### Feature
> 1. 增加根据appKey和buildID查询包信息接口

##### Version 1.8.20
##### Bugfix
> 1. 修复部分情况下代码合入以后不会触发 pipeline 的问题

##### Version 1.8.19
##### Feature
> 1. android 子仓 push 仍想保留触发 pipeline 的能力（不检查是否有 MR）

##### Version 1.8.18
##### Feature
> 1. 合入代码最后都要 trigger pipeline

##### Version 1.8.17
##### Feature
> 1. gitlab push webhook 新增对子仓 MR 是否存在的判断
> 2. 修复 hotfix 打包完成后的通知邮件中 job 链接不正确的问题

##### Version 1.8.16
##### Feature
> 1. 新增 BFS CDN 刷新接口

##### Version 1.8.15
##### Bugfix
> 1. saga 因为找不准子仓触发的 pipeline 导致无法合并的问题的逻辑修正

##### Version 1.8.14
##### Bugfix
> 1. 首次创建 MR 的判断条件修正

##### Version 1.8.13
##### Features
> 1. 增加 mtime，给 CD 前端展示时间用
> 2. 发 MR 时再触发 Pipeline
> 3. 解决主仓没有修改时，saga 因为找不准子仓触发的 pipeline 而无法合并的问题

##### Version 1.8.12
##### Bugfix
> 1. 触发 rebuild 时加上参数 PUT_CACHE_ALL，上传所有缓存

##### Version 1.8.11
##### Bugfix
> 1. 获取不到上一次 Internal Version 的报错修复

##### Version 1.8.10
##### Bugfix
> 1. Internal Version Code 补异常
> 2. 获取最后一个 Internal verison code 时加上 app key 的限制条件

##### Version 1.8.9
##### Bugfix
> 1. 修复相同 Internal Version Code 可以推送生产环境的问题

##### Version 1.8.8
##### Features
> 1. 热修包 Internal version 逻辑变更，由前端创建时输入
> 2. 热修包 Push 生产环境的时候判断 Internal version code 是否大于生产环境最后一个热修包

##### Version 1.8.7
##### Features
> 1. merge release 分支的时候提醒合并 master 分支

##### Version 1.8.6
##### Features
> 1. 修正同步 manager url 异常

##### Version 1.8.5
##### Features
> 1. 修正applist manager_plat字段不同步问题

##### Version 1.8.4
##### Features
> 1. 上传失败的laser任务不再重试  

##### Version 1.8.3
##### Features
> 1. 子仓 pipeline 从 Merge Request 的 Descriptions 迁移到 comments 中

##### Version 1.8.2
##### Features
> 1. CI 打包通知新增包类型和包大小

##### Version 1.8.1
##### Features
> 1. laser增加主动上报记录查询接口

##### Version 1.8.0
##### Features
> 1. laser增加新接口记录主动触发的任务  

##### Version 1.7.31
##### Bugfix
> 1. 修复同步Manager异常

##### Version 1.7.30
##### Features
> 1. 自建渠道相关接口

##### Version 1.7.29
##### Features
> 1. 同步Manager接口支持多APP

##### Version 1.7.28
##### Features
> 1. 刷新cdn参数优化

##### Version 1.7.27
##### Features
> 1. 添加依赖发布接口

##### Version 1.7.26
##### Features
> 1. CDN刷新接口. 渠道包发布后自动刷新

##### Version 1.7.25
##### Features
> 1. 添加接口 - 获取Config、FF修改数

##### Version 1.7.24
##### Features
> 1. FF列表添加关键字过滤

##### Version 1.7.23
##### Bugfix
> 1. 消息机器人Encode函数优化

##### Version 1.7.22
##### Bugfix
> 1. 修复通知文案异常

##### Version 1.7.21
##### Bugfix
> 1. 添加超管通知

##### Version 1.7.20
##### Features
> 1. 代码检查的 MR hook 追加 WorkInProgress 的判断
> 2. 主仓 MR webhook 追加 WorkInProgress 检查
> 3. 子仓 Push webhook 在 MR Description 中追加 pipeline 地址

##### Version 1.7.19
##### Features
> 1. 新增Conf、FF申请审核功能

##### Version 1.7.18
##### Features
> 1. 支持CI静态代码分析功能

##### Version 1.7.17
##### Bugfix
> 1. ptime复原  

##### Version 1.7.16
##### Bugfix
> 1. 代码检查的 MR hook 只在 create MR 的时候做
> 2. 修复没有 Assignee 时报错的问题

##### Version 1.7.15
##### Features
> 1. 新增代码检查的 MR hook

##### Version 1.7.14
##### Bugfix
> 1. 修复时间显示

##### Version 1.7.13
##### Bugfix
> 1. 增量包比较消耗资源，降低增量包并发数防止因为资源竞争打增量包失败

##### Version 1.7.12
##### Bugfix
> 1. 子仓 push 时，如果主仓没有同名分支则创建子仓的同名分支，此时由于主仓相当于 push 了一次，已经跑了一次 pipeline，不需要用 trigger 再跑一次
> 2. 修复用 bbgit 给主仓创建关联 MR 时，子仓的 webhook 和 bbgit 并发 Update，导致有时主仓 MR 的 description 被刷掉的问题

##### Version 1.7.11
##### Bugfix
> 1. 操作记录分页修正

##### Version 1.7.10
##### Features
> 1. 添加同步到Manager接口服务

##### Version 1.7.9
##### Features
> 1. 邮件标题显示 app 名

##### Version 1.7.8
##### Bugfix
> 1. 修复渠道包功能异常

##### Version 1.7.7
##### Features
> 1.ecode修改

##### Version 1.7.6
##### Features
> 1. rebuild 接口返回 pipeline 信息
> 2. 提供 buildstatus 接口查询 pipeline 状态

##### Version 1.7.6
##### Bugfix
> 1. 忽略删除远程分支引起的 push events，防止子仓 Merge 后自动删除分支的 events 触发主仓 pipeline

##### Version 1.7.5
##### Bugfix
> 1. 修复 push hook 分支名，去除 refs/heads/ 前缀

##### Version 1.7.4
##### Features
> 1. 忽略由 fawkes 账号联动 Merge 产生的子仓 Push events，防止一个联动 Merge 使得多个子仓 Push events 触发多次主仓 pipeline
> 2. 增加主仓重新 build 的接口，防止缓存失效和单个库缺少 32 位系统的缓存
> 3. 子仓 Push events 发生时，如果主仓不存在相应分支，则从主仓拉一个同名分支做 pipeline

##### Version 1.7.3
##### Bugfix
> 1. 修复CD发布功能异常

##### Version 1.7.2
##### Bugfix
> 1. 修复CD发布功能异常

##### Version 1.7.1
##### Bugfix
> 1. 修复关注列表读取异常

##### Version 1.7.0
##### Features
> 1. 增加升级配置开关  
> 2. 日志模块  
> 3. CD列表根据版本大小排序 

##### Version 1.6.17
##### Features
> 1. 移动端分仓 webhook

##### Version 1.6.16
##### Bugfix
> 1. CI表查询DAO修复异常

##### Version 1.6.15
##### Features
> 1. 冒烟测试、Monkey测试接入Fawkes

##### Version 1.6.14
##### Features
> 1. config配置文件添加部门ids配置，动态读取拉取用户列表

##### Version 1.6.13
##### Bugfix
> 1. 修复权限审核接口功能异常

##### Version 1.6.12
##### Features
> 1. 添加UserCache. 向指定用户推送消息通知

##### Version 1.6.11
##### Features
> 1. 添加Fawkes企业应用消息推送
> 2. 申请审核权限，添加向固定管理员推送消息推送

##### Version 1.6.10
##### Bugfix
> 1. 修正应用管理员列表接口异常

##### Version 1.6.9
##### Features
> 1. 添加权限申请审核接口

##### Version 1.6.8
##### Bugfix
> 1. 应用列表接口添加新增响应字段

##### Version 1.6.7
##### Features
> 1. 添加机器人信息设置功能

##### Version 1.6.6
##### Features
> 1. 添加机器人消息通知

##### Version 1.6.5
##### Bugfix
> 1. 修复权限管理列表接口筛选功能

##### Version 1.6.4
##### Bugfix
> 1. 只有 Android 才做增量包

##### Version 1.6.3
##### Features
> 1. hotfix 邮件通知修复
> 2. 新增机器人通知公用方法 

##### Version 1.6.2
##### Features
> 1. 应用配置版本格式修改  
> 2. 应用配置增加配出版本项 

##### Version 1.6.1
##### Features
> 1. 权限模块列表  

##### Version 1.6.0
##### Features
> 1. 权限模块 

##### Version 1.5.19
##### Features
> 1. 邮件格式调整
> 2. 解决 multipart form 超时导致邮件多次发送的问题

##### Version 1.5.18
##### Features
> 1. ci 新增 saga 通知功能
> 2. hotfix 新增邮件和 saga 通知功能
> 3. ci 新增发送群组接口
> 4. 修复 macross 因为 https 无法同步 manager 的问题
> 5. gitlab trigger 的变量 USER 修改为 FAWKES_USER，修复 USER 变量覆盖打包机本地变量的问题

##### Version 1.5.17
##### Features
> 1. ci 打包自动生成二维码
> 2. ci 构建时增加抄送邮件组功能
> 3. ci 构建完成后自动发送邮件给操作者，如构建时勾选，则抄送邮件组

##### Version 1.5.16
##### Features
> 1. macross 发送邮件接口迁移
> 2. app 新增邮件通知列表接口
> 3. app 新增更新邮件通知列表接口

##### Version 1.5.15
##### Feature
> 1. 构建包查询接口添加稳定状态字段
> 2. 添加CD构建包备注修改接口

##### Version 1.5.14
##### Bugfix
> 1. laser去掉email

##### Version 1.5.13
##### Bugfix
> 1. 优化静态渠道同步流程

##### Version 1.5.12
##### Bugfix
> 1. 修复应用申请接口异常

##### Version 1.5.11
##### Features
> 1. 添加CD渠道包自动化测试状态

##### Version 1.5.10
##### Bugfix
> 1. Config、Feature Flag历史列表发布时间修复
> 2. Channel静态渠道列表调整排序逻辑

##### Version 1.5.9
##### Features
> 1. 申请应用后. 同步静态渠道到应用渠道内

##### Version 1.5.8
##### Features
> 1. 增加权限白名单  

##### Version 1.5.7
##### Features
> 1. 增加获取laser信息的接口  

##### Version 1.5.6
##### Features
> 1. laser增加用户筛选  

##### Version 1.5.5
##### Bugfix
> 1. fix 渠道包内网下载地址  

##### Version 1.5.4
##### Bugfix
> 1. fix patch接口指针问题  

##### Version 1.5.3
##### Features
> 1. 全量laser接口  
> 2. 增量包并发改为10  

##### Version 1.5.2
##### Bugfix
> 1. 全量config日志version问题  

##### Version 1.5.1
##### Features
> 1. 增量包并发  

##### Version 1.5.0
##### Features
> 1. laser后台接口  

##### Version 1.4.4
##### Bugfix
> 1. fix all timestamp

##### Version 1.4.3
##### Features
> 1. 修复config全量历史接口  

##### Version 1.4.2
##### Features
> 1. 全量config发布历史接口增加page信息  

##### Version 1.4.1
##### Bugfix
> 1. fix ci timestamp
> 2. fix patch count

##### Version 1.4.0
##### Features
> 1. 添加静态渠道增加参数是否同步到所有app  
> 2. config增加全量发布历史接口 

##### Version 1.3.18
##### Feature
> 1. CD过滤配置去掉channel的默认值  

##### Version 1.3.17
##### Bugfix
> 1. 权限接口获取tree_path的问题

##### Version 1.3.16
##### Bugfix
> 1. fix ci sql

##### Version 1.3.15
##### Bugfix
> 1. 增加新的全量patch接口  

##### Version 1.3.14
##### Features
> 1. 新增 internal version code
> 2. 修复 cd insert r_url 错位到 r_mapping_url 的问题

##### Version 1.3.13
##### Bugfix
> 1. 修复 BuildPack 的 r_mapping_url

##### Version 1.3.12
##### Features
> 1. 新增 r_mapping_url
> 2. 粉版和国际版渠道包特殊处理

##### Version 1.3.11
##### Features
> 1. ci 推送 cd 自动生成增量包
> 2. gitlab trigger 新增 USER 参数
> 3. 增加推送 macross 服务

##### Version 1.3.10
##### Bugfix
> 1. 修复 gitlab branch 服务

##### Version 1.3.9
##### Bugfix
> 1. 修复 ci upload 超时问题

##### Version 1.3.8
##### Bugfix
> 1. 修复 ci list 排序问题

##### Version 1.3.7
##### Bugfix
> 1.修复channel sql语句拼接问题
> 2.修复静态渠道列表返回重复数据问题

##### Version 1.3.6
##### Bugfix
> 1.修复config添加版本问题 

##### Version 1.3.5
##### Features
> 1.增加app编辑接口  

##### Version 1.3.4
##### Bugfix
> 1.修复ci中sql格式问题  
> 2.ff发布放宽无效配置过滤  

##### Version 1.3.3
##### Features
> 1.ci update 的 version & version_code 改为非必传参数

##### Version 1.3.2
##### Features && Bugfix
> 1.修改sql注入问题第三弹  

##### Version 1.3.1
##### Features && Bugfix
> 1.修改sql注入问题  

##### Version 1.3.0
##### Features && Bugfix
> 1.修改sql注入问题  
> 2.修复app审核问题  

##### Version 1.2.19
##### Features
> 1.热修复升级信息路由更改名字
> 2.热修复 upload 接口增加上传CDN逻辑

##### Version 1.2.18
##### Features
> 1.FF的diff修改解析逻辑  
> 2.ci同步cd增加原包上传CDN逻辑  
> 3.修改渠道包上传路径  

##### Version 1.2.17
##### Features
> 1.ci/hotfix 本地包存储路径增加 pack 文件夹
> 2.lastPack 回传增加详细包地址等信息

##### Version 1.2.16
##### Features
> 1.FF历史接口增加total和diffs 
> 2.渠道状态更新  

##### Version 1.2.15
##### Bugfix
> 1.cd的过滤部分参数改为非必传  
> 2.android端app过审，自动增加master渠道  
> 3.master渠道禁止任何删除修改操作  
> 4.ff增加中间文件  
> 5.增加app服务树路径  
> 5.增加增量获取服务权限接口  

##### Version 1.2.14
##### Bugfix
> 1.pack全量接口赋值错误  

##### Version 1.2.13
##### Features
> 1.FF文件改结构  
> 2.渠道包逻辑兼容状态修改  
> 3.pack全量接口增加path_url  

##### Version 1.2.12
##### Features
> 1.hotfix job status 轮询刷新

##### Version 1.2.11
##### Features
> 1.渠道增加id  
> 2.config生成total的bug  

##### Version 1.2.10
##### Features
> 1.修复 cd r_url

##### Version 1.2.9
##### Features
> 1.修复 cd r_url

##### Version 1.2.8
##### Features
> 1.渠道发布超时时间调整  
> 2.修改渠道列表  
> 3.修改新建config版本的备注  

##### Version 1.2.7
##### Bugfix
> 1.ci 包 URL 变更

##### Version 1.2.6
##### Bugfix
> 1.hotfix 传包 bug fix
> 2.ci 存储路径变更

##### Version 1.2.5
##### Bugfix
> 1.ci upload 接口新增 r_name 字段，r_url 入库
> 2.ci list 新增 did_push 字段，判断是否推送过
> 3.hotfix origin/get 接口从源包构建 id 查询改为从 hotfix 的 id 查询

##### Version 1.2.5
##### Features
> 1.修复热修复数据打包状态未变更问题

##### Version 1.2.4
##### Features
> 1.config和ff版本接口改为全量  

##### Version 1.2.3
##### Features
> 1.ff增加black_list逻辑  

##### Version 1.2.2
##### Features
> 1.新增获取热修包源包下载地址接口
> 2.新增获取全部热修数据接口
> 3.修改 hotfix build 接口逻辑

##### Version 1.2.1
##### Features
> 1.app审核通过才增加默认config和ff  
> 2.ff生成tree抽出公共方法  
> 3.config和ff上传cdn增加环境一级目录  

##### Version 1.2.0
##### Features
> 1.ff的rom从最大最小改为多选  
> 2.ff的version格式校验，min和max都为空不需要传  
> 3.增加获取服务树角色接口  
> 4.增加获取服务树列表接口  

##### Version 1.1.10
##### Bugfix
> 1.修复 hotfix list 接口报 -500
> 2.修复 静态渠道 添加失败问题

##### Version 1.1.9
##### Features
> 1.hotfix upload 初版

##### Version 1.1.8
##### Features
> 1.修复 ci 列表分页问题
> 2.去除 hotfix dao 中取 short_commit 的逻辑
> 3.取消、删除 hotfix job 的补完

##### Version 1.1.7
##### Features
> 1.修复包下载地址

##### Version 1.1.6
##### Features
> 1.ff的最大最小改为连接符样式  
> 2.增加渠道包列表  

##### Version 1.1.5
##### Features
> 1.修改bfs的bucket  
> 2.增加全量获取filter接口  

##### Version 1.1.4
##### Features
> 1.ci同步cd接口换buildID字段  
> 2.增加全量获取升级数据接口  

##### Version 1.1.3
##### Bugfix
> 1.ci 废弃使用 conf.Pkg 改用 conf.LocalPath
> 2.ci record, update 接口直接更改打包状态，减轻轮询任务压力

##### Version 1.1.2
##### Features
> 1.file接口增加解密逻辑  

##### Version 1.1.1
##### Features
> 1.config发布和ff发布增加流程日志  

##### Version 1.1.0
##### Features
> 1.完整合第一版  

##### Version 1.0.19
##### Business
> 1.增量包上传  

##### Version 1.0.18
##### Features
> 1.普通包上传接口 bug fix

##### Version 1.0.17
##### Features
> 1.ci record 接口新增 build id 返回值
> 2.协程触发 gitlab trigger 使用 context.TODO()

##### Version 1.0.16
##### Features
> 1.修改config目录  
> 2.优化各种文件路径配置  

##### Version 1.0.15
##### Features
> 1.gitlab trigger 新增 build id 参数

##### Version 1.0.14
##### Features
> 1.去掉ping方法  

##### Version 1.0.13
##### Features
> 1.CI同步数据到CD的接口  
> 2.渠道包  
> 3.渠道包上传  
> 4.渠道包发布  
> 5.创建渠道包  

##### Version 1.0.12
##### Features
> 1.轮询更新 gitlab job 状态

##### Version 1.0.11
##### Features
> 1.CI 的 Dao 新增 BuildPack，通过 app_key 和 build_id 返回单条 CI 数据，给 CD 复制信息用

##### Version 1.0.10
##### Features
> 1.修改ff版本字段格式  
> 2.config外层文件不做加密  

##### Version 1.0.9
##### Features
> 1.修改 gitlab project id 逻辑，不从前端传入，改为直接取 group_name/project_name

##### Version 1.0.8
##### Features
> 1.ff wl add 批量添加  
> 2.增加获取build接口  
> 3.应用配置系统版本格式修改  

##### Version 1.0.8
##### Features
> 1.添加针对hotfix的pack version接口  

##### Version 1.0.7
##### Features
> 1.Config模块 
> 2.FF模块 
> 3.热修复

##### Version 1.0.6
##### Features
> 1.CI 增加 gitlab pipeline trigger

##### Version 1.0.5
##### Features
> 1.RSA加密

##### Version 1.0.4
##### Features
> 1.CI模块

##### Version 1.0.3
##### Features
> 1.CD模块  

##### Version 1.0.2
##### Features
> 1.渠道相关模块

##### Version 1.0.1
##### Features
> 1.首页模块  

##### Version 1.0.0
##### Features
> 1.初始化项目  

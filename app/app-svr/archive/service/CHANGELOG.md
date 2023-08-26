#### archive rpc service
##### Version 6.53.78
> 1.删除冗余日志

##### Version 6.53.77
> 1.稿件投稿属地后台管控

##### Version 6.53.76
> 1.秒开支持pcdn-修改灰度逻辑

##### Version 6.53.75
> 1.秒开支持pcdn

##### Version 6.53.74
> 1.稿件清晰度1080p+

##### Version 6.53.72
> 1.稿件自见

##### Version 6.53.71
> 1.paladin v2迁移
> 新增稿件投稿属地

##### Version 6.53.70
> 1.付费合集web

##### Version 6.53.69
> 1.4g网络 story多返回一路1080p清晰度

##### Version 6.53.68
> 1.4g story返回1080p

##### Version 6.53.67
> 1.优化缓存失败时获取稿件信息
> 2.4g返回1080p(去除该逻辑)

##### Version 6.53.66
> 1.4g返回1080p

##### Version 6.53.65
> 1.付费合集

##### Version 6.53.64
> 1.音量均衡针对低音量不下发

##### Version 6.53.63
> 1.秒开接口粉ipad清晰度更改

##### Version 6.53.62
> 1.付费稿件proto

##### Version 6.53.61
> 1.竖版缩略图

##### Version 6.53.60
> 1.查询up主首映优化

##### Version 6.53.59
> 1.支持查询首映信息

##### Version 6.53.58
> 1.稿件禁止oversea_block项迁移,关闭回源db

##### Version 6.53.57
> 1.稿件禁止oversea_block项迁移

##### Version 6.53.56
> 1.稿件缓存异常处理逻辑优化

##### Version 6.53.55
> 1.cdn白名单逻辑优化

##### Version 6.53.54
> 1.cdn策略优化,根据用户使用的网络Wi-Fi或移动网络,为用户指定ip地址

##### Version 6.53.53
> 1.ArcPlayer接入音量均衡信息

##### Version 6.53.52
> 1.自动清晰度

##### Version 6.53.51
> 1.代码回滚

##### Version 6.53.48
> 1.自动清晰度调整

##### Version 6.53.48
> 1.psb_{aid}_{cid}缓存增加过期时间
> 2.redis查询miss情况下，查询taishan成功，写入redis
> 3.查询db成功，写入redis和taishan

##### Version 6.53.47
> 1.redirect_url hotfix

##### Version 6.53.46
> 1.ogv强制跳转

##### Version 6.53.45
> 1.psb_{aid}缓存增加过期时间
> 2.videoshot老key=vst_{cid}写入删除
> 3.desc老key=desc_{aid}写入删除

##### Version 6.53.44
> 1.支持高清缩略图

##### Version 6.53.43
> 1.psb_{aid}缓存redis err 回源 taishan
> 2.回源 taishan 后将数据回填 redis

##### Version 6.53.42
> 1.a3p_{aid}缓存增加过期时间

##### Version 6.53.42
> 1.inline 地区校验

##### Version 6.53.41
> 1.增加使用taishan回填redis的prom

##### Version 6.53.40
> 1.去除不必要日志
> 2.修复arc nil panic

##### Version 6.53.39
> 1.HD videoshot proto 合入
> 2.autoplay_area_validate proto 合入
 
##### Version 6.53.38
> 1.删除往缓存里写type_name的逻辑
> 2.redis err 灰度回源 taishan
> 3.回源 taishan 后将数据回填 redis
 
##### Version 6.53.37
> 1.cdn选择恢复按省份调度

##### Version 6.53.36
> 1.倍速需求

##### Version 6.53.35
> 1.去除go-main/archive依赖

##### Version 6.53.34
> 1.稿件首帧封面逻辑

##### Version 6.53.33
> 1.杜比视界

##### Version 6.53.32
> 1.简介@能力写taishan缓存优化

##### Version 6.53.31
> iPad回粉支持拜年纪主题色

##### Version 6.53.30
> cdn调度选择5.0，扩展为大区优选

##### Version 6.53.29
> 1.修复异味

##### Version 6.53.28
> 1.未登录用户清晰度实验

##### Version 6.53.27
> 1.端上调度4.0

##### Version 6.53.26
> 1.稿件@能力archive-service

##### Version 6.53.23
> 1.删除casttype为int的类型

##### Version 6.53.22
> 1.story清晰度提升1080试验

##### Version 6.53.21
> 1.秒开接口增加第三方cdn选择能力

##### Version 6.53.20
> 1.修复short_link字段和up_from字段

##### Version 6.53.19
> 1.新增投稿来源up_from同步

##### Version 6.53.18
> 1.ArcsPlayer接口重构

##### Version 6.53.17
> 1.处理异味，部分异味需要修改逻辑代码，为方便测试，等后期独立提mr修改

##### Version 6.53.16
> 1.修改有历史记录情况下秒开信息不全

##### Version 6.53.15
> 1.非首p秒开新服务-ArcsPlayer，增加灾备开关切回老逻辑

##### Version 6.53.14
> 1.非首p秒开新服务-ArcsPlayer

##### Version 6.53.13
> 1.ArcWithStat err 复用导致 panic

##### Version 6.53.12
> 1.lint bugs修复

##### Version 6.53.11
> 1.rows改为defer写法

##### Version 6.53.10
> 1.grpc quota limit

##### Version 6.53.9
> 1.集成短链

##### Version 6.53.8
> 1.调account mids 去重

##### Version 6.53.6
> 1.秒开清晰度实验

##### Version 6.53.5
> 1.BatchPlayArg参数位置修复

##### Version 6.53.4
> 1.修复ios forcehost传参错误

##### Version 6.53.3
> 1.上提compareBatchPlayArg的位置
> 2.当BatchPlayArg参数与老参数完全一致，或业务方全面使用新参数时，对老参数重新赋值（代码迁移完成后可删除）

##### Version 6.53.2
> 1.拜年纪大型活动页需求

##### Version 6.53.1
> 1.from非必传参数

##### Version 6.53.0
> 1.新增构造秒开参数的middleware
> 2.秒开方法支持BatchPlayArg

##### Version 6.52.16
> 1.720p清晰度合并（以后对外只有qn=64）

##### Version 6.52.15
> 1.attribute、attribute_v2、access之后不可对外展示

##### Version 6.52.14
> 1.Autoplay = 0 When OnlyFavView

##### Version 6.52.13
> 1.add attribute for OnlyFavView

##### Version 6.52.12
> 1.批量接口在的db报错时返回缓存内的数据
> 2.泰山已全量，删除对应配置

##### Version 6.52.11
> 1.竖屏视频实验秒开720p

##### Version 6.52.10
> 1.redis正常时从内存中获取二级分区名称

##### Version 6.52.9
> 1.历史进度seek

##### Version 6.52.8
> 1.新版清晰度描述

##### Version 6.52.7
> 1.update go-main

##### Version 6.52.6
> 1.fix simple arc cache 

##### Version 6.52.5
> 1.所有缓存的key使用统一model

##### Version 6.52.4
> 1.修复缓存不存在时拼接的空数据

##### Version 6.52.3
> 1.根据Sonar分析，删除废弃代码

##### Version 6.52.2
> 1.秒开支持hdr清晰度

##### Version 6.52.1
> 1.批量方法validate增加min=1

##### Version 6.52.0
> 1.稿件信息获取的方法均支持缓存非ErrNil时回源taishan

##### Version 6.51.9
> 1.修改缩略图方法,不再依赖账号

##### Version 6.51.8
> 1.由于aid已无序下发删除grpc接口MaxAid
> 2.ArcsWithPlayurl接口增加日志检查aids数量超50个

##### Version 6.51.7
> 1.提供simpleArc接口，提供简版稿件信息给业务方查询

##### Version 6.51.6
> 1.upcount在zcard=0时再读db，防止冷门up缓存过期后无法更新数据

##### Version 6.51.5
> 1.ipadHD秒开支持4k
> 2.ArcsWithPlayurl GRPC接口支持传递show_pgc_url（兼容动态需要pgc秒开）

##### Version 6.51.4
> 1.秒开地址新增backup_url返回
> 2.秒开非二压清晰度如果有h265和h264同时返回

##### Version 6.51.3
> 1.proto文件删除不再使用的方法
> 2.账号昵称与头像不再作为稿件缓存的一部分
> 3.删除service、dao、model中不再使用的代码

##### Version 6.51.2
> 1.优化up主稿件列表过滤 & 删除无用方法

##### Version 6.51.1
> 1.秒开优化清晰度选择（只返回用户清晰度+秒开清晰度）

##### Version 6.51.0
> 1.archive-service的stat新增follow字段，方便stat-job回源

##### Version 6.50.9
> 1.单稿件查询接口通过灰度验证已可以全量，开为单独开关控制
> 2.增加秒开接口直接回源账号的灰度开关控制

##### Version 6.50.8
> 1.全面去除B端数据库的依赖

##### Version 6.50.7
> 1.规范pb数据类型定义(int -> int64)

##### Version 6.50.6
> 1.修复story卡片代码误修改稿件代码

##### Version 6.50.5
> 1.db 类 function 改名为 Row 开头,方便区分
> 2.增加 Arcs 方法并通过配置文件控制灰度到 Arcs 的流量占比
> 3.独立出新老秒开接口调用的Dao方法，保证其不受影响，待批量接口验证通过后再逐步迁移扩量
> 4.删除废弃的 Recommend 方法与删除不再使用的 Click 数据库配置与代码

##### Version 6.50.4
> 1.将UpperCount的计算先依赖缓存，待迁移到JOB后再优化

##### Version 6.50.3
> 1.定义ArcWithStat明确表示返回稿件与计数信息

##### Version 6.50.2
> 1.支持story秒开

##### Version 6.50.1
> 1.修复Stat3的ctx使用

##### Version 6.50.0
> 1.使用配置文件，控制Arc和View方法直接回源账号系统的比例
> 2.调整dao层目录

##### Version 6.49.42
> 1.删除从B端数据库读取videoshot库的代码
> 2.删除废弃配置

##### Version 6.49.41
> 1.更新go-common

##### Version 6.49.40
> 1.同步tag

##### Version 6.49.39
> 1.修改videoshot读取的DB表与缓存key

##### Version 6.49.38
> 1.秒开新增免流参数

##### Version 6.49.37
> 1.增加秒开请求backup_url返回数量

##### Version 6.49.36
> 1.新增获取稿件创作人接口

##### Version 6.49.35
> 1.4k增加版本控制

##### Version 6.49.34
> 1.删除2020年拜年祭相关配置

##### Version 6.49.33
> 1.秒开投屏video_project字段支持flv灰度

##### Version 6.49.32
> 1.删除 gorpc model 的依赖

##### Version 6.49.32
> 1.删除不再使用的redis和memcache代码

##### Version 6.49.31
> 1.非公开投稿不进入up投稿列表

##### Version 6.49.30
> 1.proto增加参数校验

##### Version 6.49.29
> 1.风控-秒开逻辑修改 

##### Version 6.49.28
> 1.联合投稿商业样式

##### Version 6.49.27
> 1.安卓不出2020拜年祭秒开

##### Version 6.49.26
> 1.安卓不出2020拜年祭秒开

##### Version 6.49.25
> 1.playurl-batch切grpc

##### Version 6.49.24
> 1.增加nil判断防止panic

##### Version 6.49.23
> 1.去掉stat的MC读的代码

##### Version 6.49.22
> 1.stat批量缓存灰度切redis统计修改

##### Version 6.49.21
> 1.stat批量缓存灰度切redis

##### Version 6.49.20
> 1.stat缓存灰度切redis回源支持

##### Version 6.49.19
> 1.stat缓存灰度切redis

##### Version 6.49.18
> 1.playurl batch切换走discovery

##### Version 6.49.17
> 1.增加attribute禁止后台播放

##### Version 6.49.16
> 1.增加attribute_v2字段供新业务使用

##### Version 6.49.15
> 1.删除ugc-season相关函数

##### Version 6.49.14
> 1.删除分区相关函数

##### Version 6.49.13
> 1.删除GoRPC中的Click3,Recommend3,RankArcs3,RanksArcs3,RankTopArcs3,RankAllArcs3,Video3方法

##### Version 6.49.12
> 1.调整prom的监控
> 1.删除videoshot灰度的代码

##### Version 6.49.11
> 1.删除废弃的addVideoshot代码

##### Version 6.49.10
> 1.已全量redis版本views接口，删除老代码

##### Version 6.49.9
> 1.删除addVideoshot的代码（已废弃）
> 2.删除http的videoshot接口与api文档
> 3.删除http的region分区接口

##### Version 6.49.8
> 1.删除dao/share目录的冗余代码

##### Version 6.49.7
> 1.删除DB的Prepared
> 2.删除灰度View接口的逻辑
> 3.删除desc,archive,page读写mc的逻辑
> 4.删除废弃代码
> 4.增加灰度Views接口的逻辑

##### Version 6.49.6
> 1.秒开dash格式增加size

##### Version 6.49.5
> 1.灰度page方法

##### Version 6.49.4
> 1.灰度arcCaches方法

##### Version 6.49.3
> 1.增加page接口双写

##### Version 6.49.2
> 1.构建go-common发布

##### Version 6.49.1
> 1.通过推送配置，灰度redis的Arc方法

##### Version 6.49.0
> 1.Arc和Desc方法同步增加写入redis的逻辑,为后续读取做准备

##### Version 6.48.23
> 1.删除season stat相关代码

##### Version 6.48.22
> 1.Arcs,Stats批量参数限制100个

##### Version 6.48.21
> 1.删除废弃的addShare方法

##### Version 6.48.19
> 1.新增付费pugv标志位

##### Version 6.48.18
> 1.删除videoshot灰度bfs域名的逻辑

##### Version 6.48.17
> 1.views批量限制50个

##### Version 6.48.16
> 1.剩余gorpc新增grpc方法

##### Version 6.48.15
> 1.大于24小时视频不出秒开

##### Version 6.48.14
> 1.去掉ping

##### Version 6.48.13
> 1.动态inline播放卡片增加PGC seasionID

##### Version 6.48.12
> 1.改造view和views接口，对于互动视频只返回特定的引导视频
> 2.互动视频不可inline播放和秒开
> 3.新增内部接口SteinsGateView供查询互动视频的全部分P信息

##### Version 6.48.11
> 1.修改OWNERS文件

##### Version 6.48.10
> 1.ugc剧集

##### Version 6.48.9
> 1.加grpc批量数量限制aid>0限制
> 2.数量限制先等等，重新打日志看一下

##### Version 6.48.8
> 1.加grpc批量接口日志

##### Version 6.48.7
> 1.支持4k清晰度

##### Version 6.48.6
> 1.秒开黑屏实验去掉写死480p

##### Version 6.48.5
> 1.联合投稿排序走index_order

##### Version 6.48.4
> 1.动态秒开返回pgc重试参数

##### Version 6.48.3
> 1.缩略图走i0域名
> 2.灰度i0域名策略

##### Version 6.48.2
> 1.fix intl build

##### Version 6.48.1
> 1.fix ios_b

##### Version 6.48.0
> 1.重构http的playurl接口
> 2.新增grpc的playurl接口

##### Version 6.47.13
> 1.修改ugc预览备注

##### Version 6.47.12
> 1.增加稿件ugc预览标识

##### Version 6.47.11
> 1.批量秒开返回qn配置化

##### Version 6.47.10
> 1.批量秒开返回是否非二压

##### Version 6.47.9
> 1.联合投稿按照审核库staff生成id排序

##### Version 6.47.8
> 1.up主也支持大会员清晰度

##### Version 6.47.7
> 1.批量秒开接口支持大会员鉴权

##### Version 6.47.6
> 1.iPhone 5.36版本不吐拜年祭单品稿件的秒开地址

##### Version 6.47.5
> 1.拜年祭单品视频不吐秒开地址

##### Version 6.47.4
> 1.view接口只对archive做强判断

##### Version 6.47.3
> 1.fix view

##### Version 6.47.2
> 1.迁移gorpc方法到gpc

##### Version 6.47.1
> 1.view接口增加staff信息

##### Version 6.47.0
> 1.grpc增加注释

##### Version 6.46.6
> 1.更新稿件缓存增加联合投稿部分

##### Version 6.46.5
> 1.Dislike强制0

##### Version 6.46.4
> 1.增加高能看点、bgm、联合投稿attribute

##### Version 6.46.3
> 1.删除RPC中的addShare方法

##### Version 6.46.2
> 1.分享不发databus消息

##### Version 6.46.1
> 1.fix package

##### Version 6.46.0
> 1.issues #403 大仓库项目目录结构改进

##### Version 6.45.2
> 1.加参数控制pgc吐playurl

##### Version 6.45.1
> 1.pgc不吐playurl

##### Version 6.45.0
> 1.拦截archive miss不存在的稿件

##### Version 6.44.4
> 1.接入account grpc

##### Version 6.44.3
> 1.秒开拦截辣鸡参数

##### Version 6.44.2
> 1.秒开qn白名单+大会员清晰度降级

##### Version 6.44.1
> 1.计数默认返回值修改

##### Version 6.44.0
> 1.增加UGCPay标识

##### Version 6.43.9
> 1.dash格式加codecid

##### Version 6.43.8
> 1.修复冷门up主投稿列表稿件不全的bug

##### Version 6.43.7
> 1.初始化缓存日志

##### Version 6.43.6
> 1.初始化分区缓存的时候使用context.Background

##### Version 6.43.5
> 1.修改重置up信息的缓存逻辑

##### Version 6.43.4
> 1.增加日志观察job异步databus消息是否发送成功

##### Version 6.43.3
> 1.增加日志观察账号RPC服务的返回是否正常

##### Version 6.43.2
> 1.增加日志观察账号昵称头像是否为空

##### Version 6.43.1
> 1.秒开接口空字段不吐

##### Version 6.43.0
> 1.秒开接口增加dash字段

##### Version 6.42.2
> 1.同步IsNormal,AttrVal方法

##### Version 6.42.1
> 1.接入grpc

##### Version 6.42.0
> 1.添加dao层ut

##### Version 6.42.0
> 1.调整目录

##### Version 6.41.1
> 1.fix cache ctx

##### Version 6.41.0
> 1.issue 249 metadata ip

##### Version 6.40.10
> 1.增加视频云的fnver，fnval字段返回

##### Version 6.40.9
> 1.更新bvc pb文件

##### Version 6.40.8
> 1.优化重新生成账号缓存的逻辑

##### Version 6.40.7
> 1.账号接口请求失败时，走databus慢慢更新

##### Version 6.40.6
> 1.透传视频云fnval,fnver字段

##### Version 6.40.5
> 1.账号老是不知道刷什么东西

##### Version 6.40.4
> 1.增加是否可以投屏

##### Version 6.40.3
> 1.透传投屏信息

##### Version 6.40.2
> 1.支持地区限制

##### Version 6.40.1
> 1.稿件描述走缓存

##### Version 6.40.0
> 1.稿件更新缓存bugfix
> 2.重置retag

##### Version 6.39.3
> 1.增加autoplay字段

##### Version 6.39.2
> 1.HTTP接口增加稿件分辨率字段

##### Version 6.39.1
> 1.PB接口体增加json字段输出

##### Version 6.39.0
> 1.批量MC接口优化

##### Version 6.38.1
> 1.分辨率0,0,0不做处理

##### Version 6.38.0
> 1.稿件增加分辨率字段

##### Version 6.37.19
> 1.批量稿件接口代码优化

##### Version 6.37.18
> 1.conn close fix

##### Version 6.37.17
> 1.增加缓存容错

##### Version 6.37.16
> 1.bvc灰度

##### Version 6.37.15
> 1.redis set expire -> setex

##### Version 6.37.14
> 1.无投稿的用户只缓存10分钟

##### Version 6.37.13
> 1.分享行为的databus key从aid改为mid

##### Version 6.37.12
> 1.fix row close

##### Version 6.37.11
> 1.清理share代码

##### Version 6.37.10
> 1.删除videoshot add接口

##### Version 6.37.9
> 1.RPC不需要token了  By 郝冠伟确认

##### Version 6.37.8
> 1.增加register

##### Version 6.37.7
> 1.删除冗余代码

##### Version 6.37.6
> 1.使用bm

##### Version 6.37.5
> 1.删除多余配置

##### Version 6.37.4
> 1.增加批量获取up投稿数量的http接口

##### Version 6.37.3
> 1.删除limit模块

##### Version 6.37.2
> 1.第一次分享改发databus

##### Version 6.37.1
> 1.使用account-service v7

##### Version 6.37.0
> 1.迁移到主站目录下

##### Version 6.36.16
> 1.取消强制开关

##### Version 6.36.15
> 1.提供给B+的秒开接口强行不返回playurl

##### Version 6.36.14
> 1.参数长度调整为200

##### Version 6.36.13
> 1.补充UnitTest

##### Version 6.36.12
> 1.配置文件增加开关选项，控制是否请求视频云获取播放信息

##### Version 6.36.11
> 1.firstCid只用vupload，外部源不缓存

##### Version 6.36.10
> 1.优化秒开代码

##### Version 6.36.9
> 1.Archive3结构体增加第一P的cid，供后续业务扩展使用

##### Version 6.36.8
> 1.attr增加地区限制

##### Version 6.36.7
> 1.share数双写新databus

##### Version 6.36.6
> 1.接bvc的pb接口

##### Version 6.36.5
> 1.Convey test

##### Version 6.36.4
> 1.BFS改回来

##### Version 6.36.3
> 1.BFS的封面图强制返回https

##### Version 6.36.2
> 1.删除attrbithideclick相关代码

##### Version 6.36.1
> 1.删除like的相关代码与配置

##### Version 6.36.0
> 1.删除like的相关代码与配置

##### Version 6.35.1
> 1.修改透传的player字段名

##### Version 6.35.0
> 1.attr的第十二位改成 IsPorder 私单标记

##### Version 6.34.0
> 1.增加player接口

##### Version 6.33.1
> 1.批量接口增加参数日志

##### Version 6.33.0
> 1.增加maxAID的接口

##### Version 6.32.9
> 1.删除废弃的代码

##### Version 6.32.8
> 1.删除废弃的RPC server端代码

##### Version 6.32.7
> 1.删除冗余代码

##### Version 6.32.6
> 1.增加prom db

##### Version 6.32.5
> 1.内置prom

##### Version 6.32.4
> 1.使用内置prom

##### Version 6.32.3
> 1.修复缓存miss时少吐数据的问题

##### Version 6.32.1
> 1.Video3 RPC

##### Version 6.32.0
> 1.兼容客户端传多次点赞

##### Version 6.31.3
> 1.统一修改errgroup包路径

##### Version 6.31.2
> 1.attr的第九位改成 isPGC

##### Version 6.31.1
> 1.修改Views3返回值

##### Version 6.31.0
> 1.删除非internal的对外http接口

##### Version 6.30.0
> 1.Archive3结构体改为非指针

##### Version 6.29.1
> 1.补全RPC PB接口，video3

##### Version 6.29.0
> 1.补全RPC PB接口

##### Version 6.28.1
> 1.archive增加dynamic字段

##### Version 6.28.0
> 1.增加up主推荐视频的RPC接口

##### Version 6.27.0
> 1.删除pgc相关逻辑

##### Version 6.26.0
> 1.delete Movie2 AidByCid

##### Version 6.25.0
> 1.add Page3 pb rpc

##### Version 6.24.2
> 1.upArcs & upsArcs pb

##### Version 6.24.2
> 1.rpc删除likes2接口

##### Version 6.24.1
> 1.rpc增加likes3的pb接口

##### Version 6.24.0
> 1.rpc增加stat，stats的pb接口

##### Version 6.23.0
> 1.rpc 增加archive3,archives3的pb接口
> 2.rpc 删除废弃的videos2,videosByCids2,CidByEpIDs2等方法
> 3.pgc接口只吐电影信息

##### Version 6.22.8
> 1.rpc 增加view3的pb接口

##### Version 6.22.7
> 1.http video接口走PB

##### Version 6.22.6
> 1.http archive、archives、page全量开放，异步更新page缓存

##### Version 6.22.5
> 1.http/view接口全量pb缓存预热

##### Version 6.22.4
> 1.流量扩大得到aid%10<5走PB
> 2.分P的http接口也走pb

##### Version 6.22.3
> 1.aid%10<3走pb

##### Version 6.22.2
> 1.http stat/stats接口全量走pb

##### Version 6.22.1
> 1.pb的func/model/struct/service等全面改名为数字3结尾

##### Version 6.22.0
> 1.pb bugfix

##### Version 6.21.1
> 1.archive http 接口 aid%10=1的走pb

##### Version 6.21.0
> 1.增加limiter限流

##### Version 6.20.0
> 1.增加批量views接口，aids限制为20个

##### Version 6.19.0
> 1.直播限制50个

##### Version 6.18.0
> 1.cids接口不直接return

##### Version 6.17.0
> 1.videoshot接口试水pkg/errors

##### Version 6.16.0
> 1.增加全区7天内最新稿件

##### Version 6.15.0
> 1.redis errnil return

##### Version 6.14.0
> 1.likes相关数据落库
> 2.增加likes列表的RPC接口

##### Version 6.13.0
> 1.bilibili_archive库全都读写分离

##### Version 6.12.0
> 1.upspass score bugfix

##### Version 6.11.0
> 1.upsPass接口增加copyright

##### Version 6.10.0
> 1.添加获取单P信息的http接口(包含description字段)
> 2.修改原获取单P信息的service层逻辑
> 2.添加获取长简介的http和rpc接口

##### Version 6.9.0
> 1.升级go-common

##### Version 6.8.0
> 1.升级go-common
> 2.迁移model到项目中

##### Version 6.7.0
> 1.增加主站排行榜专用接口

##### Version 6.6.1
> 1.memcache json

##### Version 6.6.0
> 1.memcache gob

##### Version 6.5.0
> 1.增加点赞相关RPC接口

##### Version 6.4.0
> 1.升级go-common&go-business
> 2.videoshot rpc 增加aid参数

##### Version 6.3.0
> 1.videosho接口增加aid参数

##### Version 6.2.4
> 1.http context fix

##### Version 6.2.3
> 1.升级go-business
> 2.manager后台变更稿件归属mid时,变更相应缓存

##### Version 6.2.2
> 1.增加upspassed rpc方法

##### Version 6.2.1
> 1.rpc video2 nil fix

##### Version 6.2.0
> 1.所有稿件&视频走新archive_result数据库
> 2.升级go-common&go-business

##### Version 6.1.21
> 1.删除SetStatCache2接口

##### Version 6.1.20
> 1.修复可能导致panic的问题

##### Version 6.1.19
> 1.增加http的typelist接口

##### Version 6.1.18
> 1.修改ci配置

##### Version 6.1.17
> 1.增加account清楚缓存时的参数

##### Version 6.1.16
> 1.增加昵称&头像更新后的缓存清理逻辑

##### Version 6.1.15
> 1.增加无脑生成view&click缓存

##### Version 6.1.14
> 1.ci配置分支

##### Version 6.1.13
> 1.删掉zlimit相关残留代码
> 2.archive和archives接口返回archive_report_result中is_show等于1的result

##### Version 6.1.12
> 1.去掉老的dede

##### Version 6.1.11
> 1.分区表走新分区
> 2.增加RPC获取所有type的方法
> 3.升级go-common和go-business

##### Version 6.1.10
> 1.升级go-common和go-business
> 2.修改prom写法

##### Version 6.1.9
> 1.修复闭包缓存

##### Version 6.1.8
> 1.重发ci

##### Version 6.1.7
> 1.修复分类缓存

##### Version 6.1.6
> 1.修复videoshot nil 导致panic

##### Version 6.1.5
> 1.修复videoshot nil 导致panic

##### Version 6.1.4
> 1.page字段走自增形式

##### Version 6.1.3
> 1.增加auth

##### Version 6.1.2
> 1.增加prom

##### Version 6.1.1
> 1.批量大小改为60

##### Version 6.1.0
> 1.计数闭包问题修复

##### Version 6.0.16
> 1.rows close bug fix

##### Version 6.0.13
> 1.mc改成永不过期

##### Version 6.0.12
> 1.计数加aid在json

##### Version 6.0.11
> 1.增加cache出错不回写
> 2.增加prom回源统计

##### Version 6.0.10
> 1.修复点击计数panic

##### Version 6.0.9
> 1.修复prom参数个数

##### Version 6.0.8
> 1.增加prom包

##### Version 6.0.7
> 1.增加memcache随机过期时间

##### Version 6.0.6
> 1.增加upcount缓存逻辑

##### Version 6.0.5
> 1.修复chan未设置长度的bug

##### Version 6.0.4
> 1.修复view2和views2

##### Version 6.0.3
> 1.修复DB prepare配置

##### Version 6.0.2
> 1.批量没有默认返回空map

##### Version 6.0.1
> 1.去掉重复的view接口

##### Version 6.0.0
> 1.重构-删除无用方法(dede等)
> 2.重构-优化批量查询
> 3.重构-优化计数信息缓存
> 4.增加批量aids获取View信息
> 5.增加单aid获取view的http接口
> 6.增加SetStat rpc方法(mc)

##### Version 5.6.14
> 1.增加rpc接口,全量更新stat数值(redis)

##### Version 5.6.13
> 1.增加internal/view

##### Version 5.6.12
> 1.cache回写逻辑

##### Version 5.6.11
> 1.去除404的header

##### Version 5.6.10
> 1.click走mysql

##### Version 5.6.9
> 1.批量计数查不到不设置空值

##### Version 5.6.8
> 1.修改identity为verify

##### Version 5.6.7
> 1.修复stat panic

##### Version 5.6.6
> 1.修复rows.next()

##### Version 5.6.5
> 1.接入新配置中心
> 2.rpc接口参数校验
> 3.去除hbase

##### Version 5.6.4
> 1.升级go-common

##### Version 5.6.3
> 1.rpc接口支持缓存的修改

##### Version 5.6.2
> 1.日志错误修复

##### Version 5.6.1
> 1.archive/page走新表,修改sql

##### Version 5.6.0
> 1.archive/page走新表

##### Version 5.5.1
> 1.videos接口增加Ptitle

##### Version 5.5.0
> 1.RPC增加根据aids获取stat接口

##### Version 5.4.0
> 1.paas发布占用

##### Version 5.3.6
> 1.RPC增加一级分区最新视频与数量接口
> 2.RPC增加Upcount方法 获取用户投稿总数
> 3.内部http接口改名
> 4.升级go-common

##### Version 5.3.5
> 1.PGC只查status=开放的

##### Version 5.3.4
> 1.ArcsNoCheck2接口校验,aid为空则直接返回参数错误

##### Version 5.3.3
> 1.monitor挪到内部接口

##### Version 5.3.2
> 1.修复批量用户动态panic的bug
> 2.增加field数量

##### Version 5.3.1
> 1.统一monitor ping接口
> 2.修复批量用户动态panic的bug
> 3.增加field数量
> 4.分页接口增加兼容性处理

##### Version 5.3.0
> 1.修复redis cache

##### Version 5.2.8
> 1.增加根据aids获取seasonid接口 rpc
> 2.更改up过审稿件sql的排序字段

##### Version 5.2.7
> 1.up主过审稿件改为pubtime排序

##### Version 5.2.6
> 1.增加RPC方法,根据mids获取最新投稿
> 2.支持attribute参数,在列表中去除展示
> 3.升级go-common

##### Version 5.2.5
> 1.增加RPC方法,根据aids获取archive聚合信息

##### Version 5.2.4
> 1.增加RPC方法根据EpID获取cid
> 2.增加RPC方法根据CID获取video信息

##### Version 5.2.3
> 1.注释PGCproc方法

##### Version 5.2.2
> 1.增加RPC的分区信息接口

##### Version 5.2.1
> 1.注释pgc方法

##### Version 5.2.0

> 1.升级go-common新版本
> 2.fix view接口，多次查单个请求改为批量请求
> 3.conf支持优先从本地加载配置

##### Version 5.1.3

> 1.archives/nocheck接口新增返回返回archive_video和archive_video_audit表数据
> 2.去掉moment逻辑

##### Version 5.1.2

> 1.router加入rpcCloser

##### Version 5.1.1

> 1.忽略video计数错误

##### Version 5.1.0

> 1.升级配置中心
> 2.使用公用identify
> 3.使用统一参数开关

##### Version 5.0.0

> 1.net/rpc升级为golang/rpcx

##### Version 4.3.0

> 1.新增rpc获取稿件点击数量
> 2.新增rpc通过cid查aid
> 3.更新go-business

##### Version 4.2.5

> 1.分享计数增加databus双写

##### Version 4.2.4

> 1.新增videoshot接口供管理后台访问

##### Version 4.2.3

> 1.videoshot接口增加稿件状态校验

##### Version 4.2.2

> 1.依赖包升级

##### Version 4.2.1

> 1.修复db使用错误

##### Version 4.2.0

> 1.添加获取视频详情rpc接口

##### Version 4.1.4

> 1.fix len(attens) == 0 不能被除

##### Version 4.1.3

> 1.更新所有匿名rpc client为默认user

##### Version 4.1.2

> 1.修改syslog日志和上报

##### Version 4.1.1

> 1.更新go-business为1.3.1

##### Version 4.1.0

> 1.支持查询pgc信息
> 2.支持查询用户关注的up主的过审稿件

##### Version 4.0.0

> 1.go vendor支持
> 2.go-common/business换成go-business包
> 3.获取本机ip注册到zk
> 4.memcache批量获取支持多连接并发
> 5.新增rpc日志

##### Version 3.6.1

> 1.修复第一次分享的topic

##### Version 3.6.0

> 1.新增稿件page信息接口

##### Version 3.5.1

> 1.修复批量获取cache出错还加入cache问题

##### Version 3.5.0

> 1.获取稿件列表不检测权限
> 2.修复稿件分区变更后二级分区最新视频转移分区

##### Version 3.4.3

> 1.修复二级分区最新视频安装pubdate排序

##### Version 3.4.2

> 1.新增update稿件cache

##### Version 3.4.1

> 1.数组越界bug

##### Version 3.4.0

> 1.增加获取up主投稿列表接口
> 2.修复增加全量分区视频时变量没有重新初始化bug
> 3.优化缓存key使均匀分布

##### Version 3.3.1

> 1.修复最新视频bug：新增视频可见过滤条件：access、attrBitNoWeb、attrBitNoMobile

##### Version 3.3.0

> 1.增加分区的视频按投稿时间排序
> 2.新增查询过审记录接口

##### Version 3.2.3

> 1.修复回复的稿件置首bug

##### Version 3.2.2

> 1.修复Archive接口cache bug

##### Version 3.2.1

> 1.修复elk日志

##### Version 3.2.0

> 1.稿件添加字段reject_reason
> 2.修改share接口
> 3.新增set_tag接口
> 4.支持trace v2

##### Version 3.1.0

> 1.新增获取stat接口
> 2.新增获取多条stat接口
> 3.新增stat更新redis接口
> 4.修改稿件获取stat的方法
> 5.增加或修改ping方法
> 6.优化部分代码

##### Version 3.0.0

> 1.context使用官方接口
> 2.添加share计数
> 3.优化部分代码

##### Version 2.5.0

> 1.新增视频缩略图版本号
> 2.支持视频缩略图更新cid
> 3.添加up主视频动态接口

##### Version 2.4.0

> 1.添加获取用户最新评论稿件以及后台job
> 2.优化配置
> 3.添加服务发现


##### Version 2.3.0

> 1.添加获取videoshot接口
> 2.rpc调用bug

##### Version 2.2.0

> 1.优化
> 2.add elk
> 3.add trace id
> 4.add haiwai api
> 5.remove noused code
> 6.add mid recommend
> 7.fix some bug

##### Version 2.1.0

> 1.add tracer

##### Version 1.1.0

> 1.基于go-common重构

##### Version 1.0.0

> 1.初始化完成稿件基础查询功能

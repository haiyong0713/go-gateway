##  内容运营服务
#### 版本说明
1. 每月一个0.1版本，直接按月份，2.1开始为2021年1月
2. 每次迭代一个0.0.1版本
3. hot fix一个0.0.0.1版本
### Version 2.12.21
>1.siderbar白名单逻辑增加ip

### Version 2.12.20
>1.多状态入口增加港澳台

### Version 2.12.19
>1.限免稿件接口

### Version 2.12.18
>1.【港澳台】垂类tab强插实验

### Version 2.12.18
> 启动优化

### Version 2.12.17
> 添加监控

### Version 2.12.16
> overlord切proxyless

### Version 2.12.15
> remove required

### Version 2.12.14
> 游戏中心屏蔽修复

### Version 2.12.13
> 游戏中心屏蔽新增动态

### Version 2.12.12
> paladin配置迁移

### Version 2.12.11
> entry panic fix

### Version 2.12.10
> 进度条装扮

### Version 2.12.8
> isUploader接口优化,降低因主动cancel而增加的错误日志

### Version 2.12.8
> 特殊小卡，pgc-ep跳转支持url参数，网关对接新增url字段

### Version 2.12.7
> 部分grpc接4040优化

### Version 2.12.6.1
> 修复2.12.5.1引入的bug，ugctab缓存avidMap的逻辑，需要考虑手工填写稿件id的情况

### Version 2.12.6
> 日志优化

### Version 2.12.5.1
> 优化ugctab缓存avidMap的逻辑，避免潜在的OOM风险

### Version 2.12.5
> 进度条ICON不下发至旧版本

### Version 2.12.4.1
> fix.bwlist接口,支持灰度实验分组白名单 fix

### Version 2.12.4
> 1.bwlist名单接口,支持灰度实验分组白名单

### Version 2.12.3
> 1.运营装扮主动装扮增加人群限制

### Version 2.12.2
> 1. 天马:根据特殊卡ID获取特殊卡信息

### Version 2.12.1
> 1.装扮

### Version 2.12.0
> 1.图标个性化

### Version 2.11.5
> 1. 主题区分人群限制

### Version 2.11.4
> 1. 更新版头菜单

### Version 2.11.3
> 1.游戏中心模糊匹配

### Version 2.11.2
> 1.MineSection支持下发多个运营位

### Version 2.11.1
> 1.内容投放:数据埋点(web端)，增加operater字段，标记消息来源

### Version 2.10.4
> 1.内容投放 PC首页改版,首焦支持inline配置

### Version 2.10.3
> 1.解决proto enum重复注册问题

### Version 2.10.2
> 1.游戏中心入口支持排除渠道号

### Version 2.10.1
> 1.会员购特殊底tab临时配置过滤

### Version 2.9.3
>1.ogv banner投放形式新增goto = static_banner

### Version 2.9.2
> 1.新增港澳台垂类tab

### Version 2.9.1
> 1.app相关推荐重构-PGC关联,为OGV侧提供接口(ogv播放页出ugc)

### Version 2.8.3.1
> 1.fix 天马多物料,物料查询接口 增加width、height

### Version 2.8.3
> 1.内容投放banner,增加封面图主色调(web端新版首页首焦吸色)

### Version 2.8.2
> 1.模块配置: param和name支持接口控制的自定义下发

### Version 2.8.1
> 1.多态入口: 增加剩余动画次数下发

### Version 2.7.7
> 1. fix版头获取生效配置并发读问题

### Version 2.7.6
> 1.天马多物料，增加物料查询接口

### Version 2.7.5.1
> 1. fix计算取兜底配置逻辑

### Version 2.7.5
> 1. fix兜底配置不算在线上配置中

### Version 2.7.4
> 1. fix版头配置无ip策略情况

### Version 2.7.3
> 1. Feature: 模块配置-我的页统一判断是否为up主和主播

### Version 2.7.2.2
> 1. fix版头并发写map问题

### Version 2.7.2.1
> 1. fix获取分区资源位方法

### Version 2.7.2
> 1. 新版头配置拉取与网关沟通接口

### Version 2.7.1.1
> 1.app特殊卡rpc接口优化，只返回app相关推荐再投特殊卡（会预加载）

### Version 2.7.1
> 1.增加resource-service的section接口上增加调用bwlist的灰度逻辑
> 2.bwlist名单接口增加灰度逻辑

### Version 2.6.3.2
> 1.增加resource-service的home section日志
> 2.hot fix: ip判断出错导致的错误显示问题

### Version 2.6.3.2
> 1. hot fix: resource-service获取sidebar基于策略组的漏洞修复

### Version 2.6.3.1
> 1.增加resource-service的home section日志

### Version 2.6.3
> 1.新增app特殊卡rpc接口（为网关提供）

#### Version 2.6.2.2
> 1.hot fix: 通用黑白名单 - 资源同时读写问题
> 2.hot fix: 通用黑白名单 - 错误返回修正

#### Version 2.6.2.1
> 1.hot fix: 通用黑白名单 - 减轻热key问题

#### Version 2.6.2
> 1.ugc tab页新增bvid匹配文件上传

#### Version 2.6.1
> 1.通用黑白名单 - 增加兜底展现配制

#### Version 2.5.4
> 1.banner新增ad inline

#### Version 2.5.3
> 1.通用黑白名单

#### Version 2.5.2.1
> 1.Fix：HomeSections类型匹配修正

#### Version 2.5.2
> 1.Feature: HomeSections中MngIcon增加开始和技术时间

#### Version 2.5.1.4
> 1.hot fix：HomeSections单独去除首页顶tab icon的红点请求，因为游戏中心的数据返回和通常的返回不一致，由网关处理

#### Version 2.5.1.3
> 1.hot fix：HomeSections修复游戏中心根据channel屏蔽的问题

#### Version 2.5.1.2
> 1.hot fix：MineSections修复直播中心问题

#### Version 2.5.1.1
> 1.hot fix: HomeSections proto定义有遗漏部分

#### Version 2.5.1
> 1.app首页改版，增加底部发布按钮和顶部分区入口

### Version 2.45.30.3
##### Fix
> 1.使MineSection白名单和红点接口的url进入日志

### Version 2.45.30.2
##### Fix
> 1.修正所有 golangcilint

### Version 2.45.30.1
##### Fix
> 1.修复异味导致的问题修复:isSongUpload

### Version 2.45.30
##### Feature
> 1.APP管理-我的页新增运营位配置

### Version 2.45.29
##### Fix
> 1.注释无用warning
> 2.去除代码异味

### Version 2.45.28
##### Feature
> 1.resource web特殊卡v2

### Version 2.45.27.2
##### Fix
> 1.修改服务内的casttype为int的代码

### Version 2.45.27.1
##### Fix
> 1.bug fix: 首页入口刷新出现问题时，不更替内存数据

### Version 2.45.27
##### Feature
> 1.banner inline弹幕配置

### Version 2.45.26
##### Feature
> 1.mysection 新增"新用户红点"配置

### Version 2.45.25
##### Feature
> 1.mysection 使用新的 ip 策略接口

### Version 2.45.24
##### Fix
> 1.sonar扫描bug修复

### Version 2.45.23
##### Features
> 1.内容运营稿件状态监控调整

## Version 2.45.22
### Fix
> 1.bugfix:Banner(内容投放) 区域限制location服务 AuthPIDs增加反转参数 InvertedMode

### Version 2.45.21
##### Features
> 1.是否为up主判断接口添加超时

### Version 2.45.20
##### Features
> 1.load special卡逻辑优化防止超时

### Version 2.45.19
##### Features
> 1.sidebar的gorpc接口增加返回字段

### Version 2.45.18
##### Features
> 1.banner id 配置化

### Version 2.45.17
##### Features
> 1.支持是否为up主判断接口

### Version 2.45.16
##### Features
> 1.传递天马 inline 配置的额外跳转目标

### Version 2.45.15
##### Fix
> 1.修正 cpm banner 部分字段
> 2.修正 topview 广告字段

### Version 2.45.13
##### Features
> 1.修复resource_id

### Version 2.45.12
##### Features
> 1.替换bangumi uri

### Version 2.45.11
##### Features
> 1.修正定向投放广告插位逻辑

### Version 2.45.10
##### Features
> 1.定向投放banner支持inline

### Version 2.45.9
##### Features
> 1.banner 资源位老接口只处理非定向物料

### Version 2.45.7
##### Features
> 1.banner 资源位资源物料查询接口

### Version 2.45.6.1
##### Fix
> 1.custom config慢
>
### Version 2.45.6
##### Fix
> 1.PC Web版本null问题

### Version 2.45.5
##### Features
> 1.运营404错误性能优化

### Version 2.45.4
##### Features
> 1.增加我的页面推荐服务显隐控制（display服务）

### Version 2.45.3
##### Features
> 1.迁移至从redis中读取线上主题配置数据

### Version 2.45.2
##### Features
> 1.活动gorpc切换到grpc

### Version 2.45.1
##### Features
> 1.hot fix: 临时屏蔽非人工运营404

### Version 2.45.0
##### Features
> 1.增加PC Web版头下发接口
> 2.将代码移动至resource下


### Version 2.44.4
##### Features
> 1.首页tab增加渐变色和图片配置能力

### Version 2.44.3
##### Features
> 1.天马弹窗配置

### Version 2.44.2
##### Features
> 1.我的页运营数据增加地区限制

### Version 2.44.1
##### Features
> 1.播放器自定义面板-免流：三期，支持运营商和优先级配置

### Version 2.44.0
##### Features
> 1.gorpc迁移grpc 

### Version 2.43.4
#### Business
> 1.版头增加美食区

### Version 2.43.3
#### Business
> 1.播放器自定义面板-免流：新增联通试看结束逻辑

### Version 2.43.2
#### Business
> 1.运营tab页查询逻辑修改

### Version 2.43.1
#### Business
> 1.播放器自定义面板-免流

### Version 2.43
#### Business
> 1.App首页常驻入口
> 2.UGC tab配置

### Version 2.42.9
#### Business
> 1.S10分品类热门数据接口增加日志报警

### Version 2.42.8
#### Business
> 1.新增获取S10分品类热门数据接口

### Version 2.42.7
#### Business
> 1.【天马】普通卡片支持物料编辑

### Version 2.42.6
#### Business
> 1.新增获取用户购买的进度条icon   

### Version 2.42.5
##### fix
> 1.天马特殊卡片获取通过的数据

### Version 2.42.4
##### fix
> 1.修复天马特殊卡片问题

### Version 2.42.4
##### Business
> 1.新增资讯区数据获取接口  

### Version 2.42.3
##### Features
> 1.首页运营资源新增点击事件

### Version 2.42.2
##### Features
> 1.新增grpc param方法

### Version 2.42.1
##### fix
> 1.修复banner拆轮后乱序的问题  
> 2.web端特殊投放失效问题

### Version 2.42.0
##### Features
> 1.app端内容投放拆轮(数据逻辑和业务逻辑同时改动)  
> 2.web端内容投放拆轮(数据逻辑兼容业务逻辑,web-show无影响)  

### Version 2.41.7
##### Features
> 1.DB查询下沉到resource-service

### Version 2.41.6
##### fix
> 1.修复未登录时模块屏蔽问题

### Version 2.41.6
##### Features
> 1.没有强运营和固定帧时，hash返回固定非空  
> 2.banner计算只依赖title、image、 uri

### Version 2.41.5
##### Features
> 1.白名单接口过滤空url

### Version 2.41.4
##### Features
> 1.新版我的页模块接口收敛resource

### Version 2.41.3
##### Features
> 1.增加副标题  

### Version 2.41.2
##### Bugfix
> 1.修复新增闪屏ID字段导致hash变化的问题 

### Version 2.41.1
##### Features
> 1.banner增加topview逻辑  

### Version 2.41.0
##### Features
> 1.新增web icon接口内包含pgc icon内容

### Version 2.40.1
##### Features
> 1.我的页增加运营icon配置 

### Version 2.40.0
##### Business
> 1.web运营数据增加强运营帧  

### Version 2.39.0
##### Bugfix
> 1.修复PGC播放器控件并发读写  

### Version 2.38.3
##### Features
> 1.游戏中心屏蔽新增侧边栏入口

### Version 2.38.2
##### Features
> 1.hash逻辑加入强插帧  

### Version 2.38.1
##### Features
> 1.防止panic  

### Version 2.38.0
##### Features
> 1.sql报错 不用返回 直接continue

### Version 2.37.2
##### Features
> 1.入口屏蔽配置entrance_hidden

### Version 2.37.1
> 1.客户端运营主题

### Version 2.37.0
##### Features
> 1.ogv season id

### Version 2.36.1
##### BugFix
> 1.修复test文件  

### Version 2.36.0
##### BugFix
> 1.修复更改hash的逻辑  

### Version 2.35.1
##### Features
> 1.修改banner hash计算逻辑  

### Version 2.35.0
##### Features
> 1.archive切GRPC  

### Version 2.34.1
##### Features
> 1.稿件databus的id换aid  

### Version 2.34.0
##### Features
> 1.label逻辑改为配置读取  

### Version 2.33.1
##### Features
> 1.location GRPC

### Version 2.33.0
##### Features
> 1.首页PC电竞模加label标记  

### Version 2.33.0
##### Features
> 1.播放器 pgc icon

### Version 2.32.1
##### Features
> 1.获取后台404播放页干预配置  

### Version 2.32.0
##### Features
> 1.增加插入帧逻辑  

### Version 2.31.3
##### Features
> 1.限制新banner帧数  

### Version 2.31.2
##### Features
> 1.内容运营增加OGV付费跳转类型 

### Version 2.31.1
##### Features
> 1.web_rcmd增加order字段  

### Version 2.31.0
##### Features
> 1.解决grpc接口panic的问题  

### Version 2.30.1
##### Features
> 1.去掉代码中的硬编码  

### Version 2.30.0
##### Features
> 1.grpc增加banners2  

### Version 2.29.0
##### Features
> 1.相关推荐添加推荐理由字段

### Version 2.28.2
##### Features
> 1.sidebar新增动画效果

### Version 2.28.1
##### Features
> 1.修改archive,location model

### Version 2.28.0
##### Features
> 1.修改应用的model

### Version 2.27.0
##### Features
> 1.H5 vlog up推荐增加标签  

### Version 2.26.12
##### Features
> 1.H5 vlog up推荐增加标签  

### Version 2.26.11
##### Features
> 1.侧边栏新增字段

### Version 2.26.10
##### Bugfxi
> 1.修复db注入问题  

### Version 2.26.9
##### Features
> 1.修复db注入问题  

### Version 2.26.7
##### Features
> 1.动态大家都在搜

### Version 2.26.6
##### Features
> 1.直播弹幕盒子添加人气相关参数  

### Version 2.26.5
##### Features
> 1.删除无用接口  

### Version 2.26.4
##### Features
> 1.内容投放增加投放活动的起止时间  

### Version 2.26.3
##### Features
> 1.修改数码区版本ID  

### Version 2.26.2
##### Features
> 1.新增数码区版本ID  

### Version 2.26.1
##### Features
> 1.playicon接口改造  

### Version 2.26.0
##### Features
> 1.增加web相关推荐接口  

### Version 2.25.5
##### Features
> 1.增加获取审核态接口  

### Version 2.25.4
##### Features
> 1.科技区右侧推广栏输出标签  

### Version 2.25.3
##### Features
> 1.sidebar增加Language  

### Version 2.25.2
##### Features
> 1.修改resource表查询语句的排序

### Version 2.25.1
##### Features
> 1.更换地区限制方法 

### Version 2.25.0
##### Bugfix
> 1.修复specialCache并发读写的问题  
> 2.不需要的err报错改为warn  

### Version 2.24.8
##### Features
> 1.接入grpc  
> 2.pgc特殊卡片、相关推荐  

### Version 2.24.7
##### Features
> 1.增加获取贴片视频cid接口  

### Version 2.24.6
> 1.location的Zone接口修改为Info  

### Version 2.24.5
##### Features
> 1.新增需要展示标签的位置  
> 2.URL监测功能针对直播URL的安全校验做兼容  

### Version 2.24.4
##### Bugfix
> 1.修改banner返回nil导致空指针异常的问题  

### Version 2.24.3
##### Features
> 1.修改abtest逻辑，大于等于改为大于  

### Version 2.24.2
##### Features
> 1.修改abtest的buvid转换方法  

### Version 2.24.1
##### Features
> 1.banner增加过滤逻辑，降低广告请求频率(参数中的version不为空且等于本地hashcache时，直接返回空)  

### Version 2.24.0
##### Features
> 1.identify为grpc，切换verify  

### Version 2.23.1
##### Features
> 1.siderbar增加红点字段  

### Version 2.23.0
##### Features
> 1.增加移动端“我的”数据接口  
> 2.获取resource数据时，去掉status筛选逻辑  
> 3.增加abtest接口逻辑  
> 4.增加rows.Err  

### Version 2.22.12
##### Features
> 1.URL监控增加限速  

### Version 2.22.11
##### Features
> 1.对被监控的稿件数据做格式兼容处理  

### Version 2.22.10
##### Features
> 1.新增需要获取稿件信息的推广位id  

### Version 2.22.9
##### Features
> 1.增加音频分区推荐卡片接口  

### Version 2.22.8
##### Features
> 1.整合告警信息  
> 2.URL监控告警逻辑简化  
> 3.移动端banner数据增加build过滤逻辑  
> 4.移动端banner输出数据增加stime字段  

### Version 2.22.7
##### Features
> 1.优化稿件自动下线和监测、URL监测告警的逻辑  

### Version 2.22.6
> 1.http default client add timeout

### Version 2.22.5 - 2018.05.25
##### Features
> 3.修改需要加label的位置ID  

### Version 2.22.4 - 2018.05.24
##### Features
> 1.完善内容运营后台的投放内容自动下线和告警的逻辑  
> 2.补充自动下线的各种日志  
> 3.增加需要加label的位置ID  

### Version 2.22.3 - 2018.05.23
##### Features
> 1.去掉投放内容自动下线逻辑中的多余log  

### Version 2.22.2 - 2018.05.23
##### Features
> 1.对URL监控功能中的URL做处理  

### Version 2.22.1 - 2018.05.22
##### Features
> 1.增加稿件和URL监控的开关  

### Version 2.22.0 - 2018.05.22
##### Features
> 1.去掉无用的http接口assignment、defbanner  
> 2.增加url类型模拟请求和告警  
> 3.告警方式变为企业微信  
> 4.根据稿件状态，自动下线内容运营数据  

### Version 2.21.0 - 2018.05.03
##### Features
> 1.http切bm  

### Version 2.20.1 - 2018.04.24
##### Features
> 1.接archive的discovery  
> 2.推广内容告警触发条件改为：稿件状态变更且变更后的状态不是开发浏览  
> 3.告警邮件标题  

### Version 2.20.0 - 2018.04.24
##### Features
> 1.rpc lient 增加discovery new方法  

### Version 2.19.1 - 2018.04.19
##### BugFix
> 1.修复creative_type赋值的问题  

### Version 2.19.0 - 2018.04.18
##### Features
> 1.增加创作中心creative_type字段  

### Version 2.18.2 - 2018.04.12
##### BugFix
> 1.修改banner排序逻辑  

### Version 2.18.1 - 2018.04.12
##### Features
> 1.接discovery，添加register接口  

### Version 2.18.0 - 2018.04.09
##### Features
> 1.增加直播弹幕盒子接口  

### Version 2.17.0 - 2018.02.27
##### Features
> 1.label恢复原有逻辑，不再使用note临时替代  
> 2.增加地区限制过滤  
> 3.修改推荐池投放的优先级  

### Version 2.16.2 - 2018.01.30
##### Features
> 1.未登录贴片增加aid和是否跳转  
> 2.番剧贴片增加aid  

### Version 2.16.1 - 2018.01.16
##### Features
> 1.优化video_ads相关逻辑，删除无用的逻辑和接口  
> 2.番剧贴片接口增加了跳转url字段  

### Version 2.16.0 - 2018.01.15
##### Features
> 1.播放器控件添加Hash  
> 2.修改resource和bannner的SQL  
> 3.各目录增加单元测试  

### Version 2.15.1 - 2018.01.05
##### BugFix & Features
> 1.优化编辑投放稿件状态变化告警邮件内容  
> 2.修复map并发读写问题  
> 3.修复推荐池会读取历史素材的问题  

### Version 2.15.0 - 2018.01.05
##### Features
> 1.增加番剧获取贴片的接口  

### Version 2.14.3 - 2018.01.02
##### BugFix
> 1.修复resource并发引起的slice越界问题  

### Version 2.14.2 - 2017.12.29
##### BugFix
> 1.兼容旧逻辑，返回给web的assignment数据的weight全部置为0  

### Version 2.14.1 - 2017.12.29
##### Features
> 1.兼容旧逻辑，修改resource_assignment表读的值(position->weight)  

### Version 2.14.0 - 2017.12.29
##### Features
> 1.添加获取播放器控件接口  

### Version 2.13.1 - 2017.12.28
##### Features
> 1.兼容部分旧逻辑  

### Version 2.13.0 - 2017.12.28
##### Features
> 1.根据新的内容运营平台修改resource和banner逻辑  

### Version 2.12.3 - 2017.12.20
##### Features
> 1.修改SQL，确保读的是旧数据。防止新版后台预发添加新数据影响线上服务  

### Version 2.12.2 - 2017.11.20
##### Features
> 1.banner限制总数修改  

### Version 2.12.1 - 2017.11.20
##### Bug
> 1.修正对RPC方法PasterAPP的err的判断  

### Version 2.12.0 - 2017.11.20
##### Bug
> 1.RPC方法DefBanner、Resource、PasterAPP增加返回值nil或err的判断  

### Version 2.11.0 - 2017.11.13
##### Features
> 1.banner接口增加aid参数、增加透传字段ad_extra(替代原来lat、lng)  
> 2.商业广告接口返回值增加透传字段extra  
> 3.提供新接口，获取首页引导图  

### Version 2.10.0 - 2017.11.7
##### Features
> 1.未登录贴片的投放目标ID(aid、season_id、type_id)进行数据库字段拆分  

### Version 2.9.0 - 2017.11.3
##### Features
> 1.未登录贴片的投放目标ID改为支持逗号分隔  

### Version 2.8.0 - 2017.11.1
##### Features
> 1.banner接口增加经纬度字段、增加open_event字段  

### Version 2.7.0 - 2017.10.25
##### Features
> 1.增加获取登录引导贴片的接口  

### Version 2.6.2 - 2017.10.17
##### Bug
> 1.修正bilibili_ads库的video_ads表aid为null的情况下，逗号分隔报错导致初始化失败的问题  

### Version 2.6.1 - 2017.10.17
##### Bug
> 1.修正bilibili_ads库的video_ads表部分数据default NULL导致panic的问题  

### Version 2.6.0 - 2017.10.16
##### Features
> 1.banner接口增加传参(version)  
> 2.修改banner的rpc和http接口返回值  
> 3.banner的dao层修改查询条件  

### Version 2.5.1 - 2017.10.13
##### Bug
> 1.修复banner逻辑初始化res的bug  

### Version 2.5.0 - 2017.10.11
##### Features
> 1.banner接口改为批量接口，接收多个resource_id  
> 2.banner接口的plat参数改为接收调用方传参  
> 3.banner接口增加is_ad参数，判断是否调用广告接口  
> 4.调用广告接口的方法去掉版本和mid判断  

### Version 2.4.0 - 2017.10.11
##### Features
> 1.banner接口不再进行签名校验  
> 2.banner接口返回值结构修改  
> 3.banner调广告接口的逻辑添加mobile_app和build的判断逻辑  
> 4.banner添加rpc接口  

### Version 2.3.0 - 2017.09.27
##### Features
> 1.去掉ecode.Init  
> 2.httpClient请求去掉app  

### Version 2.2.0 - 2017.09.07
##### Features
> 1.app-show的banner逻辑整体迁移进来  

### Version 2.1.0 - 2017.08.28
##### Features
> 1.添加获取bilibili_ads库video_ads表的方法(aid维度)  
> 2.添加获取bilibili_ads库video_ads表的方法(seasonid维度)  
> 2.广告接口调整目录(ad变为cpm)  

### Version 2.0.0 - 2017.08.23
##### Features
> 1.添加全量获取resource表数据的RPC方法  
> 2.添加全量获取resource_assignment表数据的RPC方法  
> 3.添加default_one表查询接口(http/rpc)  
> 4.添加单独和批量查询resource接口(http/rpc)  
> 5.添加单独和批量查询assignment接口(http/rpc)  

### Version 2.0.0 - 2017.07.21
##### Features
> 1.更新go-common v7    
> 2.去掉go-business依赖  

### Version 1.0.0
##### Features
> 1.广告获取接口(针对APP)  
> 2.添加prom监控  

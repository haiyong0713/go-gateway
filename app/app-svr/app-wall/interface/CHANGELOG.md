#### App wall 移动端商务对接接口

### Version 2.8.16
> 1.overlord伴生容器→proxyless改造

### Version 2.8.15
> 1.fix time.ParseInLocation

### Version 2.8.14
> 1.【免流】电信免流包激活验证方式优化

### Version 2.8.13
> 1.paladin.v2配置迁移

### Version 2.8.12
> 1.【免流】联通免流试看加入订购关系鉴权

### Version 2.8.11
> 1.福利社会员购优惠劵接口更换

### Version 2.8.10
> 1. 联通免流卡验证接口更换

### Version 2.8.9
> 1. 联通免流卡验证接口升级

### Version 2.8.8
> 1. 联通免流合并返回78115错误

### Version 2.8.7
> 1. 电信增加星卡免流套餐

### Version 2.8.6
> 1. 增加参数用于客户端多域名重试

### Version 2.8.5
> 1. 增加对联通验证接口的系统升级判断处理
> 2. 增加移动免流激活日志上报信息

### Version 2.8.4
> 1. 增加联通激活接口Product为空的判断
> 2. 去除联通fakeID长度大于24的错误

### Version 2.8.3
> 1. 修复激电信激活接口上报

### Version 2.8.2
> 1. 修复激活接口错误判断

### Version 2.8.1
> 1. 激活接口日志上报优化

### Version 2.7.104
> 1. 修复联通福利社流量兑换，请求联通接口参数错误问题

### Version 2.7.103
> 1. 修复78115

### Version 2.7.102
> 1. 联通错误消息优化

### Version 2.7.101
> 1. 电信免流激活接口日志和上报优化

### Version 2.7.100
> 1.cache.do方法替换context.Background()

### Version 2.7.99
> 1.同步订单context

### Version 2.7.98
> 1.【联通】【免流】修复异步更新缓存

### Version 2.7.97
> 1.【联通】【免流】同步订单去掉插入行数判断

### Version 2.7.96
> 1.【联通】【免流】手动激活返回增加fake_id判断

### Version 2.7.95
> 1.【联通】【免流】手动激活返回增加fake_id判断

### Version 2.7.94
> 1.【联通】【免流】手动激活返回增加usermob字段

### Version 2.7.93
> 1.【联通】【免流】增加联通订单同步fake_id为空时的特殊判断处理

### Version 2.7.92
> 1.【联通】【免流】增加联通免流包自动激活

### Version 2.7.91
> 1.【联通】【福利社】联通校验去掉日志告警

### Version 2.7.90
> 1.【联通】【福利社】福利社商品兑换增加一步校验

### Version 2.7.89
> 1.【联通】【福利社】福利社绑定&兑换业务接风控&加限制

### Version 2.7.88
> 1.免流激活查询平台结论优化

### Version 2.7.87
> 1.【免流】服务端下发tf-rules增加版本号

### Version 2.7.86
> 1.修复orderID

### Version 2.7.85
> 1.修复绑定bug

### Version 2.7.84
> 1.code smell

### Version 2.7.83
> 1.消除异味
> 2.mogul接入feature

### Version 2.7.82
> 1.lint bugs修复

### Version 2.7.81

> 1.去掉废弃的infoc上报

### Version 2.7.80

> 1.联通福利社处理会员购错误码83110005

### Version 2.7.79

> 1.电信派卡接口新增code和日志

### Version 2.7.78

> 1.电信派卡

### Version 2.7.76

> 1.激活接口减少-500的报错

### Version 2.7.75

> 1.用户激活日志上报修改

### Version 2.7.74

> 1.用户激活日志上报修改

### Version 2.7.73

> 1.新激活接口

### Version 2.7.72

> 1.修改移动关于订单是否生效的逻辑判断

### Version 2.7.71

> 1.增加移动同步告警

### Version 2.7.70

> 1.修复联通免流包订购panic

### Version 2.7.69

> 1.修复猫耳免流接口用户退订了但是未过期unicom_type为赋值问题

### Version 2.7.68

> 1.联通免流包激活失败错误文案修改,安卓crash,代码进行回滚

### Version 2.7.67

> 1.自动激活打印返回
> 2.联通免流包激活失败错误文案修改

### Version 2.7.66

> 1.修复canceltime endtime小于0

### Version 2.7.65

> 1.修复不显示错误message
> 2.修复缓存miss后订单获取不到

### Version 2.7.64

> 1.订单插入设置mtime

### Version 2.7.63

> 1.自动激活增加结果日志

### Version 2.7.62

> 1.app-wall-job 消费 canal，去掉 app-wall 的回源逻辑
> 2.免流卡和免流包逻辑分开

### Version 2.7.61

> 1.解决福利社兑换日志接口的超时

### Version 2.7.60

> 1.移动产品可配置化

### Version 2.7.56

> 1.福利社对账接口半价获取当时数据拉取超时

### Version 2.7.55

> 1.移动增加IP白名单,其他运营商支持热更新

### Version 2.7.54

> 1.订单增加日志告警

### Version 2.7.53

> 1.大会员兑换接口切grpc,更换batchid

### Version 2.7.52

> 1.免流产品增加描述

### Version 2.7.51

> 1.删除无用的代码

### Version 2.7.50

> 1.message为空,展示code配置的错误提示

### Version 2.7.49

> 1.联通福利社福利包兑换增加库存处理逻辑

### Version 2.7.48

> 免流订单未知的productid增加告警

### Version 2.7.47

> 1.过滤移动免流订单未知的productid

### Version 2.7.46

> 1.联通免流包新增20元大会员包，包退订接口做处理

### Version 2.7.45

> 1.库存不足和过期增加日志告警

### Version 2.7.44

> 1.联通新免流包

### Version 2.7.43

> 1.联通新免流包

### Version 2.7.42

> 1.福利社漫画券批量兑换

### Version 2.7.41

> 1.增加mogul上报

### Version 2.7.40

> 1.配置文件热更

### Version 2.7.39

> 1.取号器升级

### Version 2.7.38

> 3.免流卡直播礼包兑换

### Version 2.7.37

> 1.cron
> 2.删除过期代码

### Version 2.7.36

> 1.发号器切grpc

### Version 2.7.35

> 1.tf rule

### Version 2.7.34

> 1.accout grpc

### Version 2.7.33

> 1.新增移动免流产品，并配置化

### Version 2.7.32

> 1.新增移动免流产品，并配置化

### Version 2.7.31

> 1.清除16年的老代码，移动运营商ip，没有接口请求该数据，但是进程还继续加载文件数据到内存。

### Version 2.7.30

> 1.福利社用户信息返回用户伪码

### Version 2.7.29

> 1.电信接口加日志

### Version 2.7.28

> 1.修复app.bilibili.com/x/wall/unicom/order/pack/receive 接口存在并发写的问题

### Version 2.7.27

> 1.联通免流包新增包+大会员产品

### Version 2.7.26

> 1.宝藏卡

### Version 2.7.25

> 1.location-service rpc换grpc

### Version 2.7.24

> 1.更新移动流量产品编码

### Version 2.7.23

> 1.wall ip 校验接口增加服务端上报

### Version 2.7.22

> 1.新增联调自动激活接口，app-wall需要扩容

### Version 2.7.21

> 1.增加联通退订且为失效的状态

### Version 2.7.20

> 1.针对有用户反馈电信失败，现全量加日志

### Version 2.7.19

> 1.联通订购关系修改

### Version 2.7.18

> 1.老ip接口修改

### Version 2.7.13

> 1.提供给M站的ip接口修改
> 2.app请求的ip接口逻辑修改
> 3.电信接口返回值信息记录到日志里

### Version 2.7.12

> 1.hosts

### Version 2.7.11

> 1.为解决3026免流失败的问题，先埋点服务端，看有多少校验失败的qps

### Version 2.7.10

> 1.redis 写入改为SETEX

### Version 2.7.9

> 1.电信免流卡接口修改

### Version 2.7.8

> 1.联通福利点通过手机号查询信息增加当前手机号套餐是否已退订

### Version 2.7.7

> 1.H5跨域问题

### Version 2.7.6

> 1.联通福利点通过手机号查询信息
> 2.联调福利点批量发放
> 3.联通流量包计算sign

### Version 2.7.5

> 1.短信地区修改

### Version 2.7.4

> 1.配置文件下沉

### Version 2.7.3

> 1.电信大会员日志列表

### Version 2.7.2
> 1.漫读卷修改为50分换1张

### Version 2.7.1
> 1.福利包展示列表区分漫画app和主站app
> 2.增加漫读卷兑换流程

### Version 2.7.0

> 1.新增礼包原价，在活动时integral为现价

### Version 2.6.9
> 1.电信免流卡

### Version 2.6.8

> 1.福利社用户信息grpc

### Version 2.6.7

> 1.电信免流卡

### Version 2.6.6

> 1.修复联通问题数据

### Version 2.6.5

> 1.修复日志拉取日期判断

### Version 2.6.4

> 1.修复日志拉取日期判断

### Version 2.6.3

> 1.不等ipv4 return

### Version 2.6.2

> 1.多机房

### Version 2.6.1

> 1.多机房

### Version 2.5.19

> 1.福利社日志修改

### Version 2.5.18

1.去掉net/ip调用

### Version 2.5.17

> 1.福利社日志修改

### Version 2.5.16

> 1.福利社用户日志区分

### Version 2.5.15

> 1.福利社绑定状态

### Version 2.5.14

> 1.大会员requestNo int64

### Version 2.5.13

> 1.ip方法更换

### Version 2.5.12

> 1.修复福利社日志问题

### Version 2.5.11

> 1.build

### Version 2.5.10

> 1.福利社用户日志

### Version 2.5.9

> 1.csrf false

### Version 2.5.8

> 1.使用grpc auth

### Version 2.5.7

> 1.去除MC KEY找不到的错误

### Version 2.5.6

> 1.福利社绑定用户
> 2.缓存修改

### Version 2.5.5

> 1.福利社日志查询

### Version 2.5.4

> 1.M站IP查询用户伪码解密

### Version 2.5.3

> 1.M站接口增加IP判断

### Version 2.5.2

> 1.M站接口改用联通加密

### Version 2.5.1

> 1.M站接口开新路由

### Version 2.5.0

> 1. update infoc sdk

##### Version 2.4.3

> 1.广点通

##### Version 2.4.2

> 1.增加流量领取间隔时间
> 2.增加只能流量卡领取限制
> 3.增加积分限制

##### Version 2.4.1

> 1.http active

##### Version 2.4.0

> 1.广点通

##### Version 2.3.9

> 1.seq server

##### Version 2.3.8

>1.修复 url


##### Version 2.3.7

>1.fix bug

##### Version 2.3.6

>2.慢查询

##### Version 2.3.5

>2.gdt重构

##### Version 2.3.4

>2.gdt response返回ret 0

##### Version 2.3.3

>2.gdt 新增 advertiser_id

##### Version 2.3.2

> 1.部分接口bm.CORS

##### Version 2.3.1

> 1.bm cors

##### Version 2.3.0

> 1.http层换成BM

##### Version 2.2.9

> 1.流量卡不能订购流量包

##### Version 2.2.8

> 1.联通流量包强行删除缓存

##### Version 2.2.7

> 1.联通IP异步同步下沉

##### Version 2.2.6

> 1.异步消费逻辑放到job

##### Version 2.2.5

> 1.修复积分异常添加问题

##### Version 2.2.4

> 1.增加礼包领取日志

##### Version 2.2.3

> 1.修复联通礼包BUG
> 2.修复联通订购关系问题

##### Version 2.2.2

> 1.修复联通订购关系BUG
> 2.fix头条重复激活

##### Version 2.2.1

> 1.fix头条重复激活

##### Version 2.2.0

> 1.联通礼包

##### Version 2.1.9

> 1.头条fix

##### Version 2.1.6

> 1.广点通广告投放点击上报

##### Version 2.1.5

> 1.今日头条广告投放点击上报

##### Version 2.1.4

> 1.drop statsd

##### Version 2.1.3

> 1.移动新订单默认流量100%

##### Version 2.1.2

> 1.增加运营商数据回掉结果日志

##### Version 2.1.1

> 1.增加红点

##### Version 2.1.0

> 1.联通订购数据查询修改

##### Version 2.0.9

> 1.移动流量包增加产品类型字段

##### Version 2.0.8

> 1.联通IP同步改为异步更新

##### Version 2.0.7

> 1.联通流量包订购关系接口下沉到服务端
> 2.去除手机号

##### Version 2.0.6

> 1.更新缓存逻辑修改
> 2.删除无用的日志

##### Version 2.0.5

> 1.运营商用户数据缓存修改

##### Version 2.0.4

> 1.缓存增加回去

##### Version 2.0.3

> 1.直播礼包切回原地址

##### Version 2.0.2

> 1.直播接口切换

##### Version 2.0.1

> 1.IOS客户端需要线上测试，暂时屏蔽缓存逻辑

##### Version 2.0.0

> 1.修复电信文档与实际请求不符合问题 

##### Version 1.9.9

> 1.判断电信流量包是否是有效的

##### Version 1.9.8

> 1.errgroup error return
> 2.验证电信接口状态是否正确

##### Version 1.9.7

> 1.电信接口增加errgroup减少接口超时时间  

##### Version 1.9.6

> 1.删除多余的 ecode.NoLogin  
> 2.电信端口改成string转int
> 3.增加错误日志

##### Version 1.9.5

> 1.电信增加提示  
> 2.电信增加短信模板  
> 3.增加流水号和手机号缓存   

##### Version 1.9.4

> 1.修复error没有return问题   

##### Version 1.9.3

> 1.电信接口返回error修改 

##### Version 1.9.2

> 1.电信用户状态接口改成Get请求  
> 2.电信用户许可新增状态  
> 3.电信支付接口返回订单流水号  

##### Version 1.9.1

> 1.电信接口修改   

##### Version 1.9.0

> 1.监控图名字修改  

##### Version 1.8.9

> 1.监控图名字修改  

##### Version 1.8.8

> 1.电信数据同步地址修改   

##### Version 1.8.7

> 1.增加缓存监控  

##### Version 1.8.6

> 1.联通、电信、移动用户增加缓存  
> 2.新增电信流量包服务  

##### Version 1.8.5

> 1.联通直播礼包接口日志修改  

##### Version 1.8.4

> 1.联通直播礼包接口  

##### Version 1.8.3

> 1.联通相关接口infoc数据上报   

##### Version 1.8.2

> 1.联通直播礼包实时查询数据库     

##### Version 1.8.1

> 1.移动用户IP判断  

##### Version 1.8.0

> 1.联通用户IP判断  

##### Version 1.7.9

> 1.移动逻辑修改  

##### Version 1.7.8

> 1.移动接口数据同步合并成一个接口   

##### Version 1.7.7

> 1.中国移动流量包   

##### Version 1.7.6

> 1.实时查库去除异步读数据库   
> 2.打印联通流量包状态日志   

##### Version 1.7.5

> 1.ip限制放入配置文件  

##### Version 1.7.4

> 1.修复BUG   

##### Version 1.7.3

> 1.修复message="0"的bug   

##### Version 1.7.2

> 1.合并大仓库   

##### Version 1.7.1

> 1.dotinapp渠道  

##### Version 1.7.0

> 1.添加IP白名单  

##### Version 1.6.9

> 1.联通用户状态处理  

##### Version 1.6.8

> 1.联通用户状态处理    

##### Version 1.6.7

> 1.添加IP白名单        

##### Version 1.6.6

> 1.unicomtype在其他接口不返回      

##### Version 1.6.5

> 1.修改显示订购状态   

##### Version 1.6.4

> 1.修改显示订购状态   

##### Version 1.6.3

> 1.增加状态接口   

##### Version 1.6.2

> 1.增加状态接口   

##### Version 1.6.1

> 1.增加预开户数据同步接口   
> 2.更新vendor  

##### Version 1.6.0

> 1.增加数据同步日志  

##### Version 1.5.9

> 1.修复bug  

##### Version 1.5.8

> 1.增加错误吗提示  

##### Version 1.5.7

> 1.增加H5查看用户状态，修改   

##### Version 1.5.6

> 1.增加H5查看用户状态，修改   

##### Version 1.5.5

> 1.更新vendor  

##### Version 1.5.4

> 1.增加H5查看用户状态，改成GET   

##### Version 1.5.3

> 1.增加同步接口IP白名单  

##### Version 1.5.2

> 1.增加H5查看用户状态   

##### Version 1.5.1

> 1.升级vendor  

##### Version 1.5.0

> 1.接入新的配置中心

##### Version 1.4.9

> 1.接入新的配置中心

##### Version 1.4.8

> 1.实时查询订购关系逻辑修改     

##### Version 1.4.7

> 1.MonitorPing   

##### Version 1.4.6

> 1.本地TW   

##### Version 1.4.5

> 1.vendor升级   

##### Version 1.4.4

> 1.IP同步SQL逻辑修改   

##### Version 1.4.3

> 1.增加数据同步白名单IP  

##### Version 1.4.2

> 1.增加message提示  

##### Version 1.4.1

> 1.增加message提示  

##### Version 1.4.0

> 1.增加错误message  

##### Version 1.3.9

> 1.增加ecode  

##### Version 1.3.8

> 1.判断是否是联通IP  
> 2.增加特权礼包接口  

##### Version 1.3.7

> 1.接入平滑发布  

##### Version 1.3.6

> 1.增加spid对应卡的类型   

##### Version 1.3.5

> 1.vendor升级   

##### Version 1.3.4

> 1.增加错误码返回   

##### Version 1.3.3

> 1.增加日志信息   

##### Version 1.3.2

> 1.ordertype int改成string   

##### Version 1.3.1

> 1.增加返回字段spid   

##### Version 1.3.0

> 1.升级go-business   

##### Version 1.2.9

> 1.修改项目上报  

##### Version 1.2.8

> 1.修改联通数据同步缓存    

##### Version 1.2.7

> 1.vendor升级   
> 2.增加过期判断   
> 3.增加返回字段   
> 4.增加用户状态接口   

##### Version 1.2.6

> 1.ci更新服务镜像   

##### Version 1.2.5

> 1.配置文件支持本地读取  

##### Version 1.2.4

> 1.删除Inner  
> 2.升级vendor     

##### Version 1.2.3

> 1.删除无用的接口      

##### Version 1.2.2

> 1.vendor升级   

##### Version 1.2.1

> 1.vendor升级   

##### Version 1.2.0

> 1.vendor升级   

##### Version 1.1.9

> 1.Success int改成string  

##### Version 1.1.8

> 1.shike接口改成post请求  
> 2.vendor升级  

##### Version 1.1.7

> 1.增加联通IP      

##### Version 1.1.6

> 1.联通IP同步接口Ipbegion改成ipbegin   
> 2.增加联通IP      

##### Version 1.1.5

> 1.联通IP同步接口开始ip和结束ip字段改成string   

##### Version 1.1.4

> 1.联通IP同步接口   

##### Version 1.1.3

> 1.联通流量接口请求改成post json     

##### Version 1.1.2

> 1.usermob解密  

##### Version 1.1.1

> 1.推广渠道编号可以为空  

##### Version 1.1.0

> 1.联通流量包接口对接  

##### Version 1.0.1

> 1.修复BUG  
> 2.联通信息同步接口  

##### Version 1.0.0

> 1.初始化项目  
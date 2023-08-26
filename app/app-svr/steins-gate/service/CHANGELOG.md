
### v2.3.23
1.ios繁体版支持互动视频

### v2.3.22
1. gRPC 限流

### v2.3.18

1.安卓 HD 支持互动视频

### v2.3.17

1. 修正所有 golangcilint

### v2.3.15

1. 删除 proto typecast

### v2.3.14

1. quick lint

### v2.3.13

1. 支持安卓国际版播放

### v2.3.12

1. lint bugs修复

### v2.3.10

1. 提升节点数至 600

### v2.3.9

1. 统一变更 defaultForceHost

### v2.3.8
1. 互动视频游戏排行榜接口
2. 只记录用户历史最高成绩
3. 分数更新时清理缓存

### v2.3.5
1. 互动视频中插seek进度条

### v2.3.4
1. 关闭进度回溯的情况下隐藏变量仅记入缓存

### v2.3.3
1. 选项帧动画新增空缓存

### v2.3.2
1. 新增 edge group 为 0 时的 preload 逻辑

### v2.3.1
1. 接口支持bvid请求

### v2.3.0
1. 新增 edge 帧动画定义和拜年祭相关功能

### v2.2.8
1. 调整互动视频调用视频云的报错日志

### v2.2.7
1. 修正 edgeinfo v2 preview 接口隐藏变量问题

### v2.2.6
1. 修正 edgeinfo v2 preview 接口依赖稿件状态的问题

### v2.2.5
1. MarkEvaluations接口支持非登录场景

### v2.2.4
1. 增加一个edgeinfo_v2/preview接口

### v2.2.3
1. 修正 edgeinfo_v2 接口里无选项时缺失 edges 字段的问题

### v2.2.2
1. 增加一个批量获取评分的grpc接口

### v2.2.1
1. 处理add mark可能panic问题

### v2.2.0
1. diffMsg取值隐藏变量方法内外逻辑不一致修复

### v2.1.8
1. mark支持mid分表写入
2. ecode更新

### v2.1.7
1. condition导致不出edges的prom上报修正

### v2.1.6
1. 非中插增加非叶子节点因为condition导致不出edges的prom上报
2. 隐藏变量的存档过期时间修改为时间段

### v2.1.5
1. 服务端报错拦截优化

### v2.1.4
1.播放器提示隐藏数值

### v2.1.3
1.兼容客户端条件判断逻辑

### v2.1.2
1.允许未声明即使用局部变量，但是只存储全局变量

### v2.1.1
1.修复view接口中插树报错问题

### v2.1.0
1.新增skin/list列表接口  
2.nodeinfo,edgeinfo 接口新增皮肤信息字段  
3.graph/save 接口新增皮肤设置

### v2.0.0
1. 支持中插选项
2. 增加edgeinfo_v2接口
3. condition和action等字段下放给客户端
4. 隐藏变量支持浮点数计算和展示

### v1.9.3
1. 剧情树的diff结果json若超出1000上限则结束diff
2. 修复待审的diff应该回溯最近一次非删除的

### v1.9.2
1. 提供allowPlay方法，针对version=2（表达式/中插）的graph进行playurl接口的报错
2. 在nodeinfo接口针对version=2的graph进行提示升级的报错

### v1.9.1
1. preview已无node树，相关逻辑删除
2. 删除node2edge相关灰度逻辑

### v1.9.0
1. 存档、隐藏变量存档增加redis过期逻辑

### v1.8.9
1. 对客户端重复portal=0的情况记录prom

### v1.8.8
1. 对客户端重复portal=0请求进行兼容

### v1.8.7
1. 开环容器维度灰度发布时，不信任cursor_choice，清洗数据
2. 存档类型增加prom上报
3. 针对存档缓存的key增加开关，保证开环灰度过程中存档无损后，再强推配置使得存档缓存一次性击穿保证准确性

### v1.8.6
1. 支持有向有环图（开环逻辑）

### v1.8.5
1.修复node2edge带来的隐藏变量缓存回源上升问题

### v1.8.4
1.node_info接口增加SAR字段

### v1.8.3
1. 隐藏数值支持控制前台是否可见

### v1.8.2
1. preview开启edge树增加开关

### v1.8.1
1. 隐藏变量的缓存root edge

### v1.8.0
1. dao层拆分为小块儿，方便进行更细化的改动
2. nodeinfo & grpc接口涉及隐藏变量和存档的支持node_id字段意义改为edge_id
3. 稿件过审时新增灰度和up主白名单逻辑，逐步放量使得新graph的version=1，即edge导向图
4. 整体新增一套edge导向的逻辑

### v1.6.8
1. steins-gate-service的proto调用去掉archive的依赖

### v1.6.7
1. 针对<=5秒的视频，强制下发qte_style=1，避免客户端来不及加载
2. 在下发start_pos的时候，也应用上述逻辑，强制改为qte_style的逻辑

### v1.6.6
1. 评分展示四舍五入

### v1.6.5
1. 评分上报新增时间戳字段

### v1.6.4
1. grpc的View接口新增展示互动视频的平均分数

### v1.6.3
1. 新增节点到达节点检查

### v1.6.2
1. 预览下掉回跳逻辑

### v1.6.1
1. 进度回溯向前跳转改为2秒

### v1.6.1
1. 修正剧情树的修复待审状态定义

### v1.6.0
1. 增加评分逻辑

### v1.5.9
1. 链接配置化

### v1.5.8
1. 新增ogv使用的批量接口
2. nodeinfo接口新增ogv播放鉴权
3. 修复nodeinfo可能写入脏存档的问题

### v1.5.7
1. 私信查archive-service改为查审核库

### v1.5.6
547需求：
1. 在story_list的节点中新增is_current字段标明当前需要点亮的节点，逻辑上收回至服务端
2. 倒计时选项，新增枚举值字段，1=老的最后一帧出，2=从倒数第几秒开始出

### v1.5.5
1. 增加获取视频分辨率失败的错误
2. 去掉AegisPub配置

### v1.5.4
1. 剧情树提交进审改为同步，取消重试

### v1.5.3
1. graphPassPub & steinsCidPub & aegisAdd & aegisCancel 新增10次重试

### v1.5.2
1. dimensions请求切片50

### v1.5.1
1. dimension指针判空

### v1.5.0
1. 剧情图对接审核
2. 视频云接口改为批量

### v1.4.9
1. 新增GraphView grpc接口

### v1.4.8
1. 修复raw node问题

### v1.4.7
1. node len change to 200

### v1.4.6
1. 微信配置、视频云配置修改

### v1.4.5
1. 配置切换

### v1.4.4
1. 修改变量判断条件限制个数

### v1.4.3
1. nodeinfo接口新增预加载逻辑

### v1.4.2
1. 新增vid3o_info接口，用于编辑器中的定点位选项编辑操作
2. 将预览和提交图的操作从-20提前到分发完成
3. 修改nodeinfo中的dimension取值
4. 技改：memcache.Pool改为memcache.Memcache、dao中的error改为errors.Wrap，在外层打印、抽象videoUpView的鉴权逻辑到service/auth.go中简化代码
5. playurl预览接口增加稿件信息鉴权

### v1.4.2
1. 最近稿件列表接口增加stime和etime参数

### v1.4.1
1. 新增运营使用的稿件列表接口，查询最近4小时互动视频

### v1.4.0
1. graph attrs 增加回源逻辑

### v1.3.9
1. graph show 接口支持预发白名单查看

### v1.3.8
1. filter包

### v1.3.7
1. 修改scipt长度

### v1.3.6
1.增加cookie的prom数据上报

### v1.3.5
1. 增加infoc字段

### v1.3.4
1. 增加return

### v1.3.3
1. 用户记录改走redis

### v1.3.2
1. wechat 独立http client

### v1.3.1
1. fix send wx msg param

### v1.3.0
1. 提供审核能看节点名和分支名的接口
2. 剧情图过审的时候给审核发送过审消息

### v1.2.2
1. web端根节点请求时，如果没带buvid优先尝试从cookie中获取buvid3

### v1.2.1
1. 和前端统一字符长度计算规则

### v1.2.0
1. 新增回溯时刻判断nodeID曾经到达过

### v1.1.9
1. bm框架方法上添加sign校验

### v1.1.8
1. 去掉return

### v1.1.7
1. 针对app端请求nodeinfo新增签名校验

### v1.1.6
1. 增加隐藏变量写入-400的mobiApp的日志
2. 增加不带buvid的写入存档的mobiApp的日志

### v1.1.5
1. game record分表
2. record回源后添加缓存增加buvid
3. 增加缓存回源的prom上报

### v1.1.4
1. graph save 过滤词等级修改

### v1.1.3
1. nodeinfo/preview新增隐藏变量的展示

### v1.1.2
1. preview 保存不设日期拦截

### v1.1.1
1. writeRecord查询record有error时直接return，无记录时继续往下走

### v1.1.0
1. 编辑器新增支持隐藏变量
2. 编辑器新增剧情预览功能，及预览存档功能
3. nodeinfo新增支持隐藏变量功能

### v1.0.2
1. ecode fix

### v1.0.1
1. 判断cid是否可播判断加稿件状态

### v1.0.0
1. 上线互动视频基本功能





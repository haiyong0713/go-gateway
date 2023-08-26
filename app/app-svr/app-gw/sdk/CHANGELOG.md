## app-gw-sdk

### v0.1.15
1. 修正没有优先使用pattern的问题

### v0.1.14
1. 限流应用参数从当前环境获取

### v0.1.13
1. 修正所有 golangcilint

### v0.1.12

1. gRPC 支持获取原始 error

### v0.1.11

1. 修复sonar扫描的bug
2. http新增referer域名限流

### v0.1.10

1. gRPC 代理支持多规则模式

### v0.1.9

1. 支持 multipart 请求转发

### v0.1.8

1. sentry SDK 更新

### v0.1.7

1. 普罗米修斯 path使用encoded path

### v0.1.6
1. gRPC SDK profile接口修复，新增version

### v0.1.5
1. gRPC SDK 新增profile、config、digest接口

### v0.1.4
1. gRPC SDK 配置增加 Service 级 ClientInfo 配置

### v0.1.3
1. 优化 gRPC 重试超时时间算法，使用剩余 80% 的时间
2. 降低最小重试间隔时间至 10ms

### v0.1.1
1. 支持 gRPC 服务代理功能
2. 跟进 rate quota 改动

### v0.0.29
1. 最大重试次数区分未配置与配置0次

### v0.0.28
1. http-sdk 集成限流服务

### v0.0.27
1. 严格校验 ab 中的条件解析逻辑

### v0.0.26
1. client_info新增最大重试次数与超时时间

### v0.0.25
1. configs接口新增ProxyConfig头

### v0.0.24
1. 监控数据排序

### v0.0.23
1. 统一监控数据获取接口

### v0.0.22
1. 新增配置reload接口，扩展profile接口展示digest
2. 本地数据降级支持 jsonp 请求

### v0.0.21
1. http client 监控行为纠正
2. 监控数据标示统一使用 SafeMetricURI 方法
3. 默认禁用 retry

### v0.0.20
1. 转发请求时显式设置 host 字段

### v0.0.19
1. 转发时使用EscapedPath

### v0.0.18
1. grpc-sdk熔断降级测试

### v0.0.17
1. 新增configs接口
2. 修改metrics调用接口

### v0.0.16
1. 优化 grpc-sdk 接入体验

### v0.0.15
1. 降级采样逻辑测试

### v0.0.14
1. 修正降级采样逻辑

### v0.0.13
1. grpc-sdk 的示例与无 ab 环境下默认命中匹配策略
2. grpc-sdk interceptor 构造使用 ensureConfig

### v0.0.12
1. 代理规则的匹配测试

### v0.0.11
1. 修正代理规则的匹配顺序

### v0.0.10
1. 熔断降级的测试

### v0.0.9
1. 实现 grpc-sdk 下的请求流程抽象

### v0.0.8
1. 规范 http 默认数据降级的 content-type

### v0.0.7
1. 实例监控uptime修复

### v0.0.6
1. 监控界面调整

### v0.0.5
1. 监控界面调整，新增实例名、uptime等

### v0.0.4
1. 实现熔断与降级功能

### v0.0.3
1. response nil判断

### v0.0.2
1. 实例监控

### v0.0.1
1. 初始化 SDK 包
2. 实现 http-sdk 下的请求流程抽象
3. 实现了 http-sdk 下 request 包的单元测试

# appstatic-job

### v1.2.5
1. 修复代码异味
2. fix log misuse

### v1.2.4
1. 去掉无用的redis代码

### v1.2.3
1. 小于20M的文件也上传到boss

### v1.2.2
1. 去掉预热

### v1.2.1
1. change boss bucket

### v1.2.0
1.优化静态资源文件下载地址生成规则 

### v1.1.0
1.broadcast 推送不走job

### v1.0.3
1. 每次推送完成后强制sleep 5分钟，不进行连续推送
2. 新增开关控制是否进行推送

### v1.0.2
1. 刷完app-resource的GRPC后sleep 2秒

### v1.0.1
1. 新增接broadcast逻辑：
* 拆分dao层，分出cal_diff（增量包计算）和push（请求broadcast推送）两个dao的包出来
* 对接app-resource，计算增量包完成后请求app-resource刷新缓存，成功后再请求broadcast推送

### v1.0.0
1. 项目初始化，从appstatic-admin中迁移出增量包计算逻辑


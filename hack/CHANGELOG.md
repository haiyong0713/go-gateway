# v3.2.3
1. 更新diff-reference脚本，修改评论样式

# v3.2.2
1. 更新diff-reference脚本，增加gitlab评论功能

# v3.2.1
1. 更新diff-reference v0.0.15

# v3.2.0
1. 修改build任务脚本，app-svr/app-wall, app-svr/up-archive服务走自己的go mod

# v3.1.9
1. 删除mod lint

# v3.1.8
1. app-svr/app-feed添加jenkins

# v3.1.7
1. 过滤presubmit的jenkins job.
2. 添加postsubmit的jenkins job.

# v3.1.6
1. 更新go版本至1.16

# v3.1.5
1. correct prow job golong lint image version。

# v3.1.4
1. 修复lint相关pj配置问题。

# v3.1.3
1. 增加jenkins job.

# v3.1.2
1. 优化代码规范检查逻辑

# v3.1.1
1. 增加错误码重复校验任务

# v3.1.0
1. 初始化prow

# v3.1.0
1. 升级 golang 版本为 1.13.8
2. 升级 test 镜像

# v3.0.5
1.修改owner文件

# v3.0.4
1. 升级ut任务配置
2. 修复 test prowjob 格式错误问题
3. 修复 mod change 问题

# v3.0.3
1. go-common基础库支持avalon环境升级

# v3.0.2
1. Prow 支持 unit_test_all 和 unit_test_restrictive 参数
2. 修复 env 为字符串问题
3. app/community/thumbup/admin 任务放到 always
4. 调用 hub 使用 https 协议
5. 修复 daemon.json 配置文件挂载错误问题
6. 修复 daemon.json 配置文件挂载错误问题

# v3.0.1
1. Prow 自动生成工具支持项目在任意目录
2. 升级 go-main 镜像
3. 升级 go-main 的 dind 镜像

# v2.1.1
1. 简化 prowjob, 删除不必要 default 配置

# v2.0.2
1. Prow 支持 started.json 和 finished.json 日志
2. 修复 prowjob 配置
3. go-main 支持 started.json 和 finished.json 日志

# v2.0.1
1. UT 正式支持 DinD, 修改测试 Prow job
2. 回滚 UT dind 
3. 添加 UT 发布到 PROW-TEST 节点机器上策略
4. 修改 UT dind 运行参数
5. 修改 UT 任务配置
6. 将测试任务拉取镜像规则改为 always
7. 升级测试镜像版本为 0.0.5
8. 正式发布测试镜像版本 v1.0.0

# v1.0.8
1. 支持相对黑名单和 base 黑名单
2. community/thumbup/admin 任务使用 dind
3. community/thumbup/service|job 任务使用 dind
4. 修复任务配置格式错误问题
5. 修改 dind-test 任务脚本
6. 支持 ut 项目 dind

# v1.0.7
1. 单元测试任务加入dind部分脚本

# v1.0.6
1. 升级 prow tools 支持统一黑名单, 并与 .gitignore 文件一致
2. prow tools 忽略 common 目录
3. 修复当黑明单里面内容为非目录时, 跳过所有了 base 目录问题

# v1.0.5
1. 所有部门大仓全量接入ut. 

# v1.0.4
1. go mod 变更仅仅只构建变更内容依赖包
2. 升级 verify-mod.dep.sh 
3. 修复 UT 上传不规范目录的项目出现的问题

# v1.0.3
1. 修复 go mod tidy 和 build 生成的 go.mod 文件不一致问题
2. 升级单测镜像为 "hub.bilibili.co/k8s-prow/go-main-test:v20190802141544-6a2233de9-dirty"

# v1.0.2
1. 配置单测 job 的 dns
2. 单元测试脚本增加gcflags=-l去除inline编译参数

# v1.0.1
1. 修复 go mod 检查没有挂载 go mod 缓存磁盘问题
2. 将 lint 任务单位从文件改为项目
3. 修复 lint 任务
4. 单元测试脚本增加多仓库支持
5. 在测试镜像中添加 overlord 代理
6. 修复测试镜像中添加 overlord 代理问题
7. 清理没有意义 labels
8. go.mod 或者 go.sum 文件变更之后会运行所有任务

# v1.0.0
1. go-main 接入 prow
2. 将 prowjob 模板的 run_pr_push 去掉
3. 修改 label 定时任务配置
4. 删除 go build 详细日志
5. 修改 changelog 脚本
6. 修改 label 镜像
7. 修改 golint 脚本
8. 修复 lint 脚本参数问题
9. lint 脚本支持 go mod, 更新 update-prow 提示信息
10. 升级 build 脚本,支持项目没有 cmd 目录, 直接 build 包
11. 注释 lint 任务
12. 修复[部门任务](prow/template/auto/department/template.yaml)配置中 trusted_labels 错误问题
13. 清理没有意义 label
14. 添加 go.mod 和 go.sum 检查任务
15. 清理 ep 部门下面 owner labels
16. 清理所有 OWNER 文件里面没有意义的 label

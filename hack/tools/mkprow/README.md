# mkprow

根据目录生成 Prow 任务配置。主要分为三种任务类型。

- 仓库 JOB

1. prow-lint prow 任务配置检查
2. changelog-lint changelog 检查
3. security-lint 安全关键字检查
4. code-lint golang 代码检查

- 部门 JOB

1. build 构建任务。

- 项目 JOB

1. BUILD 构建任务。
2. TEST 单测任务。
#labels 工具

##介绍
prow 是根据 mr 上面 label 来驱动工作流程, 项目都有不同的 label, 而且 label 比较多, 而且复杂.
当前新增 labels 工具, 自动解析生成 label, 并将所有 label 写入 [labels.yaml](../../prow/labels.yaml)

##原理
1. 解析仓库下面所有的目录 OWNERS 文件, 获取所有 labels
2. 解析仓库下面 app-department 目录, 自动生成 `new-project/${department}` label
3. 将所有的 labels, 写入到 labels.yaml
4. Prow 中新增定时任务将 [labels.yaml](../../prow/labels.yaml) 文件 label 同步到仓库

#变更
用老的 [labels.yaml](../../prow/labels.yaml) 与最新对比, 如果有变更退出 1, 异常退出 2, 正常退出 0.

| 退出状态 | 描述   | 
| ----- | --------- |
| 0 | 无变更|    
| 1  | 有变更    | 
| 2 | 异常退出 |

#为什么需要 [labels.yaml](../../prow/labels.yaml)
Gitlab 的 MR 或者 Issue 添加 label 必须 label 是在当前项目的 label 库里面, 否则不可以添加, 所以新增 labels 解析所有 owner 文件
里面的 labels 写入到 [labels.yaml](../../prow/labels.yaml).


## 功能

初始化git hook，在git commit提交代码时针对本次变更代码文件自动完成以下几件事

1. 执行 go generate 更新mock代码
2. 用 gofmt 和 goimports 格式化代码
3. 用 go vet 和 golint 检查代码缺陷

如果代码规范检查不通过，本次git commit会被拒绝，只需要按照提示解决规范问题后，重新git add & git commit即可

## 使用方法

1. 如果本地没有hook工具代码，先下载 https://git.bilibili.co/platform/go-gateway/-/tree/master/app/web-svr/activity/tools/command/hook ，我们约定下载到本地的hook目录地址为$git_common_hook
2. 选择两种环境初始化方案中的其中一种，根据文档执行命令

### 环境初始化

提供两种初始化方案：

1. 全局初始化方案：只需要执行一次$git_common_hook/install操作，后续按照正常开发流程操作即可
2. 项目初始化方案：每次git clone后，进入项目目录执行$git_common_hook/install，针对当前项目初始化环境

默认(建议)：使用全局初始化，因为pre-commit时会检查是否有go文件变更，如果没有go文件变更pre-commit不做任何操作，因此对非go项目没有任何影响，可全局配置，减少重复$git_common_hook/install操作

#### 全局初始化方案

在$GOPATH/src/git.bilibili.co目录下执行，一次性对存量项目升级hook(其他git项目目录也可以执行，但是只针对执行目录项目生效hook检查机制)
```
$git_common_hook/install
```
$git_common_hook/install主要完成以下几件事

1. 将$git_common_hook/git目录配置为git初始化模板目录
2. 如果当前目录为git项目目录，则同时更新当前项目hook配置
3. 如果当前目录不是git项目目录，递归更新子目录中是git.bilibili.co下载的同时包含go文件的项目目录hook
4. 检查本地是否安装goimports和golint，未安装则自动安装

全局初始化方案采用git template在git clone时进行hook文件拷贝，因此本地已拷贝的项目默认不生效hook机制，对于本地已经clone的项目有两种方案更新hook

1. 在项目目录执行$git_common_hook/install
2. 在项目目录执行git init（注意init只会从无到有升级，不支持升级hook版本）
3. 删除项目目录，重新git clone

#### 项目初始化方案

在git clone下载代码仓库后进入仓库目录执行 
```
$git_common_hook/install -l
```

本地已有仓库也可以进入仓库目录执行以下命令初始化当前仓库hook环境
```
$git_common_hook/install -l
```

$git_common_hook/install主要用来初始化单个仓库hook环境，可重复执行覆盖更新，通常建议在每次git clone一个仓库完成后，立即进入仓库目录执行$git_common_hook/install完成hook初始化再开始编码

$git_common_hook/install主要完成以下几件事

1. 更新当前仓库hook文件
2. 检查本地是否安装goimports和golint，未安装则自动安装

### 代码提交

环境初始化完成后只需要和往常一样正常提交代码就可以了。在git commit时因编码规范检查导致commit失败后，根据提示解决问题后重新git add & git commit即可
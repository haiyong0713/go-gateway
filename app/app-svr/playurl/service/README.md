# playurl-service

# 项目简介
1.提供playurl服务

# 编译环境
> 请只用golang v1.10.x以上版本编译执行。

# 依赖包
> 1.公共包go-common

# 编译执行
go run ./cmd/main.go -t test.toml

# 云控信息
### ArcConf:稿件维度按钮的显影
is_support:显影状态 disabled:禁用状态（按钮置灰）
### CloudConf:设备维度按钮的状态
show:客户端已不使用 field_value:显影状态 conf_value:配置的值

# 云控按钮添加步骤
## proto:
1.增加一个conf_type
2.ArcConf和CloudConf中增加conf_type对应的字段
## playConf和playView:
1.在fromAbilityConf()方法内增加对应的conf_type代码
2.如果有稿件维度的控制则需要单独在playView接口中填充ArcConf对应的值
## playConfEdit:
1.在convertConfValueToAnys()方法中增加对应的conf_type的取值逻辑


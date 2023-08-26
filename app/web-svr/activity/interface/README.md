#### activity

##### 项目简介
> 1.活动相关高并发接口

##### 编译环境
> 请使用golang v1.7.x以上版本编译执行。  

##### 依赖包
> 1.公共包go-common  

##### 编译执行
> 在主目录执行go build。   
> 编译后可执行
> 连接 uat 数据库启动
```
./activity -conf configs -appid activity.service
```
> 连接 fat 数据库启动
```
./activity -conf configs -appid activity.service -dev_conf configs/fat
```   
> 也可执行 
```
./activity -conf_appid=activity -conf_version=v2.1.0 -conf_host=172.16.33.134:9011 -conf_path=/data/conf/activity -conf_env=10 -conf_token=SEHXM8x1vYhIUaZvQUmyWnMYJrF9jHJY 
```
使用配置中心测试环境配置启动服务，如无法启动，可检查token是否正确。  
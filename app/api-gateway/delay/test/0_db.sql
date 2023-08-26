create database bilibili_lego;
use bilibili_lego;

CREATE TABLE `workflow_list` (
  `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `api_name` varchar(128) NOT NULL DEFAULT '' COMMENT '发布的应用名',
  `boss` varchar(255) NOT NULL DEFAULT '' COMMENT 'boss地址 /{bucket}/xxx',
  `version` varchar(128) NOT NULL DEFAULT '' COMMENT '版本 {api_name}-{时间}',
  `wf_name` varchar(128) NOT NULL DEFAULT '' COMMENT 'workflow name',
  `discovery_id` varchar(255) NOT NULL DEFAULT '' COMMENT 'discovery_id',
  `image` varchar(512) NOT NULL DEFAULT '' COMMENT '镜像地址',
  `state` tinyint(4) NOT NULL DEFAULT 0 COMMENT '0-正常 1-结单 2-失败 3-手动停止',
  `display_name` varchar(32) NOT NULL DEFAULT '' COMMENT '发布的阶段',
  `display_state` tinyint(4) NOT NULL DEFAULT 0 COMMENT '1-运行中 2-成功 3-失败',
  `log` text COMMENT '生成代码的日志',
  `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_path` (`api_name`, `version`),
  KEY `ix_mtime` (`mtime`)
) CHARSET=utf8 COMMENT = 'workflow表';

CREATE TABLE `api_list` (
  `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `discovery_id` varchar(128) NOT NULL DEFAULT '' COMMENT '仅grpc接口用',
  `protocol` tinyint(4) NOT NULL DEFAULT 0 COMMENT '协议类型 0-grpc 1-http',
  `service` varchar(64) NOT NULL DEFAULT '' COMMENT '所属服务名称 仅grpc接口用',
  `method` varchar(16) NOT NULL DEFAULT '' COMMENT 'http请求方法',
  `path` varchar(255) NOT NULL DEFAULT '' COMMENT 'grpc路径或http链接',
  `header` varchar(1024) NOT NULL DEFAULT '' COMMENT '请求的header',
  `params` varchar(1024) NOT NULL DEFAULT '' COMMENT '请求的query',
  `form_body` varchar(1024) NOT NULL DEFAULT '' COMMENT 'form格式的body',
  `json_body` varchar(2048) NOT NULL DEFAULT '' COMMENT 'json格式的body',
  `output` varchar(2048) NOT NULL DEFAULT '' COMMENT '接口输出',
  `state` tinyint(4) NOT NULL DEFAULT 0 COMMENT '接口状态 0-正常 1-删除',
  `description` varchar(2048) NOT NULL DEFAULT '' COMMENT '接口说明',
  `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `ix_path` (`path`),
  KEY `ix_mtime` (`mtime`)
) CHARSET=utf8 COMMENT = '聚合网关api列表';

CREATE TABLE `protos` (
  `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `file_path` varchar(255) NOT NULL DEFAULT '' COMMENT 'bapis文件路径',
  `go_path` varchar(255) NOT NULL DEFAULT '' COMMENT 'bapis-go文件路径',
  `discovery_id` varchar(128) NOT NULL DEFAULT '',
  `alias` varchar(255) NOT NULL DEFAULT '' COMMENT '文件别名',
  `package` varchar(255) NOT NULL DEFAULT '' COMMENT '文件包名',
  `file` text COMMENT '文件',
  `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '最后修改时间',
  `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_discovery_id` (`discovery_id`),
  KEY `ix_mtime` (`mtime`)
) CHARSET=utf8 COMMENT = 'proto存储表';

CREATE TABLE `auto_subscribe_season_detail` (
    `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `team_id` int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '战队id',
    `mid` int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户id',
    `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    `is_deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0: 未删除， 1：已删除',
    PRIMARY KEY (`id`),
    KEY `ix_mtime` (`mtime`),
    UNIQUE KEY `ix_mtId` (`mid`, `team_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT = '一键订阅赛季/战队用户明细';

CREATE TABLE `auto_subscribe_seasons` (
    `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `season_id` int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '战队id',
    `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    `is_deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0: 未删除， 1：已删除',
    PRIMARY KEY (`id`),
    KEY `ix_mtime` (`mtime`),
    UNIQUE KEY `ix_seasonId` (`season_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT = '一键订阅赛季列表';
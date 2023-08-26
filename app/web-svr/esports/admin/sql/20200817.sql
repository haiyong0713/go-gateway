CREATE TABLE `contest_series` (
    `id` int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `parent_title` varchar(128) NOT NULL DEFAULT '' COMMENT '系列赛父阶段标题',
    `child_title` varchar(128) NOT NULL DEFAULT '' COMMENT '系列赛父阶段标题',
    `score_id` varchar(64) NOT NULL DEFAULT '' COMMENT 'score系列赛id',
    `season_id` int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '赛季id',
    `start_time` int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '比赛开始时间',
    `end_time` int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '比赛结束时间',
    `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    `is_deleted` tinyint(1) NOT NULL DEFAULT 0 COMMENT '0: 未删除， 1：已删除',
    PRIMARY KEY (`id`),
    KEY `ix_mtime` (`mtime`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT = '系列赛阶段';

ALTER TABLE es_contests ADD COLUMN series_id bigint(20) UNSIGNED NOT NULL DEFAULT 0 comment 'series id';

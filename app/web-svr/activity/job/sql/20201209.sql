CREATE TABLE `bnj_reserve_reward_rule` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `count` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '预约人数',
    `reward_id` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '奖励物品ID',
    `activity_id` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '奖池ID',
    `start_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '生效开启时间',
    `end_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '生效开启时间',
    `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `ix_count` (`count`),
    KEY `ix_mtime` (`mtime`)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '拜年纪预约发奖规则';

CREATE TABLE `bnj_reserve_reward_receive_last_id` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `count` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '预约人数',
    `last_received_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '最后一次更新记录',
    `suffix` char(2) NOT NULL DEFAULT '' COMMENT '分表后缀',
    `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `ix_count_suffix` (`count`, `suffix`),
    KEY `ix_mtime` (`mtime`)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '拜年纪预约发奖分表最后更新的id记录';

CREATE TABLE `bnj_reserve_reward_00` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `mid` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户id',
    `count` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '预约达标人数',
    `received` tinyint(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '0: 未抽奖 1: 已抽奖 2：已发放',
    `reward` TEXT NOT NULL DEFAULT '' COMMENT '中奖信息',
    `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_mtime` (`mtime`),
    UNIQUE KEY `ix_mid_count` (`mid`, `count`)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '拜年纪预约抽奖明细';

CREATE TABLE `bnj_reserve_reward_01` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_02` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_03` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_04` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_05` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_06` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_07` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_08` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_09` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_10` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_11` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_12` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_13` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_14` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_15` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_16` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_17` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_18` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_19` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_20` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_21` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_22` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_23` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_24` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_25` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_26` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_27` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_28` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_29` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_30` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_31` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_32` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_33` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_34` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_35` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_36` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_37` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_38` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_39` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_40` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_41` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_42` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_43` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_44` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_45` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_46` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_47` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_48` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_49` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_50` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_51` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_52` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_53` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_54` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_55` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_56` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_57` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_58` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_59` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_60` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_61` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_62` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_63` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_64` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_65` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_66` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_67` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_68` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_69` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_70` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_71` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_72` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_73` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_74` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_75` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_76` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_77` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_78` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_79` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_80` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_81` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_82` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_83` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_84` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_85` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_86` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_87` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_88` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_89` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_90` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_91` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_92` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_93` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_94` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_95` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_96` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_97` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_98` LIKE `bnj_reserve_reward_00`;
CREATE TABLE `bnj_reserve_reward_99` LIKE `bnj_reserve_reward_00`;

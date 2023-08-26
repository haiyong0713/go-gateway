CREATE TABLE `bnj_live_lottery_receive_last_id` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `duration` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户观看时长',
    `last_received_id` int(10) unsigned NOT NULL DEFAULT 0 COMMENT '最后一次更新记录',
    `suffix` char(2) NOT NULL DEFAULT '' COMMENT '分表后缀',
    `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `ix_duration_suffix` (`duration`, `suffix`),
    KEY `ix_mtime` (`mtime`)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '拜年纪直播间发奖分表最后更新的id记录';

CREATE TABLE `bnj_live_lottery_rule` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `duration` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户观看时长',
    `start_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '生效开启时间',
    `end_time` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '生效开启时间',
    `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_mtime` (`mtime`)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '拜年纪直播间发奖规则';

CREATE TABLE `bnj_live_user_00` (
    `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    `mid` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户id',
    `duration` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户观看时长',
    `received` tinyint(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '0: 未抽奖 1: 已抽奖 2：已发放',
    `reward` TEXT NOT NULL DEFAULT '' COMMENT '中奖信息',
    `unique_id` char(128) NOT NULL DEFAULT '' COMMENT '消息唯一ID，用于幂等',
    `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
    PRIMARY KEY (`id`),
    KEY `ix_mtime` (`mtime`),
    KEY `ix_mid` (`mid`),
    UNIQUE KEY `ix_uniqID` (`unique_id`)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '拜年纪直播间抽奖明细';

CREATE TABLE `bnj_live_user_01` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_02` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_03` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_03` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_04` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_05` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_06` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_07` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_08` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_09` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_10` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_11` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_12` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_13` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_13` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_14` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_15` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_16` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_17` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_18` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_19` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_20` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_21` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_22` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_23` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_23` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_24` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_25` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_26` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_27` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_28` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_29` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_30` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_31` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_32` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_33` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_33` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_34` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_35` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_36` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_37` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_38` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_39` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_40` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_41` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_42` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_43` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_43` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_44` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_45` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_46` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_47` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_48` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_49` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_50` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_51` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_52` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_53` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_53` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_54` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_55` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_56` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_57` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_58` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_59` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_60` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_61` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_62` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_63` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_63` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_64` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_65` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_66` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_67` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_68` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_69` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_70` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_71` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_72` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_73` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_73` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_74` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_75` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_76` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_77` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_78` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_79` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_80` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_81` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_82` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_83` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_83` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_84` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_85` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_86` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_87` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_88` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_89` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_90` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_91` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_92` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_93` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_93` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_94` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_95` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_96` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_97` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_98` LIKE `bnj_live_user_00`;
CREATE TABLE `bnj_live_user_99` LIKE `bnj_live_user_00`;
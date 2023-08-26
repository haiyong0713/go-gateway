CREATE TABLE `bnj_ar_exchange_rule` (
	`id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
	`score` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT 'AR得分档位',
	`coupon` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '得分对应奖券数量',
	`ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
	`mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
	`is_deleted` tinyint(1) UNSIGNED NOT NULL DEFAULT 0 COMMENT '逻辑删除状态',
	PRIMARY KEY (`id`),
	KEY `ix_mtime` (`mtime`)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '拜年纪AR兑换规则';

CREATE TABLE `bnj_ar_log_00` (
	`id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
	`mid` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户id',
	`date_str` char(8) NOT NULL DEFAULT '' COMMENT '记录时间日期',
	`score` SMALLINT(5) UNSIGNED NOT NULL DEFAULT 0 COMMENT '游戏得分',
	`log_index` tinyint(1) UNSIGNED NOT NULL DEFAULT 1 COMMENT '当天第x次记录',
	`ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
	`mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
	PRIMARY KEY (`id`),
	KEY `ix_mtime` (`mtime`),
	UNIQUE KEY `mid_date_index` (`mid`, `date_str`, `log_index`)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '拜年纪AR参与记录';

CREATE TABLE `bnj_ar_log_01` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_02` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_03` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_04` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_05` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_06` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_07` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_08` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_09` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_10` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_11` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_12` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_13` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_14` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_15` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_16` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_17` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_18` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_19` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_20` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_21` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_22` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_23` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_24` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_25` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_26` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_27` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_28` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_29` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_30` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_31` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_32` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_33` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_34` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_35` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_36` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_37` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_38` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_39` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_40` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_41` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_42` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_43` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_44` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_45` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_46` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_47` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_48` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_49` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_50` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_51` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_52` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_53` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_54` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_55` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_56` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_57` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_58` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_59` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_60` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_61` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_62` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_63` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_64` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_65` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_66` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_67` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_68` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_69` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_70` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_71` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_72` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_73` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_74` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_75` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_76` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_77` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_78` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_79` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_80` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_81` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_82` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_83` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_84` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_85` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_86` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_87` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_88` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_89` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_90` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_91` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_92` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_93` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_94` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_95` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_96` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_97` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_98` LIKE `bnj_ar_log_00`;
CREATE TABLE `bnj_ar_log_99` LIKE `bnj_ar_log_00`;

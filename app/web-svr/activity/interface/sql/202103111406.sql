CREATE TABLE `user_commit_manuscript_tmp_00` (
  `id` int(10) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
  `mid` int(10) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户id',
  `activity_id` int(10) NOT NULL DEFAULT 0 COMMENT '活动id',
  `content` TEXT NOT NULL DEFAULT '' COMMENT '用户上报信息',
  `bvid` varchar(128) NOT NULL DEFAULT '' COMMENT '稿件标识',
  `ctime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `mtime` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '修改时间',
  PRIMARY KEY (`id`),
  KEY `ix_mtime` (`mtime`),
  UNIQUE KEY `ix_mid_bvid` (`mid`, `activity_id`, `bvid`)
) ENGINE = InnoDB CHARSET = utf8 COMMENT '用户上报活动xxx内容数据分表';

CREATE TABLE `user_commit_manuscript_tmp_01` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_02` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_03` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_04` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_05` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_06` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_07` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_08` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_09` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_10` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_11` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_12` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_13` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_14` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_15` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_16` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_17` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_18` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_19` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_20` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_21` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_22` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_23` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_24` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_25` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_26` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_27` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_28` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_29` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_30` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_31` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_32` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_33` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_34` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_35` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_36` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_37` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_38` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_39` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_40` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_41` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_42` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_43` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_44` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_45` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_46` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_47` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_48` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_49` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_50` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_51` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_52` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_53` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_54` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_55` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_56` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_57` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_58` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_59` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_60` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_61` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_62` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_63` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_64` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_65` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_66` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_67` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_68` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_69` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_70` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_71` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_72` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_73` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_74` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_75` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_76` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_77` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_78` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_79` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_80` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_81` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_82` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_83` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_84` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_85` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_86` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_87` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_88` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_89` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_90` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_91` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_92` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_93` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_94` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_95` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_96` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_97` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_98` LIKE `user_commit_manuscript_tmp_00`;
CREATE TABLE `user_commit_manuscript_tmp_99` LIKE `user_commit_manuscript_tmp_00`;
DROP TABLE rewards_cdkey;

CREATE TABLE IF NOT EXISTS rewards_cdkey (
	id int(11) UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '自增ID',
	mid int(11) UNSIGNED NOT NULL DEFAULT 0 COMMENT '用户id',
	activity_id int(11) unsigned NOT NULL DEFAULT 0 COMMENT '活动id',
	cdkey_name varchar(50) NOT NULL DEFAULT '0' COMMENT 'cdkey名称',
	cdkey_content varchar(50) NOT NULL DEFAULT '0' COMMENT 'cdkey内容',
	unique_id varchar(50) NOT NULL DEFAULT '0' COMMENT '幂等ID',
	is_used tinyint(4) NOT NULL DEFAULT '0' COMMENT '0 未使用 1 已使用',
	ctime datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
	mtime datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
	PRIMARY KEY (id),
	UNIQUE KEY ix_mid_name (mid, unique_id),
    KEY ix_cdkey_mid (cdkey_name, mid),
	KEY ix_mtime (mtime)
) ENGINE = InnoDB CHARSET = utf8 COMMENT 'cdkey发放表';
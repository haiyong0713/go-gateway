package rewards

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCdKey(t *testing.T) {
	ctx := context.Background()

	{
		_, err := testDao.db.Exec(ctx, "drop table if exists rewards_cdkey")
		assert.Equal(t, nil, err)
		_, err = testDao.db.Exec(ctx, `
CREATE TABLE rewards_cdkey (
id int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '自增ID',
mid int(11) unsigned NOT NULL DEFAULT 0 COMMENT '用户id',
activity_id int(11) unsigned NOT NULL DEFAULT 0 COMMENT '活动id',
cdkey_name varchar(50) NOT NULL DEFAULT '0' COMMENT 'cdkey名称',
cdkey_content varchar(50) NOT NULL DEFAULT '0' COMMENT 'cdkey内容',
unique_id varchar(50) NOT NULL DEFAULT '0' COMMENT '幂等ID',
ctime datetime NOT NULL DEFAULT current_timestamp() COMMENT '创建时间',
mtime datetime NOT NULL DEFAULT current_timestamp() ON UPDATE current_timestamp() COMMENT '修改时间',
PRIMARY KEY (id),
KEY ix_mid_name (mid, cdkey_name),
KEY ix_mtime (mtime)
) ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='cdkey发放表';
`)
		assert.Equal(t, nil, err)
		if !assert.Equal(t, nil, err) {
			t.FailNow()
		}
	}

	keys := make([]string, 0)
	for i := 0; i < 1000; i++ {
		keys = append(keys, fmt.Sprintf("cd-key-%v", i))
	}
	_, err = testDao.SendCdKey(ctx, 216761, 1, "test-1", "unique_id_error")
	assert.NotEqual(t, nil, err)
	for i := 0; i < 1000; i++ {
		_, err = testDao.SendCdKey(ctx, 216761, 1, "test", fmt.Sprintf("unique_id_%v", i))
		assert.Equal(t, nil, err)
	}

	//out of stock
	_, err = testDao.SendCdKey(ctx, 216761, 1, "test", "unique_id_error")
	assert.NotEqual(t, nil, err)

	id, err := testDao.SendCdKey(ctx, 88888, 1, "test-1", fmt.Sprintf("unique_id_888"))
	assert.Equal(t, nil, err)
	t.Logf("cd key id is %v", id)
	res, err := testDao.GetCdKeyById(ctx, 88888, id)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "abc", res[0].Cdkey)
	for _, r := range res {
		t.Logf("%+v\n", r)
	}
}

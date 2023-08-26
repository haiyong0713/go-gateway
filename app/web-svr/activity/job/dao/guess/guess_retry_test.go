package guess

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/activity/job/model/guess"

	"github.com/stretchr/testify/assert"
)

// DEPLOY_ENV="" go test -v dao_test.go dao.go guess_retry_test.go guess_retry.go im_msg.go guess.go
func TestGuessRetry(t *testing.T) {
	ctx := context.Background()
	_, err := d.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS act_finish_error_guess 
             ( 
                          id          INT(11) UNSIGNED NOT NULL auto_increment comment '主键', 
                          main_id      INT(11) UNSIGNED NOT NULL DEFAULT '0' comment '竞猜主表id', 
                          result_id    INT(11) UNSIGNED NOT NULL DEFAULT '0' comment '选择竞猜从表id结算', 
                          business     INT(11) UNSIGNED NOT NULL DEFAULT '0' comment '业务类型', 
                          oid          INT(11) UNSIGNED NOT NULL DEFAULT '0' comment '业务数据源id', 
                          table_index  INT(11) UNSIGNED NOT NULL DEFAULT '0' comment '对应用户拆分表索引', 
                          odds         INT(11) UNSIGNED NOT NULL DEFAULT '0' comment '赔率', 
                          retry_status TINYINT(4) UNSIGNED NOT NULL DEFAULT '0' comment '重试状态: 未重试0，重试成功1', 
                          ctime        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP comment '创建时间', 
                          mtime        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP on 
             UPDATE CURRENT_TIMESTAMP comment '修改时间', 
                    PRIMARY KEY (id), 
                    KEY ix_mtime (mtime) 
             ) 
             engine = innodb charset = utf8 comment '竞猜结算失败记录表'`)
	assert.Equal(t, nil, err)
	err = d.AddFinishGuessFailTask(ctx, guess.FinishGuessFailTask{
		MainID:     1,
		ResultID:   1,
		Business:   1,
		Oid:        1,
		TableIndex: 1,
		Odds:       1,
	})
	assert.Equal(t, nil, err)
	list, err := d.GetAllFinishGuessFailTask(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, 1, len(list))

	err = d.MarkFinishGuessFailTaskAsDone(ctx, 1)
	assert.Equal(t, nil, err)
	//should not contain done task
	list, err = d.GetAllFinishGuessFailTask(ctx)
	assert.Equal(t, nil, err)
	assert.Equal(t, 0, len(list))
}

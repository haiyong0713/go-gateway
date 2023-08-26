package dao

import (
	"context"
	"testing"

	"go-gateway/app/web-svr/appstatic/admin/model"

	"github.com/stretchr/testify/assert"
)

func TestDao_CreatePackageAudit(t *testing.T) {
	id, err := d.CreatePackageAudit(context.Background(), &model.PackageAudit{
		Operator: "aaa",
		Behavior: "just for test",
	})
	assert.NoError(t, err)
	assert.NotEqual(t, 0, id)
	t.Log(id)
}

func TestDao_GetPackageAuditInfo(t *testing.T) {
	reply, err := d.GetPackageAuditInfo(context.Background(), 1)
	assert.NoError(t, err)
	assert.NotNil(t, reply)
	assert.Equal(t, "just for test", reply.Behavior)
}

func TestDao_ShowPackageInfoListByAppKeyAndServiceKey(t *testing.T) {
	reply, err := d.ShowPackageInfoList(context.Background(), "app_key", "service_key")
	assert.NoError(t, err)
	for _, v := range reply {
		assert.NotNil(t, v)
		assert.Equal(t, "app_key", v.AppKey)
		t.Log(v)
	}
}

func TestDao_AuditList(t *testing.T) {
	reply, err := d.AuditList(context.Background(), "app_key", "service_key")
	assert.NoError(t, err)
	assert.NotNil(t, reply)
	for _, v := range reply {
		assert.NotNil(t, v)
		assert.Equal(t, int64(0), v.AuditStatus)
		t.Log(v)
	}
}

func TestDao_FetchAllPackageByAppAndService(t *testing.T) {
	reply, err := d.FetchAllPackageByAppAndService()
	assert.NoError(t, err)
	assert.NotEqual(t, len(reply), 0)
	for k, infos := range reply {
		t.Logf("package info key is (%s)", k)
		for _, info := range infos {
			t.Logf("info message is (%+v)", info)
		}
	}
}

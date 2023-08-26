package chronosv2test

import (
	"context"
	"flag"
	"math/rand"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"go-gateway/app/web-svr/appstatic/admin/conf"
	"go-gateway/app/web-svr/appstatic/admin/model"
	"go-gateway/app/web-svr/appstatic/admin/service"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

var srv *service.Service

func init() {
	dir, _ := filepath.Abs("../../cmd/appstatic-admin-test.toml")
	flag.Set("conf", dir)
	conf.Init()
	srv = service.New(conf.Conf)
	time.Sleep(time.Second)
}

func TestShowAppInfoList(t *testing.T) {
	infos, err := srv.ShowAppInfoList(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, infos)
	for _, info := range infos {
		t.Log(info)
	}
}

func TestShowAppInfoDetail(t *testing.T) {
	AppKey := "Integration_" + strconv.FormatInt(rand.Int63n(100), 10)
	appCreateInfo := &model.AppInfo{
		AppKey: AppKey,
		AppID:  "app_id",
		Name:   "name_test",
	}
	err := srv.SaveAppInfo(context.Background(), appCreateInfo)
	assert.NoError(t, err)
	//create
	info, err := srv.ShowAppInfoDetail(context.Background(), AppKey)
	assert.NoError(t, err)
	assert.Equal(t, appCreateInfo.Name, info.Name)
	assert.Equal(t, appCreateInfo.AppID, info.AppID)
	appUpdateInfo := &model.AppInfo{
		ID:     info.ID,
		AppKey: AppKey,
		AppID:  "app_id",
		Name:   "name_test_up",
	}
	err = srv.SaveAppInfo(context.Background(), appUpdateInfo)
	assert.NoError(t, err)
	info, err = srv.ShowAppInfoDetail(context.Background(), AppKey)
	assert.NoError(t, err)
	assert.NotEqual(t, appCreateInfo.Name, info.Name)
	assert.Equal(t, appUpdateInfo.Name, info.Name)
	assert.Equal(t, appUpdateInfo.AppID, info.AppID)
	err = srv.DeleteAppInfo(context.Background(), AppKey)
	assert.NoError(t, err)
	info, err = srv.ShowAppInfoDetail(context.Background(), AppKey)
	assert.NotNil(t, err)
	assert.Nil(t, info)
}

func TestCreateAndUpdateAuditPackage(t *testing.T) {
	packageInfo := &model.PackageInfo{
		UUID:       uuid.NewV4().String(),
		Name:       "test_",
		Rank:       1,
		AppKey:     "app_key",
		ServiceKey: "service_key",
	}
	reply, err := srv.SavePackageToAudit(context.Background(), packageInfo, "bili_test")
	assert.NoError(t, err)
	assert.NotNil(t, reply)
	t.Logf("=============auditID is (%d)============", reply.AuditID)
	err = srv.AuditApproved(context.TODO(), reply.AuditID)
	assert.NoError(t, err)
	packageInfoFromDB := &model.PackageInfo{}
	packageInfoFromDB, err = srv.ShowPackageInfoDetail(context.TODO(), packageInfo.UUID)
	assert.NoError(t, err)
	assert.NotNil(t, packageInfoFromDB)
	assert.Equal(t, packageInfo.UUID, packageInfoFromDB.UUID)
	//create:version is unchanged
	assert.Equal(t, packageInfoFromDB.Version, packageInfo.Version)

	packageInfoFromDB.Name = "test_bili"
	//update:version is should add
	replyV2 := &model.PackageOpReply{}
	replyV2, err = srv.SavePackageToAudit(context.Background(), packageInfoFromDB, "bilitest")
	assert.NoError(t, err)
	assert.NotNil(t, reply)
	t.Logf("=============auditID is (%d)============", replyV2.AuditID)
	err = srv.AuditApproved(context.TODO(), replyV2.AuditID)
	assert.NoError(t, err)
	packageInfoFromDBV2 := &model.PackageInfo{}
	packageInfoFromDBV2, err = srv.ShowPackageInfoDetail(context.TODO(), packageInfoFromDB.UUID)
	assert.NoError(t, err)
	assert.NotNil(t, packageInfoFromDBV2)
	assert.Equal(t, packageInfoFromDB.UUID, packageInfoFromDBV2.UUID)
	assert.Greater(t, packageInfoFromDBV2.Version, packageInfoFromDB.Version)
}

func TestBatchSavePakcage(t *testing.T) {
	var (
		rawPackages  []*model.PackageInfo
		appkey       = "android_test"
		servicekey   = "service_danmaku"
		uuidToCreate = uuid.NewV4().String()
	)
	t.Run("先取db中的数据以备造数据使用", func(t *testing.T) {
		var err error
		rawPackages, err = srv.ShowPackageInfoList(context.Background(), appkey, servicekey)
		assert.NoError(t, err)
	})
	var (
		uuidToUpdate string
		idToUpdate   int64
		deleteUUID   = make(map[string]struct{})
	)
	for _, v := range rawPackages {
		if uuidToUpdate == "" {
			idToUpdate = v.ID
			uuidToUpdate = v.UUID
			continue
		}
		//delete
		deleteUUID[v.UUID] = struct{}{}
	}
	packages := []*model.PackageInfo{
		//update
		{
			ID:         idToUpdate,
			UUID:       uuidToUpdate,
			AppKey:     appkey,
			ServiceKey: servicekey,
			Rank:       1,
			Name:       "android_test_update",
		},
		//create
		{
			UUID:       uuidToCreate,
			AppKey:     appkey,
			ServiceKey: servicekey,
			Rank:       2,
		},
	}
	t.Run("BatchSave", func(t *testing.T) {
		err := srv.BatchSavePackage(context.Background(), packages, appkey, servicekey)
		assert.NoError(t, err)
	})
	t.Run("取出BatchSave过后的数据，比较uuid结果", func(t *testing.T) {
		var err error
		packages, err = srv.ShowPackageInfoList(context.Background(), appkey, servicekey)
		assert.NoError(t, err)
		var created bool
		for _, v := range packages {
			//应该删除
			_, ok := deleteUUID[v.UUID]
			assert.Equal(t, false, ok)
			if v.UUID == uuidToCreate {
				created = true
			}
			if v.UUID == uuidToUpdate {
				assert.Equal(t, v.Name, "android_test_update")
			}
		}
		assert.Equal(t, true, created)
	})
}

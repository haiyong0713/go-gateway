package dao

import (
	"math/rand"
	"testing"

	"go-gateway/app/web-svr/appstatic/admin/model"

	uuid2 "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestResolution(t *testing.T) {
	var (
		uuid = uuid2.NewV1().String()
	)
	t.Run("新建 dolby whitelist", func(t *testing.T) {
		err := d.AddDolbyWhiteList(&model.DolbyWhiteList{
			Model:       "testModel",
			Brand:       "TB138FC",
			BFSPath:     uuid,
			BFSPathHash: "30ed168ca3673e0faaa2ed9a04143014",
		})
		assert.NoError(t, err)
	})
	var (
		id int64
	)
	t.Run("查询 dolby whitelist", func(t *testing.T) {
		lists, err := d.FetchDolbyWhiteList()
		assert.NoError(t, err)
		for _, v := range lists {
			if v.BFSPath == uuid {
				id = v.ID
			}
		}
		assert.NotEqual(t, 0, id)
	})
	t.Run("修改 dolby whitelist", func(t *testing.T) {
		err := d.SaveDolbyWhiteList(&model.DolbyWhiteList{
			ID:          id,
			Model:       "testModelV2",
			Brand:       "TB138FC",
			BFSPath:     uuid,
			BFSPathHash: "30ed168ca3673e0faaa2ed9a04143014",
		})
		assert.NoError(t, err)
	})
	t.Run("查询 dolby whitelist", func(t *testing.T) {
		lists, err := d.FetchDolbyWhiteList()
		assert.NoError(t, err)
		for _, v := range lists {
			if v.ID == id {
				assert.Equal(t, "testModelV2", v.Model)
			}
		}
	})
}

func TestQn(t *testing.T) {
	var (
		uuid = uuid2.NewV1().String()
	)
	t.Run("新建 qn black list", func(t *testing.T) {
		err := d.AddQnBlackList(&model.QnBlackList{
			Model:  "testModel",
			Brand:  uuid,
			QnList: "1,2,3,4",
		})
		assert.NoError(t, err)
	})
	var id int64
	t.Run("查询 qn black list", func(t *testing.T) {
		lists, err := d.FetchQnBlackList()
		assert.NoError(t, err)
		for _, v := range lists {
			if v.Brand == uuid {
				id = v.ID
			}
		}
		assert.NotEqual(t, 0, id)
	})
	t.Run("修改 qn black list", func(t *testing.T) {
		err := d.SaveQnBlackList(&model.QnBlackList{
			ID:     id,
			Model:  "testModelV2",
			Brand:  uuid,
			QnList: "1,2,3,4",
		})
		assert.NoError(t, err)
	})
	t.Run("验证修改结果 qn black list", func(t *testing.T) {
		reply, err := d.FetchQnBlackList()
		assert.NoError(t, err)
		for _, v := range reply {
			if v.ID == id {
				assert.Equal(t, "testModelV2", v.Model)
			}
		}
	})
}

func TestLimitFree(t *testing.T) {
	var aid = rand.Int63n(1000000000)
	t.Run("新建 limit free", func(t *testing.T) {
		err := d.AddLimitFreeInfo(&model.LimitFreeInfo{
			Aid:      aid,
			Subtitle: "xxx",
		})
		assert.NoError(t, err)
	})
	var id int64
	t.Run("查询 limit free", func(t *testing.T) {
		reply, err := d.FetchLimitFreeList()
		assert.NoError(t, err)
		for _, v := range reply {
			if v.Aid == aid {
				id = v.ID
			}
		}
		assert.NotEqual(t, 0, id)
	})
	t.Run("修改 limit free", func(t *testing.T) {
		err := d.EditLimitFreeInfo(&model.LimitFreeInfo{
			Aid:      aid,
			ID:       id,
			Subtitle: "xx",
		})
		assert.NoError(t, err)
	})
	t.Run("验证是否修改成功 limit free", func(t *testing.T) {
		reply, err := d.FetchLimitFree(id)
		assert.NoError(t, err)
		assert.Equal(t, "xx", reply.Subtitle)
		assert.Equal(t, aid, reply.Aid)
	})
	t.Run("test FetchLimitFreeByAid", func(t *testing.T) {
		_, err := d.FetchLimitFreeByAid(123)
		assert.NoError(t, err)
	})
	t.Run("test delete", func(t *testing.T) {
		err := d.DeleteLimitFreeInfo(id)
		assert.NoError(t, err)
	})
	t.Run("test FetchLimitFreeByAid", func(t *testing.T) {
		reply, err := d.FetchLimitFreeByAid(aid)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), reply.ID)
	})
}

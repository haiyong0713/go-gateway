package laser

import (
	"context"

	laserProto "go-gateway/app/app-svr/fawkes/service/api/laser"
	"go-gateway/app/app-svr/fawkes/service/model/app"
	log "go-gateway/app/app-svr/fawkes/service/tools/logger"

	_type "git.bilibili.co/bapis/bapis-go/push/service/broadcast/type"
	"github.com/gogo/protobuf/types"
)

// LaserPushLogUpload Broadcast Channel Push Message - LogUpload
func (s *Service) LaserPushLogUpload(c context.Context, appKey string, laser *app.Laser) (msgId int64, err error) {
	var (
		body        *types.Any
		bodyMessage *laserProto.LaserLogUploadResp
		toMessage   *_type.Message
	)
	bodyMessage = &laserProto.LaserLogUploadResp{
		Taskid: laser.ID,
		Date:   laser.LogDate,
	}
	if body, err = types.MarshalAny(bodyMessage); err != nil {
		log.Error("types.MarshalAny error: %v", err)
		return
	}
	toMessage = &_type.Message{
		TargetPath: s.c.BroadcastGrpc.Laser.TargetPath,
		Body:       body,
	}
	msgId, err = s.fkDao.BroadcastPushOne(c, appKey, laser.Buvid, laser.MID, toMessage, 7*24*3600)
	return
}

// LaserPushCommand Broadcast Channel Push Message - Command
func (s *Service) LaserPushCommand(c context.Context, appKey string, cmd *app.LaserCmd) (err error) {
	log.Warn("[laser-broadcast] 用户(%v/%v/%v/%v). 准备推送消息", cmd.MID, cmd.Buvid, cmd.MobiApp, cmd.ID)
	var (
		body        *types.Any
		bodyMessage *laserProto.LaserEventResp
		toMessage   *_type.Message
	)
	bodyMessage = &laserProto.LaserEventResp{
		Taskid: cmd.ID,
		Action: cmd.Action,
		Params: cmd.Params,
	}
	if body, err = types.MarshalAny(bodyMessage); err != nil {
		log.Warn("[laser-broadcast] 用户(%v/%v/%v/%v). 数据格式化失败", cmd.MID, cmd.Buvid, cmd.MobiApp, cmd.ID)
		return
	}
	toMessage = &_type.Message{
		TargetPath: s.c.BroadcastGrpc.LaserCommand.TargetPath,
		Body:       body,
	}
	_, err = s.fkDao.BroadcastPushOne(c, appKey, cmd.Buvid, cmd.MID, toMessage, 7*24*3600)
	return
}

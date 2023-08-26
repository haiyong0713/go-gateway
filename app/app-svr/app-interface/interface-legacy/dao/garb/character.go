package garb

import (
	"context"

	live2dgrpc "git.bilibili.co/bapis/bapis-go/vas/garb/live2d/service"
)

func (d *Dao) GetUserSpaceCharacterList(ctx context.Context, req *live2dgrpc.GetUserSpaceCharacterListReq) (*live2dgrpc.GetUserSpaceCharacterListResp, error) {
	return d.live2dClient.GetUserSpaceCharacterList(ctx, req)
}

func (d *Dao) SetUserSpaceCharacter(ctx context.Context, req *live2dgrpc.SetUserSpaceCharacterReq) (*live2dgrpc.SetUserSpaceCharacterResp, error) {
	return d.live2dClient.SetUserSpaceCharacter(ctx, req)
}

func (d *Dao) RemoveUserSpaceCharacter(ctx context.Context, req *live2dgrpc.RemoveUserSpaceCharacterReq) (*live2dgrpc.RemoveUserSpaceCharacterResp, error) {
	return d.live2dClient.RemoveUserSpaceCharacter(ctx, req)
}

func (d *Dao) GetUserSpaceCharacterInfo(ctx context.Context, req *live2dgrpc.GetUserSpaceCharacterInfoReq) (*live2dgrpc.GetUserSpaceCharacterInfoResp, error) {
	return d.live2dClient.GetUserSpaceCharacterInfo(ctx, req)
}

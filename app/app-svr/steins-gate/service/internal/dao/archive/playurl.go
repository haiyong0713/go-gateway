package archive

import (
	"context"

	"go-gateway/app/app-svr/playurl/service/api"
)

// Playurl get playurl service.
func (d *Dao) Playurl(c context.Context, req *api.SteinsPreviewReq) (res *api.SteinsPreviewReply, err error) {
	return d.playurlClient.SteinsPreview(c, req)

}

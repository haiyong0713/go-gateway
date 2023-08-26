package result

import (
	"context"
	"fmt"
	"net/http"

	"go-common/library/conf/env"
	"go-common/library/ecode"
	"go-common/library/log"
)

// checkVideoShot is
func (d *Dao) CheckVideoShot(c context.Context, cid, cnt int64) error {
	host := "i0.hdslb.com"
	if env.DeployEnv == env.DeployEnvUat {
		host = "uat-" + host
	}
	for i := int64(0); i < cnt; i++ {
		imgPath := fmt.Sprintf("http://%s/bfs/videoshot/%d-%d.jpg@50q.webp", host, cid, i)
		if i == 0 {
			imgPath = fmt.Sprintf("http://%s/bfs/videoshot/%d.jpg@50q.webp", host, cid)
		}
		req, err := http.NewRequest("GET", imgPath, nil)
		if err != nil {
			log.Error("CheckVideoShot cid(%d) cnt(%d) imgPath(%s) http.NewRequest err(%+v)", cid, cnt, imgPath, err)
			return err
		}
		resp, _, err := d.client.RawResponse(c, req, "http://i0.hdslb.com/videoshot/%d.jpg@50q.webp")
		if err != nil {
			log.Error("CheckVideoShot cid(%d) cnt(%d) imgPath(%s) d.client.RawResponse err(%+v)", cid, cnt, imgPath, err)
			return err
		}
		// nolint:gomnd
		if resp.StatusCode != 200 {
			log.Error("CheckVideoShot cid(%d) cnt(%d) imgPath(%s) resp.StatusCode!=200 resp(%+v)", cid, cnt, imgPath, resp)
			return ecode.NothingFound
		}
	}
	return nil
}

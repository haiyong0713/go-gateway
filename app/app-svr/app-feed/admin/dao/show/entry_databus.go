package show

import (
	"context"
	"go-common/library/log"
)

//nolint:bilirailguncheck
func (d *Dao) PubEntryState(id string, data interface{}) (err error) {
	log.Warn("Databus Send success(%v)", id)
	if err = d.Producer.Send(context.Background(), id, &data); err != nil {
		log.Error("Databus Send error(%v)", err)
	}
	return
}

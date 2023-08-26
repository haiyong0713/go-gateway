package kvstore

import (
	"fmt"
)

type DummyMessage struct {
	HasData bool
	Payload []byte
}

func (dm *DummyMessage) Reset() {
	dm.HasData = false
	dm.Payload = dm.Payload[:0]
}
func (dm *DummyMessage) String() string {
	return fmt.Sprintf("<HasData:%+v Payload:%q>", dm.HasData, dm.Payload)
}
func (dm *DummyMessage) ProtoMessage() {}
func (dm *DummyMessage) Marshal() ([]byte, error) {
	if dm.HasData {
		return dm.Payload, nil
	}
	return nil, nil
}
func (dm *DummyMessage) Unmarshal(in []byte) error {
	dm.HasData = false
	dm.Payload = append(dm.Payload[:0], in...)
	if len(dm.Payload) > 0 {
		dm.HasData = true
	}
	return nil
}

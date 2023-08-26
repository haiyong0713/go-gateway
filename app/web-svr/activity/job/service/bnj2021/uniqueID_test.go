package bnj2021

import (
	"context"
	"sync"
	"testing"
	"time"
)

// go test -v --count=1  uniqueID_test.go live_lottery.go reserve_lottery.go service.go biz_limit_tool.go exam_stats.go
func TestGenUniqID(t *testing.T) {
	var wg sync.WaitGroup
	var m sync.Map
	for i := 0; i < 100000; i++ {
		wg.Add(1)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer func() {
				cancel()
				wg.Done()
			}()

			for {
				select {
				case <-ctx.Done():
					return
				default:
					uniqueID := genUniqueIDByUnixTime()
					if uniqueID == "" {
						t.Error(time.Now(), "uniqueID is nil")
						continue
					}

					if _, ok := m.Load(uniqueID); ok {
						t.Error(time.Now(), "uniqueID is existed", uniqueID)
					} else {
						m.Store(uniqueID, 1)
					}
				}
			}
		}()
	}

	wg.Wait()
}

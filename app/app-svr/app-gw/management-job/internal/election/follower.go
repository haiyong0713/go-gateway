package election

import (
	"context"
	"time"

	"go-common/library/log"

	"github.com/pkg/errors"
)

// Follower can follow an election in real-time and push notifications whenever
// there is a change in leadership.
type Follower struct {
	client taishanLockClient
	key    string

	leader   string
	leaderCh chan string
	errCh    chan error
}

// NewFollower creates a new follower.
func NewFollower(client taishanLockClient, key string) *Follower {
	return &Follower{
		client: client,
		key:    key,
	}
}

// Leader returns the current leader.
func (f *Follower) Leader() string {
	return f.leader
}

// FollowElection starts monitoring the election.
func (f *Follower) FollowElection(ctx context.Context) (<-chan string, <-chan error) {
	f.leaderCh = make(chan string)
	f.errCh = make(chan error)

	//nolint:biligowordcheck
	go f.follow(ctx)

	return f.leaderCh, f.errCh
}

func (f *Follower) follow(ctx context.Context) {
	defer close(f.leaderCh)
	defer close(f.errCh)

	f.leader = ""
	for {
		select {
		case <-ctx.Done():
			break
		default:
			func() {
				defer time.Sleep(5 * time.Second)
				curr, err := f.client.GetKey(context.Background(), f.key)
				if err != nil {
					log.Error("Failed to get key(%s) error(%+v)", f.key, err)
					return
				}
				if string(curr) == f.leader {
					return
				}
				f.leader = string(curr)
				f.leaderCh <- f.leader
				//nolint:gosimple
				return
			}()
		}
	}

	// Channel closed, we return an error
	//nolint:govet
	f.errCh <- errors.New("Leader Election: watch leader channel closed, the store may be unavailable...")
}

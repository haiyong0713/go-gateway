package election

import (
	"context"
	"time"

	"go-common/library/log"
)

const (
	defaultRecoverTime = 10 * time.Second
)

func SetupReplication(ctx context.Context, candidate *Candidate, follower *Follower) {
	//nolint:biligowordcheck
	go func() {
		for {
			run(ctx, candidate)
			time.Sleep(defaultRecoverTime)
		}
	}()

	//nolint:biligowordcheck
	go func() {
		for {
			follow(ctx, follower)
			time.Sleep(defaultRecoverTime)
		}
	}()
}

func run(ctx context.Context, candidate *Candidate) {
	electedCh, errCh := candidate.RunForElection(ctx)
	for {
		select {
		case isElected := <-electedCh:
			if isElected {
				log.Info("Leader Election: leadership acquired")
			} else {
				log.Info("Leader Election: leadership lost")
				// TODO(nishanttotla): perhaps EventHandler for subscription events should
			}

		case err := <-errCh:
			log.Error("Failed to run election:%+v", err)
			return
		}
	}
}

func follow(ctx context.Context, follower *Follower) {
	leaderCh, errCh := follower.FollowElection(ctx)
	for {
		select {
		case leader := <-leaderCh:
			if leader == "" {
				continue
			} else {
				log.Info("New leader elected: %s", leader)
			}

		case err := <-errCh:
			log.Error("Failed to follow election: %+v", err)
			return
		}
	}
}

func NewCandidateAndFollower(client taishanLockClient, key string, value []byte, leaderTTL uint32) (*Candidate, *Follower) {
	candidate := NewCandidate(client, key, value, leaderTTL)
	follower := NewFollower(client, key)
	return candidate, follower
}

type taishanLockClient interface {
	TryLock(ctx context.Context, key string, oldVal, newVal []byte, ttl uint32) error
	GetKey(ctx context.Context, key string) ([]byte, error)
	DelKey(ctx context.Context, key string) error
}

package election

import (
	"context"
	"sync"

	"go-common/library/log"
)

// Candidate runs the leader election algorithm asynchronously
type Candidate struct {
	client        taishanLockClient
	key           string
	instanceValue []byte

	electedCh chan bool
	lock      sync.Mutex
	lockTTL   uint32
	leader    bool
	stopCh    chan struct{}
	resignCh  chan bool
	errCh     chan error
}

// NewCandidate creates a new Candidate
func NewCandidate(client taishanLockClient, key string, value []byte, ttl uint32) *Candidate {
	return &Candidate{
		client:        client,
		key:           key,
		instanceValue: value,
		leader:        false,
		lockTTL:       ttl,
		resignCh:      make(chan bool),
		stopCh:        make(chan struct{}),
	}
}

// IsLeader returns true if the candidate is currently a leader.
func (c *Candidate) IsLeader() bool {
	return c.leader
}

// RunForElection starts the leader election algorithm. Updates in status are
// pushed through the ElectedCh channel.
//
// ElectedCh is used to get a channel which delivers signals on
// acquiring or losing leadership. It sends true if we become
// the leader, and false if we lose it.
func (c *Candidate) RunForElection(ctx context.Context) (<-chan bool, <-chan error) {
	c.electedCh = make(chan bool)
	c.errCh = make(chan error)

	//nolint:biligowordcheck
	go c.campaign(ctx)

	return c.electedCh, c.errCh
}

// Stop running for election.
func (c *Candidate) Stop() {
	close(c.stopCh)
}

// Resign forces the candidate to step-down and try again.
// If the candidate is not a leader, it doesn't have any effect.
// Candidate will retry immediately to acquire the leadership. If no-one else
// took it, then the Candidate will end up being a leader again.
func (c *Candidate) Resign() {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.leader {
		c.resignCh <- true
	}
}

func (c *Candidate) update(status bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.leader = status
	c.electedCh <- status
}

func (c *Candidate) campaign(ctx context.Context) {
	defer close(c.electedCh)
	defer close(c.errCh)

	for {
		// Start as a follower.
		c.update(false)

		tsLock := NewLock(c.client, c.key, c.instanceValue, c.lockTTL)
		lostCh, err := tsLock.Lock(nil)
		if err != nil {
			c.errCh <- err
			return
		}
		// Hooray! We acquired the lock therefore we are the new leader.
		c.update(true)

		select {
		case <-c.resignCh:
			// We were asked to resign, give up the lock and go back
			// campaigning.
			if err := c.client.DelKey(ctx, c.key); err != nil {
				log.Error("Failed to resign delKey: %+v", err)
			}
		case <-c.stopCh:
			// Give up the leadership and quit.
			if c.leader {
				if err := c.client.DelKey(ctx, c.key); err != nil {
					log.Error("Failed to give up the leadership delKey: %+v", err)
				}
			}
			return
		case <-lostCh:
			// We lost the lock. Someone else is the leader, try again.
		}
	}
}

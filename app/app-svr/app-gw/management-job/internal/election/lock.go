package election

import (
	"context"
	"time"

	"go-common/library/log"
)

type taishanLock struct {
	client taishanLockClient
	key    string
	value  []byte
	ttl    uint32
	// Closed when the caller wants to stop renewing the lock. I'm not sure
	// why this is even used - you could just call the Unlock() method.
	stopRenew chan struct{}
	// When the lock is held, this function will cancel the locked context.
	// This is called both by the Unlock() method in order to stop the
	// background holding goroutine and in a deferred call in that background
	// holding goroutine in case the lock is lost due to an error or the
	// stopRenew channel is closed. Calling this function also closes the chan
	// returned by the Lock() method.
	cancel context.CancelFunc
	// Used to sync the Unlock() call with the background holding goroutine.
	// This channel is closed when that background goroutine exits, signalling
	// that it is okay to conditionally delete the key.
	doneHolding chan struct{}
}

func NewLock(taishan taishanLockClient, key string, value []byte, ttl uint32) *taishanLock {
	lock := &taishanLock{
		client:    taishan,
		key:       key,
		value:     value,
		ttl:       ttl,
		stopRenew: make(chan struct{}),
	}
	return lock
}

func (l *taishanLock) Lock(stopChan chan struct{}) (<-chan struct{}, error) {
	for {
		err := l.client.TryLock(context.Background(), l.key, []byte{}, l.value, l.ttl)
		if err == nil {
			lockedCtx, cancel := context.WithCancel(context.Background())
			l.cancel = cancel
			l.doneHolding = make(chan struct{})
			//nolint:biligowordcheck
			go l.holdLock(lockedCtx)

			return lockedCtx.Done(), nil
		}

		// Need to wait for the lock key to expire or be deleted.
		if err := l.waitLock(stopChan); err != nil {
			return nil, err
		}
	}
}

func (l *taishanLock) holdLock(ctx context.Context) {
	defer close(l.doneHolding)
	defer l.cancel()

	update := time.NewTicker(time.Second * 5)
	defer update.Stop()

	ttl := l.ttl
	for {
		select {
		case <-update.C:
			if err := l.client.TryLock(ctx, l.key, l.value, l.value, ttl); err != nil {
				log.Error("Failed to lock key: %q: %+v", string(l.value), err)
				return
			}
		case <-l.stopRenew:
			return
		case <-ctx.Done():
			return
		}
	}
}

type taishanError interface {
	GetMsg() string
}

// WaitLock simply waits for the key to be available for creation.
func (l *taishanLock) waitLock(stopWait <-chan struct{}) error {
	waitCtx, waitCancel := context.WithCancel(context.Background())
	defer waitCancel()
	//nolint:biligowordcheck
	go func() {
		select {
		case <-stopWait:
			// If the caller closes the stopWait, cancel the wait context.
			waitCancel()
		case <-waitCtx.Done():
			// No longer waiting.
		}
	}()

	for {
		err := func() error {
			defer time.Sleep(time.Second * 5)
			if _, err := l.client.GetKey(waitCtx, l.key); err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			tsErr, ok := err.(taishanError)
			if !ok {
				return err
			}
			if tsErr.GetMsg() == "KeyNotFoundError" {
				return nil
			}
			return err
		}
	}

	//nolint:govet
	return nil
}

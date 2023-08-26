package tool

import (
	"context"
	"errors"
	"sync"

	"go-common/library/log"

	"github.com/fsnotify/fsnotify"
)

type FileWatchCallback func(ctx context.Context, event fsnotify.Event)

var (
	FileWatcher     *fsnotify.Watcher
	registerFuncMap sync.Map
)

func init() {
	registerFuncMap = sync.Map{}

	var err error
	FileWatcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Error("init file watcher failed, err: %v", err)

		return
	}

	go func() {
		for {
			select {
			case event := <-FileWatcher.Events:
				if d, ok := registerFuncMap.Load(event.Name); ok {
					if f, ok := d.(FileWatchCallback); ok {
						f(context.Background(), event)
					}
				}
			}
		}
	}()
}

func RegisterWatchHandlerV1(filename string, f FileWatchCallback) (err error) {
	if FileWatcher == nil {
		err = errors.New("FileWatcher is not initialized")

		return
	}

	err = FileWatcher.Add(filename)
	if err != nil {
		return
	}

	registerFuncMap.Store(filename, f)

	return
}

func DeRegisterWatchHandler(filename string) {
	registerFuncMap.Delete(filename)
}

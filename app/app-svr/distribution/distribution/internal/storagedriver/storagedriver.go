package storagedriver

import (
	"context"
	"sync"

	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"

	"go-common/library/log"
	"go-common/library/sync/errgroup.v2"

	"github.com/pkg/errors"
)

var (
	globalStorageDriverStore = storageDriverStore{
		store: map[string]Driver{},
	}
)

func Register(driver Driver) {
	if _, ok := globalStorageDriverStore.store[driver.Name()]; ok {
		panic(errors.Errorf("duplicate storage driver: %q", driver.Name()))
	}
	globalStorageDriverStore.store[driver.Name()] = driver
	log.Info("New storage driver %q is registered", driver.Name())
}

func GetDriver(name string) (Driver, bool) {
	driver, ok := globalStorageDriverStore.store[name]
	if !ok {
		return nil, false
	}
	return driver, true
}

type storageDriverStore struct {
	store map[string]Driver
}

type Setter interface {
	SetUserPreference(ctx context.Context, preferences []*preferenceproto.Preference) error
}

type Getter interface {
	GetUserPreference(ctx context.Context, metas []*preferenceproto.PreferenceMeta) ([]*preferenceproto.Preference, error)
}

type Driver interface {
	Setter
	Getter

	Name() string
}

func DispatchSetUserPreference(ctx context.Context, preferences []*preferenceproto.Preference) error {
	separateByDriver := map[string][]*preferenceproto.Preference{}
	for _, p := range preferences {
		separateByDriver[p.Meta.StorageDriver()] = append(separateByDriver[p.Meta.StorageDriver()], p)
	}
	eg := errgroup.WithContext(ctx)
	for driverName, subPreferences := range separateByDriver {
		driverName := driverName
		subPreferences := subPreferences

		driver, ok := GetDriver(driverName)
		if !ok {
			log.Error("Unrecognized storage driver: %q", driverName)
			continue
		}
		eg.Go(func(ctx context.Context) error {
			if err := driver.SetUserPreference(ctx, subPreferences); err != nil {
				log.Error("Failed to set user preference with driver: %q: %+v", driver.Name(), err)
				return nil
			}
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return err
	}
	return nil
}

func DispatchGetUserPreference(ctx context.Context, metas []*preferenceproto.PreferenceMeta) (map[string]*preferenceproto.Preference, error) {
	separateByDriver := map[string][]*preferenceproto.PreferenceMeta{}
	for _, m := range metas {
		separateByDriver[m.StorageDriver()] = append(separateByDriver[m.StorageDriver()], m)
	}

	lock := sync.RWMutex{}
	preferencesSeparateByDriver := map[string][]*preferenceproto.Preference{}
	eg := errgroup.WithContext(ctx)
	for driverName, subMetas := range separateByDriver {
		driverName := driverName
		subMetas := subMetas

		driver, ok := GetDriver(driverName)
		if !ok {
			log.Error("Unrecognized storage driver: %q", driverName)
			continue
		}
		eg.Go(func(ctx context.Context) error {
			preferences, err := driver.GetUserPreference(ctx, subMetas)
			if err != nil {
				log.Error("Failed to get user preference with driver: %q: %+v", driver.Name(), err)
				return nil
			}
			lock.Lock()
			defer lock.Unlock()
			preferencesSeparateByDriver[driver.Name()] = preferences
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	out := make(map[string]*preferenceproto.Preference, len(metas))
	for _, subPreferences := range preferencesSeparateByDriver {
		for _, p := range subPreferences {
			out[p.Meta.ProtoDesc.GetFullyQualifiedName()] = p
		}
	}
	return out, nil
}

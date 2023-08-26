package dao

import (
	"github.com/BurntSushi/toml"
	"sync"
)

type AuthUsers struct {
	Gateway []string
	DS      []string
}

type SideBarUsers struct {
	sync.RWMutex
	authUsers *AuthUsers
}

func newSideBarUsers() *SideBarUsers {
	return &SideBarUsers{authUsers: &AuthUsers{}}
}

func (d *dao) GetAuthUsers() *AuthUsers {
	return d.sideBarUsers.Get()
}

func (d *SideBarUsers) Get() *AuthUsers {
	d.RLock()
	defer d.RUnlock()
	return d.authUsers
}

func (d *SideBarUsers) Set(rawCfg string) error {
	d.Lock()
	defer d.Unlock()
	temp := &AuthUsers{}
	err := toml.Unmarshal([]byte(rawCfg), temp)
	d.authUsers = temp
	return err
}

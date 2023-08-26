package fawkes

import (
	"context"
	"flag"
	"fmt"
	"os"
	"testing"
	"time"

	"go-common/library/database/sql"

	"github.com/BurntSushi/toml"

	"go-gateway/app/app-svr/fawkes/service/conf"
)

var c *conf.Config
var gitmergeDao *Dao

func init() {
	var confPath string
	strings := flag.CommandLine.Args()
	fmt.Printf("%s", strings)
	args := os.Args[1:]
	for i, v := range args {
		if v == "-conf" {
			args = args[i:]
			break
		}
	}
	tf := flag.NewFlagSet("test", flag.ContinueOnError)
	tf.StringVar(&confPath, "conf", "", "config")
	if err := tf.Parse(args); err != nil {
		return
	}
	_, _ = toml.DecodeFile(confPath, &c)
	gitmergeDao = New(c)
}

func TestDao_MergeInfoInsert(t *testing.T) {
	type fields struct {
		c  *conf.Config
		db *sql.DB
	}
	type args struct {
		c           context.Context
		mergeId     int64
		appKey      string
		path        string
		state       string
		action      string
		requestUser string
		lastCommit  string
		mrStartTime time.Time
		mergedTime  time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantId  int64
		wantErr bool
	}{
		{name: "case1",
			fields: fields{c: c, db: gitmergeDao.db},
			args: args{
				c:           context.Background(),
				mergeId:     1,
				appKey:      "test",
				state:       "opened",
				action:      "open",
				requestUser: "wdlu",
				lastCommit:  "-lgm",
				mrStartTime: time.Time{},
				mergedTime:  time.Time{},
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dao{
				c:  c,
				db: gitmergeDao.db,
			}
			gotId, err := d.MergeInfoInsert(tt.args.c, tt.args.mergeId, tt.args.appKey, tt.args.path, tt.args.state, tt.args.action, tt.args.requestUser, tt.args.lastCommit, tt.args.mrStartTime, tt.args.mergedTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeInfoInsert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotId <= 0 {
				t.Errorf("MergeInfoInsert() gotId = %v", gotId)
			}
		})
	}
}

func TestDao_MergedTimeUpdate(t *testing.T) {
	type fields struct {
		c  *conf.Config
		db *sql.DB
	}
	type args struct {
		c          context.Context
		mergeId    int64
		state      string
		action     string
		mergedTime time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantId  int64
		wantErr bool
	}{
		{
			name:   "case1",
			fields: fields{c: c, db: gitmergeDao.db},
			args: args{
				c:          context.Background(),
				mergeId:    3,
				state:      "merged",
				action:     "merge",
				mergedTime: time.Now(),
			},
			wantId:  0,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dao{
				c:  c,
				db: gitmergeDao.db,
			}
			_, err := d.MergedTimeUpdate(tt.args.c, tt.args.mergeId, tt.args.state, tt.args.action, tt.args.mergedTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergedTimeUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestDao_MergeStartTimeUpdate(t *testing.T) {
	type fields struct {
		c  *conf.Config
		db *sql.DB
	}
	type args struct {
		c              context.Context
		mergeId        int64
		mergeStartTime time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantId  int64
		wantErr bool
	}{
		{
			name:   "case1",
			fields: fields{c: c, db: gitmergeDao.db},
			args: args{
				c:              context.Background(),
				mergeId:        3,
				mergeStartTime: time.Now(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Dao{
				c:  tt.fields.c,
				db: tt.fields.db,
			}
			_, err := d.MergeStartTimeUpdate(tt.args.c, tt.args.mergeId, tt.args.mergeStartTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeStartTimeUpdate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

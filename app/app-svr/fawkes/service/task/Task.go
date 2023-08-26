package task

import (
	"context"

	bm "go-common/library/net/http/blademaster"
	"go-common/library/railgun"

	"go-gateway/app/app-svr/fawkes/service/conf"
	"go-gateway/app/app-svr/fawkes/service/dao/fawkes"
	"go-gateway/app/app-svr/fawkes/service/task/bender"
	"go-gateway/app/app-svr/fawkes/service/task/event"
	"go-gateway/app/app-svr/fawkes/service/task/pack"
	"go-gateway/app/app-svr/fawkes/service/task/user"
)

type FawkesTask interface {
	HandlerFunc(ctx context.Context) railgun.MsgPolicy
	TaskName() string
}

type Task struct {
	BMHandler func(ctx *bm.Context)
	Close     func()
}

type Service struct {
	conf                         *conf.Config
	fkDao                        *fawkes.Dao
	UserReloadTask               *Task
	CIDeleteTask                 *Task
	ChannelDeleteTask            *Task
	PatchDeleteTask              *Task
	MoveTribeTask                *Task
	EventMonitorTask             *Task
	EventMonitorNotifyConfigTask *Task
	EventCompletionTask          *Task
	VedaUpdateTask               *Task
	BenderFileSyncTask           *Task
}

func New(c *conf.Config) *Service {
	fkDao := fawkes.New(c)
	svr := &Service{
		conf:  c,
		fkDao: fkDao,
	}
	svr.UserReloadTask = Register(user.NewReloadTask(c, fkDao, "用户信息刷新任务"))

	svr.CIDeleteTask = Register(pack.NewCIDeleteTask(c, fkDao, "CI过期包清理任务"))
	svr.PatchDeleteTask = Register(pack.NewPatchDeleteTask(fkDao, "过期热修复包清理"))
	svr.ChannelDeleteTask = Register(pack.NewChannelDeleteTask(fkDao, "过期渠道包清理"))

	svr.MoveTribeTask = Register(pack.NewMoveTribeTask(fkDao, "tribe产物移动任务"))

	svr.EventMonitorTask = Register(event.NewMonitorTask(c, fkDao, "技术埋点监控任务"))
	svr.EventMonitorNotifyConfigTask = Register(event.NewMonitorNotifyConfigTask(c, fkDao, "技术埋点监测通知配置更新"))
	svr.EventCompletionTask = Register(event.NewCompletionTask(c, fkDao, "技术埋点数据补全任务"))

	svr.VedaUpdateTask = Register(pack.NewVedaUpdateTask(c, fkDao, "Veda崩溃堆栈自动标记已解决"))

	svr.BenderFileSyncTask = Register(bender.NewTask(c, fkDao, "bender资源同步任务"))
	return svr
}

func Register(f FawkesTask) *Task {
	var t = &Task{}
	gun := railgun.NewRailGun(f.TaskName(), nil, railgun.NewRemoteCronInputer(nil), railgun.NewCronProcessor(nil, f.HandlerFunc))
	t.BMHandler = gun.BMHandler
	t.Close = gun.Close
	gun.Start()
	return t
}

func (s Service) Ping(ctx context.Context) error {
	err := s.fkDao.Ping(ctx)
	if err != nil {
		panic(err)
	}
	return nil
}

func (s Service) Close() error {
	s.fkDao.Close()
	s.UserReloadTask.Close()
	s.PatchDeleteTask.Close()
	s.CIDeleteTask.Close()
	s.ChannelDeleteTask.Close()
	s.MoveTribeTask.Close()
	s.EventMonitorTask.Close()
	s.EventCompletionTask.Close()
	s.EventMonitorNotifyConfigTask.Close()
	s.VedaUpdateTask.Close()
	s.BenderFileSyncTask.Close()
	return nil
}

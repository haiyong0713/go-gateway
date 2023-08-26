package http

import (
	bm "go-common/library/net/http/blademaster"
)

// 用户信息更新
func userInfoReloadTask(ctx *bm.Context) {
	s.TaskSvr.UserReloadTask.BMHandler(ctx)
}

// nas盘ci包清理
func nasCIPackDeleteTask(ctx *bm.Context) {
	s.TaskSvr.CIDeleteTask.BMHandler(ctx)
}

// nas盘patch包清理
func nasPatchPackDeleteTask(ctx *bm.Context) {
	s.TaskSvr.PatchDeleteTask.BMHandler(ctx)
}

// nas盘渠道包清理
func nasChannelPackDeleteTask(ctx *bm.Context) {
	s.TaskSvr.ChannelDeleteTask.BMHandler(ctx)
}

// tribe产物移动
func tribePackMoveTask(ctx *bm.Context) {
	s.TaskSvr.MoveTribeTask.BMHandler(ctx)
}

// 技术埋点监控
func apmEventMonitorTask(ctx *bm.Context) {
	s.TaskSvr.EventMonitorTask.BMHandler(ctx)
}

// 技术埋点监测通知配置
func apmEventMonitorNotifyConfigTask(ctx *bm.Context) {
	s.TaskSvr.EventMonitorNotifyConfigTask.BMHandler(ctx)
}

// 技术埋点补全
func apmEventCompletionTask(ctx *bm.Context) {
	s.TaskSvr.EventCompletionTask.BMHandler(ctx)
}

// 修改堆栈解析状态
func apmVedaStatusUpdate(ctx *bm.Context) {
	s.TaskSvr.VedaUpdateTask.BMHandler(ctx)
}

// bender资源同步
func benderResourceSyncTask(ctx *bm.Context) {
	s.TaskSvr.BenderFileSyncTask.BMHandler(ctx)
}

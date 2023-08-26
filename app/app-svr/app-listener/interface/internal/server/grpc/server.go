package grpc

import (
	"context"

	mauth "go-common/component/auth/middleware/grpc"
	authmeta "go-common/component/metadata/auth"
	"go-common/component/metadata/device"
	restriction "go-common/component/restriction/middleware/grpc"
	"go-common/library/conf/paladin.v2"
	"go-common/library/ecode"
	"go-common/library/net/rpc/warden"
	"go-gateway/app/app-svr/app-listener/interface/api/v1"
	"go-gateway/app/app-svr/app-listener/interface/conf"

	"google.golang.org/grpc"
)

// New new a grpc server.
func New(svc v1.ListenerServer, music v1.MusicServer) (ws *warden.Server, err error) {
	var (
		cfg     warden.ServerConfig
		ct      paladin.TOML
		appAuth conf.AppAuth
	)
	if err = paladin.Get("grpc.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Server").UnmarshalTOML(&cfg); err != nil {
		return
	}
	if err = ct.Get("AppAuth").UnmarshalTOML(&appAuth); err != nil {
		return
	}

	ws = warden.NewServer(&cfg)
	v1.RegisterListenerServer(ws.Server(), svc)
	v1.RegisterMusicServer(ws.Server(), music)
	// 注册中间件
	auth := mauth.New(nil)
	myauth := newAuthMiddleware()
	ws.Add("/bilibili.app.listener.v1.Listener/PlayURL", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/BKArcDetails", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/Playlist", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/PlaylistAdd", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/PlaylistDel", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/PlayHistory", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/PlayHistoryAdd", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/PlayHistoryDel", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/TripleLike", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/ThumbUp", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/CoinAdd", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/RcmdPlaylist", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor(), restriction.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/PlayActionReport", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/FavItemAdd", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/FavItemDel", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/FavFolderList", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/FavFolderDetail", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/FavFolderCreate", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/FavFolderDelete", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/FavItemBatch", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/FavoredInAnyFolders", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Listener/PickFeed", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/PickCardDetail", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Listener/Event", auth.UnaryServerInterceptor(true))
	ws.Add("/bilibili.app.listener.v1.Listener/Medialist", auth.UnaryServerInterceptor(true), myauth.UnaryServerInterceptor())

	// 添加旧音频的路由中间件
	legacyMusicMiddleware(ws, auth, myauth, &appAuth)

	ws, err = ws.Start()
	return
}

func legacyMusicMiddleware(ws *warden.Server, auth *mauth.Auth, myauth *authMiddleware, appAuth *conf.AppAuth) {
	ws.Add("/bilibili.app.listener.v1.Music/FavTabShow", appAuth.UnaryServerInterceptor())
	ws.Add("/bilibili.app.listener.v1.Music/MainFavMusicSubTabList", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Music/MainFavMusicMenuList", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Music/MenuEdit", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Music/MenuDelete", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Music/MenuSubscribe", auth.UnaryServerInterceptor(false))
	ws.Add("/bilibili.app.listener.v1.Music/Click", myauth.UnaryServerInterceptor())
}

// 确保 buvid 和 mid 至少设置一个
type authMiddleware struct {
}

func newAuthMiddleware() *authMiddleware {
	return &authMiddleware{}
}

func (m *authMiddleware) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		au, _ := authmeta.FromContext(ctx)
		dev, _ := device.FromContext(ctx)
		if len(dev.Buvid) == 0 && au.Mid == 0 {
			return nil, ecode.NoLogin
		}
		return handler(ctx, req)
	}
}

package component

import (
	"fmt"
	tunnelapi "git.bilibili.co/bapis/bapis-go/platform/service/tunnel"

	arcclient "git.bilibili.co/bapis/bapis-go/archive/service"
	tagrpc "git.bilibili.co/bapis/bapis-go/community/interface/tag"
	favclient "git.bilibili.co/bapis/bapis-go/community/service/favorite"
	actClient "go-gateway/app/web-svr/activity/interface/api"
	espClient "go-gateway/app/web-svr/esports/interface/api/v1"
	"go-gateway/app/web-svr/esports/job/conf"
	espServiceClient "go-gateway/app/web-svr/esports/service/api/v1"
)

var (
	ActivityClient   actClient.ActivityClient
	FavClient        favclient.FavoriteClient
	EspClient        espClient.EsportsClient
	EspServiceClient espServiceClient.EsportsServiceClient
	ArcClient        arcclient.ArchiveClient
	TunnelClient     tunnelapi.TunnelClient
	TagClient        tagrpc.TagRPCClient
)

func InitClients() error {
	client, err := actClient.NewClient(conf.Conf.ActClient)
	if err != nil {
		fmt.Println("InitClients actClient >>> ", err)
		return err
	}
	tmpFavClient, err := favclient.NewClient(conf.Conf.FavClient)
	if err != nil {
		fmt.Println("InitClients favclient >>> ", err)
		return err
	}
	tmpEspClient, err := espClient.NewClient(conf.Conf.EspClient)
	if err != nil {
		fmt.Println("InitClients espClient >>> ", err)
		return err
	}
	tmpEspServiceClient, err := espServiceClient.NewClient(conf.Conf.EspServiceClient)
	if err != nil {
		fmt.Println("InitClients espClient >>> ", err)
		return err
	}
	if ArcClient, err = arcclient.NewClient(conf.Conf.ArcClient); err != nil {
		panic(err)
	}
	if TunnelClient, err = tunnelapi.NewClient(conf.Conf.TunnelClient); err != nil {
		panic(err)
	}
	if TagClient, err = tagrpc.NewClient(conf.Conf.TagRPC); err != nil {
		panic(err)
	}
	ActivityClient = client
	FavClient = tmpFavClient
	EspClient = tmpEspClient
	EspServiceClient = tmpEspServiceClient
	return nil
}

package dao

import (
	"context"
	"fmt"
	"go-common/library/conf/paladin"
	"go-common/library/database/sql"
	"go-common/library/log"
	pb "go-gateway/app/api-gateway/api-manager/api"

	"go-gateway/app/api-gateway/api-manager/internal/model"
)

func NewDB() (db *sql.DB, cf func(), err error) {
	var (
		cfg sql.Config
		ct  paladin.TOML
	)
	if err = paladin.Get("db.toml").Unmarshal(&ct); err != nil {
		return
	}
	if err = ct.Get("Client").UnmarshalTOML(&cfg); err != nil {
		return
	}
	db = sql.NewMySQL(&cfg)
	cf = func() { _ = db.Close() }
	return
}

const (
	_insertApiSql = "INSERT INTO api_list(discovery_id,protocol,service,method,path,header,params,form_body,json_body,output,description) VALUES (?,?,?,?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE state=0," +
		"discovery_id=VALUES(discovery_id),protocol=VALUES(protocol),service=VALUES(service),header=VALUES(header),params=VALUES(params),form_body=VALUES(form_body),json_body=VALUES(json_body),output=VALUES(output),description=VALUES(description)"
	_selectHttpSql = "SELECT id,discovery_id,protocol,service,method,path,header,params,form_body,json_body,output,state,description,mtime,ctime FROM api_list WHERE protocol=1 AND state=0 AND id>? LIMIT 1000"
	_selectGrpcSql = "SELECT id,discovery_id,protocol,service,method,path,header,params,form_body,json_body,output,state,description,mtime,ctime FROM api_list WHERE protocol=0 AND discovery_id=? AND state=0 LIMIT 1000"
	_updateHttpSql = "UPDATE api_list SET state=1 WHERE id=?"
	_selectService = "SELECT discovery_id,service FROM api_list WHERE discovery_id IN (%s)"
	_selectHttp    = "SELECT path,json_body,output FROM api_list WHERE path IN (%s)"

	_insertProtoSql = "INSERT INTO protos(file_path,go_path,discovery_id,alias,package,file) VALUES (?,?,?,?,?,?) ON DUPLICATE KEY UPDATE " +
		"file_path=VALUES(file_path),go_path=VALUES(go_path),discovery_id=VALUES(discovery_id),package=VALUES(package),file=VALUES(file)"
	_selectAllProtoSql = "SELECT id,file_path,go_path,discovery_id,alias,package,file,mtime,ctime FROM protos WHERE id>? LIMIT 100"
	_selectProtoSql    = "SELECT id,file_path,go_path,discovery_id,alias,package,file,mtime,ctime FROM protos WHERE discovery_id=?"
	_selectProto       = "SELECT discovery_id,alias,go_path FROM protos WHERE discovery_id IN (%s)"
)

func (d *dao) AddApi(c context.Context, api *model.ApiRawInfo) (err error) {
	if api == nil {
		return
	}
	_, err = d.db.Exec(c, _insertApiSql, api.DiscoveryID, api.Protocol, api.ApiService, api.Method, api.ApiPath,
		api.ApiHeader, api.ApiParams, api.FormBody, api.JsonBody, api.Output, api.Description)
	if err != nil {
		log.Errorc(c, "AddApisError: %+v ", err)
	}
	return
}

func (d *dao) UpApi(c context.Context, id int64) (err error) {
	_, err = d.db.Exec(c, _updateHttpSql, id)
	if err != nil {
		log.Errorc(c, "UpApiError: %+v ", err)
	}
	return
}

func (d *dao) GetHttpApis(c context.Context) (res []*model.ApiRawInfo, err error) {
	res = make([]*model.ApiRawInfo, 0)
	var idx, maxID int64 = 0, -1
	for {
		if maxID == idx {
			break
		}
		idx = maxID
		var rows *sql.Rows
		rows, err = d.db.Query(c, _selectHttpSql, idx)
		if err != nil {
			log.Errorc(c, "GetHttpApisError: %+v ", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			a := &model.ApiRawInfo{}
			if err = rows.Scan(&a.ID, &a.DiscoveryID, &a.Protocol, &a.ApiService, &a.Method, &a.ApiPath,
				&a.ApiHeader, &a.ApiParams, &a.FormBody, &a.JsonBody, &a.Output, &a.State,
				&a.Description, &a.Mtime, &a.Ctime); err != nil {
				log.Errorc(c, "GetHttpApisError: %+v ", err)
				return
			}
			maxID = a.ID
			res = append(res, a)
		}
		err = rows.Err()
	}
	return
}

func (d *dao) GetGrpcApis(c context.Context, discoveryID string) (res []*model.ApiRawInfo, err error) {
	res = make([]*model.ApiRawInfo, 0)
	var rows *sql.Rows
	rows, err = d.db.Query(c, _selectGrpcSql, discoveryID)
	if err != nil {
		log.Errorc(c, "GetGrpcApisError: %+v ", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &model.ApiRawInfo{}
		if err = rows.Scan(&a.ID, &a.DiscoveryID, &a.Protocol, &a.ApiService, &a.Method, &a.ApiPath,
			&a.ApiHeader, &a.ApiParams, &a.FormBody, &a.JsonBody, &a.Output, &a.State,
			&a.Description, &a.Mtime, &a.Ctime); err != nil {
			log.Errorc(c, "GetGrpcApisError: %+v ", err)
			return
		}
		res = append(res, a)
	}
	err = rows.Err()
	return
}

func (d *dao) GetHttpApisByPath(c context.Context, paths []string) (res map[string]*pb.ApiInfo, err error) {
	if len(paths) == 0 {
		return
	}
	res = make(map[string]*pb.ApiInfo)
	var rows *sql.Rows
	subStr := ""
	for i, p := range paths {
		if i == 0 {
			subStr = fmt.Sprintf("'%s'", p)
			continue
		}
		subStr = fmt.Sprintf(",'%s'", p)
	}
	rows, err = d.db.Query(c, fmt.Sprintf(_selectHttp, subStr))
	if err != nil {
		log.Errorc(c, "GetHttpApisByPathError: %+v ", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &pb.ApiInfo{}
		var url string
		if err = rows.Scan(&url, &a.Input, &a.Output); err != nil {
			log.Errorc(c, "GetHttpApisByPathError: %+v ", err)
			return
		}
		res[url] = a
	}
	err = rows.Err()
	return
}

func (d *dao) GetServiceName(c context.Context, discoveryIDs []string) (res map[string][]string, err error) {
	if len(discoveryIDs) == 0 {
		return
	}
	res = make(map[string][]string)
	var rows *sql.Rows
	subStr := ""
	for i, p := range discoveryIDs {
		if i == 0 {
			subStr = fmt.Sprintf("'%s'", p)
			continue
		}
		subStr = fmt.Sprintf(",'%s'", p)
	}
	rows, err = d.db.Query(c, fmt.Sprintf(_selectService, subStr))
	if err != nil {
		log.Errorc(c, "GetServiceNameError: %+v ", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		var (
			discoveryID string
			service     string
		)
		if err = rows.Scan(&discoveryID, &service); err != nil {
			log.Errorc(c, "GetServiceNameError: %+v ", err)
			return
		}
		if _, ok := res[discoveryID]; !ok {
			res[discoveryID] = make([]string, 0)
		}
		res[discoveryID] = append(res[discoveryID], service)
	}
	err = rows.Err()
	return
}

func (d *dao) AddProto(c context.Context, pro *model.ProtoInfo) (err error) {
	if pro == nil {
		return
	}
	_, err = d.db.Exec(c, _insertProtoSql, pro.FilePath, pro.GoPath, pro.DiscoveryID,
		pro.Alias, pro.Package, pro.File)
	if err != nil {
		log.Errorc(c, "AddProtoError: %+v ", err)
	}
	return
}

func (d *dao) GetAllProtos(c context.Context) (res []*model.ProtoInfo, err error) {
	res = make([]*model.ProtoInfo, 0)
	var idx, maxID int64 = 0, -1
	for {
		if maxID == idx {
			break
		}
		idx = maxID
		var rows *sql.Rows
		rows, err = d.db.Query(c, _selectAllProtoSql, idx)
		if err != nil {
			log.Errorc(c, "GetAllProtosError: %+v ", err)
			return
		}
		defer rows.Close()
		for rows.Next() {
			a := &model.ProtoInfo{}
			if err = rows.Scan(&a.ID, &a.FilePath, &a.GoPath, &a.DiscoveryID, &a.Alias, &a.Package, &a.File, &a.Mtime, &a.Ctime); err != nil {
				log.Errorc(c, "GetAllProtosError: %+v ", err)
				return
			}
			maxID = a.ID
			res = append(res, a)
		}
		err = rows.Err()
		if err != nil {
			log.Errorc(c, "GetAllProtosError: %+v ", err)
			return
		}
	}
	return
}

func (d *dao) GetProto(c context.Context, discoveryID string) (res []*model.ProtoInfo, err error) {
	res = make([]*model.ProtoInfo, 0)
	var rows *sql.Rows
	rows, err = d.db.Query(c, _selectProtoSql, discoveryID)
	if err != nil {
		log.Errorc(c, "GetProtoError: %+v ", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &model.ProtoInfo{}
		if err = rows.Scan(&a.ID, &a.FilePath, &a.GoPath, &a.DiscoveryID, &a.Alias, &a.Package, &a.File, &a.Mtime, &a.Ctime); err != nil {
			log.Errorc(c, "GetProtoError: %+v ", err)
			return
		}
		res = append(res, a)
	}
	err = rows.Err()
	return
}

func (d *dao) GetProtoByDis(c context.Context, discoveryIDs []string) (res map[string]*pb.ApiInfo, err error) {
	if len(discoveryIDs) == 0 {
		return
	}
	res = make(map[string]*pb.ApiInfo)
	var rows *sql.Rows
	subStr := ""
	for i, p := range discoveryIDs {
		if i == 0 {
			subStr = fmt.Sprintf("'%s'", p)
			continue
		}
		subStr = fmt.Sprintf(",'%s'", p)
	}
	rows, err = d.db.Query(c, fmt.Sprintf(_selectProto, subStr))
	if err != nil {
		log.Errorc(c, "GetHttpApisByPathError: %+v ", err)
		return
	}
	defer rows.Close()
	for rows.Next() {
		a := &pb.ApiInfo{}
		var discoveryID string
		if err = rows.Scan(&discoveryID, &a.PbAlias, &a.PbPath); err != nil {
			log.Errorc(c, "GetHttpApisByPathError: %+v ", err)
			return
		}
		res[discoveryID] = a
	}
	err = rows.Err()
	return
}

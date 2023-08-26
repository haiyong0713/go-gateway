package dao

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"

	"go-common/library/database/taishan"
	"go-common/library/ecode"
	"go-common/library/log"
	"go-common/library/net/metadata"
	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model"

	"github.com/pkg/errors"
)

type taishanError interface {
	GetMsg() string
}

func deploymentIDKey(in int64) string {
	return strconv.FormatInt(math.MaxInt64-in, 10) + "/" + strconv.FormatInt(in, 10)
}

func deploymentMetaKey(deploymentType, node, gateway, id string) string {
	builder := &strings.Builder{}
	builder.WriteString("{%s-deployment-%s-%s}")
	args := []interface{}{deploymentType, node, gateway}
	if id != "" {
		builder.WriteString("/%s")
		args = append(args, id)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func deploymentConfirmKey(node, gateway, id string) string {
	return fmt.Sprintf("{deployment-%s-%s-%s}/confirm", node, gateway, id)
}

func deploymentActionLogKey(node, gateway, id string, createdAt int64) string {
	builder := &strings.Builder{}
	builder.WriteString("{deployment-%s-%s-%s}/action-log")
	args := []interface{}{node, gateway, id}
	if createdAt != 0 {
		builder.WriteString("/%s")
		args = append(args, createdAt)
	}
	return fmt.Sprintf(builder.String(), args...)
}

func (d *dao) CreateDeploymentMeta(ctx context.Context, meta *pb.DeploymentMeta) error {
	key := deploymentMetaKey(meta.DeploymentType, meta.Node, meta.Gateway, meta.DeploymentId)
	value, err := meta.Marshal()
	if err != nil {
		return err
	}
	casReq := d.taishan.NewCASReq([]byte(key), []byte{}, value)
	return d.taishan.CAS(ctx, casReq)
}

func (d *dao) SetDeploymentMeta(ctx context.Context, meta *pb.DeploymentMeta) error {
	key := deploymentMetaKey(meta.DeploymentType, meta.Node, meta.Gateway, meta.DeploymentId)
	value, err := meta.Marshal()
	if err != nil {
		return err
	}
	putReq := d.taishan.NewPutReq([]byte(key), value, 0)
	return d.taishan.Put(ctx, putReq)
}

func (d *dao) UpdateDeploymentState(ctx context.Context, src, dst *pb.DeploymentMeta) error {
	key := deploymentMetaKey(dst.DeploymentType, dst.Node, dst.Gateway, dst.DeploymentId)
	oldVal, err := src.Marshal()
	if err != nil {
		return err
	}
	newVal, err := dst.Marshal()
	if err != nil {
		return err
	}
	casReq := d.taishan.NewCASReq([]byte(key), oldVal, newVal)
	return d.taishan.CAS(ctx, casReq)
}

func (d *dao) SetDeploymentConfirm(ctx context.Context, req *pb.DeploymentReq, confirm *pb.DeploymentConfirm) error {
	key := deploymentConfirmKey(req.Node, req.Gateway, req.DeploymentId)
	value, err := confirm.Marshal()
	if err != nil {
		return err
	}
	casReq := d.taishan.NewCASReq([]byte(key), []byte{}, value)
	return d.taishan.CAS(ctx, casReq)
}

func (d *dao) GetDeploymentMeta(ctx context.Context, req *pb.DeploymentReq) (*pb.DeploymentMeta, error) {
	key := deploymentMetaKey(req.DeploymentType, req.Node, req.Gateway, req.DeploymentId)
	getReq := d.taishan.NewGetReq([]byte(key))
	record, err := d.taishan.Get(ctx, getReq)
	if err != nil {
		return nil, err
	}
	reply := &pb.DeploymentMeta{}
	if err := reply.Unmarshal(record.Columns[0].Value); err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *dao) GetDeploymentConfirm(ctx context.Context, req *pb.DeploymentReq) (*pb.DeploymentConfirm, error) {
	key := deploymentConfirmKey(req.Node, req.Gateway, req.DeploymentId)
	getReq := d.taishan.NewGetReq([]byte(key))
	record, err := d.taishan.Get(ctx, getReq)
	if err != nil {
		return nil, err
	}
	reply := &pb.DeploymentConfirm{}
	if err := reply.Unmarshal(record.Columns[0].Value); err != nil {
		return nil, err
	}
	return reply, nil
}

func (d *dao) DeploymentIsConfirmed(ctx context.Context, req *pb.DeploymentReq) (bool, error) {
	key := deploymentConfirmKey(req.Node, req.Gateway, req.DeploymentId)
	getReq := d.taishan.NewGetReq([]byte(key))
	_, err := d.taishan.Get(ctx, getReq)
	if err != nil {
		return checkErr(err)
	}
	return true, nil
}

func checkErr(err error) (bool, error) {
	tsErr, ok := errors.Cause(err).(taishanError)
	if !ok {
		return false, err
	}
	if tsErr.GetMsg() == "KeyNotFoundError" {
		return false, nil
	}
	return false, err
}

func (d *dao) GetDeploymentActionLog(ctx context.Context, req *pb.DeploymentReq) ([]*pb.ActionLog, error) {
	key := deploymentActionLogKey(req.Node, req.Gateway, req.DeploymentId, 0)
	start, end := fullRange(key)
	out := []*pb.ActionLog{}
	scanReq := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	for {
		reply, err := d.taishan.Scan(ctx, scanReq)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			actionLog := &pb.ActionLog{}
			if err := actionLog.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal action log: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, actionLog)
		}
		if !reply.HasNext {
			break
		}
		scanReq.StartRec = &taishan.Record{
			Key: append(reply.NextKey, 0x00),
		}
	}
	return out, nil
}

func (d *dao) ReloadConfig(ctx context.Context, req *model.ReloadConfigReq) (*model.ReloadConfigReply, error) {
	var res struct {
		Code int                      `json:"code"`
		Data *model.ReloadConfigReply `json:"data"`
	}
	u, err := url.Parse(req.Host)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to parse url: %s", req.Host)
	}
	u.Path = "/_/reload"
	if req.IsGRPC {
		u.Path = "/_/grpc-reload"
	}
	params := url.Values{}
	params.Set("digest", req.Digest)
	params.Set("content", req.Content)
	params.Set("original_digest", req.OriginalDigest)
	if err := d.http.Post(ctx, u.String(), metadata.String(ctx, metadata.RemoteIP), params, &res); err != nil {
		return nil, err
	}
	if res.Code != 0 {
		return nil, errors.Wrap(ecode.Int(res.Code), u.String())
	}
	return res.Data, nil
}

func (d *dao) AddActionLog(ctx context.Context, req *pb.AddActionLogReq) {
	key := deploymentActionLogKey(req.Node, req.Gateway, req.DeploymentId, req.ActionLog.CreatedAt)
	value, err := req.ActionLog.Marshal()
	if err != nil {
		log.Error("Failed to marshal action log: %+v", err)
		return
	}
	putReq := d.taishan.NewPutReq([]byte(key), value, 0)
	if err = d.taishan.Put(ctx, putReq); err != nil {
		log.Error("Failed to put taishan: %+v", err)
	}
}

func (d *dao) ListDeployment(ctx context.Context, req *pb.ListDeploymentReq) ([]*pb.DeploymentMeta, error) {
	start := deploymentMetaKey(req.DeploymentType, req.Node, req.Gateway, deploymentIDKey(req.Etime))
	end := deploymentMetaKey(req.DeploymentType, req.Node, req.Gateway, deploymentIDKey(req.Stime))
	scanReq := d.taishan.NewScanReq([]byte(start), []byte(end), 100)
	out := []*pb.DeploymentMeta{}
	for {
		reply, err := d.taishan.Scan(ctx, scanReq)
		if err != nil {
			return nil, err
		}
		for _, r := range reply.Records {
			dm := &pb.DeploymentMeta{}
			if err := dm.Unmarshal(r.Columns[0].Value); err != nil {
				log.Error("Failed to unmarshal deployment: %+v", errors.WithStack(err))
				continue
			}
			out = append(out, dm)
		}
		if !reply.HasNext {
			break
		}
		scanReq.StartRec = &taishan.Record{
			Key: append(reply.NextKey, 0x00),
		}
	}
	return out, nil
}

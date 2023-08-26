package service

import (
	"context"
	"strings"

	pb "go-gateway/app/app-svr/app-gw/baas/api"
	"go-gateway/app/app-svr/app-gw/baas/utils/sets"
)

func (s *Service) AuthZ(ctx context.Context, req *pb.AuthZReq) (*pb.AuthZReply, error) {
	nodes, err := s.dao.FetchRoleTree(ctx, req.Username, req.Cookie)
	if err != nil {
		return nil, err
	}
	projectSet := sets.NewString()
	for _, n := range nodes {
		projectSet.Insert(nodeDir(n.Path))
	}

	reply := &pb.AuthZReply{}
	for _, p := range projectSet.List() {
		reply.Projects = append(reply.Projects, &pb.Project{
			ProjectName: p,
			Node:        p,
		})
	}
	return reply, nil
}

// Trim "bilibili.main.web-svr.playlist-job" as "main.web-svr"
//nolint:gomnd
func nodeDir(path string) string {
	path = strings.TrimPrefix(path, "bilibili.")
	parts := strings.SplitN(path, ".", 3)
	if len(parts) < 3 {
		return path
	}
	return strings.Join(parts[0:2], ".")
}

// Extract "playlist-job" from "bilibili.main.web-svr.playlist-job"
// nolint: unused,deadcode,gomnd
func nodeAppName(path string) (string, bool) {
	path = strings.TrimPrefix(path, "bilibili.")
	parts := strings.SplitN(path, ".", 3)
	if len(parts) < 3 {
		return "", false
	}
	return parts[len(parts)-1], true
}

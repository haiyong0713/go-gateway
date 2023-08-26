package service

import (
	"context"
	"strings"

	pb "go-gateway/app/app-svr/app-gw/management/api"
	"go-gateway/app/app-svr/app-gw/management/internal/model/sets"
)

func (s *CommonService) AuthZ(ctx context.Context, req *pb.AuthZReq) (*pb.AuthZReply, error) {
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
// nolint:gomnd
func nodeAppName(path string) (string, bool) {
	path = strings.TrimPrefix(path, "bilibili.")
	parts := strings.SplitN(path, ".", 3)
	if len(parts) < 3 {
		return "", false
	}
	return parts[len(parts)-1], true
}

func (s *CommonService) AuthZSidebar(ctx context.Context, username string) ([]string, error) {
	var privilegeSlice = make([]string, 0)
	authUsers := s.dao.GetAuthUsers()
	if authUsers == nil {
		return nil, nil
	}
	if contains(authUsers.Gateway, username) {
		privilegeSlice = append(privilegeSlice, "gateway")
	}
	if contains(authUsers.DS, username) {
		privilegeSlice = append(privilegeSlice, "ds")
	}
	return privilegeSlice, nil
}

func contains(strSlice []string, str string) bool {
	for _, temp := range strSlice {
		if temp == str {
			return true
		}
	}
	return false
}

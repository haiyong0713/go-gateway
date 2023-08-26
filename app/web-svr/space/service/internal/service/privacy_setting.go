package service

import (
	"context"

	"go-common/library/ecode"
	"go-common/library/sync/errgroup.v2"
	pb "go-gateway/app/web-svr/space/service/api"
	"go-gateway/app/web-svr/space/service/internal/model"

	"github.com/golang/protobuf/ptypes/empty"
)

func (s *Service) PrivacySetting(ctx context.Context, req *pb.PrivacySettingReq) (*pb.PrivacySettingReply, error) {
	reply := &pb.PrivacySettingReply{}
	req.PrivacyOption = s.covertPrivacyOption(req.PrivacyOption)
	var (
		res          []*model.MemberPrivacy
		livePlayback bool
	)
	g := errgroup.WithContext(ctx)
	g.Go(func(ctx context.Context) error {
		var err error
		res, err = s.dao.PrivacySetting(ctx, req)
		return err
	})
	g.Go(func(ctx context.Context) error {
		// 未设置过直播回放开关，如果在白名单内，则需要默认为关
		_, livePlayback = s.livePlaybackWhitelist[req.Mid]
		return nil
	})
	if err := g.Wait(); err != nil {
		return nil, err
	}
	reply.Settings = make(map[string]*pb.PrivacySetting, len(res))
	stateFunc := func(status int64, option pb.PrivacyOption, set bool) pb.PrivacyState { // 新用户，则返回新用户开关默认为开
		switch status {
		case 1:
			return pb.PrivacyState_opened
		case 0:
			return pb.PrivacyState_closed
		default:
		}
		switch option {
		case pb.PrivacyOption_live_playback:
			if livePlayback {
				return pb.PrivacyState_closed
			}
			return pb.PrivacyState_opened
		default:
			if set {
				return pb.PrivacyState_opened
			}
			return pb.PrivacyState_closed
		}
	}
	var set bool
	for _, r := range res { // 数据库有记录，表示用户设置过开关，其他未设置的开关默认开
		if r.ID > 0 {
			set = true
			break
		}
	}
	for _, option := range req.PrivacyOption {
		var ok bool
		for _, r := range res {
			set = set || r.NewUser // 或者用户是新用户，其他未设置的开关默认开
			if r.Privacy != option.String() {
				continue
			}
			reply.Settings[option.String()] = &pb.PrivacySetting{Option: option, State: stateFunc(r.Status, option, set)}
			ok = true
			break
		}
		if !ok {
			reply.Settings[option.String()] = &pb.PrivacySetting{Option: option, State: stateFunc(-1, option, set)}
		}
	}
	return reply, nil
}

func (s *Service) UpdatePrivacySetting(ctx context.Context, req *pb.UpdatePrivacysReq) (*empty.Empty, error) {
	for _, setting := range req.Settings {
		if setting == nil {
			return nil, ecode.Error(ecode.RequestErr, "setting is nil")
		}
		if _, ok := pb.PrivacyOption_name[int32(setting.Option)]; !ok {
			return nil, ecode.Error(ecode.RequestErr, "setting option is illegal")
		}
		if _, ok := pb.PrivacyState_name[int32(setting.State)]; !ok {
			return nil, ecode.Error(ecode.RequestErr, "setting state is illegal")
		}
	}
	if err := s.dao.UpdatePrivacySetting(ctx, req); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func (s *Service) covertPrivacyOption(option []pb.PrivacyOption) []pb.PrivacyOption {
	m := make(map[pb.PrivacyOption]struct{}, len(option))
	var res []pb.PrivacyOption
	for _, privacyOption := range option {
		if _, ok := m[privacyOption]; ok {
			continue
		}
		res = append(res, privacyOption)
		m[privacyOption] = struct{}{}
	}
	if len(res) != 0 {
		return res
	}
	for key := range pb.PrivacyOption_name {
		res = append(res, pb.PrivacyOption(key))
	}
	return res
}

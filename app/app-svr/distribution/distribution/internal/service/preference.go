package service

import (
	"context"
	"sort"
	"strings"
	"time"

	pb "go-gateway/app/app-svr/distribution/distribution/api"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/defaultvalue"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/lastmodified"
	"go-gateway/app/app-svr/distribution/distribution/internal/extension/mergepreference"
	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	"go-gateway/app/app-svr/distribution/distribution/internal/storagedriver"

	"go-common/library/log"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/jhump/protoreflect/dynamic"
)

const googleApis = "type.googleapis.com/"

func trimGoogleApis(in string) string {
	return strings.TrimPrefix(in, googleApis)
}

func (s *Service) UserPreference(ctx context.Context, req *pb.UserPreferenceReq) (*pb.UserPreferenceReply, error) {
	metas := []*preferenceproto.PreferenceMeta{}
	for _, meta := range preferenceproto.ALLPreference() {
		if meta.Disabled() {
			continue
		}
		metas = append(metas, meta)
	}
	preferences, err := s.fetchUserPreference(ctx, metas, InitialNotExistedPreference(true), WithDefaultValue(true))
	if err != nil {
		return nil, err
	}
	defer s.loggingPreferenceExp(ctx, time.Now(), preferences)
	reply := &pb.UserPreferenceReply{}
	for _, p := range preferences {
		untyped, err := ptypes.MarshalAny(p.Message)
		if err != nil {
			log.Error("Failed to marshal as any on message: %+v: %+v", p, err)
			continue
		}
		reply.Preference = append(reply.Preference, untyped)
	}
	sort.Slice(reply.Preference, func(i, j int) bool {
		return reply.Preference[i].TypeUrl < reply.Preference[j].TypeUrl
	})
	return reply, nil
}

func (s *Service) SetUserPreference(ctx context.Context, req *pb.SetUserPreferenceReq) (*pb.SetUserPreferenceReply, error) {
	now := time.Now()
	metas := map[string]*preferenceproto.PreferenceMeta{}
	metaSlice := []*preferenceproto.PreferenceMeta{}
	for _, untyped := range req.Preference {
		meta, ok := preferenceproto.TryGetPreference(trimGoogleApis(untyped.TypeUrl))
		if !ok {
			log.Warn("Failed to get preference meta: %q", untyped.TypeUrl)
			continue
		}
		metas[untyped.TypeUrl] = meta
		metaSlice = append(metaSlice, meta)
	}
	preferences, err := s.fetchUserPreference(ctx, metaSlice, InitialNotExistedPreference(true), WithDefaultValue(true))
	if err != nil {
		return nil, err
	}

	toUpdate := []*preferenceproto.Preference{}
	for _, untyped := range req.Preference {
		meta, ok := metas[untyped.TypeUrl]
		if !ok {
			continue
		}
		ctr := dynamic.NewMessage(meta.ProtoDesc)
		if err := ctr.Unmarshal(untyped.Value); err != nil {
			log.Warn("Failed to unmarshal to proto desc container: %+v", err)
			continue
		}

		func() {
			originPreference, ok := preferences[meta.ProtoDesc.GetFullyQualifiedName()]
			if !ok {
				lastmodified.SetPreferenceValueLastModified(ctr, now)
				return
			}
			lastmodified.CompareAndSetPreferenceValueLastModified(ctr, originPreference.Message, now)
			mergepreference.MergePreference(ctr, originPreference.Message)
		}()

		toUpdate = append(toUpdate, &preferenceproto.Preference{
			Meta:    *meta,
			Message: ctr,
		})
	}

	defer s.loggingPreferenceUpdate(ctx, now, toUpdate...)
	if err := storagedriver.DispatchSetUserPreference(ctx, toUpdate); err != nil {
		return nil, err
	}
	return &pb.SetUserPreferenceReply{}, nil
}

func (s *Service) GetUserPreference(ctx context.Context, req *pb.GetUserPreferenceReq) (*pb.GetUserPreferenceReply, error) {
	metas := []*preferenceproto.PreferenceMeta{}
	for _, typeURL := range req.TypeUrl {
		meta, ok := preferenceproto.TryGetPreference(trimGoogleApis(typeURL))
		if !ok {
			log.Warn("Failed to get preference meta: %q", typeURL)
			continue
		}
		metas = append(metas, meta)
	}

	preferences, err := s.fetchUserPreference(ctx, metas, InitialNotExistedPreference(true), WithDefaultValue(true))
	if err != nil {
		return nil, err
	}
	reply := &pb.GetUserPreferenceReply{}
	for _, p := range preferences {
		untyped, err := ptypes.MarshalAny(p.Message)
		if err != nil {
			log.Error("Failed to marshal as any on message: %+v: %+v", p, err)
			continue
		}
		reply.Value = append(reply.Value, untyped)
	}
	sort.Slice(reply.Value, func(i, j int) bool {
		return reply.Value[i].TypeUrl < reply.Value[j].TypeUrl
	})
	return reply, nil
}

func (s *Service) fetchUserPreference(ctx context.Context, metas []*preferenceproto.PreferenceMeta, opts ...FetchOption) (map[string]*preferenceproto.Preference, error) {
	cfg := &fetchConfig{}
	cfg.Apply(opts...)

	preferences, err := storagedriver.DispatchGetUserPreference(ctx, metas)
	if err != nil {
		return nil, err
	}

	if cfg.initialNotExistedPreference {
		for _, meta := range metas {
			if _, ok := preferences[meta.ProtoDesc.GetFullyQualifiedName()]; ok {
				continue
			}
			preferences[meta.ProtoDesc.GetFullyQualifiedName()] = &preferenceproto.Preference{
				Meta:    *meta,
				Message: dynamic.NewMessage(meta.ProtoDesc),
			}
		}
	}

	if cfg.withDefaultValue {
		for _, preference := range preferences {
			defaultvalue.InitializeWithDefaultValue(preference.Message)
		}
	}

	return preferences, nil
}

type fetchConfig struct {
	initialNotExistedPreference bool
	withDefaultValue            bool
}

func (f *fetchConfig) Apply(opts ...FetchOption) {
	for _, opt := range opts {
		opt(f)
	}
}

type FetchOption func(*fetchConfig)

func InitialNotExistedPreference(initial bool) FetchOption {
	return func(cfg *fetchConfig) {
		cfg.initialNotExistedPreference = initial
	}
}

func WithDefaultValue(with bool) FetchOption {
	return func(cfg *fetchConfig) {
		cfg.withDefaultValue = with
	}
}

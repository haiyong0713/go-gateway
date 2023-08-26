package service

import (
	"context"
	"time"

	"go-gateway/app/app-svr/distribution/distribution/internal/preferenceproto"
	"go-gateway/app/app-svr/distribution/distribution/internal/sessioncontext"

	"go-common/library/log"
	"go-common/library/log/infoc.v2"

	"github.com/golang/protobuf/jsonpb"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/pkg/errors"
)

func messagejsonify(in *dynamic.Message) string {
	bytes, err := in.MarshalJSON()
	if err != nil {
		return ""
	}
	return string(bytes)
}

func withallfieldsjsonify(in *desc.MessageDescriptor) string {
	marshaler := jsonpb.Marshaler{
		EmitDefaults: true,
	}
	dm := dynamic.NewMessage(in)
	jsonstring, err := marshaler.MarshalToString(dm)
	if err != nil {
		return ""
	}
	return jsonstring
}

func (s *Service) NewPreferenceLog(ctx context.Context, at time.Time, preferences ...*preferenceproto.Preference) ([]*infoc.Payload, error) {
	ssCtx, ok := sessioncontext.FromContext(ctx)
	if !ok {
		return nil, errors.Errorf("Session context is required")
	}

	payloads := make([]*infoc.Payload, 0, len(preferences))
	for _, p := range preferences {
		payload := infoc.NewLogStreamV(s.cfg.PreferenceLogID,
			log.KVString("buvid", ssCtx.Device().Buvid),
			log.KVInt64("mid", ssCtx.Mid()),
			log.KVString("ctime", at.Format(time.RFC3339)),
			log.KVString("platform", ssCtx.Device().RawPlatform),
			log.KVInt64("build", ssCtx.Device().Build),
			log.KVString("brand", ssCtx.Device().Brand),
			log.KVString("mobi_app", ssCtx.Device().RawMobiApp),
			log.KVString("preference", messagejsonify(p.Message)),
			log.KVString("name", p.Meta.ProtoDesc.GetName()),
			log.KVString("meta", withallfieldsjsonify(p.Meta.ProtoDesc)),
		)
		payloads = append(payloads, &payload)
	}
	return payloads, nil
}

func (s *Service) AsyncLogging(ctx context.Context, payload ...*infoc.Payload) {
	_ = s.fanout.Do(ctx, func(ctx context.Context) {
		for _, p := range payload {
			if err := s.infoc.Info(ctx, *p); err != nil {
				log.Error("Failed to async logging payload: %+v: %+v", p, err)
				continue
			}
		}
	})
}

func (s *Service) loggingPreferenceUpdate(ctx context.Context, at time.Time, preferences ...*preferenceproto.Preference) {
	payloads, err := s.NewPreferenceLog(ctx, at, preferences...)
	if err != nil {
		log.Error("Failed to construct preference log: %+v", err)
		return
	}
	s.AsyncLogging(ctx, payloads...)
}

func (s *Service) loggingPreferenceExp(ctx context.Context, at time.Time, preferences map[string]*preferenceproto.Preference) {
	payloads, err := s.NewPreferenceLogExp(ctx, at, preferences)
	if err != nil {
		log.Error("Failed to construct preference log: %+v", err)
		return
	}
	s.AsyncLogging(ctx, payloads...)
}

func (s *Service) NewPreferenceLogExp(ctx context.Context, at time.Time, preferences map[string]*preferenceproto.Preference) ([]*infoc.Payload, error) {
	ssCtx, ok := sessioncontext.FromContext(ctx)
	if !ok {
		return nil, errors.Errorf("Session context is required")
	}

	payloads := make([]*infoc.Payload, 0, len(preferences))
	for name, p := range preferences {
		if _, ok := expLogPreference[name]; !ok {
			continue
		}
		payload := infoc.NewLogStreamV(s.cfg.ExpConfigLogID,
			log.KVInt64("mid", ssCtx.Mid()),
			log.KVString("config_location", name),
			log.KVString("config", messagejsonify(p.Message)),
			log.KVString("mobi_app", ssCtx.Device().RawMobiApp),
			log.KVInt64("build", ssCtx.Device().Build),
			log.KVString("buvid", ssCtx.Device().Buvid),
			log.KVString("ctime", at.Format(time.RFC3339)),
			log.KVString("platform", ssCtx.Device().RawPlatform),
			log.KVString("brand", ssCtx.Device().Brand),
			log.KVString("fp_local", ssCtx.Device().FpLocal),
		)
		payloads = append(payloads, &payload)
	}
	return payloads, nil
}

var expLogPreference = map[string]struct{}{
	"bilibili.app.distribution.experimental.v1.MultipleTusConfig": {},
}

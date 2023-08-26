package dao

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"go-common/library/log"
	"go-common/library/queue/databus"
	"go-gateway/app/app-svr/app-feed/ng-clarify-job/internal/model"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

func parseSession(msg *databus.Message) (interface{}, error) {
	out := &model.IndexSession{}
	if err := json.Unmarshal(msg.Value, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func shardingSession(msg *databus.Message, data interface{}) int {
	session, ok := data.(*model.IndexSession)
	if !ok {
		return 0
	}
	return int(session.Time)
}

func (d *dao) processSession(msg []interface{}) {
	for _, m := range msg {
		session, ok := m.(*model.IndexSession)
		if !ok {
			continue
		}
		if session.ID == "" {
			continue
		}
		log.Info("processing session message: %+v", session)
		if err := d.sessionRecorder.saveSession(session); err != nil {
			log.Error("Failed to save session: %+v", err)
			continue
		}
	}
}

func (sr *sessionRecorder) saveSession(session *model.IndexSession) error {
	raw, err := json.Marshal(session)
	if err != nil {
		return err
	}
	sr.Lock()
	defer sr.Unlock()
	sr.buf.Write(raw)
	sr.buf.WriteByte('\n')
	log.Info("succeed to record session to buffer, buffer len: %d and max buffer size: %d", sr.buf.Len(), sr.maxBufferSize)
	if sr.buf.Len() >= sr.maxBufferSize {
		return sr.doArchive()
	}
	return nil
}

// nolint:unparam
func gzipBytes(name string, in []byte) []byte {
	gzipBuf := &bytes.Buffer{}
	zw := gzip.NewWriter(gzipBuf)
	_, _ = zw.Write(in)
	_ = zw.Close()
	return gzipBuf.Bytes()
}

func (sr *sessionRecorder) doArchive() error {
	// ensure locked
	defer sr.buf.Reset()
	now := time.Now()
	log.Info("archvie session file to boss at: %d", now.UnixNano())

	name := fmt.Sprintf("%d-%s.txt", now.Unix(), uuid.New().String())
	gziped := gzipBytes(name, sr.buf.Bytes())
	path := fmt.Sprintf("/feed-session-v1/%s/%s", now.Format("2006-01-02"), name)
	if err := sr.putObject(context.Background(), path, gziped); err != nil {
		log.Error("failed to put object to boss at: %d: %+v", now.UnixNano(), err)
		return err
	}
	log.Info("succeed to put object to boss at: %d: %q", now.UnixNano(), path)
	if err := sr.saveArchvieIndex(context.Background(), now, path, sr.buf.Len(), len(gziped)); err != nil {
		log.Error("failed to save archive object index at: %d: %+v", now.UnixNano(), err)
		return err
	}
	log.Info("succeed to save archive object index at: %d: %q", now.UnixNano(), path)
	return nil
}

func (sr *sessionRecorder) putObject(ctx context.Context, path string, payload []byte) error {
	body := bytes.NewReader(payload)
	req := &s3.PutObjectInput{
		Bucket:        aws.String(sr.bucket),
		Key:           aws.String(path),
		Body:          body,
		ContentLength: aws.Int64(int64(len(payload))),
		ContentType:   aws.String(http.DetectContentType(payload)),
	}
	if _, err := sr.s3.PutObjectWithContext(ctx, req); err != nil {
		return errors.WithStack(err)
	}
	return nil
}

type sessionRecorder struct {
	sync.Mutex
	buf bytes.Buffer

	taishan       *Taishan
	s3            *s3.S3
	bucket        string
	maxBufferSize int
}

func newSessionRecorder(taishan *Taishan, s3 *s3.S3, bucket string, maxBufferSize int) *sessionRecorder {
	return &sessionRecorder{
		buf:           bytes.Buffer{},
		taishan:       taishan,
		s3:            s3,
		bucket:        bucket,
		maxBufferSize: maxBufferSize,
	}
}

// nolint:unused
func (sr *sessionRecorder) headObject(ctx context.Context, path string) (*s3.HeadObjectOutput, error) {
	req := &s3.HeadObjectInput{
		Bucket: aws.String(sr.bucket),
		Key:    aws.String(path),
	}
	reply, err := sr.s3.HeadObjectWithContext(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return reply, nil
}

func (sr *sessionRecorder) presignedURL(path string, duration time.Duration) (string, error) {
	req, _ := sr.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(sr.bucket),
		Key:    aws.String(path),
	})
	urlStr, err := req.Presign(duration)
	if err != nil {
		return "", errors.WithStack(err)
	}
	return urlStr, nil
}

func archiveIndexKey(at time.Time, path string) string {
	return fmt.Sprintf("{feed_session_%s}/%d/archive/%s", at.Format("2006-01-02"), math.MaxInt64-at.Unix(), strings.TrimLeft(path, "/"))
}

func (sr *sessionRecorder) saveArchvieIndex(ctx context.Context, at time.Time, path string, rawSize int, gzipedSize int) error {
	key := archiveIndexKey(at, path)
	value := &model.ArchiveIndex{
		Path:       path,
		CreatedAt:  at.UnixNano(),
		RawSize:    int64(rawSize),
		GzipedSize: int64(gzipedSize),
	}
	valBytes, _ := json.Marshal(value)
	req := sr.taishan.NewPutReq([]byte(key), valBytes, 0)
	return sr.taishan.Put(ctx, req)
}

func (sr *sessionRecorder) scanArchvieIndex(ctx context.Context, startTS, endTS int64, lastKey string) (*model.ScanArchiveIndexReply, error) {
	startTime := time.Unix(startTS, 0)
	endTime := time.Unix(endTS, 0)
	start := []byte(fmt.Sprintf("{feed_session_%s}/%d/archive/", startTime.Format("2006-01-02"), math.MaxInt64-startTS))
	end := []byte(fmt.Sprintf("{feed_session_%s}/%d/archive/", endTime.Format("2006-01-02"), math.MaxInt64-endTS))
	if lastKey != "" {
		start = append([]byte(lastKey), 0x00)
	}
	req := sr.taishan.NewScanReq(start, end, 100)
	out := []*model.ArchiveIndex{}
	reply, err := sr.taishan.Scan(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, r := range reply.Records {
		ai := &model.ArchiveIndex{}
		if err := json.Unmarshal(r.Columns[0].Value, ai); err != nil {
			log.Error("Failed to unmarshal archive index: %+v", errors.WithStack(err))
			continue
		}
		out = append(out, ai)
	}
	indexReply := &model.ScanArchiveIndexReply{
		Index:   out,
		NextKey: string(reply.NextKey),
		HasNext: reply.HasNext,
	}
	return indexReply, nil
}

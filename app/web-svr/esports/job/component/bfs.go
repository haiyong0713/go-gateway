package component

import (
	"context"
	"io/ioutil"
	"net/http"
	"time"

	"go-common/library/database/bfs"
)

const (
	BFSBucket              = "esport"
	BFSDir4Default         = "default"
	BFSDir4S10             = "LOL/S10"
	BFSDir4S10CountryImage = "LOL/S10/country"
	BFSDir4S10TeamImage    = "LOL/S10/team"
	BFSDir4S10PlayerImage  = "LOL/S10/player"
	BFSDir4S10HeroImage    = "LOL/S10/hero"
	bfsContentTypeOfJson   = "application/json"

	UploadType4ImageOfCountry = iota
	UploadType4ImageOfTeam
	UploadType4ImageOfPlayer
	UploadType4ImageOfHero
)

var (
	bfsClient         *bfs.BFS
	defaultHttpClient http.Client
)

func init() {
	bfsClient = bfs.New(nil)
	defaultHttpClient = http.Client{
		Timeout: 5 * time.Second,
	}
}

func UploadBFSResource(ctx context.Context, req *bfs.Request) (string, error) {
	return bfsClient.Upload(ctx, req)
}

func UploadBFSImageResourceByUrl(ctx context.Context, url, filename string, uploadType int) (string, error) {
	if url == "" {
		return "", nil
	}

	resp, err := defaultHttpClient.Get(url)
	if err != nil {
		return "", err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	bs := make([]byte, 0)
	if bs, err = ioutil.ReadAll(resp.Body); err != nil {
		return "", err
	}

	req := &bfs.Request{
		Bucket:      BFSBucket,
		Dir:         BFSDir4Default,
		Filename:    filename,
		ContentType: http.DetectContentType(bs),
		File:        bs,
	}

	switch uploadType {
	case UploadType4ImageOfTeam:
		req.Dir = BFSDir4S10TeamImage
	case UploadType4ImageOfPlayer:
		req.Dir = BFSDir4S10PlayerImage
	case UploadType4ImageOfHero:
		req.Dir = BFSDir4S10HeroImage
	case UploadType4ImageOfCountry:
		req.Dir = BFSDir4S10CountryImage
	}

	return UploadBFSResource(ctx, req)
}

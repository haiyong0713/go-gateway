package appstoreconnect

import (
	"net/http"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// AppStoreVersionsService Manage versions of your app that are available in App Store.
type AppStoreVersionsService struct {
	// HTTP client used to communicate with the API.
	client *Client
}

const appStoreVersionServicePath = "appStoreVersions"

// Build Get the build that is attached to a specific App Store version.
func (s *AppStoreVersionsService) Build(appKey string, appStoreVerID string) (*BuildResponse, *http.Response, error) {
	u := appStoreVersionServicePath + "/" + appStoreVerID + "/build"
	req, err := s.client.NewRequest(appKey, http.MethodGet, u, nil)
	if err != nil {
		log.Error("Build: %v", err)
		return nil, nil, err
	}
	r := &BuildResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("Build: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

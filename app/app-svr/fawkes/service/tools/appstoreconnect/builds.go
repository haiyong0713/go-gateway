package appstoreconnect

import (
	"net/http"
	"net/url"
	"time"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// BuildsService handles communication with review submission methods of the AppStoreConenct API.
type BuildsService struct {
	// HTTP client used to communicate with the API.
	client *Client
}

const buildsServicePath = "builds"

// BuildsResponse ...
type BuildsResponse struct {
	Builds []Build            `json:"data,omitempty"`
	Links  PagedDocumentLinks `json:"links,omitempty"`
	Meta   PagingInformation  `json:"meta,omitempty"`
}

// BuildResponse ...
type BuildResponse struct {
	Build Build         `json:"data,omitempty"`
	Links DocumentLinks `json:"links,omitempty"`
}

// BuildsOptions ...
type BuildsOptions struct {
	AppFilter             string `url:"filter[app],omitempty"`
	ExpiredFilter         string `url:"filter[expired],omitempty"`
	IDFilter              string `url:"filter[id],omitempty"`
	ProcessingStateFilter string `url:"filter[processingState],omitempty"`
	VersionFilter         string `url:"filter[version],omitempty"`
	Limit                 int    `url:"limit,omitempty"`
	Cursor                string `url:"cursor,omitempty"`
	Next                  string `url:"-"`
}

// BuildAttributes ...
type BuildAttributes struct {
	Expired                 bool       `json:"expired,omitempty"`
	IconAssetToken          ImageAsset `json:"iconAssetToken,omitempty"`
	MinOsVersion            string     `json:"minOsVersion,omitempty"`
	ProcessingState         string     `json:"processingState,omitempty"`
	Version                 string     `json:"version,omitempty"`
	UsesNonExemptEncryption bool       `json:"usesNonExemptEncryption,omitempty"`
	UploadedDate            time.Time  `json:"uploadedDate,omitempty"`
	ExpirationDate          time.Time  `json:"expirationDate,omitempty"`
}

// BuildRelationships ...
type BuildRelationships struct {
	App                      Relationship `json:"app,omitempty"`
	AppEncryptionDeclaration Relationship `json:"appEncryptionDeclaration,omitempty"`
	IndividualTesters        Relationship `json:"individualTesters ,omitempty"`
	PreReleaseVersion        Relationship `json:"preReleaseVersion ,omitempty"`
	BetaBuildLocalizations   Relationship `json:"betaBuildLocalizations ,omitempty"`
	BetaGroups               Relationship `json:"betaGroups ,omitempty"`
	BuildBetaDetail          Relationship `json:"buildBetaDetail ,omitempty"`
	BetaAppReviewSubmission  Relationship `json:"betaAppReviewSubmission ,omitempty"`
}

// Build ...
type Build struct {
	Attributes    BuildAttributes    `json:"attributes,omitempty"`
	ID            string             `json:"id,omitempty"`
	Relationships BuildRelationships `json:"relationships,omitempty"`
	Type          string             `json:"type,omitempty"`
	Links         ResourceLinks      `json:"links,omitempty"`
}

// BuildUpdateRequest ...
type BuildUpdateRequest struct {
	Data BuildUpdateRequestData `json:"data,omitempty"`
}

// BuildUpdateRequestData ...
type BuildUpdateRequestData struct {
	Attributes BuildUpdateAttributes `json:"attributes,omitempty"`
	ID         string                `json:"id,omitempty"`
	Type       string                `json:"type,omitempty"`
}

// BuildUpdateAttributes ...
type BuildUpdateAttributes struct {
	Expired                 bool `json:"expired"`
	UsesNonExemptEncryption bool `json:"usesNonExemptEncryption"`
}

// BuildUpdateRelationships ...
type BuildUpdateRelationships struct {
	AppEncryptionDeclaration Relationship `json:"appEncryptionDeclaration,omitempty"`
}

// Builds list builds of an app
func (s *BuildsService) Builds(appKey string, opt *BuildsOptions) (*BuildsResponse, *http.Response, error) {
	if opt != nil && opt.Next != "" {
		u, err := url.Parse(opt.Next)
		if err != nil {
			log.Error("Builds: %v", err)
			return nil, nil, err
		}
		cursor := u.Query().Get("cursor")
		opt.Cursor = cursor
	}
	u := buildsServicePath
	u, err := addOptions(u, opt)
	if err != nil {
		log.Error("Builds: %v", err)
		return nil, nil, err
	}
	req, err := s.client.NewRequest(appKey, http.MethodGet, u, nil)
	if err != nil {
		log.Error("Builds: %v", err)
		return nil, nil, err
	}
	r := &BuildsResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("Builds: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

// ModifyBuild Expire a build or change its encryption exemption setting.
func (s *BuildsService) ModifyBuild(appKey string, buildID string, expired bool, useNonExemptEncryption bool) (*BuildResponse, *http.Response, error) {
	u := buildsServicePath + "/" + buildID
	body := BuildUpdateRequest{
		Data: BuildUpdateRequestData{
			Attributes: BuildUpdateAttributes{
				Expired:                 expired,
				UsesNonExemptEncryption: useNonExemptEncryption,
			},
			ID:   buildID,
			Type: "builds",
		},
	}
	req, err := s.client.NewRequest(appKey, http.MethodPatch, u, body)
	if err != nil {
		log.Error("ModifyBuild: %v", err)
		return nil, nil, err
	}
	r := &BuildResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("ModifyBuild: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

// BetaAppReviewSubmission Get the beta app review submission status for a specific build.
func (s *BuildsService) BetaAppReviewSubmission(appKey, buildID string) (*BetaAppReviewSubmissionResponse, *http.Response, error) {
	u := buildsServicePath + "/" + buildID + "/betaAppReviewSubmission"
	req, err := s.client.NewRequest(appKey, http.MethodGet, u, nil)
	if err != nil {
		log.Error("BetaAppReviewSubmission: %v", err)
		return nil, nil, err
	}
	r := &BetaAppReviewSubmissionResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("Builds: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

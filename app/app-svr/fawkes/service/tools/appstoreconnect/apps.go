package appstoreconnect

import (
	"net/http"
	"net/url"
	"time"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// AppsService handles communication with apps methods of the AppStoreConenct API.
type AppsService struct {
	// HTTP client used to communicate with the API.
	client *Client
}

const appsServicePath = "apps"

// AppResponse ...
type AppResponse struct {
	App   App           `json:"data,omitempty"`
	Links DocumentLinks `json:"links,omitempty"`
}

// AppsResponse ...
type AppsResponse struct {
	Apps  []App              `json:"data,omitempty"`
	Links PagedDocumentLinks `json:"links,omitempty"`
	Meta  PagingInformation  `json:"meta,omitempty"`
}

// BetaGroupsResponse ...
type BetaGroupsResponse struct {
	BetaGroups []BetaGroup        `json:"data,omitempty"`
	Links      PagedDocumentLinks `json:"links,omitempty"`
	Meta       PagingInformation  `json:"meta,omitempty"`
}

// AppsOptions ...
type AppsOptions struct {
	BundleIDFilter string `url:"filter[bundleId],omitempty"`
	IDFilter       string `url:"filter[id],omitempty"`
	NameFilter     string `url:"filter[name],omitempty"`
	SKUFilter      string `url:"filter[sku],omitempty"`

	Limit  int    `url:"limit,omitempty"`
	Cursor string `url:"cursor,omitempty"`
	Next   string `url:"-"`
}

// AppRelationships ...
type AppRelationships struct {
	BetaLicenseAgreement Relationship `json:"betaLicenseAgreement,omitempty"`
	PreReleaseVersions   Relationship `json:"preReleaseVersions,omitempty"`
	BetaAppLocalizations Relationship `json:"betaAppLocalizations,omitempty"`
	BetaGroups           Relationship `json:"betaGroups,omitempty"`
	BetaTesters          Relationship `json:"betaTesters,omitempty"`
	Builds               Relationship `json:"builds,omitempty"`
	BetaAppReviewDetail  Relationship `json:"betaAppReviewDetail,omitempty"`
}

// AppAttributes ...
type AppAttributes struct {
	BundleID      string `json:"bundleId,omitempty"`
	Name          string `json:"name,omitempty"`
	PrimaryLocale string `json:"primaryLocale,omitempty"`
	Sku           string `json:"sku,omitempty"`
}

// App ...
type App struct {
	Attributes    AppAttributes    `json:"attributes,omitempty"`
	ID            string           `json:"id,omitempty"`
	Type          string           `json:"type,omitempty"`
	Relationships AppRelationships `json:"relationships,omitempty"`
	Links         ResourceLinks    `json:"links,omitempty"`
}

// AppStoreVersionsOption ...
type AppStoreVersionsOption struct {
	IDFilter            string `url:"filter[id],omitempty"`
	VersionStringFilter string `url:"filter[versionString],omitempty"`
	PlatformFilter      string `url:"filter[platform],omitempty"`
	AppStoreStateFilter string `url:"filter[appStoreState],omitempty"`

	Limit  int    `url:"limit,omitempty"`
	Cursor string `url:"cursor,omitempty"`
	Next   string `url:"-"`
}

// AppStoreVersionAttributes ...
type AppStoreVersionAttributes struct {
	Platform            string    `json:"platform,omitempty"`
	AppStoreState       string    `json:"appStoreState,omitempty"`
	Copyright           string    `json:"copyright,omitempty"`
	EarliestReleaseDate time.Time `json:"earliestReleaseDate,omitempty"`
	ReleaseType         string    `json:"releaseType,omitempty"`
	UsesIdfa            bool      `json:"usesIdfa,omitempty"`
	VersionString       string    `json:"versionString,omitempty"`
	CreatedDate         time.Time `json:"createdDate,omitempty"`
	Downloadable        bool      `json:"downloadable,omitempty"`
}

// AppStoreVersionRelationships ...
type AppStoreVersionRelationships struct {
	App                          Relationship `json:"app,omitempty"`
	AgeRatingDeclaration         Relationship `json:"ageRatingDeclaration,omitempty"`
	AppStoreReviewDetail         Relationship `json:"appStoreReviewDetail,omitempty"`
	AppStoreVersionLocalizations Relationship `json:"appStoreVersionLocalizations,omitempty"`
	AppStoreVersionPhasedRelease Relationship `json:"appStoreVersionPhasedRelease,omitempty"`
	AppStoreVersionSubmission    Relationship `json:"appStoreVersionSubmission,omitempty"`
	Build                        Relationship `json:"build,omitempty"`
	IdfaDeclaration              Relationship `json:"idfaDeclaration,omitempty"`
	RoutingAppCoverage           Relationship `json:"routingAppCoverage,omitempty"`
}

// AppStoreVersion ...
type AppStoreVersion struct {
	Attributes    AppStoreVersionAttributes    `json:"attributes,omitempty"`
	ID            string                       `json:"id,omitempty"`
	Type          string                       `json:"type,omitempty"`
	RelationShips AppStoreVersionRelationships `json:"relationships,omitempty"`
	Links         ResourceLinks                `json:"links,omitempty"`
}

// AppStoreVersionsResponse ...
type AppStoreVersionsResponse struct {
	AppStoreVersion []*AppStoreVersion `json:"data,omitempty"`
	Links           PagedDocumentLinks `json:"links,omitempty"`
	Meta            PagingInformation  `json:"meta,omitempty"`
}

// Apps Find and list apps added in App Store Connect.
func (s *AppsService) Apps(appKey string, opt *AppsOptions) (*AppsResponse, *http.Response, error) {
	if opt != nil && opt.Next != "" {
		u, err := url.Parse(opt.Next)
		if err != nil {
			log.Error("Apps: %v", err)
			return nil, nil, err
		}
		cursor := u.Query().Get("cursor")
		opt.Cursor = cursor
	}
	u := appsServicePath
	u, err := addOptions(u, opt)
	if err != nil {
		log.Error("Apps: %v", err)
		return nil, nil, err
	}
	req, err := s.client.NewRequest(appKey, http.MethodGet, u, nil)
	if err != nil {
		log.Error("Apps: %v", err)
		return nil, nil, err
	}
	r := &AppsResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("Apps: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

// AppInfo Get information about a specific app.
func (s *AppsService) AppInfo(appKey, ID string, opt *AppsOptions) (*AppResponse, *http.Response, error) {
	if opt != nil && opt.Next != "" {
		u, err := url.Parse(opt.Next)
		if err != nil {
			log.Error("AppInfo: %v", err)
			return nil, nil, err
		}
		cursor := u.Query().Get("cursor")
		opt.Cursor = cursor
	}
	u := appsServicePath + "/" + ID
	u, err := addOptions(u, opt)
	if err != nil {
		log.Error("AppInfo: %v", err)
		return nil, nil, err
	}
	req, err := s.client.NewRequest(appKey, http.MethodGet, u, nil)
	if err != nil {
		log.Error("AppInfo: %v", err)
		return nil, nil, err
	}
	r := &AppResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("AppInfo: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

// AppStoreVersions list app store versions for an app.
func (s *AppsService) AppStoreVersions(appKey, ID string, opt *AppStoreVersionsOption) (*AppStoreVersionsResponse, *http.Response, error) {
	if opt != nil && opt.Next != "" {
		u, err := url.Parse(opt.Next)
		if err != nil {
			log.Error("AppStoreVersions: %v", err)
			return nil, nil, err
		}
		cursor := u.Query().Get("cursor")
		opt.Cursor = cursor
	}
	u := appsServicePath + "/" + ID + "/appStoreVersions"
	u, err := addOptions(u, opt)
	if err != nil {
		log.Error("AppStoreVersions: %v", err)
		return nil, nil, err
	}
	req, err := s.client.NewRequest(appKey, http.MethodGet, u, nil)
	if err != nil {
		log.Error("AppStoreVersions: %v", err)
		return nil, nil, err
	}
	r := &AppStoreVersionsResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("AppStoreVersions: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

// BetaGroups Get a list of beta groups associated with a specific app.
func (s *AppsService) BetaGroups(appKey, ID string) (*BetaGroupsResponse, *http.Response, error) {
	u := appsServicePath + "/" + ID + "/betaGroups"
	req, err := s.client.NewRequest(appKey, http.MethodGet, u, nil)
	if err != nil {
		log.Error("BetaGrops: %v", err)
		return nil, nil, err
	}
	r := &BetaGroupsResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("BetaGrops: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

// RemoveBetaTesters Remove one or more beta testers' access to test any builds of a specific app.
func (s *AppsService) RemoveBetaTesters(appKey, ID string, betaTestersID []string) (*http.Response, error) {
	body := &LinkagesRequest{}
	for _, testerID := range betaTestersID {
		newData := &RelationshipData{
			ID:   testerID,
			Type: "betaTesters",
		}
		body.Data = append(body.Data, *newData)
	}
	u := appsServicePath + "/" + ID + "/relationships/betaTesters"
	req, err := s.client.NewRequest(appKey, http.MethodDelete, u, body)
	if err != nil {
		log.Error("RemoveBetaTesters: %v", err)
		return nil, err
	}
	resp, err := s.client.Do(req, nil)
	if err != nil {
		log.Error("RemoveBetaTesters: %v", err)
	}
	return resp, err
}

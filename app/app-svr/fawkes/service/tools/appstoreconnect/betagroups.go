package appstoreconnect

import (
	"net/http"
	"time"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// BetaGroupsService handles communication with review submission methods of the AppStoreConenct API.
type BetaGroupsService struct {
	// HTTP client used to communicate with the API.
	client *Client
}

const betaGroupsServicePath = "betaGroups"

// BetaGropAttributes ...
type BetaGropAttributes struct {
	IsInternalGroup        bool      `json:"isInternalGroup,omitempty"`
	Name                   string    `json:"name,omitempty"`
	PublicLink             string    `json:"publicLink,omitempty"`
	PublicLinkEnabled      bool      `json:"publicLinkEnabled,omitempty"`
	PublicLinkID           string    `json:"publicLinkId,omitempty"`
	PublicLinkLimit        int       `json:"publicLinkLimit,omitempty"`
	PublicLinkLimitEnabled bool      `json:"publicLinkLimitEnabled,omitempty"`
	CreatedDate            time.Time `json:"createdDate,omitempty"`
	FeedbackEnabled        bool      `json:"feedbackEnabled,omitempty"`
}

// BetaGropRelationships ...
type BetaGropRelationships struct {
	App         Relationship    `json:"app,omitempty"`
	BetaTesters RelationshipArr `json:"betaTesters,omitempty"`
	Builds      RelationshipArr `json:"builds,omitempty"`
}

// BetaGroup ...
type BetaGroup struct {
	Attributes    BetaGropAttributes    `json:"attributes,omitempty"`
	ID            string                `json:"id,omitempty"`
	Relationships BetaGropRelationships `json:"relationships,omitempty"`
	Type          string                `json:"type,omitempty"`
	Links         ResourceLinks         `json:"links,omitempty"`
}

// BetaGroupResponse ...
type BetaGroupResponse struct {
	BetaGroup BetaGroup     `json:"data,omitempty"`
	Links     DocumentLinks `json:"links,omitempty"`
}

// BetaGroupCreateAttributes ...
type BetaGroupCreateAttributes struct {
	Name                   string `json:"name,omitempty"`
	PublicLinkEnabled      bool   `json:"publicLinkEnabled,omitempty"`
	PublicLinkLimit        int    `json:"publicLinkLimit,omitempty"`
	PublicLinkLimitEnabled bool   `json:"publicLinkLimitEnabled,omitempty"`
	FeedbackEnabled        bool   `json:"feedbackEnabled,omitempty"`
}

// BetaGropCreateRelationships ...
type BetaGropCreateRelationships struct {
	App Relationship `json:"app,omitempty"`
}

// BetaGroupCreateRequestData ...
type BetaGroupCreateRequestData struct {
	Attributes    BetaGroupCreateAttributes   `json:"attributes,omitempty"`
	Relationships BetaGropCreateRelationships `json:"relationships,omitempty"`
	Type          string                      `json:"type,omitempty"`
}

// BetaGroupCreateRequest ...
type BetaGroupCreateRequest struct {
	Data BetaGroupCreateRequestData `json:"data,omitempty"`
}

// BetaTesterRelationships ...
type BetaTesterRelationships struct {
	Apps       RelationshipArr `json:"apps,omitempty"`
	BetaGroups RelationshipArr `json:"betaGroups,omitempty"`
	Builds     RelationshipArr `json:"builds,omitempty"`
}

// BetaTesterAttributes ...
type BetaTesterAttributes struct {
	Email      string `json:"email,omitempty"`
	Firstname  string `json:"firstName,omitempty"`
	InviteType string `json:"inviteType,omitempty"`
	LastName   string `json:"lastName,omitempty"`
}

// BetaTester ...
type BetaTester struct {
	Attributes    BetaTesterAttributes    `json:"attributes,omitempty"`
	ID            string                  `json:"id,omitempty"`
	Relationships BetaTesterRelationships `json:"relationships,omitempty"`
	Type          string                  `json:"type,omitempty"`
	Links         ResourceLinks           `json:"links,omitempty"`
}

// BetaTestersResponse ...
type BetaTestersResponse struct {
	BetaTesters []BetaTester       `json:"data,omitempty"`
	Links       PagedDocumentLinks `json:"links,omitempty"`
	Meta        PagingInformation  `json:"meta,omitempty"`
}

// CreateBetaGroup Create a beta group associated with an app, optionally enabling TestFlight public links.
func (s *BetaGroupsService) CreateBetaGroup(appKey, ID string, name string) (*BetaGroupResponse, *http.Response, error) {
	u := betaGroupsServicePath
	body := BetaGroupCreateRequest{
		Data: BetaGroupCreateRequestData{
			Attributes: BetaGroupCreateAttributes{
				Name:                   name,
				PublicLinkEnabled:      true,
				PublicLinkLimit:        10000,
				PublicLinkLimitEnabled: true,
				FeedbackEnabled:        true,
			},
			Relationships: BetaGropCreateRelationships{
				App: Relationship{
					Data: RelationshipData{
						ID:   ID,
						Type: "apps",
					},
				},
			},
			Type: "betaGroups",
		},
	}
	req, err := s.client.NewRequest(appKey, http.MethodPost, u, body)
	if err != nil {
		log.Error("CreateBetaGroup: %v", err)
		return nil, nil, err
	}
	r := &BetaGroupResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("CreateBetaGroup: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

// BetaTesters Get a list of beta testers contained in a specific beta group.
func (s *BetaGroupsService) BetaTesters(appKey, groupID string) (*BetaTestersResponse, *http.Response, error) {
	u := betaGroupsServicePath + "/" + groupID + "/betaTesters?limit=200"
	req, err := s.client.NewRequest(appKey, http.MethodGet, u, nil)
	if err != nil {
		log.Error("ListBetaTesters: %v", err)
		return nil, nil, err
	}
	r := &BetaTestersResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("ListBetaTesters: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

// RemoveBetaTesters Remove a specific beta tester from a one or more beta groups, revoking their access to test builds associated with those groups.
func (s *BetaGroupsService) RemoveBetaTesters(appKey, groupID string, betaTestersID []string) (*http.Response, error) {
	body := &LinkagesRequest{}
	for _, testerID := range betaTestersID {
		newData := &RelationshipData{
			ID:   testerID,
			Type: "betaTesters",
		}
		body.Data = append(body.Data, *newData)
	}
	u := betaGroupsServicePath + "/" + groupID + "/relationships/betaTesters"
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

// AddBuilds Associate builds with a beta group to enable the group to test the builds.
func (s *BetaGroupsService) AddBuilds(appKey, groupID string, buildsID []string) (*http.Response, error) {
	body := &LinkagesRequest{}
	for _, buildID := range buildsID {
		newData := &RelationshipData{
			ID:   buildID,
			Type: "builds",
		}
		body.Data = append(body.Data, *newData)
	}
	u := betaGroupsServicePath + "/" + groupID + "/relationships/builds"
	req, err := s.client.NewRequest(appKey, http.MethodPost, u, body)
	if err != nil {
		log.Error("AddBuilds: %v", err)
		return nil, err
	}
	resp, err := s.client.Do(req, nil)
	if err != nil {
		log.Error("AddBuilds: %v", err)
	}
	return resp, err
}

// RemoveBuilds Remove access to test one or more builds from beta testers in a specific beta group.
func (s *BetaGroupsService) RemoveBuilds(appKey, groupID string, buildsID []string) (*http.Response, error) {
	body := &LinkagesRequest{}
	for _, buildID := range buildsID {
		newData := &RelationshipData{
			ID:   buildID,
			Type: "builds",
		}
		body.Data = append(body.Data, *newData)
	}
	u := betaGroupsServicePath + "/" + groupID + "/relationships/builds"
	req, err := s.client.NewRequest(appKey, http.MethodDelete, u, body)
	if err != nil {
		log.Error("RemoveBuilds: %v", err)
		return nil, err
	}
	resp, err := s.client.Do(req, nil)
	if err != nil {
		log.Error("RemoveBuilds: %v", err)
	}
	return resp, err
}

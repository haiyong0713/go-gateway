package appstoreconnect

import (
	"net/http"

	log "go-gateway/app/app-svr/fawkes/service/tools/logger"
)

// SubmissionsService The submissions of builds for beta app review, including the status of submissions.
type SubmissionsService struct {
	// HTTP client used to communicate with the API.
	client *Client
}

const submissionsServicePath = "betaAppReviewSubmissions"

// BetaAppReviewSubmissionAttributes ...
type BetaAppReviewSubmissionAttributes struct {
	BetaReviewState string `json:"betaReviewState,omitempty"`
}

// BetaAppReviewSubmissionRelationships ...
type BetaAppReviewSubmissionRelationships struct {
	Build Relationship `json:"build,omitempty"`
}

// BetaAppReviewSubmission ...
type BetaAppReviewSubmission struct {
	Attributes    BetaAppReviewSubmissionAttributes    `json:"attributes,omitempty"`
	ID            string                               `json:"id,omitempty"`
	Relationships BetaAppReviewSubmissionRelationships `json:"relationships,omitempty"`
	Type          string                               `json:"type,omitempty"`
	Links         ResourceLinks                        `json:"links,omitempty"`
}

// SubmissionCreateRequestRelationships ...
type SubmissionCreateRequestRelationships struct {
	Build Relationship `json:"build,omitempty"`
}

// SubmissionCreateRequestData ...
type SubmissionCreateRequestData struct {
	Relationships SubmissionCreateRequestRelationships `json:"relationships,omitempty"`
	Type          string                               `json:"type,omitempty"`
}

// BetaAppReviewSubmissionCreateRequest ...
type BetaAppReviewSubmissionCreateRequest struct {
	Data SubmissionCreateRequestData `json:"data,omitempty"`
}

// BetaAppReviewSubmissionResponse ...
type BetaAppReviewSubmissionResponse struct {
	Data  BetaAppReviewSubmission `json:"data,omitempty"`
	Links DocumentLinks           `json:"links,omitempty"`
}

// SubmitForBetaReview Submit an app for beta app review to allow external testing.
func (s *SubmissionsService) SubmitForBetaReview(appKey, buildID string) (*BetaAppReviewSubmissionResponse, *http.Response, error) {
	u := submissionsServicePath
	body := BetaAppReviewSubmissionCreateRequest{
		Data: SubmissionCreateRequestData{
			Relationships: SubmissionCreateRequestRelationships{
				Build: Relationship{
					Data: RelationshipData{
						ID:   buildID,
						Type: "builds",
					},
				},
			},
			Type: "betaAppReviewSubmissions",
		},
	}
	req, err := s.client.NewRequest(appKey, http.MethodPost, u, body)
	if err != nil {
		log.Error("SubmitForBetaReview: %v", err)
		return nil, nil, err
	}
	r := &BetaAppReviewSubmissionResponse{}
	resp, err := s.client.Do(req, r)
	if err != nil {
		log.Error("SubmitForBetaReview: %v", err)
		return nil, resp, err
	}
	return r, resp, err
}

package metadata

// ClientInfo wraps immutable data from the client.Client structure.
type ClientInfo struct {
	AppID      string
	Endpoint   string
	Timeout    int64
	MaxRetries *int64
}

// Identifier is
func (ci ClientInfo) Identifier() string {
	if ci.AppID != "" {
		return ci.AppID
	}
	return ci.Endpoint
}

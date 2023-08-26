package metadata

// ClientInfo wraps immutable data from the client.Client structure.
type ClientInfo struct {
	Timeout    int64
	MaxRetries *int64
}

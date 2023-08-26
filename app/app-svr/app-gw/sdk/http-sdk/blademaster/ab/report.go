package ab

// Reporter is for reporting env variables, experiment states and used flags.
type Reporter interface {
	Log(states []State, flags []string, kv []KV)
}

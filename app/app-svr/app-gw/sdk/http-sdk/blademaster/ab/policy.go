package ab

// DivertPolicy takes diversion key and returns its hash value.
type DivertPolicy func(string) uint32

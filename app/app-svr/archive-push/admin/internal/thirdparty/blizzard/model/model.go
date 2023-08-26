package model

type Config struct {
	Secret *SecretConfig
	Host   *HostConfig
}

type SecretConfig struct {
	Key string
}

type HostConfig struct {
	Host string
}

package config

type Config struct {
	Port string
	Host string
}

func NewConfig(hostFlag, portFlag string) Config {
	host := hostFlag
	port := portFlag
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "8080"
	}
	return Config{
		Host: host,
		Port: port,
	}
}

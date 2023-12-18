package config

type Config struct {
	Port string
	Host string
}

func NewConfig(hostFlag, portFlag string) Config {
	if hostFlag == "" {
		hostFlag = "http://localhost"
	}
	if portFlag == "" {
		portFlag = "8080"
	}
	return Config{
		Host: hostFlag,
		Port: portFlag,
	}
}

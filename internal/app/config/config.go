package config

type ServerConfig struct {
	ServerAddr string
}

type HandlerConfig struct {
	HandlerURL string
}

func NewHandlerConfig(urlFlag string) HandlerConfig {
	url := urlFlag
	if url == "" {
		url = "http://localhost:8080"
	}
	return HandlerConfig{
		HandlerUrl: url,
	}
}

func NewServerConfig(addrFlag string) ServerConfig {
	addr := addrFlag
	if addr == "" {
		addr = "localhost:8080"
	}
	return ServerConfig{
		ServerAddr: addr,
	}
}

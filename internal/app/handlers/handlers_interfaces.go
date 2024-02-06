package handlers

type Repository interface {
	Save(url, hash string) error
	Get(hash string) (string, error)
	Ping() error
}

type Generator interface {
	MakeHash(s string) (string, error)
}

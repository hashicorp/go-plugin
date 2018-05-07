package api

type Host interface {
	HelloHost(string) error
}

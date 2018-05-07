package api

type Extender interface {
	HelloExtension(string) error
}

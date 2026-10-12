package interfaces

type GetterSetter interface {
	Set(key, value string) error
	Get(key string) (string, error)
}

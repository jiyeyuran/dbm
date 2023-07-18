package dbm

type Adapter interface {
	Build(migration interface{}) string
	MapError(error) error
}

package cachex

// ICache ...
type ICache[T any] interface {
	Set(*T) error
	Get() (*T, bool, error)
	Update(func(*T)) error

	MustSet(*T)
	MustGet() (*T, bool)
	MustUpdate(func(*T))
}

type interCache[T any] interface {
	Set(*T) error
	Get() (*T, bool, error)
	Update(func(*T)) error
}

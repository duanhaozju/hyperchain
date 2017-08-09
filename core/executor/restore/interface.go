package restore

type Restorer interface {
	Restore(string) error
}

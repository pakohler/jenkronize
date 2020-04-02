package notifications

type Notifier interface {
	Post(string) error
}

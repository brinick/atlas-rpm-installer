package installer

type emailer interface {
	send(string, []string)
}

// Send an email upon failure
func email(e emailer) error {
	return nil
}

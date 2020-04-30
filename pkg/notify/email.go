package notify

import (
	"fmt"
	"strings"
	"time"

	"github.com/brinick/shell"
)

// NewEmail permits sending an email from the given address
// to the comma-separated to string of emails
func NewEmail(from string, to string) *email {
	// No check if addresses are valid...
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" || to == "" {
		return nil
	}

	var toEmails []string
	for _, toAddr := range strings.Split(to, ",") {
		toEmails = append(toEmails, strings.TrimSpace(toAddr))
	}

	if len(toEmails) == 0 {
		return nil
	}

	return &email{from: from, to: toEmails, exe: "/bin/mailx"}
}

// email is for sending an email
type email struct {
	from    string
	to      []string
	timeout uint
	cancel  <-chan struct{}
	exe     string
}

func (e *email) WithTimeout(secs uint) *email {
	e.timeout = secs
	return e
}

func (e *email) WithCancel(stop <-chan struct{}) *email {
	e.cancel = stop
	return e
}

func (e *email) Send(subject, body string) *shell.Result {
	cmd := fmt.Sprintf(
		"echo '%s' | %s -r %s -s '%s' %s",
		body,
		e.exe,
		e.from,
		subject,
		strings.Join(e.to, " "),
	)

	var opts []shell.Option
	if e.timeout > 0 {
		opts = append(opts, shell.Timeout(time.Duration(e.timeout)*time.Second))
	}

	if e.cancel != nil {
		opts = append(opts, shell.Cancel(e.cancel))
	}
	return shell.Run(cmd, opts...)
}

package config

import (
	"flag"
	"fmt"
	"strings"
)

type adminOpts struct {
	Email         string
	DontSendEmail bool
	SudoUser      string
}

func (o *adminOpts) flags() {
	flag.StringVar(
		&o.SudoUser,
		"admin.sudo-user",
		"cvatlasnightlies",
		"The name of the sudo user to install nightlies on CVMFS",
	)

	flag.StringVar(
		&o.Email,
		"admin.email",
		strings.Join([]string{
			"oana.boeriu@cern.ch",
			"brinick.simmons@cern.ch",
		}, ","),
		"Comma-separated list of emails to contact in case of failure",
	)

	flag.BoolVar(
		&o.DontSendEmail,
		"admin.no-email",
		false,
		"Do not send an email on failure (default: do send email)",
	)
}

func (o *adminOpts) validate() error {
	emails := strings.Split(o.Email, ",")
	for _, email := range emails {
		// Very basic check
		if !strings.Contains(email, "@") {
			return fmt.Errorf("Invalid email address %s", email)
		}
	}

	return nil
}

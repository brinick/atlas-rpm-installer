package config

import (
	"flag"
	"fmt"
	"strings"
)

// AdminOpts are options for certain meta variables
type AdminOpts struct {
	Email         string
	DontSendEmail bool
	SudoUser      string
	Monitor       bool
}

func (o *AdminOpts) flags() {
	// TODO: should this be in cvmfs opts?
	flag.StringVar(
		&o.SudoUser,
		"admin.sudo-user",
		"cvatlasnightlies",
		"The name of the sudo user to install nightlies on CVMFS",
	)

	flag.BoolVar(
		&o.Monitor,
		"admin.monitor",
		false,
		"Expose expvar monitoring variables",
	)

	flag.StringVar(
		&o.Email,
		"admin.email",
		strings.Join([]string{
			// TODO: push this into a separate input
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

func (o *AdminOpts) validate() error {
	emails := strings.Split(o.Email, ",")
	for _, email := range emails {
		// Very basic check
		if !strings.Contains(email, "@") {
			return fmt.Errorf("Invalid email address %s", email)
		}
	}

	return nil
}

func (o *AdminOpts) String() string {
	args := []string{
		"- Admin Options:",
		fmt.Sprintf("   - Sudo user = %s", o.SudoUser),
		fmt.Sprintf("   - Email recipients = %s", o.Email),
		fmt.Sprintf("   - Send email on failure: %t", !o.DontSendEmail),
	}

	return strings.Join(args, "\n")
}

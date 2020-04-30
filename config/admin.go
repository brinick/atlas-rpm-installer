package config

import (
	"flag"
	"fmt"
	"strings"
)

// AdminOpts are options for certain meta variables
type AdminOpts struct {
	EmailTo       string
	EmailFrom     string
	DontSendEmail bool
	Monitor       bool
}

func (o *AdminOpts) flags() {
	flag.BoolVar(
		&o.Monitor,
		"admin.monitor",
		false,
		"Switch on metrics monitoring (default false i.e. switched off)",
	)

	flag.StringVar(
		&o.EmailFrom,
		"admin.email-from",
		"atnight@mail.cern.ch",
		"Send failure emails from this address",
	)

	flag.StringVar(
		&o.EmailTo,
		"admin.email-to",
		strings.Join([]string{
			// TODO: push this into a separate input
			"oana.boeriu@cern.ch",
			"brinick.simmons@cern.ch",
		}, ","),
		"Send failure emails to this comma-separated list of addresses",
	)

	flag.BoolVar(
		&o.DontSendEmail,
		"admin.no-email",
		false,
		"Do not send an email on failure (default false i.e do send email)",
	)
}

func (o *AdminOpts) validate() error {
	emails := strings.Split(o.EmailTo, ",")
	emails = append(emails, o.EmailFrom)
	for _, email := range emails {
		// Very basic check
		if strings.Count(email, "@") != 1 {
			return fmt.Errorf("Invalid email address %s", email)
		}
	}

	return nil
}

func (o *AdminOpts) String() string {
	args := []string{
		"- Admin Options:",
		fmt.Sprintf("   - Failure email sender = %s", o.EmailFrom),
		fmt.Sprintf("   - Failure email recipients = %s", o.EmailTo),
		fmt.Sprintf("   - Send email on failure: %t", !o.DontSendEmail),
	}

	return strings.Join(args, "\n")
}

package notify

import "testing"

func TestNewEmail(t *testing.T) {
	email := NewEmail("sender@from.me", "addr1@hello.org, addr2@world.com,   addr3@inter.net")

	toEmails := []string{"addr1@hello.org", "addr2@world.com", "addr3@inter.net"}

	for i, em := range email.to {
		if em != toEmails[i] {
			t.Errorf(
				"failed to split multiple to addresses correctly, expected %s, got %s",
				toEmails[i],
				em,
			)

			break
		}
	}
}

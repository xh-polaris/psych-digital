package util

import "testing"

func TestEmail(t *testing.T) {
	err := AlertEMail()
	if err != nil {
		t.Errorf("AlertEMail() error: %v", err)
	}
}

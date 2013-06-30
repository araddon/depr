package deprlib

import (
	"testing"
)

func TestChomp(t *testing.T) {
	if chompnl([]byte("This is a line\n")) != "This is a line" {
		t.Fail()
	}
	if chompnl([]byte("This is a line")) != "This is a line" {
		t.Fail()
	}
}

package gotox

import (
	"testing"
)

func TestConstants(t *testing.T) {
	if UserStatusNone != 0 {
		t.Errorf("derp")
	}
}

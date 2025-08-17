package test

import (
	"testing"
)

func TestReceiver(t *testing.T) {
	data := DBMessageRecord{}
	A := DBActions{
		Res: &data,
	}

	if err := A.Execute(); err != nil {
		t.Errorf("expected non nil value")
	}
}

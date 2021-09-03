package util

import (
	"github.com/rs/zerolog/log"
	"testing"
)

func TestUsd(t *testing.T) {

	f1 := 0.100246
	a1 := FloatToUsd(f1)
	e1 := "$0.100246"
	if a1 != e1 {
		log.Debug().
			Str("expected", e1).
			Str("actual", a1).
			Send()
		t.Fail()
	}

	f2 := 3420.84
	a2 := FloatToUsd(f2)
	e2 := "$3,420.84"
	if a2 != e2 {
		log.Debug().
			Str("expected", e2).
			Str("actual", a2).
			Send()
		t.Fail()
	}

}

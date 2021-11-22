package model

import (
	"nuchal-api/util"
	"testing"
)

func TestNewSim(t *testing.T) {

	sim, err := NewSim(uint(20), alpha, omega)
	if err != nil {
		t.Fail()
	}

	util.PrettyPrint(sim.Analysis)
}

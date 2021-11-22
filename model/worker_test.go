package model

import (
	"testing"
)

func TestPerformAllJobs(t *testing.T) {
	if err := PerformAllJobs(uint(1)); err != nil {
		t.Fail()
	}
}

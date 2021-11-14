package model

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"os"
	"testing"
)

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
}

func TestStartSellSession(t *testing.T) {

	StartSellSession(2.0579, 1.0, "TRAC-USD")

}

func TestGetSessions(t *testing.T) {

	util.PrettyPrint(GetSessions(userID))
}

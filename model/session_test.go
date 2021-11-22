package model

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"nuchal-api/util"
	"os"
	"testing"
	"time"
)

func init() {

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
}

func TestStartSellSession(t *testing.T) {

	StartSellSession(0.23482, 4, "REQ-USD")
	time.Sleep(time.Second * 5)

}

func TestGetSessions(t *testing.T) {

	util.PrettyPrint(GetSessions(userID))
}

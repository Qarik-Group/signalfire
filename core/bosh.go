package core

import (
	"time"

	"github.com/starkandwayne/signalfire/bosh"
)

//BOSH assembles the properties required for the core to coordinate with a BOSH
// director
type BOSH struct {
	Client       *bosh.Client
	PollInterval time.Duration
}

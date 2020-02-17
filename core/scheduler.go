package core

import (
	"github.com/starkandwayne/signalfire/log"
	"time"
)

type Scheduler struct {
	Boshes []BOSH
	Cache  *Cache
	Logger *log.Logger
}

func (s *Scheduler) Start() {
	//This is pretty garbage, but it's a start
	for _, b := range s.Boshes {
		go func(thisBOSH BOSH) {
			s.scrapeBOSH(thisBOSH)
			for range time.Tick(thisBOSH.PollInterval) {
				s.scrapeBOSH(thisBOSH)
			}
		}(b)
	}
}

func (s *Scheduler) scrapeBOSH(b BOSH) {
	deps, err := b.Client.Deployments()
	if err != nil {
		s.Logger.Error("Could not get deployments from BOSH with name `%s': %s", b.Client.Name(), err)
	}
	toPush := CacheEnvironment{
		Name: b.Client.Name(),
		UUID: b.Client.UUID(),
	}
	for _, dep := range deps {
		depToPush := CacheDeployment{
			Name: dep.Name,
		}

		for _, rel := range dep.Releases {
			relToPush := CacheRelease{
				Name:    rel.Name,
				Version: rel.Version,
			}

			depToPush.Releases = append(depToPush.Releases, relToPush)
		}

		toPush.Deployments = append(toPush.Deployments, depToPush)
	}
	s.Cache.UpdateEnvironment(toPush)
}

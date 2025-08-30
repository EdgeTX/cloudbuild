package targets

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func updateDefsSingleton(repoURL string) {
	defs := targetsDef.Load()
	if defs.update {
		newDefs, err := ReadTargetsDef(defs.sourceURL, repoURL)
		if err != nil {
			log.Errorf("could not update targets: %s", err)
		} else {
			targetsDef.Store(newDefs)
		}
	} else {
		newDefs := *defs
		if err := newDefs.updateRefs(repoURL, false); err != nil {
			log.Errorf("could not update targets: %s", err)
		}
		targetsDef.Store(&newDefs)
	}
}

func Updater(interval time.Duration, repoURL string) {
	for {
		time.Sleep(interval)
		updateDefsSingleton(repoURL)
	}
}

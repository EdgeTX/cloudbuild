package targets

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func UpdateRefs(defs *TargetsDef) {
	for k := range defs.Releases {
		v := defs.Releases[k]
		if v.Remote == nil {
			continue
		}
		sha, err := v.Remote.Fetch()
		if err != nil {
			log.Errorf("failed to fetch ref '%v': %s", k, err.Error())
			continue
		}
		v.SHA = sha
		defs.Releases[k] = v
	}
}

func updateDefsSingleton() {
	newDefs := *targetsDef.Load()
	UpdateRefs(&newDefs)
	targetsDef.Store(&newDefs)
}

func Updater(interval time.Duration) {
	for {
		time.Sleep(interval)
		updateDefsSingleton()
	}
}

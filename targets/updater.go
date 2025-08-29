package targets

import (
	"time"

	log "github.com/sirupsen/logrus"
)

func UpdateRefs(defs *TargetsDef, repoURL string) {
	var (
		tags map[string]string
		err  error
	)

	if repoURL != "" {
		log.Debugf("Listing tags...")
		tags, err = ListTags(repoURL)
		if err != nil {
			log.Errorf("could not list tags from %s: %s", repoURL, err)
		}
	} else {
		tags = make(map[string]string)
	}

	for k := range defs.Releases {
		v := defs.Releases[k]
		if v.update {
			tag := k.String()
			if sha, ok := tags[tag]; ok {
				v.SHA = sha
				log.Debugf("%s -> %s", tag, v.SHA)
			} else {
				log.Errorf("could not update %s from %s", tag, repoURL)
			}
		}
	}
}

func updateDefsSingleton(repoURL string) {
	newDefs := *targetsDef.Load()
	UpdateRefs(&newDefs, repoURL)
	targetsDef.Store(&newDefs)
}

func Updater(interval time.Duration, repoURL string) {
	for {
		time.Sleep(interval)
		updateDefsSingleton(repoURL)
	}
}

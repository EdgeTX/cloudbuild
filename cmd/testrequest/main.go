package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/server"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	data := server.BuildRequest{
		CommitHash: "fac0eed66c2e7c4eb3d4602dd52b30564c4a4a16",
		Flags: []firmware.BuildFlag{
			firmware.NewFlag("DISABLE_COMPANION", "YES"),
			firmware.NewFlag("CMAKE_BUILD_TYPE", "Release"),
			firmware.NewFlag("TRACE_SIMPGMSPACE", "NO"),
			firmware.NewFlag("VERBOSE_CMAKELISTS", "YES"),
			firmware.NewFlag("CMAKE_RULE_MESSAGES", "OFF"),
			firmware.NewFlag("PCB", "X10"),
			firmware.NewFlag("PCBREV", "TX16S"),
			firmware.NewFlag("INTERNAL_MODULE_CRSF", "ON"),
			firmware.NewFlag("ENABLE_SERIAL_PASSTHROUGH", "YES"),
		},
	}

	body, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("failed to build json: %s", err)
	}

	url := "http://api:3000/jobs"
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		log.Fatalf("failed to create request: %s", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("failed to send request: %s", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	log.Infof("data: %s", responseBody)
}

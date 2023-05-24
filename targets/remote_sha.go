package targets

import (
	"encoding/json"
	"io"
	"net/http"
)

type RemoteSHA struct {
	URL  string `json:"url"`
	Path string `json:"json_path"`
}

type githubRefObject struct {
	SHA string `json:"sha"`
}

type githubResponse struct {
	Object githubRefObject `json:"object"`
}

func (r *RemoteSHA) Fetch() (string, error) {
	resp, err := http.Get(r.URL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var ghResp githubResponse
	err = json.Unmarshal(body, &ghResp)
	if err != nil {
		return "", err
	}

	return ghResp.Object.SHA, nil
}

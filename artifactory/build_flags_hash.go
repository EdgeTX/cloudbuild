package artifactory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"sort"

	"github.com/edgetx/cloudbuild/firmware"
)

type byBuildFlag []firmware.BuildFlag

func (a byBuildFlag) Len() int           { return len(a) }
func (a byBuildFlag) Less(i, j int) bool { return a[i].Key < a[j].Key }
func (a byBuildFlag) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

/*
	Build flags SHA256 hash is created by sorting and joining build flags array.
*/
func HashBuildFlags(buildFlags []firmware.BuildFlag) string {
	flags := make([]firmware.BuildFlag, len(buildFlags))
	copy(flags, buildFlags)
	sort.Sort(byBuildFlag(flags))

	var data bytes.Buffer
	for i := range flags {
		data.WriteString(flags[i].Format())
	}
	hash := sha256.New()
	hash.Write(data.Bytes())
	md := hash.Sum(nil)
	return hex.EncodeToString(md)
}

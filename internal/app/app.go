package app

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const releasesURL = "https://api.github.com/repos/gameap/daemon/releases"

func Run(_ []string) {
	http.HandleFunc("/", handler) // each request calls handler
	err := http.ListenAndServe("0.0.0.0:8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	j, err := http.Get("https://api.github.com/repos/gameap/daemon/releases")

	w.Header().Add("Content-Type", "text/plain")

	if err != nil {
		log.Print(err)
		_, _ = w.Write([]byte(`github unavailable`))
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	query := r.URL.Query()
	if len(query["os"]) == 0 || len(query["arch"]) == 0 {
		_, _ = w.Write([]byte(`invalid request`))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	link, err := findRelease(j.Body, query["os"][0], query["arch"][0])
	if err != nil {
		log.Print(err)
		_, _ = w.Write([]byte(`failed to find archive file`))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if link == "" {
		_, _ = w.Write([]byte(`archive file not found`))
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Location", link)
	w.WriteHeader(http.StatusFound)
}

type releases struct {
	TagName string `json:"tag_name"`
	Assets  []asset  `json:"assets"`
}

type asset struct {
	Url                string `json:"url"`
	Name               string `json:"name"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

func findRelease(reader io.Reader, os string, arch string) (string, error) {
	r := []releases{}
	d := json.NewDecoder(reader)
	err := d.Decode(&r)
	if err != nil {
		return "", err
	}

	for _, release := range r {
		archiveName := fmt.Sprintf("gameap-daemon-%s-%s-%s.tar.gz", release.TagName, os, arch)
		archiveNameWindows := fmt.Sprintf("gameap-daemon-%s-%s-%s.zip", release.TagName, os, arch)

		for _, asset := range release.Assets {
			if asset.Name == archiveName {
				return asset.BrowserDownloadUrl, nil
			}

			if os == "windows" && asset.Name == archiveNameWindows {
				return asset.BrowserDownloadUrl, nil
			}
		}
	}

	return "", nil
}

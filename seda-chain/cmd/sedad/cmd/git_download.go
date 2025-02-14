package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const (
	OrgName  = "sedaprotocol"
	RepoName = "seda-networks"
)

type GitFile struct {
	Name string `json:"name"`
	Path string `json:"path"`
	URL  string `json:"download_url"`
	Type string `json:"type"`
}

func DownloadGitFiles(path, downloadPath string) error {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s", OrgName, RepoName, path)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	} else if resp.StatusCode == 401 {
		return errors.New("github authorization failure")
	}
	defer resp.Body.Close()

	var files []GitFile
	err = json.NewDecoder(resp.Body).Decode(&files)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.Type == "dir" {
			err := DownloadGitFiles(file.Path, downloadPath)
			if err != nil {
				return err
			}
		} else {
			err := downloadGitFile(file, downloadPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func downloadGitFile(file GitFile, downloadPath string) error {
	resp, err := http.Get(file.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Split the path into parts
	pathParts := strings.Split(file.Path, "/")

	// Rejoin everything except the first part
	subPath := strings.Join(pathParts[1:], "/")

	// Create all directories in the path
	localPath := filepath.Join(downloadPath, subPath)
	err = os.MkdirAll(filepath.Dir(localPath), os.ModePerm)
	if err != nil {
		return err
	}

	out, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
)

func getVideoAspectRatio(filePath string) (string, error) {

	type FileAspects struct {
		Streams []struct {
			Width       int    `json:"width"`
			Height      int    `json:"height"`
			AspectRatio string `json:"display_aspect_ratio"`
		} `json:"streams"`
	}

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var stdOut bytes.Buffer

	cmd.Stdout = &stdOut

	err := cmd.Run()

	if err != nil {
		return "", err
	}
	fileAspects := FileAspects{}
	err = json.Unmarshal(stdOut.Bytes(), &fileAspects)

	if err != nil {
		return "", err
	}

	return fileAspects.Streams[0].AspectRatio, nil
}

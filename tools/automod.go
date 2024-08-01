package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/config"
	"os/exec"
)

type AutoMod struct {
	HasIssues bool `json:"HasIssues"`
}

func RunAutoMod(directory string) (*AutoMod, error) {
	cmd := exec.Command(
		config.Instance.QuaverToolsPath,
		"-automod",
		directory)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return nil, fmt.Errorf("%v\n\n```%v```", err, stderr.String())
	}

	var automod AutoMod

	if err := json.Unmarshal(out.Bytes(), &automod); err != nil {
		return nil, err
	}

	return &automod, nil
}

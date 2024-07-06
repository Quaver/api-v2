package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/config"
	"os/exec"
	"strconv"
)

type DifficultyCalc struct {
	Difficulty struct {
		OverallDifficulty float64 `json:"OverallDifficulty"`
		Version           string  `json:"Version"`
	} `json:"Difficulty"`
}

// RunDifficultyCalculator Runs the difficulty calculator from Quaver.Tools and returns the value
func RunDifficultyCalculator(path string, mods int64) (*DifficultyCalc, error) {
	cmd := exec.Command(config.Instance.QuaverToolsPath, "-calcdiff", path, strconv.FormatInt(mods, 10))

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return nil, fmt.Errorf("%v\n\n```%v```", err, stderr.String())
	}

	var calc DifficultyCalc

	if err := json.Unmarshal(out.Bytes(), &calc); err != nil {
		return nil, err
	}

	return &calc, nil
}

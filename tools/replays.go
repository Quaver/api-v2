package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"os/exec"
	"strconv"
)

// BuildReplay Builds a full replay file from a headerless one.
func BuildReplay(user *db.User, score *db.Score, headerlessPath string, outputPath string) error {
	if score.QuaverVersion == "" {
		score.QuaverVersion = "0.0.1"
	}

	cmd := exec.Command(config.Instance.QuaverToolsPath, "-buildreplay",
		headerlessPath, outputPath, score.QuaverVersion, score.MapMD5,
		strconv.FormatInt(score.TimePlayEnd, 10),
		strconv.Itoa(int(score.Map.GameMode)),
		strconv.FormatInt(score.Modifiers, 10),
		strconv.Itoa(score.TotalScore),
		fmt.Sprintf("%f", score.Accuracy),
		strconv.Itoa(score.MaxCombo),
		strconv.Itoa(score.CountMarvelous),
		strconv.Itoa(score.CountPerfect),
		strconv.Itoa(score.CountGreat),
		strconv.Itoa(score.CountGood),
		strconv.Itoa(score.CountOkay),
		strconv.Itoa(score.CountMiss),
		strconv.Itoa(score.PauseCount),
		user.Username)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return fmt.Errorf("%v\n\n```%v```", err, stderr.String())
	}

	return nil
}

type VirtualReplayPlayer struct {
	Score      int      `json:"Score"`
	Accuracy   float64  `json:"Accuracy"`
	MaxCombo   int      `json:"MaxCombo"`
	CountMarv  int      `json:"CountMarv"`
	CountPerf  int      `json:"CountPerf"`
	CountGreat int      `json:"CountGreat"`
	CountGood  int      `json:"CountGood"`
	CountOkay  int      `json:"CountOkay"`
	CountMiss  int      `json:"CountMiss"`
	Hits       []string `json:"Hits"`
}

// PlayReplayVirtually Plays a replay virtually and returns the result
func PlayReplayVirtually(quaPath string, replayPath string, mods int64) (*VirtualReplayPlayer, error) {
	cmd := exec.Command(config.Instance.QuaverToolsPath,
		"-virtualreplay",
		replayPath,
		quaPath,
		strconv.FormatInt(mods, 10),
		"-hl")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()

	if err != nil {
		return nil, fmt.Errorf("%v\n\n```%v```", err, stderr.String())
	}

	type Output struct {
		Player VirtualReplayPlayer `json:"VirtualReplayPlayer"`
	}

	var data Output

	if err := json.Unmarshal(out.Bytes(), &data); err != nil {
		return nil, err
	}

	return &data.Player, nil
}

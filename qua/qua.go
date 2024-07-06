package qua

import (
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
)

type Qua struct {
	AudioFile                      string `yaml:"AudioFile"`
	SongPreviewTime                int    `yaml:"SongPreviewTime"`
	BackgroundFile                 string `yaml:"BackgroundFile"`
	BannerFile                     string `yaml:"BannerFile"`
	MapId                          int    `yaml:"MapId"`
	MapSetId                       int    `yaml:"MapSetId"`
	RawMode                        string `yaml:"Mode"`
	Mode                           enums.GameMode
	Title                          string  `yaml:"Title"`
	Artist                         string  `yaml:"Artist"`
	Source                         string  `yaml:"Source"`
	Tags                           string  `yaml:"Tags"`
	Creator                        string  `yaml:"Creator"`
	DifficultyName                 string  `yaml:"DifficultyName"`
	Description                    string  `yaml:"Description"`
	Genre                          string  `yaml:"Genre"`
	BPMDoesNotAffectScrollVelocity bool    `yaml:"BPMDoesNotAffectScrollVelocity"`
	InitialScrollVelocity          float32 `yaml:"InitialScrollVelocity"`
	HasScratchKey                  bool    `yaml:"HasScratchKey"`
	RawBytes                       []byte  `yaml:"-,omitempty"`
}

// Parse Parses and returns a Qua file
func Parse(file []byte) (*Qua, error) {
	qua := Qua{RawBytes: file}

	if err := yaml.Unmarshal(file, &qua); err != nil {
		logrus.Error(err)
		return nil, err
	}

	switch qua.RawMode {
	case "Keys4":
	case "1":
		qua.Mode = enums.GameModeKeys4
		break
	case "Keys7":
	case "2":
		qua.Mode = enums.GameModeKeys7
		break
	}

	return &qua, nil
}

func (q *Qua) ReplaceIds(mapsetId int, mapId int) string {
	q.MapSetId = mapsetId
	q.MapId = mapId

	fileStr := string(q.RawBytes)

	fileStr = regexp.MustCompile(`MapSetId:\s*-?\d+`).ReplaceAllStringFunc(fileStr, func(match string) string {
		return fmt.Sprintf("MapSetId: %v", mapsetId)
	})

	fileStr = regexp.MustCompile(`MapId:\s*-?\d+`).ReplaceAllStringFunc(fileStr, func(match string) string {
		return fmt.Sprintf("MapId: %v", mapId)
	})

	q.RawBytes = []byte(fileStr)
	return fileStr
}

// Writes the .qua to a file
func (q *Qua) Write(path string) error {
	return os.WriteFile(path, q.RawBytes, 0644)
}

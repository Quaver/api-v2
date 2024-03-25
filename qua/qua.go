package qua

import (
	"github.com/Quaver/api2/enums"
	"github.com/goccy/go-yaml"
	"github.com/sirupsen/logrus"
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
}

// Parse Parses and returns a Qua file
func Parse(file []byte) (*Qua, error) {
	qua := Qua{}

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

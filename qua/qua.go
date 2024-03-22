package qua

import (
	"github.com/Quaver/api2/enums"
	"github.com/goccy/go-yaml"
)

type Qua struct {
	AudioFile                      string         `yaml:"AudioFile"`
	SongPreviewTime                int            `yaml:"SongPreviewTime"`
	BackgroundFile                 string         `yaml:"BackgroundFile"`
	BannerFile                     string         `yaml:"BannerFile"`
	MapId                          int            `yaml:"MapId"`
	MapSetId                       int            `yaml:"MapSetId"`
	Mode                           enums.GameMode `yaml:"Mode"`
	Title                          string         `yaml:"Title"`
	Artist                         string         `yaml:"Artist"`
	Source                         string         `yaml:"Source"`
	Tags                           string         `yaml:"Tags"`
	Creator                        string         `yaml:"Creator"`
	DifficultyName                 string         `yaml:"DifficultyName"`
	Description                    string         `yaml:"Description"`
	Genre                          string         `yaml:"Genre"`
	BPMDoesNotAffectScrollVelocity bool           `yaml:"BPMDoesNotAffectScrollVelocity"`
	InitialScrollVelocity          float32        `yaml:"InitialScrollVelocity"`
	HasScratchKey                  bool           `yaml:"HasScratchKey"`
}

// Parse Parses and returns a Qua file
func Parse(file []byte) (*Qua, error) {
	var qua *Qua

	if err := yaml.Unmarshal(file, &qua); err != nil {
		return nil, err
	}

	return qua, nil
}

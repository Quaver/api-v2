package qua

import (
	"fmt"
	"github.com/Quaver/api2/enums"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"math"
	"os"
	"regexp"
)

type Qua struct {
	RawBytes                       []byte           `yaml:"-"`
	AudioFile                      string           `yaml:"AudioFile"`
	SongPreviewTime                int              `yaml:"SongPreviewTime"`
	BackgroundFile                 string           `yaml:"BackgroundFile"`
	BannerFile                     string           `yaml:"BannerFile"`
	MapId                          int              `yaml:"MapId"`
	MapSetId                       int              `yaml:"MapSetId"`
	RawMode                        string           `yaml:"Mode"`
	Mode                           enums.GameMode   `yaml:"-"`
	Title                          string           `yaml:"Title"`
	Artist                         string           `yaml:"Artist"`
	Source                         string           `yaml:"Source"`
	Tags                           string           `yaml:"Tags"`
	Creator                        string           `yaml:"Creator"`
	DifficultyName                 string           `yaml:"DifficultyName"`
	Description                    string           `yaml:"Description"`
	Genre                          string           `yaml:"Genre"`
	LegacyLNRendering              bool             `yaml:"LegacyLNRendering"`
	BPMDoesNotAffectScrollVelocity bool             `yaml:"BPMDoesNotAffectScrollVelocity"`
	InitialScrollVelocity          float32          `yaml:"InitialScrollVelocity"`
	HasScratchKey                  bool             `yaml:"HasScratchKey"`
	EditorLayers                   []EditorLayer    `yaml:"EditorLayers"`
	Bookmarks                      []Bookmark       `yaml:"Bookmarks"`
	SoundEffects                   []SoundEffect    `yaml:"SoundEffects"`
	TimingPoints                   []TimingPoint    `yaml:"TimingPoints"`
	ScrollVelocities               []ScrollVelocity `yaml:"SliderVelocities"`
	HitObjects                     []HitObject      `yaml:"HitObjects"`
}

// Parse Parses and returns a Qua file
func Parse(file []byte) (*Qua, error) {
	qua := Qua{RawBytes: file}

	if err := yaml.Unmarshal(file, &qua); err != nil {
		logrus.Error(err)
		return nil, err
	}

	var err error
	qua.Mode, err = enums.GetModeFromString(qua.RawMode)

	if err != nil {
		return nil, err
	}

	return &qua, nil
}

// MapLength Returns the length of the map
func (q *Qua) MapLength() int {
	var length int

	for _, hitObject := range q.HitObjects {
		length = int(math.Max(float64(length), float64(hitObject.StartTime)))
		length = int(math.Max(float64(length), float64(hitObject.EndTime)))
	}

	return length
}

// CommonBPM Returns the most common BPM in the map
func (q *Qua) CommonBPM() float32 {
	if len(q.TimingPoints) == 0 {
		return 0
	}

	if len(q.TimingPoints) == 1 {
		return q.TimingPoints[0].BPM
	}

	var durations = map[float32]float32{}
	lastTime := float32(q.MapLength())

	for i := len(q.TimingPoints) - 1; i >= 0; i-- {
		point := q.TimingPoints[i]

		if point.StartTime > lastTime {
			continue
		}

		duration := lastTime - point.StartTime
		durations[point.BPM] += duration

		lastTime = point.StartTime
	}

	var maxBpm float32
	var maxDuration float32
	firstIter := true

	for bpm, duration := range durations {
		if firstIter || duration > maxDuration {
			maxBpm = bpm
			maxDuration = duration
			firstIter = false
		}
	}

	return maxBpm
}

// CountHitObjectNormal Returns the count of normal hit objects in the map
func (q *Qua) CountHitObjectNormal() int {
	var count int

	for _, hitObject := range q.HitObjects {
		if !hitObject.IsLongNote() {
			count++
		}
	}

	return count
}

// CountHitObjectLong Returns the count of long notes in the map
func (q *Qua) CountHitObjectLong() int {
	var count int

	for _, hitObject := range q.HitObjects {
		if hitObject.IsLongNote() {
			count++
		}
	}

	return count
}

// MaxCombo Returns the max combo achievable in the map
func (q *Qua) MaxCombo() int {
	return q.CountHitObjectLong()*2 + q.CountHitObjectNormal()
}

// ReplaceIds Replaces the ids of the map and sets Qua.RawBytes. Returns a string of the new file.
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

// FileName Returns the file name of the qua (map_id.qua)
func (q *Qua) FileName() string {
	return fmt.Sprintf("%v.qua", q.MapId)
}

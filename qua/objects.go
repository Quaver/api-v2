package qua

type EditorLayer struct {
	Name     string `yaml:"Name"`
	Hidden   bool   `yaml:"Hidden"`
	ColorRGB string `yaml:"ColorRgb"`
}

type Bookmark struct {
	StartTime int    `yaml:"StartTime"`
	Note      string `yaml:"Note"`
}

type CustomAudioSample struct {
	Path             string `yaml:"Path"`
	UnaffectedByRate bool   `yaml:"UnaffectedByRate"`
}

type SoundEffect struct {
	StartTime float32 `yaml:"StartTime"`
	Sample    int     `yaml:"Sample"`
	Volume    int     `yaml:"Volume"`
}

type TimingPoint struct {
	StartTime     float32 `yaml:"StartTime"`
	BPM           float32 `yaml:"Bpm"`
	TimeSignature int8    `yaml:"TimeSignature"`
	Hidden        bool    `yaml:"Hidden"`
}

type ScrollVelocity struct {
	StartTime  float32 `yaml:"StartTime"`
	Multiplier float32 `yaml:"Multiplier"`
}

type KeySound struct {
	Sample int `yaml:"Sample"`
	Volume int `yaml:"Volume"`
}

type HitObject struct {
	StartTime   int        `yaml:"StartTime"`
	Lane        int        `yaml:"Lane"`
	EndTime     int        `yaml:"EndTime"`
	HitSound    int        `yaml:"HitSound"`
	KeySounds   []KeySound `yaml:"KeySounds"`
	EditorLayer int        `yaml:"EditorLayer"`
}

// IsLongNote Returns if the hit object is a long note
func (h HitObject) IsLongNote() bool {
	return h.EndTime > 0
}

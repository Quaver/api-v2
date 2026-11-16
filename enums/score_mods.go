package enums

type ScoreMods struct {
	NoMiss             bool
	FullLN             bool
	Inverse            bool
	NoLongNotes        bool
	NoSliderVelocities bool
	Mirror             bool
	SpeedRate          int
}

var ScoreSpeedRates = [...]struct {
	Mod  Mods
	Rate int
}{
	{ModSpeed05X, 50},
	{ModSpeed055X, 55},
	{ModSpeed06X, 60},
	{ModSpeed065X, 65},
	{ModSpeed07X, 70},
	{ModSpeed075X, 75},
	{ModSpeed08X, 80},
	{ModSpeed085X, 85},
	{ModSpeed09X, 90},
	{ModSpeed095X, 95},
	{ModSpeed105X, 105},
	{ModSpeed11X, 110},
	{ModSpeed115X, 115},
	{ModSpeed12X, 120},
	{ModSpeed125X, 125},
	{ModSpeed13X, 130},
	{ModSpeed135X, 135},
	{ModSpeed14X, 140},
	{ModSpeed145X, 145},
	{ModSpeed15X, 150},
	{ModSpeed155X, 155},
	{ModSpeed16X, 160},
	{ModSpeed165X, 165},
	{ModSpeed17X, 170},
	{ModSpeed175X, 175},
	{ModSpeed18X, 180},
	{ModSpeed185X, 185},
	{ModSpeed19X, 190},
	{ModSpeed195X, 195},
	{ModSpeed20X, 200},
}

func GetScoreMods(mods Mods) ScoreMods {
	metadata := ScoreMods{
		NoMiss:             IsModActivated(mods, ModNoMiss),
		FullLN:             IsModActivated(mods, ModFullLN),
		Inverse:            IsModActivated(mods, ModInverse),
		NoLongNotes:        IsModActivated(mods, ModNoLongNotes),
		NoSliderVelocities: IsModActivated(mods, ModNoSliderVelocities),
		Mirror:             IsModActivated(mods, ModMirror),
		SpeedRate:          100,
	}
	for _, speed := range ScoreSpeedRates {
		if IsModActivated(mods, speed.Mod) {
			metadata.SpeedRate = speed.Rate
			break
		}
	}
	return metadata
}

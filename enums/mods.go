package enums

import (
	"strings"
)

type Mods int64

const (
	ModNoSliderVelocities Mods = 1 << iota
	ModSpeed05X
	ModSpeed06X
	ModSpeed07X
	ModSpeed08X
	ModSpeed09X
	ModSpeed11X
	ModSpeed12X
	ModSpeed13X
	ModSpeed14X
	ModSpeed15X
	ModSpeed16X
	ModSpeed17X
	ModSpeed18X
	ModSpeed19X
	ModSpeed20X
	ModStrict
	ModChill
	ModNoPause
	ModAutoplay
	ModPaused
	ModNoFail
	ModNoLongNotes
	ModRandomize
	ModSpeed055X
	ModSpeed065X
	ModSpeed075X
	ModSpeed085X
	ModSpeed095X
	ModInverse
	ModFullLN
	ModMirror
	ModCoop
	ModSpeed105X
	ModSpeed115X
	ModSpeed125X
	ModSpeed135X
	ModSpeed145X
	ModSpeed155X
	ModSpeed165X
	ModSpeed175X
	ModSpeed185X
	ModSpeed195X
	ModHealthAdjust
	ModNoMiss
	ModEnumMaxValue // This is only in place for looping purposes (i < ModEnumMaxValue - 1; i++)
)

var RankedMods = []Mods{
	ModSpeed05X,
	ModSpeed055X,
	ModSpeed06X,
	ModSpeed065X,
	ModSpeed07X,
	ModSpeed075X,
	ModSpeed08X,
	ModSpeed085X,
	ModSpeed09X,
	ModSpeed095X,
	ModSpeed105X,
	ModSpeed11X,
	ModSpeed115X,
	ModSpeed12X,
	ModSpeed125X,
	ModSpeed13X,
	ModSpeed135X,
	ModSpeed14X,
	ModSpeed145X,
	ModSpeed15X,
	ModSpeed155X,
	ModSpeed16X,
	ModSpeed165X,
	ModSpeed17X,
	ModSpeed175X,
	ModSpeed18X,
	ModSpeed185X,
	ModSpeed19X,
	ModSpeed195X,
	ModSpeed20X,
	ModMirror,
	ModNoMiss,
}

var ModStrings = map[Mods]string{
	ModNoSliderVelocities: "NSV",
	ModSpeed05X:           "0.5x",
	ModSpeed06X:           "0.6x",
	ModSpeed07X:           "0.7x",
	ModSpeed08X:           "0.8x",
	ModSpeed09X:           "0.9x",
	ModSpeed11X:           "1.1x",
	ModSpeed12X:           "1.2x",
	ModSpeed13X:           "1.3x",
	ModSpeed14X:           "1.4x",
	ModSpeed15X:           "1.5x",
	ModSpeed16X:           "1.6x",
	ModSpeed17X:           "1.7x",
	ModSpeed18X:           "1.8x",
	ModSpeed19X:           "1.9x",
	ModSpeed20X:           "2.0x",
	ModStrict:             "Strict",
	ModChill:              "Chill",
	ModNoPause:            "No Pause",
	ModAutoplay:           "Autoplay",
	ModPaused:             "Paused",
	ModNoFail:             "No Fail",
	ModNoLongNotes:        "No Long Notes",
	ModRandomize:          "Randomize",
	ModSpeed055X:          "0.55x",
	ModSpeed065X:          "0.65x",
	ModSpeed075X:          "0.75x",
	ModSpeed085X:          "0.85x",
	ModSpeed095X:          "0.95x",
	ModInverse:            "Inverse",
	ModFullLN:             "Full Long Notes",
	ModMirror:             "Mirror",
	ModCoop:               "Co-op",
	ModSpeed105X:          "1.05x",
	ModSpeed115X:          "1.15x",
	ModSpeed125X:          "1.25x",
	ModSpeed135X:          "1.35x",
	ModSpeed145X:          "1.45x",
	ModSpeed155X:          "1.55x",
	ModSpeed165X:          "1.65x",
	ModSpeed175X:          "1.75x",
	ModSpeed185X:          "1.85x",
	ModSpeed195X:          "1.95x",
	ModHealthAdjust:       "Health Adjustments",
	ModNoMiss:             "NM",
	ModEnumMaxValue:       "INVALID!",
}

// IsModActivated Returns if a given mod is activated in a mod combo
func IsModActivated(modCombo Mods, mod Mods) bool {
	return modCombo&mod != 0
}

// IsModComboRanked Returns if a combination of mods is ranked
func IsModComboRanked(modCombo Mods) bool {
	if modCombo == 0 {
		return true
	}

	for i := 0; (1 << i) < ModEnumMaxValue-1; i++ {
		mod := Mods(1 << i)

		if !IsModActivated(modCombo, mod) {
			continue
		}

		if !isModRanked(mod) {
			return false
		}
	}

	return true
}

// IsUnrankedModComboAllowed Returns if a combination of mods is allowed in score submission
func IsUnrankedModComboAllowed(modCombo Mods) bool {
	if modCombo == 0 {
		return true
	}

	for i := 0; (1 << i) < ModEnumMaxValue-1; i++ {
		mod := Mods(1 << i)

		if !IsModActivated(modCombo, mod) {
			continue
		}

		if !isUnrankedModAllowed(mod) && !isModRanked(mod) {
			return false
		}
	}

	return true
}

// GetModsString Gets a stringified version of a mod combination
func GetModsString(modCombo Mods) string {
	if modCombo == 0 {
		return "None"
	}

	mods := []string{}

	for i := 0; (1 << i) < ModEnumMaxValue-1; i++ {
		mod := Mods(1 << i)

		if !IsModActivated(modCombo, mod) {
			continue
		}

		mods = append(mods, ModStrings[mod])
	}

	return strings.Join(mods[:], ", ")
}

// HasIncompatibleModifiers Checks if the combination of modifiers is incompatible
func HasIncompatibleModifiers(modCombo Mods) bool {
	for i := 0; (1 << i) < ModEnumMaxValue-1; i++ {
		mod := Mods(1 << i)

		if !IsModActivated(modCombo, mod) {
			continue
		}

		// Go through each modifier
		for j := 0; (1 << j) < ModEnumMaxValue-1; j++ {
			modToCheck := Mods(1 << j)

			if !IsModActivated(modCombo, modToCheck) || mod == modToCheck {
				continue
			}

			// Both modifiers are speed mods
			if isSpeedModifier(mod) && isSpeedModifier(modToCheck) {
				return true
			}

			// Both modifiers change long notes in some way
			if isLongNoteModifier(mod) && isLongNoteModifier(modToCheck) {
				return true
			}
		}
	}

	return false
}

// isModRanked Returns if a particular mod is ranked
func isModRanked(mod Mods) bool {
	if mod == 0 {
		return true
	}

	for _, rankedMod := range RankedMods {
		if rankedMod == mod {
			return true
		}
	}

	return false
}

// isUnrankedModAllowed  Returns if a particular mod is allowed to be submitted.
func isUnrankedModAllowed(mod Mods) bool {
	if mod == 0 {
		return true
	}

	switch mod {
	case ModNoLongNotes, ModFullLN, ModInverse, ModNoSliderVelocities:
		return true
	}

	return false
}

// Returns if the modifier is a speed modifier
func isSpeedModifier(mod Mods) bool {
	return (mod >= ModSpeed05X && mod <= ModSpeed20X) ||
		(mod >= ModSpeed055X && mod <= ModSpeed095X) ||
		(mod >= ModSpeed105X && mod <= ModSpeed195X)
}

// Returns if a modifier changes the long notes within the map
func isLongNoteModifier(mod Mods) bool {
	return mod == ModFullLN || mod == ModInverse || mod == ModNoLongNotes
}

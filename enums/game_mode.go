package enums

type GameMode int

const (
	None GameMode = iota
	GameModeKeys4
	GameModeKeys7

	// GameModeKeys1 New game modes so they start counting from 3
	GameModeKeys1 = 3
	GameModeKeys2 = iota
	GameModeKeys3
	GameModeKeys5
	GameModeKeys6
	GameModeKeys8
	GameModeKeys9
	GameModeKeys10
	GameModeEnumMaxValue
)

// GetGameModeString Returns a game mode int in its stringified name
func GetGameModeString(mode GameMode) string {
	switch mode {
	case GameModeKeys1:
		return "keys1"
	case GameModeKeys2:
		return "keys2"
	case GameModeKeys3:
		return "keys3"
	case GameModeKeys4:
		return "keys4"
	case GameModeKeys5:
		return "keys5"
	case GameModeKeys6:
		return "keys6"
	case GameModeKeys7:
		return "keys7"
	case GameModeKeys8:
		return "keys8"
	case GameModeKeys9:
		return "keys9"
	case GameModeKeys10:
		return "keys10"
	default:
		return "not_implemented"
	}
}

// GetShorthandGameModeString Gets a short-handed version of a game mode
func GetShorthandGameModeString(mode GameMode) string {
	switch mode {
	case GameModeKeys1:
		return "1K"
	case GameModeKeys2:
		return "2K"
	case GameModeKeys3:
		return "3K"
	case GameModeKeys4:
		return "4K"
	case GameModeKeys5:
		return "5K"
	case GameModeKeys6:
		return "6K"
	case GameModeKeys7:
		return "7K"
	case GameModeKeys8:
		return "8K"
	case GameModeKeys9:
		return "9K"
	case GameModeKeys10:
		return "10K"
	default:
		return "not_implemented"
	}
}

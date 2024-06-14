package enums

type GameMode int

const (
	None GameMode = iota
	GameModeKeys4
	GameModeKeys7
)

// GetGameModeString Returns a game mode int in its stringified name
func GetGameModeString(mode GameMode) string {
	switch mode {
	case GameModeKeys4:
		return "keys4"
	case GameModeKeys7:
		return "keys7"
	default:
		return "not_implemented"
	}
}

// GetShorthandGameModeString Gets a short-handed version of a game mode
func GetShorthandGameModeString(mode GameMode) string {
	switch mode {
	case GameModeKeys4:
		return "4K"
	case GameModeKeys7:
		return "7K"
	default:
		return "not_implemented"
	}
}

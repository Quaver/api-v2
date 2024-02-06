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

package main

import (
	"flag"
	"github.com/Quaver/api2/cmd/migrations/commands"
	"github.com/sirupsen/logrus"
)

func main() {
	cmd := flag.String("cmd", "", "The migration command to execute")
	flag.Parse()

	switch *cmd {
	case "playlist-mapsets":
		commands.RunPlaylistMapset()
	default:
		logrus.Fatal("You must provide a valid migration command. See the README for a list of commands.")
	}
}

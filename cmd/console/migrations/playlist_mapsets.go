package migrations

import (
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
)

var MigrationPlaylistMapsetCmd = &cobra.Command{
	Use:   "migration:playlist:mapsets",
	Short: "Migrates playlists from v1 to v2",
	Run: func(cmd *cobra.Command, args []string) {
		RunPlaylistMapset()
	},
}

func RunPlaylistMapset() {
	playlists, err := db.GetAllPlaylists()

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Found: %v playlists", len(playlists))

	for index, playlist := range playlists {
		// Delete playlist mapsets that have zero maps in them
		for _, playlistMapset := range playlist.Mapsets {
			if len(playlistMapset.Maps) == 0 {
				if err := db.DeletePlaylistMapset(playlist.Id, playlistMapset.MapsetId); err != nil {
					logrus.Fatal(err)
				}

				logrus.Infof("Deleted playlist mapset: %v from playlist %v", playlistMapset.Id, playlist.Id)
			}
		}

		if len(playlist.Maps) == 0 {
			continue
		}

		logrus.Infof("[%v/%v] Updating Playlist", index+1, len(playlists))

		for _, songMap := range playlist.Maps {
			// Check if there's a playlist_mapset with the playlist id and mapset id
			if songMap.Map == nil {
				continue
			}

			playlistMapset, err := db.GetPlaylistMapsetByIds(playlist.Id, songMap.Map.MapsetId)

			if err != nil && err != gorm.ErrRecordNotFound {
				logrus.Fatal(err)
			}

			// Create new playlist mapset
			if playlistMapset == nil {
				playlistMapset = &db.PlaylistMapset{
					PlaylistId: playlist.Id,
					MapsetId:   songMap.Map.MapsetId,
				}

				if err := playlistMapset.Insert(); err != nil {
					logrus.Fatal(err)
				}

				logrus.Infof("Inserted playlist mapset for %v - %v", playlist.Id, songMap.Map.MapsetId)
			}

			if err := songMap.UpdateMapsetId(playlistMapset.Id); err != nil {
				logrus.Fatal(err)
			}
		}

	}
}

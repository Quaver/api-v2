package commands

import (
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func RunPlaylistMapset() {
	if err := config.Load("../../config.json"); err != nil {
		logrus.Panic(err)
	}

	db.ConnectMySQL()
	defer db.CloseMySQL()

	playlists, err := db.GetAllPlaylists()

	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Found: %v playlists", len(playlists))

	for index, playlist := range playlists {
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

			// Update the playlist map with the playlist_mapsets_id
			songMap.PlaylistsMapsetId = playlistMapset.Id

			if err := db.SQL.Save(&songMap).Error; err != nil {
				logrus.Fatal(err)
			}
		}
	}
}

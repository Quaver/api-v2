package cache

import (
	"errors"
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/go-resty/resty/v2"
	"github.com/sirupsen/logrus"
)

func RemoveCacheServerFile(container string, id int) error {
	resp, err := resty.New().R().
		SetBody(map[string]interface{}{
			"key": config.Instance.CacheServer.Key,
		}).
		Post(fmt.Sprintf("%v/%v/%v/delete", config.Instance.CacheServer.URL, container, id))

	log := fmt.Sprintf("Failed to remove cache server file %v #%v with status: %v", container,
		id, resp.StatusCode())

	if err != nil {
		logrus.Warning(log)
		return errors.New(log)
	}

	if resp.IsError() {
		logrus.Warn(log)
		return errors.New(log)
	}

	logrus.Infof("Removed cache server file: %v with status: %v", container, id)
	return nil
}

func RemoveCacheServerMapsetBanner(id int) error {
	return RemoveCacheServerFile("mapsets", id)
}

func RemoveCacheServerPlaylistCover(id int) error {
	return RemoveCacheServerFile("playlists", id)
}

func RemoveCacheServerBadge(id int) error {
	return RemoveCacheServerFile("badges", id)
}

func RemoveCacheServerProfileCover(id int) error {
	return RemoveCacheServerFile("profile-covers", id)
}

func RemoveCacheServerAudioPreview(id int) error {
	return RemoveCacheServerFile("audio-previews", id)
}

func RemoveCacheServerClanAvatar(id int) error {
	return RemoveCacheServerFile("clan-avatars", id)
}

func RemoveCacheServerClanBanner(id int) error {
	return RemoveCacheServerFile("clan-banners", id)
}

func RemoveCacheServerMusicArtistAvatar(id int) error {
	return RemoveCacheServerFile("music-artist-avatars", id)
}

func RemoveCacheServerMusicArtistBanner(id int) error {
	return RemoveCacheServerFile("music-artist-banners", id)
}

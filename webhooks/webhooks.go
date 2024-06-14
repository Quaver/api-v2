package webhooks

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	rankingQueue  webhook.Client
	rankedMapsets webhook.Client
)

const (
	quaverLogo string = "https://i.imgur.com/DkJhqvT.jpg"
)

func InitializeWebhooks() {
	rankingQueue, _ = webhook.NewWithURL(config.Instance.RankingQueue.Webhook)
	rankedMapsets, _ = webhook.NewWithURL(config.Instance.RankingQueue.RankedWebhook)
}

// SendQueueSubmitWebhook Sends a webhook displaying that the user submitted a mapset to the ranking queue
func SendQueueSubmitWebhook(user *db.User, mapset *db.Mapset) error {
	if rankingQueue == nil {
		return nil
	}

	embed := discord.NewEmbedBuilder().
		SetAuthorName(user.Username).
		SetAuthorURLf("https://quavergame.com/user/%v", user.Id).
		SetAuthorIcon(*user.AvatarUrl).
		AddField("Ranking Queue Action", "Submitted", true).
		AddField("Mapset",
			fmt.Sprintf("[%v - %v](https://quavergame.com/mapsets/%v)", mapset.Artist, mapset.Title, mapset.Id), false).
		SetDescription("").
		SetThumbnail(quaverLogo).
		SetFooter("Quaver", quaverLogo).
		SetTimestamp(time.Now()).
		SetColor(0x00FFFF).
		Build()

	_, err := rankingQueue.CreateEmbeds([]discord.Embed{embed})

	if err != nil {
		logrus.Error("Failed to send ranking queue submit webhook")
		return err
	}

	return nil
}

// SendQueueWebhook Sends a ranking queue webhook
func SendQueueWebhook(user *db.User, mapset *db.Mapset, action db.RankingQueueAction) error {
	if rankingQueue == nil {
		return nil
	}

	actionStr := ""
	color := 0x000000

	switch action {
	case db.RankingQueueActionComment:
		actionStr = "Commented"
		color = 0x808080
	case db.RankingQueueActionDeny:
		actionStr = "Denied"
		color = 0xFF0000
	case db.RankingQueueActionBlacklist:
		actionStr = "Blacklisted"
		color = 0x000000
	case db.RankingQueueActionOnHold:
		actionStr = "On Hold"
		color = 0xFFFF00
	case db.RankingQueueActionVote:
		actionStr = "Voted"
		color = 0x00FF00
	}

	embed := discord.NewEmbedBuilder().
		SetAuthorName(user.Username).
		SetAuthorURLf("https://quavergame.com/user/%v", user.Id).
		SetAuthorIcon(*user.AvatarUrl).
		AddField("Ranking Queue Action", actionStr, true).
		AddField("Mapset",
			fmt.Sprintf("[%v - %v](https://quavergame.com/mapsets/%v)", mapset.Artist, mapset.Title, mapset.Id), false).
		SetDescription("").
		SetThumbnail(quaverLogo).
		SetFooter("Quaver", quaverLogo).
		SetTimestamp(time.Now()).
		SetColor(color).
		Build()

	_, err := rankingQueue.CreateEmbeds([]discord.Embed{embed})

	if err != nil {
		logrus.Error("Failed to send ranking queue action webhook")
		return err
	}

	return nil
}

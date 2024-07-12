package webhooks

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/webhook"
	"github.com/sirupsen/logrus"
	"slices"
	"strings"
	"time"
)

var (
	rankingQueue  webhook.Client
	rankedMapsets webhook.Client
	orders        webhook.Client
)

const (
	quaverLogo string = "https://i.imgur.com/DkJhqvT.jpg"
)

func InitializeWebhooks() {
	rankingQueue, _ = webhook.NewWithURL(config.Instance.RankingQueue.Webhook)
	rankedMapsets, _ = webhook.NewWithURL(config.Instance.RankingQueue.RankedWebhook)
	orders, _ = webhook.NewWithURL(config.Instance.OrdersWebhook)
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
			fmt.Sprintf("[%v](https://quavergame.com/mapsets/%v)", mapset.String(), mapset.Id), false).
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
	case db.RankingQueueActionResolved:
		actionStr = "Resolved"
		color = 0xFFA500
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
			fmt.Sprintf("[%v](https://quavergame.com/mapsets/%v)", mapset.String(), mapset.Id), false).
		SetThumbnail(quaverLogo).
		SetFooter("Quaver", quaverLogo).
		SetTimestamp(time.Now()).
		SetColor(color).
		Build()

	_, err := rankedMapsets.CreateMessage(discord.WebhookMessageCreate{
		Content: getUserPingText(mapset),
		Embeds:  []discord.Embed{embed},
	})

	if err != nil {
		logrus.Error("Failed to send ranking queue action webhook")
		return err
	}

	return nil
}

// SendRankedWebhook Sends a webhook that a new mapset was ranked
func SendRankedWebhook(mapset *db.Mapset, votes []*db.MapsetRankingQueueComment) error {
	if rankedMapsets == nil {
		return nil
	}

	votedBy := ""

	for index, voter := range votes {
		votedBy = fmt.Sprintf("[%v](https://quavergame.com/user/%v)", voter.User.Username, voter.UserId)

		if index != len(votes)-1 {
			votedBy += ", "
		}
	}

	var minDiff float64 = 0
	var maxDiff float64 = 0
	var gameModes []string

	for index, currMap := range mapset.Maps {
		if index == 0 {
			minDiff = currMap.DifficultyRating
			maxDiff = currMap.DifficultyRating
		} else {
			if currMap.DifficultyRating < minDiff {
				minDiff = currMap.DifficultyRating
			}

			if currMap.DifficultyRating > maxDiff {
				maxDiff = currMap.DifficultyRating
			}
		}

		mode := enums.GetShorthandGameModeString(currMap.GameMode)

		if !slices.Contains(gameModes, mode) {
			gameModes = append(gameModes, mode)
		}
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("✅ New Mapset Ranked!").
		SetDescription("A new mapset has been ranked and is now available to get scores on.").
		AddField("Song",
			fmt.Sprintf("[%v](https://quavergame.com/mapsets/%v)", mapset.String(), mapset.Id), true).
		AddField("Creator",
			fmt.Sprintf("[%v](https://quavergame.com/user/%v)", mapset.CreatorUsername, mapset.CreatorID), true).
		AddField("Game Modes", strings.Join(gameModes, ", "), true).
		AddField("Difficulty Range",
			fmt.Sprintf("%.2f - %.2f", minDiff, maxDiff), true).
		AddField("Ranked By", votedBy, true).
		SetImagef("https://cdn.quavergame.com/mapsets/%v.jpg", mapset.Id).
		SetThumbnail(quaverLogo).
		SetFooter("Quaver", quaverLogo).
		SetTimestamp(time.Now()).
		SetColor(0x00FF00).
		Build()

	_, err := rankedMapsets.CreateMessage(discord.WebhookMessageCreate{
		Content: getUserPingText(mapset),
		Embeds:  []discord.Embed{embed},
	})

	if err != nil {
		logrus.Error("Failed to send ranking queue action webhook")
		return err
	}

	return nil
}

func SendOrderWebhook(purchasedOrders []*db.Order) error {
	description := "**A new order has been purchased!**"

	for _, order := range purchasedOrders {
		description += fmt.Sprintf("\n- %v", order.Description)
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("💰 New Order Incoming").
		SetDescription(description).
		AddField("Purchaser", fmt.Sprintf("[User Profile](https://quavergame.com/user/%v)", purchasedOrders[0].UserId), true).
		AddField("Receiver", fmt.Sprintf("[User Profile](https://quavergame.com/user/%v)", purchasedOrders[0].ReceiverUserId), true).
		SetThumbnail(quaverLogo).
		SetFooter("Quaver", quaverLogo).
		SetTimestamp(time.Now()).
		SetColor(0x00FF00).
		Build()

	_, err := orders.CreateMessage(discord.WebhookMessageCreate{
		Embeds: []discord.Embed{embed},
	})

	if err != nil {
		logrus.Error("Failed to send order webhook: ", err)
		return err
	}

	return nil
}

func getUserPingText(mapset *db.Mapset) string {
	content := ""

	if mapset.User != nil && mapset.User.MiscInformation != nil && mapset.User.MiscInformation.NotifyMapsetActions {
		if mapset.User.DiscordId == nil {
			return ""
		}

		content = fmt.Sprintf("<@%v>", *mapset.User.DiscordId)
	}

	return content
}

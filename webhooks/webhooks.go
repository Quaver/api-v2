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
	rankingQueue    webhook.Client
	rankedMapsets   webhook.Client
	events          webhook.Client
	teamAnnounce    webhook.Client
	clansFirstPlace webhook.Client
	clansMapRanked  webhook.Client
)

const (
	QuaverLogo string = "https://i.imgur.com/DkJhqvT.jpg"
)

func InitializeWebhooks() {
	rankingQueue, _ = webhook.NewWithURL(config.Instance.RankingQueue.Webhook)
	rankedMapsets, _ = webhook.NewWithURL(config.Instance.RankingQueue.RankedWebhook)
	events, _ = webhook.NewWithURL(config.Instance.EventsWebhook)
	teamAnnounce, _ = webhook.NewWithURL(config.Instance.TeamAnnounceWebhook)
	clansFirstPlace, _ = webhook.NewWithURL(config.Instance.ClansFirstPlaceWebhook)
	clansMapRanked, _ = webhook.NewWithURL(config.Instance.ClansMapRankedWebhook)
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
		SetThumbnail(QuaverLogo).
		SetFooter("Quaver", QuaverLogo).
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
		SetThumbnail(QuaverLogo).
		SetFooter("Quaver", QuaverLogo).
		SetTimestamp(time.Now()).
		SetColor(color).
		Build()

	_, err := rankingQueue.CreateMessage(discord.WebhookMessageCreate{
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
		votedBy += fmt.Sprintf("[%v](https://quavergame.com/user/%v)", voter.User.Username, voter.UserId)

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
		SetThumbnail(QuaverLogo).
		SetFooter("Quaver", QuaverLogo).
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
	if events == nil {
		return nil
	}

	description := "**A new order has been purchased!**"

	for _, order := range purchasedOrders {
		description += fmt.Sprintf("\n- %v", order.Description)
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("💰 New Order Incoming").
		SetDescription(description).
		AddField("Purchaser", fmt.Sprintf("[User Profile](https://quavergame.com/user/%v)", purchasedOrders[0].UserId), true).
		AddField("Receiver", fmt.Sprintf("[User Profile](https://quavergame.com/user/%v)", purchasedOrders[0].ReceiverUserId), true).
		SetThumbnail(QuaverLogo).
		SetFooter("Quaver", QuaverLogo).
		SetTimestamp(time.Now()).
		SetColor(0x00FF00).
		Build()

	_, err := events.CreateMessage(discord.WebhookMessageCreate{
		Embeds: []discord.Embed{embed},
	})

	if err != nil {
		logrus.Error("Failed to send order webhook: ", err)
		return err
	}

	return nil
}

func SendBackupWebhook(successful bool, failureError ...error) error {
	if events == nil {
		return nil
	}

	var title string
	var description string
	var color int

	if successful {
		title = "✅ Database Backup Complete!"
		description = "A database backup has been successfully created."
		color = 0x00FF00
	} else {
		title = "❌ Database Backup Failed!"
		description = fmt.Sprintf("Database backup failed.\n"+
			"Reason: \n"+
			"```%v```", failureError[0])
		color = 0xFF0000
	}

	embed := discord.NewEmbedBuilder().
		SetTitle(title).
		SetDescription(description).
		SetThumbnail(QuaverLogo).
		SetFooter("Quaver", QuaverLogo).
		SetTimestamp(time.Now()).
		SetColor(color).
		Build()

	msg := discord.WebhookMessageCreate{
		Embeds: []discord.Embed{embed},
	}

	if !successful {
		msg.Content = "@everyone"
	}

	_, err := events.CreateMessage(msg)

	if err != nil {
		logrus.Error("Failed to send backup webhook: ", err)
		return err
	}

	return nil
}

func SendSupervisorActivityWebhook(results map[*db.User]int, timeStart int64, timeEnd int64) error {
	if teamAnnounce == nil {
		return nil
	}

	description := fmt.Sprintf("Below, you can find last week's supervisor activity report. "+
		"1 week of donator has been automatically given to users with at least %v actions.\n\n",
		config.Instance.RankingQueue.WeeklyRequiredSupervisorActions)

	for user, actionCount := range results {
		var emoji string

		if actionCount >= config.Instance.RankingQueue.WeeklyRequiredSupervisorActions {
			emoji = "✅"
		} else {
			emoji = "❌"
		}

		description += fmt.Sprintf("- %v **%v**: %v\n", emoji, user.Username, actionCount)
	}

	link := fmt.Sprintf("[View](%v/ranking-queue/supervisors/actions?start=%v&end=%v)",
		config.Instance.WebsiteUrl, timeStart, timeEnd)

	embed := discord.NewEmbedBuilder().
		SetTitle("📝 Ranking Supervisor Activity Report").
		SetDescription(description).
		AddField("Detailed Report", link, false).
		SetThumbnail(QuaverLogo).
		SetFooter("Quaver", QuaverLogo).
		SetTimestamp(time.Now()).
		SetColor(0x49E6EF).
		Build()

	_, err := teamAnnounce.CreateMessage(discord.WebhookMessageCreate{
		Embeds: []discord.Embed{embed},
	})

	if err != nil {
		logrus.Error("Failed to send supervisor webhook: ", err)
		return err
	}

	return nil
}

func SendClanFirstPlaceWebhook(clan *db.Clan, mapQua *db.MapQua, newScore *db.ClanScore, oldScore *db.ClanScore) error {
	if clansFirstPlace == nil {
		return nil
	}

	embed := discord.NewEmbedBuilder().
		SetAuthor(fmt.Sprintf("%v | %v", clan.Tag, clan.Name),
			fmt.Sprintf("https://two.quavergame.com/clan/%v", newScore.ClanId), QuaverLogo).
		SetDescription("🏆 Achieved a new first place clan score!").
		AddField("Map", fmt.Sprintf("[%v](https://quavergame.com/mapset/map/%v)", mapQua, mapQua.Id), false).
		AddField("Overall Rating", fmt.Sprintf("%.2f", newScore.OverallRating), true).
		AddField("Overall Accuracy", fmt.Sprintf("%.2f%%", newScore.OverallAccuracy), true).
		SetImagef("https://cdn.quavergame.com/mapsets/%v.jpg", mapQua.MapsetId).
		SetThumbnail(QuaverLogo).
		SetFooter("Quaver", QuaverLogo).
		SetTimestamp(time.Now()).
		SetColor(0x00FF00)

	if oldScore != nil {
		embed.AddField("Previous #1 Holder", fmt.Sprintf("[%v](%v)", oldScore.Clan.Name, oldScore.ClanId), true)
	}

	_, err := clansFirstPlace.CreateMessage(discord.WebhookMessageCreate{
		Embeds: []discord.Embed{embed.Build()},
	})

	if err != nil {
		logrus.Error("Failed to send supervisor webhook: ", err)
		return err
	}

	return nil
}

func SendClanRankedWebhook(mapQua *db.MapQua) error {
	if clansMapRanked == nil {
		return nil
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("✅ New Map Clan Ranked!").
		SetDescription("A new map has been clan ranked and is now available to get scores on.").
		AddField("Song", fmt.Sprintf("[%v](https://quavergame.com/mapset/map/%v)", mapQua.String(), mapQua.Id), true).
		AddField("Creator", fmt.Sprintf("[%v](https://quavergame.com/user/%v)", mapQua.CreatorUsername, mapQua.CreatorId), true).
		AddField("Game Mode", enums.GetShorthandGameModeString(mapQua.GameMode), true).
		AddField("Difficulty ", fmt.Sprintf("%.2f", mapQua.DifficultyRating), true).
		SetImagef("https://cdn.quavergame.com/mapsets/%v.jpg", mapQua.MapsetId).
		SetThumbnail(QuaverLogo).
		SetFooter("Quaver", QuaverLogo).
		SetTimestamp(time.Now()).
		SetColor(0x00FF00).
		Build()

	_, err := clansMapRanked.CreateMessage(discord.WebhookMessageCreate{
		Embeds: []discord.Embed{embed},
	})

	if err != nil {
		logrus.Error("Failed to send clan ranked webhook")
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

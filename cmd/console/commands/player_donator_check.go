package commands

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/Quaver/api2/db"
	"github.com/Quaver/api2/enums"
	"github.com/Quaver/api2/internal/discord"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const discordPremiumDonatorDuration = time.Hour

var playerDonatorCheckLock sync.Mutex

var PlayerDonatorCheckCmd = &cobra.Command{
	Use:   "player:donator:check",
	Short: "Expires donator status and renews it for Discord premium members",
	Run: func(cmd *cobra.Command, args []string) {
		if !playerDonatorCheckLock.TryLock() {
			logrus.Info("Donator check is already running")
			return
		}
		defer playerDonatorCheckLock.Unlock()

		ctx := context.Background()
		if cmd != nil {
			ctx = cmd.Context()
		}

		client := discord.NewClient(&http.Client{})
		if err := checkPlayerDonators(ctx, client); err != nil {
			logrus.WithError(err).Error("Donator check failed")
		}
	},
}

func checkPlayerDonators(ctx context.Context, client *discord.Client) error {
	cutoff := time.Now().UnixMilli()
	var users []db.User
	premiumLookupAvailable := true
	return db.SQL.WithContext(ctx).
		Session(&gorm.Session{SkipHooks: true}).
		Select("id", "usergroups", "donator_end_time", "discord_id").
		Where("(usergroups & ? <> 0 AND donator_end_time < ?) OR (usergroups & ? = 0 AND discord_id IS NOT NULL)",
			enums.UserGroupDonator, cutoff, enums.UserGroupDonator).
		FindInBatches(&users, 1000, func(batch *gorm.DB, _ int) error {
			for _, user := range users {
				if err := ctx.Err(); err != nil {
					return err
				}
				if err := checkPlayerDonator(ctx, batch, client, user, cutoff, &premiumLookupAvailable); err != nil {
					logrus.WithError(err).WithField("user_id", user.Id).Error("Could not update donator status")
				}
			}
			return nil
		}).Error
}

func checkPlayerDonator(ctx context.Context, database *gorm.DB, client *discord.Client, user db.User, cutoff int64, premiumLookupAvailable *bool) error {
	removeRole := false
	if *premiumLookupAvailable && user.DiscordId != nil && discord.ValidID(*user.DiscordId) {
		hasPremium, err := client.HasPremiumRole(ctx, *user.DiscordId)
		switch {
		case errors.Is(err, discord.ErrMemberNotFound):
		case err != nil:
			*premiumLookupAvailable = false
			logrus.WithError(err).Warn("Discord premium lookup unavailable for this check")
		case hasPremium:
			return renewDiscordPremiumDonator(database, user, cutoff)
		default:
			removeRole = true
		}
	}

	if !enums.HasUserGroup(user.UserGroups, enums.UserGroupDonator) {
		return nil
	}

	expired, err := expirePlayerDonator(database, user, cutoff)
	if err != nil || !expired || !removeRole {
		return err
	}

	if err := client.RemoveDonatorRole(ctx, *user.DiscordId); err != nil && !errors.Is(err, discord.ErrMemberNotFound) {
		logrus.WithError(err).WithField("user_id", user.Id).Warn("Could not remove Discord donator role after expiry")
	}
	return nil
}

func renewDiscordPremiumDonator(database *gorm.DB, user db.User, cutoff int64) error {
	result := database.Model(&db.User{}).
		Where("id = ? AND discord_id = ?", user.Id, *user.DiscordId).
		Where("(usergroups & ? = 0 OR donator_end_time < ?)", enums.UserGroupDonator, cutoff).
		Updates(map[string]any{
			"usergroups":       gorm.Expr("usergroups | ?", enums.UserGroupDonator),
			"donator_end_time": time.Now().Add(discordPremiumDonatorDuration).UnixMilli(),
		})
	if result.RowsAffected > 0 {
		logrus.WithField("user_id", user.Id).Info("Granted one hour of donator for Discord premium membership")
	}
	return result.Error
}

func expirePlayerDonator(database *gorm.DB, user db.User, cutoff int64) (bool, error) {
	expired := false
	err := database.Transaction(func(tx *gorm.DB) error {
		var current db.User
		err := tx.Select("id", "discord_id").
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ? AND usergroups & ? <> 0 AND donator_end_time < ?", user.Id, enums.UserGroupDonator, cutoff).
			Take(&current).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		if err != nil {
			return err
		}

		// Recheck on the next run if the linked account changed during the Discord lookup.
		if !sameDiscordAccount(current.DiscordId, user.DiscordId) {
			return nil
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"usergroups":       gorm.Expr("usergroups & ~?", enums.UserGroupDonator),
			"donator_end_time": 0,
		}).Error; err != nil {
			return err
		}

		notification := db.NewDonatorExpiredNotification(user.Id)
		notification.Timestamp = time.Now().UnixMilli()
		if err := tx.Create(notification).Error; err != nil {
			return err
		}
		expired = true
		return nil
	})
	return expired, err
}

func sameDiscordAccount(current, previous *string) bool {
	if current == nil || previous == nil {
		return current == previous
	}
	return *current == *previous
}

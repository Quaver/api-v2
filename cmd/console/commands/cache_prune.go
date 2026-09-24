package commands

import (
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/files"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var CachePruneCmd = &cobra.Command{
	Use:   "cache:prune",
	Short: "Finds cached files unused for 90 days",
	RunE: func(cmd *cobra.Command, args []string) error {
		deleteFiles, err := cmd.Flags().GetBool("delete")
		if err != nil {
			return err
		}
		return RunCachePrune(deleteFiles)
	},
}

func init() {
	CachePruneCmd.Flags().Bool("delete", false, "delete old cached files instead of performing a dry run")
}

func RunCachePrune(deleteFiles bool) error {
	logrus.Infof("Cache prune scanning %s", config.Instance.Cache.DataDirectory)
	result, err := files.PruneCache(deleteFiles)
	megabytes := result.Bytes / 1_000_000

	if deleteFiles {
		logrus.Infof("Cache prune deleted %d files (%d MB)", result.Files, megabytes)
	} else {
		logrus.Infof("Cache prune dry run found %d files (%d MB)", result.Files, megabytes)
	}

	if err != nil {
		logrus.Errorf("Error pruning cached files: %v", err)
	}
	return err
}

package main

import (
	"flag"
	"github.com/Quaver/api2/cmd/console/commands"
	"github.com/Quaver/api2/config"
	"github.com/Quaver/api2/db"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	configPath := flag.String("config", "../../config.json", "path to config file")
	flag.Parse()

	if err := config.Load(*configPath); err != nil {
		logrus.Panic(err)
	}

	db.ConnectMySQL()
	db.InitializeRedis()
	db.InitializeElasticSearch()

	c := cron.New()
	jobs := config.Instance.Cron

	registerCronJob(c, jobs.DonatorCheck.Job, func() { commands.WeeklyMostPlayedMapsetsCmd.Run(nil, nil) })
	registerCronJob(c, jobs.ElasticIndexMapsets.Job, func() { commands.ElasticIndexMapsets.Run(nil, nil) })
	registerCronJob(c, jobs.WeeklyMostPlayed.Job, func() { commands.WeeklyMostPlayedMapsetsCmd.Run(nil, nil) })

	c.Start()

	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM)
	<-quitChannel

	logrus.Info("Exiting...")
}

func registerCronJob(c *cron.Cron, job config.Job, f func()) {
	if !job.Enabled {
		logrus.Warningf("Ignoring job: %v, as it is disabled", job.Name)
		return
	}

	c.AddFunc(job.Schedule, func() { f() })
	logrus.Infof("Registered job: `%v` on schedule: `%v`", job.Name, job.Schedule)
}

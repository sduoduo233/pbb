package cron

import (
	"log/slog"

	"github.com/go-co-op/gocron/v2"
	"github.com/sduoduo233/pbb/update"
)

func Init() {
	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	s.NewJob(gocron.CronJob("*/10 * * * *", false), gocron.NewTask(sample))
	s.NewJob(gocron.CronJob("*/30 * * * *", false), gocron.NewTask(clean))
	s.NewJob(gocron.CronJob("*/30 * * * * *", true), gocron.NewTask(incident))
	s.NewJob(gocron.CronJob("0 3 * * *", false), gocron.NewTask(func() {
		err := update.AutoUpdate("https://dl.exec.li/install-hub.sh")
		if err != nil {
			slog.Error("auto update error", "error", err)
		}
	}))

	s.Start()
}

package cron

import "github.com/go-co-op/gocron/v2"

func Init() {
	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	s.NewJob(gocron.CronJob("*/1 * * * *", false), gocron.NewTask(sample))
	s.NewJob(gocron.CronJob("*/30 * * * *", false), gocron.NewTask(clean))

	s.Start()
}

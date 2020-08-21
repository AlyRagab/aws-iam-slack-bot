package main

import (
	"iambot/bot"
	"iambot/config"
	"iambot/provider"
	"os"

	"github.com/robfig/cron"
)

var (
	slackClient bot.SlackClient
	svc         provider.SVC
)

func init() {
	svc.InitSession()
	slackClient.CreateSlackClient(os.Getenv("SLACK_ACCESS_TOKEN"), svc)
	config.DB = config.ConnectMySQL()
}

func run() {
	c := cron.New()
	c.AddFunc("0 5 * * *", func() {
		users := svc.ListAWSUsersWithGroups()
		slackClient.SendToSlack(users, "warning")
	})
	c.AddFunc("0 6 * * *", func() {
		svc.Revoke(slackClient.SlackChannel)
	})
	c.Start()
	slackClient.RespondToEvents()
}
func main() {

	run()
}

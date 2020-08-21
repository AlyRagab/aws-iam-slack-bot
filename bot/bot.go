package bot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"iambot/config"
	"iambot/provider"

	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/slack-go/slack"
)

const (
	helpMessage = `[Usage] ::.. A warning will be sent every 4h if you don't press keep <username> then after 1h all users FA access will be revoked`
)

// SlackClient for init sessions
type SlackClient struct {
	SlackClient  *slack.RTM
	SlackChannel string
	svc          provider.SVC
}

// SlackRequestBody which will sent to the channel
type slackRequestBody struct {
	Text  string `json:"text"`
	Users string `json:"users"`
}

// CreateSlackClient connection by bot access token
func (slackClient *SlackClient) CreateSlackClient(token string, svc provider.SVC) {
	slackClient.SlackClient = slack.New(token).NewRTM()
	slackClient.SlackChannel = os.Getenv("SLACK_CHANNEL_WEBHOOK")
	slackClient.svc = svc
	go slackClient.SlackClient.ManageConnection()
}

// BuildText to be sent to Slack
func buildText(AWSUsers []provider.AWSUser) string {
	var text string
	if len(AWSUsers) == 0 {
		text = fmt.Sprint("There are no Users in the FA_Groups")
		return text
	}
	text = fmt.Sprint("<@channel> The below users will be removed from the Admin_Groups: \n ")
	for _, awsUser := range AWSUsers {
		text += fmt.Sprintf("User : \t %s will be removed from the following Groups:\t %v \n ", awsUser.UserName, awsUser.Groups)
	}

	return text
}

// SendToSlack function
func (slackClient *SlackClient) SendToSlack(usersGroups []provider.AWSUser, option string) {
	var data slackRequestBody
	switch option {
	case "warning":
		data.Text = buildText(provider.FilterUsersGroups(usersGroups))
	}

	// Slack Request Body to send to the SlackAPI
	slackBody, _ := json.Marshal(data)
	req, err := http.NewRequest(http.MethodPost, slackClient.SlackChannel, bytes.NewBuffer(slackBody))
	if err != nil {
		log.Printf("Slack Request failed %s ", err.Error())
	}

	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Sending to Slack timed out %s ", err.Error())
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	if buf.String() != "ok" {
		log.Println("Non-ok response returned from Slack")
	}
}

// RespondToEvents interact with events and mentioning and tags:
func (slackClient *SlackClient) RespondToEvents() {
	for msg := range slackClient.SlackClient.IncomingEvents {
		log.Println("Event Received ..")
		switch event := msg.Data.(type) {
		case *slack.MessageEvent:
			botTagString := fmt.Sprintf("<@%s>", slackClient.SlackClient.GetInfo().User.ID)
			if strings.Contains(event.Msg.Text, botTagString) {
				continue
			}
			message := strings.Replace(event.Msg.Text, botTagString, "", -1)
			slackClient.botResponse(message, event.Channel)
			//slackClient.revokeAll(message, event.Channel)
		}
	}
}

// botResponse for both keep and help
func (slackClient *SlackClient) botResponse(message, slackChannel string) {
	// Sending HelpMessage
	if strings.ToLower(message) == "help" {
		slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage(helpMessage, slackChannel))
	}
	// Keeping users
	keepMessage := strings.Split(message, " ")
	if keepMessage[0] == "keep" {
		// Checking if there are warned users or no , if yes then will allow writing keep, else so nothing !!
		rows, err := config.DB.Query(`SELECT * FROM iambot`)
		if err != nil {
			log.Printf("Select Query is not succeeded %s", err.Error())
		}
		if rows.Next() {
			for _, user := range keepMessage[1:] {
				_, err := config.DB.Query(`UPDATE iambot SET keep = ? WHERE username = ?`, 1, user)
				if err != nil {
					log.Printf("Update Query table is not succeeded %s", err.Error())
				}
			}
			slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage("User/s kept successfully .. ", slackChannel))
		} else {
			slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage("Please Wait till the Warning message is sent :) ", slackChannel))
		}
	}
	// Assigning an Access to a User
	var user, group string
	assignMessage := strings.Split(message, " ")

	if assignMessage[0] == "assign" {
		user = assignMessage[1]
		group = assignMessage[2]
		//userGroups := slackClient.svc.ListAWSUsersWithGroups()
		// Check if the user and group/s are existed
		isExist, _ := slackClient.svc.ListAllAdminGroups()
		//for _, users := range isExist. {
		//	if users.UserName == user {
		for _, groups := range isExist.Groups {
			if *groups.GroupName == group {
				input := &iam.AddUserToGroupInput{
					GroupName: &group,
					UserName:  &user,
				}
				_, err := slackClient.svc.Svc.AddUserToGroup(input)
				if err != nil {
					slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage("Not Valid User ...", slackChannel))
					return
				}
				slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage("Access Assigned Successfully ...", slackChannel))
				return
			}
		}
		slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage("Not Valid Group ...", slackChannel))
	}

	// Revoke an Access to a User
	revokeMessage := strings.Split(message, " ")
	if revokeMessage[0] == "revoke" {
		revokeUser := revokeMessage[1]
		revokeGroup := revokeMessage[2]
		removeInput := &iam.RemoveUserFromGroupInput{
			GroupName: &revokeGroup,
			UserName:  &revokeUser,
		}
		_, err := slackClient.svc.Svc.RemoveUserFromGroup(removeInput)
		if err != nil {
			log.Println(err)
		}
		slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage("Access Revoked Successfully ...", slackChannel))
		return
	}

	// Revoke all accesses from user/s
	// List all groups for the user/s
	revokeAllMessage := strings.Split(message, " ")
	if revokeAllMessage[0] == "revoke-all-from" {
		toRevoke := slackClient.svc.ListAWSUsersWithGroups()
		username := revokeAllMessage[1]
		for _, u := range toRevoke {
			if username == u.UserName {
				for _, group := range u.Groups {
					revokeAction := &iam.RemoveUserFromGroupInput{
						GroupName: &group,
						UserName:  &u.UserName,
					}
					_, err := slackClient.svc.Svc.RemoveUserFromGroup(revokeAction)
					if err != nil {
						log.Printf("Failed to Remove User: %s from Group/s: %s Error %s", u.UserName, group, err.Error())
						slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage("Can not Revoked Access from User: "+u.UserName, slackChannel))
					}
				}
				slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage("Access Revoked Successfully from User: "+u.UserName, slackChannel))
				return
			}
		}
		slackClient.SlackClient.SendMessage(slackClient.SlackClient.NewOutgoingMessage("Wrong User : "+username, slackChannel))
	}
}

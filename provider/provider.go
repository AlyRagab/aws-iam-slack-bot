/*Package provider is meant to :
- List all Users with their associated group
- List all IAM users
- List all "admin" groups
- Revoke access "RemoveUser/sFromGroup/s"
- Assign access "AddUserToGroup"
*/
package provider

import (
	"fmt"
	"iambot/config"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

// AWSUser struct for slack API
type AWSUser struct {
	UserName string
	Groups   []string
}

// SVC struct with AWS
type SVC struct {
	Svc *iam.IAM
}

// InitSession function for creating the session with AWS
func (svc *SVC) InitSession() {
	sess, _ := session.NewSession(&aws.Config{Region: aws.String(os.Getenv("AWS_REGION"))})
	svc.Svc = iam.New(sess)
	return
}

// ListAllAdminGroups func
func (svc *SVC) ListAllAdminGroups() (*iam.ListGroupsOutput, error) {
	input := &iam.ListGroupsInput{}
	result, err := svc.Svc.ListGroups(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case iam.ErrCodeServiceFailureException:
				fmt.Println(iam.ErrCodeServiceFailureException, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return nil, err
	}
	return result, nil
}

// ListAWSUsersWithGroups IAM users:
func (svc *SVC) ListAWSUsersWithGroups() []AWSUser {
	result, _ := svc.Svc.ListUsers(&iam.ListUsersInput{
		MaxItems: aws.Int64(20),
	})
	users := make([]AWSUser, 0)
	for _, user := range result.Users {
		result, err := svc.Svc.ListGroupsForUser(&iam.ListGroupsForUserInput{
			MaxItems: aws.Int64(20),
			UserName: user.UserName,
		})
		if err != nil { //&& strings.Contains(err.Error(), iam.ErrCodeNoSuchEntityException) {
			log.Printf("User is not Existed %s", err.Error())
		}
		users = append(users, AWSUser{UserName: *user.UserName,
			Groups: getUsersGroups(result.Groups)})
	}

	return users
}

func getUsersGroups(groups []*iam.Group) []string {
	var g []string
	for _, group := range groups {
		g = append(g, *group.GroupName)
	}
	return g
}

// FilterUsersGroups is used to be revoked from users
func FilterUsersGroups(usersGroups []AWSUser) []AWSUser {
	// Checking any group staring with admin
	regex := regexp.MustCompile("^admin")

	var usersGroupsToBeRevoked []AWSUser
	for _, user := range usersGroups {
		groupSlice := make([]string, 0)
		for _, userGroup := range user.Groups {
			if regex.MatchString(userGroup) {
				groupSlice = append(groupSlice, userGroup)
			}
		}
		user.Groups = groupSlice
		if len(user.Groups) != 0 {
			_, err := config.DB.Query(`INSERT INTO iambot (username, groupname, keep) VALUES (?, ?, ?)`, user.UserName, strings.Join(user.Groups, ","), 0)
			if err != nil {
				log.Printf("Insert into table failed %s", err.Error())
			}
			usersGroupsToBeRevoked = append(usersGroupsToBeRevoked, user)
		}
	}

	return usersGroupsToBeRevoked
}

// Revoke users as input from Slack
func (svc *SVC) Revoke(slackChannel string) {
	var username, groupname string
	rows, err := config.DB.Query(`SELECT username, groupname FROM iambot WHERE keep = 0`)
	if err != nil {
		log.Printf("Select statement failed %s", err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&username, &groupname)
		if err != nil {
			log.Printf("Scanning rows failed %s", err.Error())
		}
		groups := strings.Split(groupname, ",")
		for _, group := range groups {
			revokeAction := &iam.RemoveUserFromGroupInput{
				GroupName: &group,
				UserName:  &username,
			}
			_, err := svc.Svc.RemoveUserFromGroup(revokeAction)
			if err != nil {
				log.Printf("Failed to Remove User: %s from Group: %s Error %s", username, group, err.Error())
			}
		}
	}
	// Truncate the table to be prepared to the next warning round again :)
	_, err = config.DB.Query(`TRUNCATE TABLE iambot`)
	if err != nil {
		log.Printf("Failed to Truncate %s", err.Error())
	}
}

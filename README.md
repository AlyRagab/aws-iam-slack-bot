# AWS Access Management by Slack Bot

## Slack-Bot is Managing AWS Access process as below : 
- A warning message will be sent every day at 5PM listing all user/s with their associated group/s.
- The revoke action will be done every day at 6PM.

### Actions :
- `keep <username>` or `keep <username1> <username2>` to keep the access to user/s between the warning message time and the revoke which is probably one hour
- `assign <username> <groupname>` to assign an access to a user at any time in the day.
- `revoke <username> <groupname>` to revoke specific access from a user at any time in the day.
- `revoke-all-from <username>` to revoke all accesses from a user at any time in the day.

### Notes:
- The AWS Access Management is a group based access , so considering creating a group per AWS resource and attach the related access policy document to this particular group.
- A limitation from AWS that the groups should not have any numbers as per this [Reference](https://docs.aws.amazon.com/IAM/latest/APIReference/API_AddUserToGroup.html)
- The Bot is removing users from admin groups by checking "^admin" regex only like `admin_dns , admin_RDS , admin_compute ..etc`.
- After the Revoke action is done , It lists all users again with their current group/s.
- When writing the command `keep <username>`or `keep <username1> <username2>` , it updates a column called `keep` with (1) integer which means to keep.
- All users which ready to be revoked has the column “keep“ with value 0 by default.
- Value 0 == revoke , 1 == keep
- By the end of the revoke function execution , it will truncate the content of the tables , to be ready to get inserted by the warning again.

### Prerequisites :
- The table should be created as below as a prerequisites:
```
create table iambot (
    username varchar(60),
    groupname varchar(100),
    keep int NOT NULL,
    PRIMARY KEY (username)
);
```
- Set the below environment variables as below:
```
SLACK_CHANNEL_WEBHOOK
SLACK_ACCESS_TOKEN
DB_NAME
DB_PORT
DB_PASSWORD
DB_ADDRESS
DB_USER
AWS_REGION
```
- you probably can access the AWS CLI by the `aws configure` command before.
- As recommended you just need to set the proper access permissions the bot just needs as below:
```
- iam:List
- iam:AddUserToGroup
- iam:RemoveUserFromGroup
- iam:Get 
```
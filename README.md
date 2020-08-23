# IAM Slack-bot

## Slack-Bot is Managing AWS Access process as below : 
- Warning message will be sent every day at 5PM.
- If we write `keep <username>` or `keep <username1> <username2>` , the user/s will be kept till next day.
- The revoke action will be done every day at 6PM.
- At any time we can revoke all accesses by the command `revoke-all-from <username>`.
- At any time we can assign a specific access to a user by the command `assign <username> <groupname>`
- At any time we can revoke a specific permission from a user by `revoke <username> <groupname>`

- A limitation from AWS that the groups should not have any numbers as per this [Reference](https://docs.aws.amazon.com/IAM/latest/APIReference/API_AddUserToGroup.html)
- The Bot is removing users from admin groups by checking "^admin" regex only like `admin_dns , admin_RDS , admin_compute ..etc`.
- After the Revoke action is done , It lists all users again with their current group/s.
- When writing the command `keep <username>`or `keep <username1> <username2>` , it updates a column called `keep` with (1) integer which means to keep.
- All users which ready to be revoked has the column “keep“ with value 0 by default.
- Value 0 == revoke , 1 == keep
- By the end of the revoke function execution , it will truncate the content, to be ready to get inserted by the warning again.

### The table should be created as below:

```
create table iambot (
    username varchar(60),
    groupname varchar(100),
    keep int NOT NULL,
    PRIMARY KEY (username)
);
```

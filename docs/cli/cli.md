## leopards 

A tool for ORM database struct generate.


## 命令帮助

```shell
leopards -h
An funny tool for DB schema

Usage:
  leopards [flags]
  leopards [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  mysql       An MySQL schema generate tool for leopards

Flags:
  -h, --help   help for leopards

Use "leopards [command] --help" for more information about a command.
```

## 结构体内嵌

> [!WARNING]
> 只支持结构体内嵌，不支持结构体指针内嵌

```go

type User struct {
	Id int `json:"id"`
	Name string `json:"name"`
}

type UserMobile struct {
	User
	Mobile string `json:"mobile"`
}


var users []UserMobile
err := db.Query().Select(`id`, `name`, `mobile`).From(`users`).Where(leopards.LTE(`id`, 10)).Scan(context.TODO(), &users)
if err != nil {
	panic(err)
}

for _, user := range users {
	println(user.Id, user.Name, user.Mobile)
}


``` 
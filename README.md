## leopards 

An ORM for Goer, forking from facebook's `ent` framework.


## leopards command

leopards 工具用来生成操作数据库的表结构信息, 安装命令：

```shell
go install github.com/liqiongfan/leopards/cmd/leopards@latest
```

> [!TIP]
> Table struct and table name should be maintained by leopards. 

[click to view docs](https://github.com/liqiongfan/leopards/tree/main/cmd/leopards)


## ORM实例

```go

package main

import (
	"context"
	
	"github.com/liqiongfan/leopards"
	
	"time"
)

// User 表结构, 理论上表结构应该使用 leopards 命令生成，避免维护问题
type User struct {
    Id int `json:"id"`
	Name string `json:"name"`
	Age int `json:"age"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func main() {
	orm, err := leopards.OpenOptions{
		User:     "用户名",
		Password: "密码",
		Host:     "账号",
		Port:     "3306",
		Database: "数据库名",
		Debug:    true, // 是否开启调试，开启调试会输出SQL到标准输出
		Dialect:  leopards.MySQL,
	}.Open()
	if err != nil {
        panic(err)
	}
	
	users := make([]User, 0, 10)
	err = orm.Query().Select().From(`user`).Scan(context.TODO(), &users)
	if err != nil {
        panic(err)
	}
	
	for _, user := range users {
		println(`ID:`, user.Id, ` Name:`, user.Name, ` Age:`, user.Age)
    }
}

```

## 章节明细

+ [leopards cli command](docs/cli/cli.md)
+ [SQL query statement](docs/query/query.md)
+ [SQL delete statement](docs/delete/delete.md)
+ [SQL insert statement](docs/insert/insert.md)
+ [SQL update statement](docs/update/update.md)
+ [interceptors](docs/interceptors/interceptors.md)


## 








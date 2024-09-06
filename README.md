## leopards 

An ORM for Goer, forking from facebook's `ent` framework.


```Open title="hello.go"

	db, err := leopards.OpenOptions{
		User:     "用户名",
		Password: "密码",
		Host:     "账号",
		Port:     "3306",
		Database: "数据库名",
		Debug:    true, // 是否开启调试，开启调试会输出SQL到标准输出
		Dialect:  leopards.MySQL,
	}.Open()

```


### 中间件

![img.png](img.png)


+ InterceptorsQuery 添加一个前置查询中间件
+ InterceptorsAfterQuery 添加一个后置查询中间件
+ InterceptorsInsert 添加一个前置插入中间件
+ InterceptorsAfterInsert 添加一个后置插入中间件
+ InterceptorsUpdate 添加一个前置更新中间件
+ InterceptorsAfterUpdate 添加一个后置更新中间件
+ InterceptorsDelete 添加一个前置删除中间件
+ InterceptorsAfterDelete 添加一个后置删除中间件


### 增删改查

+ Query  查询
+ Update 更新
+ Insert 插入
+ Delete 删除


### 实例

```go
dest := make([]User, 0, 30)
err := db.Query().Select(`id`, `count(*) as count`).From(`user`).Scan(context.TODO(), &dest)
```





















## leopards

A tool for generate database struct for database.

+ [x] MySQL      yes
+ [x] Postgres   yes
+ [ ] SQLite     planing

### MySQL

if you want to set database connection info into environment, leopards will read then from env.

+ MYSQL_USER
  用户名

+ MYSQL_PASSWORD
  密码

+ MYSQL_HOST
  主机地址

+ MYSQL_PORT
  端口


otherwise you can set this info from cli. such as:

```shell
leopards mysql --host=xxx --port=xxx --user=xxx --pasword=xxx database tables --out=指定输出目录和文件
```


### PostgresSQL

if you want to set database connection info into environment, leopards will read then from env.

+ PG_USER
  用户名

+ PG_PASSWORD
  密码

+ PG_HOST
  主机地址

+ PG_PORT
  端口


otherwise you can set this info from cli. such as:

```shell
leopards postgres --host=xxx --port=xxx --user=xxx --pasword=xxx database tables --out=指定输出目录和文件
```

#### 表

`tables` 参数支持逗号分隔，或者 `*` 标识全部表


> [!TIP]
> Linux & mac 下 * 需要用 \* 替代 以免展开
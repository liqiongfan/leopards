## leopards

A tool for generate database struct for database.

+ [x] MySQL      yes
+ [ ] Postgres   planing
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


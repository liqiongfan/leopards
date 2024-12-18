## leopards SQL select 帮助手册

### WHERE 语句

```go

err := orm.Query().Select().From(`user`).Where(*Predicate).Scan(context.TODO(), &dest)

```

#### 1. WhereMap

+ WhereMap(map[string]any)

```go
WhereMap(map[string]any{`id`: 10, `age`: 28})
```

#### 2. Predicate 可选项

+ NotExists(Query): WHERE NOT EXISTS (select * from x )

```go
leopards.NotExists(orm.Query().From(`mobile`).Select())
```

+ ExprP(expr, args...):

```go
leopards.ExprP(`id > ? and id < ?`, 5, 10)
```

+ IsNull(col)

```go
leopards.IsNull(`mobile`)
```

+ Not(*Predicate)

```go
leopards.Not(leopards.EQ(`id`, 10))
```

+ NotNull(col)

```go
leopards.NotNull(`mobile`)
```

+ In(col, args...)

```go
leopards.In(`id`, []any{1,2,3})
```

+ EQ(col, value)

```go
leopards.EQ(`id`, 1)
```

+ Contains(col, value)

```go
leopards.Contains(`name`, `LL`)
```

+ LT(col, value)

```go
leopards.LT(`id`, 10)
```

+ Between(col, v1, v2)

```go
leopards.Between(`id`, 10, 20)
```

+ LTE(col, value)

```go
leopards.LTE(`id`, 10)
```

+ GTE(col, value)

```go
leopards.GTE(`id`, 10)
```

+ GT(col, value)

```go
leopards.GT(`id`, 10)
```

+ HasPrefix(col, value)

```go
leopards.HasPrefix(`name`, `Y`)
```

+ NEQ(col, value)

```go
leopards.NEQ(`name`, `Y`)
```

+ IsFalse(col)

```go
leopards.IsFalse(`name`)
```

+ IsTrue(col)

```go
leopards.IsTrue(`name`)
```

+ NotIn(col, args...)

```go
leopards.NotIn(`name`, []any{`aa`, `bb`})
```








### GROUP BY 语句

+ GroupBy(columns...)

```go
GroupBy(`id`, `DATE_FORMAT("%Y-%m-%d")`)
```

### ORDER BY 语句

+ OrderBy(columns)

```go
OrderBy(leopards.Asc(`id`))
```

### HAVING 语句

+ Having(*Predicate)

```go
用法与 `Where` 一致
```

### LIMIT 语句

+ Limit(count)

```go
Limit(10)
```

### OFFSET 语句

+ Offset(count)

```go
Offset(10)
```
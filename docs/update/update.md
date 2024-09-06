## leopards SQL update 帮助手册

### Set(column, value)

```go
Set(`id`, 10).Set(`name`, `Go`)
```

### SetMap(kv map[string]any)

```go
SetMap(map[string]any{
	`id`: 100,
	`name`: `Go`
})
```


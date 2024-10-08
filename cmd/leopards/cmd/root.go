package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

type Info struct {
	User, Password, Host, Port, Charset string
}

func enum(columnName *string, dataType *string, columnType string) string {
	if columnName == nil || dataType == nil {
		return ``
	}

	w := strings.Builder{}

	switch *dataType {
	case `enum`:
		columnType = columnType[5 : len(columnType)-1]
	case `set`:
		columnType = columnType[4 : len(columnType)-1]
	default:
		return w.String()
	}

	columnType = strings.ReplaceAll(columnType, `'`, ``)

	w.WriteString("\n" + `type ` + camel(columnName) + `Type ` + Type(dataType, columnType, `NO`))
	w.WriteString("\n")
	w.WriteString("const (\n")
	for _, name := range strings.Split(columnType, `,`) {
		w.WriteString("\t" + camel(columnName) + camel(&name) + ` ` + camel(columnName) + `Type = "` + name + `"`)
		w.WriteString("\n")
	}
	w.WriteString(")\n")

	return w.String()
}

func snake(str string) string {

	first := true
	b := strings.Builder{}
	for _, c := range str {
		switch {
		case c >= 'A' && c <= 'Z' && !first:
			b.WriteRune('_')
			b.WriteRune('_')
		}
		b.WriteRune(c)
		first = false
	}

	return b.String()
}

func camel(str *string) string {
	v := ``
	if str != nil {
		v = *str
	}
	var b strings.Builder
	var upperNext = true
	for _, c := range v {
		switch {
		case c == '！' || c == '￥' || c == '…' || c == '（' || c == '）' || c == '【' || c == '】' ||
			c == '、' || c == '？' || c == '《' || c == '》' || c == '“' || c == '：':
			continue
		case c == '_' && upperNext:
			upperNext = false
		case c == '_' && !upperNext:
			upperNext = true
			continue
		case upperNext:
			if c >= 'a' && c <= 'z' {
				c ^= 0x20
			}
			upperNext = false
		}
		b.WriteRune(c)
	}

	return b.String()
}

func Type(typ *string, columnType, isNullable string) string {
	tn := ``
	if typ != nil {
		tn = *typ
	}
	r := ``
	switch tn {
	case `bit`:
		r = `[]byte`

	case `tinyint`:
		r = `int8`
	case `int`:
		r = `int32`
	case `smallint`:
		r = `int16`
	case `mediumint`:
		r = `int32`
	case `bigint`:
		r = `int64`

	case `float`:
		r = `float32`
		if strings.Contains(columnType, `unsigned`) {
			r = `float64`
		}
		return r
	case `double`:
		r = `float64`
		return r
	case `decimal`:
		r = `string`
		return r
	case `date`:
		r = `time.Time`
	case `year`:
		r = `string`
	case `time`:
		r = `string`
	case `timestamp`:
		r = `time.Time`
	case `datetime`:
		r = `time.Time`

	case `enum`:
		r = `string`
	case `set`:
		r = `string`

	case `char`:
		r = `string`
	case `varchar`:
		r = `string`
	case `text`:
		r = `string`
	case `tinytext`:
		r = `string`
	case `smalltext`:
		r = `string`
	case `mediumtext`:
		r = `string`
	case `longtext`:
		r = `string`

	case `json`:
		r = `string`

	case `blob`:
		r = `[]byte`
	case `longblob`:
		r = `[]byte`
	default:
		panic(`unknown type:` + tn)
	}

	if strings.Contains(columnType, `unsigned`) {
		r = `u` + r
	}

	if isNullable == `YES` {
		r = `*` + r
	}
	return r
}

func tag(name, comment string, padLength int) string {
	for _, c := range "\r\n" {
		comment = strings.ReplaceAll(comment, string(c), ` `)
	}
	padLength -= len(name)
	padStr := strings.Repeat(` `, padLength)
	return fmt.Sprintf("`json:\"%s\"` %s// %s", name, padStr, comment)
}

func tagPtr(name string, comment *string, padLength int) string {
	newComment := ``
	if comment != nil {
		for _, c := range "\r\n" {
			*comment = strings.ReplaceAll(*comment, string(c), ` `)
		}
		newComment = *comment
	}
	padLength -= len(name)
	padStr := strings.Repeat(` `, padLength)
	return fmt.Sprintf("`json:\"%s\"` %s// %s", name, padStr, newComment)
}

func pad(name string, length int) string {
	if length-len(name) <= 0 {
		return ``
	}
	return strings.Repeat(` `, length-len(name))
}

func trace(err error) error {
	if err != nil {
		fmt.Printf("trace: %v\n", err)
	}
	return nil
}

func getPGInfo(cmd *cobra.Command) *Info {
	i := &Info{
		User:     os.Getenv(`PG_USER`),
		Password: os.Getenv(`PG_PASSWORD`),
		Host:     os.Getenv(`PG_HOST`),
		Port:     os.Getenv(`PG_PORT`),
		Charset:  os.Getenv(`PG_CHARSET`),
	}

	flags := []string{`user`, `password`, `host`, `port`, `charset`}
	for _, flag := range flags {
		v, _ := cmd.Flags().GetString(flag)
		switch flag {
		case `user`:
			i.User = v
		case `password`:
			i.Password = v
		case `host`:
			i.Host = v
		case `port`:
			i.Port = v
		case `charset`:
			i.Charset = v
		}
	}

	return i
}

func getInfo(cmd *cobra.Command) *Info {
	r := &Info{
		User:     os.Getenv(`MYSQL_USER`),
		Password: os.Getenv(`MYSQL_PASSWORD`),
		Host:     os.Getenv(`MYSQL_HOST`),
		Port:     os.Getenv(`MYSQL_PORT`),
		Charset:  os.Getenv(`MYSQL_CHARSET`),
	}

	u, _ := cmd.Flags().GetString(`user`)
	if u != `` {
		r.User = u
	}

	p, _ := cmd.Flags().GetString(`password`)
	if p != `` {
		r.Password = p
	}

	h, _ := cmd.Flags().GetString(`host`)
	if h != `` {
		r.Host = h
	}

	port, _ := cmd.Flags().GetString(`port`)
	if port != `` {
		r.Port = port
	}

	charset, _ := cmd.Flags().GetString(`charset`)
	if port != `` {
		r.Charset = charset
	}

	return r
}

var RootCMD = &cobra.Command{
	Use:   `leopards`,
	Short: `A funny tool for DB schema`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	RootCMD.AddCommand(mysqlCMD)
	RootCMD.AddCommand(postgresCMD)
}

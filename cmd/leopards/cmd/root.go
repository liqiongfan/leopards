package cmd

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
)

type Info struct {
	User, Password, Host, Port string
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

func camel(str *string) string {
	v := ``
	if str != nil {
		v = *str
	}
	var upper = true
	var b strings.Builder
	for _, c := range v {
		switch {
		case unicode.IsLetter(c):
			switch upper {
			case true:
				switch {
				case c >= 'a' && c <= 'z':
					b.WriteRune(c ^ 0x20)
				default:
					b.WriteRune(c)
				}
				upper = false
			default:
				b.WriteRune(c)
			}
		case unicode.IsDigit(c):
			b.WriteRune(c)
			continue
		default:
			upper = true
			continue
		}
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

func tag(name, comment string) string {
	for _, c := range "\r\n" {
		comment = strings.ReplaceAll(comment, string(c), ` `)
	}
	return fmt.Sprintf("`json:\"%s\"` // %s", name, comment)
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

func getInfo(cmd *cobra.Command) *Info {
	r := &Info{
		User:     os.Getenv(`MYSQL_USER`),
		Password: os.Getenv(`MYSQL_PASSWORD`),
		Host:     os.Getenv(`MYSQL_HOST`),
		Port:     os.Getenv(`MYSQL_PORT`),
	}

	if r.User == `` {
		r.User, _ = cmd.Flags().GetString(`user`)
	}

	if r.Password == `` {
		r.Password, _ = cmd.Flags().GetString(`password`)
	}

	if r.Host == `` {
		r.Host, _ = cmd.Flags().GetString(`host`)
	}

	if r.Port == `` {
		r.Port, _ = cmd.Flags().GetString(`port`)
	}

	return r
}

var RootCMD = &cobra.Command{
	Use:   `leopards`,
	Short: `An funny tool for DB schema`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	RootCMD.AddCommand(mysqlCMD)
}

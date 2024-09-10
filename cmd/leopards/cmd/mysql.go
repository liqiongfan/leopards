package cmd

import (
	bytes2 "bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"text/template"

	"github.com/liqiongfan/leopards"
	"github.com/spf13/cobra"
)

type Column struct {
	TableSchema            string  `json:"TABLE_SCHEMA"`
	TableName              string  `json:"TABLE_NAME"`
	ColumnName             *string `json:"COLUMN_NAME"`
	IsNullable             string  `json:"IS_NULLABLE"`
	DataType               *string `json:"DATA_TYPE"`
	CharacterMaximumLength *int64  `json:"CHARACTER_MAXIMUM_LENGTH"`
	CharacterOctetLength   *int64  `json:"CHARACTER_OCTET_LENGTH"`
	NumericPrecision       *int64  `json:"NUMERIC_PRECISION"`
	NumericScale           *int64  `json:"NUMERIC_SCALE"`
	ColumnType             string  `json:"COLUMN_TYPE"`
	ColumnKey              string  `json:"COLUMN_KEY"`
	Extra                  *string `json:"EXTRA"`
	ColumnComment          string  `json:"COLUMN_COMMENT"`
	CamelName              *string
}

const TemplateStruct = `
//
// created by leopards at: {{ .Time }}
// 

{{ range $key, $value := .Data }}
{{- range .Columns }} {{ enum .ColumnName .DataType .ColumnType }}    
{{- end }}

// {{ camel $value.TableName }}Table {{ $value.Comment }}
const {{ camel $value.TableName }}Table = "{{ $value.TableName }}"

// {{ camel $value.TableName }} {{ $value.Comment }}
type {{ camel $value.TableName }} struct { 
{{ range .Columns }}    {{ camel .CamelName }}{{ pad (camel .ColumnName) $value.MaxColumnLength }} {{ type .DataType .ColumnType .IsNullable }}{{ pad (type .DataType .ColumnType .IsNullable) $value.MaxTypeLength }}  {{ tag .ColumnName .ColumnComment  }}
{{ end -}} 
}
{{ end }}
`

type Table struct {
	TableName                      string `json:"TABLE_NAME"`
	Comment                        string `json:"TABLE_COMMENT"`
	Columns                        []Column
	MaxColumnLength, MaxTypeLength int
}

func generate(cmd *cobra.Command, args []string) error {
	info := getInfo(cmd)
	dsn := fmt.Sprintf(
		`%s:%s@tcp(%s:%s)/information_schema?interpolateParams=true&loc=Local&parseTime=True&timeTruncate=1s`,
		info.User,
		info.Password,
		info.Host,
		info.Port,
	)

	db, err := leopards.Open(leopards.MySQL, dsn)
	if err != nil {
		return err
	}

	query := db.Query().
		From(`tables`).
		Where(leopards.EQ(`TABLE_SCHEMA`, args[0]))
	if args[1] != `*` {
		tableNames := strings.Split(args[1], `,`)
		ins := make([]any, 0, 20)
		for _, name := range tableNames {
			ins = append(ins, strings.TrimSpace(name))
		}
		query.Where(leopards.In(`TABLE_NAME`, ins...))
	}

	tables := make([]Table, 0, 20)
	err = query.
		Scan(cmd.Context(), &tables)
	if err != nil {
		return err
	}

	needImportTime := false

	for i, table := range tables {
		columns := make([]Column, 0, 20)
		err = db.Query().
			From(`columns`).
			Where(leopards.EQ(`TABLE_SCHEMA`, args[0])).
			Where(leopards.EQ(`TABLE_NAME`, table.TableName)).
			OrderBy(leopards.Asc(`ORDINAL_POSITION`)).
			Scan(cmd.Context(), &columns)
		if err != nil {
			return err
		}
		flags := make(map[string]struct{}, len(columns))
		for j, column := range columns {
			camelName := camel(column.ColumnName)
			column.CamelName = column.ColumnName
			columns[j].CamelName = column.ColumnName
			if _, ok := flags[camelName]; ok {
				tName := snake(camelName)
				column.ColumnName = &tName
				columns[j].CamelName = &tName
			} else {
				flags[camelName] = struct{}{}
			}

			if Type(column.DataType, column.ColumnComment, column.IsNullable) == `time.Time` {
				needImportTime = true
			}
			if length := len(camel(column.CamelName)); length > tables[i].MaxColumnLength {
				tables[i].MaxColumnLength = length
			}
			if length := len(Type(column.DataType, column.ColumnType, column.IsNullable)); length > tables[i].MaxTypeLength {
				tables[i].MaxTypeLength = length
			}
		}
		tables[i].Columns = columns
	}

	output, err := cmd.Flags().GetString(`out`)
	if err != nil {
		return err
	}

	if output == `` {
		return trace(errors.New(`--out: output file not specified`))
	}

	t := template.New(`template`)

	t = t.Funcs(map[string]any{
		`camel`: camel,
		`type`:  Type,
		`tag`:   tag,
		`pad`:   pad,
		`enum`:  enum,
	})
	t, err = t.Parse(TemplateStruct)
	if err != nil {
		return trace(err)
	}

	bytes := bytes2.NewBuffer(nil)

	err = t.Execute(bytes, map[string]any{
		`Time`: time.Now().Format(`2006-01-02 15:04:05`),
		`Data`: tables,
	})
	if err != nil {
		return err
	}

	f, err := os.OpenFile(output, os.O_CREATE|os.O_RDWR|os.O_APPEND, os.ModePerm)
	if err != nil {
		return err
	}
	fi, err := f.Stat()
	if err != nil {
		return err
	}

	if fi.Size() == 0 {
		packageName := ``
		if filepath.Dir(output) == `.` {
			wd, _ := os.Getwd()
			packageName = filepath.Base(wd)
		} else {
			packageName = filepath.Base(filepath.Dir(output))
			_ = os.MkdirAll(filepath.Dir(output), os.ModePerm)
		}

		_, _ = f.WriteString(`package ` + packageName)
		if needImportTime {
			_, _ = f.WriteString("\n\nimport \"time\"")
		}
	}

	defer f.Close()
	_, _ = f.WriteString(bytes.String())

	return nil
}

var mysqlCMD = &cobra.Command{
	Use:   `mysql database table [-h]`,
	Short: `A MySQL schema generate tool for leopards`,
	RunE: func(cmd *cobra.Command, args []string) error {

		return generate(cmd, args)
	},
}

func init() {
	mysqlCMD.Flags().StringP(`user`, `U`, ``, `mysql database username`)
	mysqlCMD.Flags().StringP(`password`, `P`, ``, `mysql database password`)
	mysqlCMD.Flags().StringP(`host`, `H`, ``, `mysql database host`)
	mysqlCMD.Flags().StringP(`port`, `p`, `3306`, `mysql database port`)
	mysqlCMD.Flags().StringP(`out`, `o`, ``, `output path`)
}

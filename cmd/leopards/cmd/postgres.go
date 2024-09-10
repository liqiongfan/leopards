package cmd

import (
	bytes2 "bytes"
	"errors"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/liqiongfan/leopards"
	"github.com/spf13/cobra"
)

type PgTable struct {
	TableCatalog    string  `json:"table_catalog"`
	TableSchema     string  `json:"table_schema"`
	TableName       string  `json:"table_name"`
	TableType       string  `json:"table_type"`
	Comment         *string `json:"description"`
	Columns         []PgColumn
	MaxColumnLength int
	MaxTypeLength   int
}

type PgColumn struct {
	TableCatalog           string  `json:"table_catalog"`
	TableSchema            string  `json:"table_schema"`
	TableName              string  `json:"table_name"`
	ColumnName             *string `json:"column_name"`
	ColumnDefault          *string `json:"column_default"`
	IsNullAble             string  `json:"is_nullable"`
	DataType               string  `json:"data_type"`
	CharacterMaximumLength *int    `json:"character_maximum_length"`
	CharacterOctetLength   *int    `json:"character_octet_length"`
	NumericPrecision       *int    `json:"numeric_precision"`
	NumericPrecisionRadix  *int    `json:"numeric_precision_radix"`
	NumericScale           *int    `json:"numeric_scale"`
	UdtName                string  `json:"udt_name"`
	UdtCatalog             string  `json:"udt_catalog"`
	IsUpdatable            string  `json:"is_updatable"`
	Comment                *string `json:"description"`
	CamelName              *string
}

const TemplatePGStruct = `
//
// created by leopards at: {{ .Time }}
// 

{{ range $key, $value := .Data }}

// {{ camel $value.TableName }}Table {{ emit $value.Comment }}
const {{ camel $value.TableName }}Table = "{{ $value.TableName }}"

// {{ camel $value.TableName }} {{ emit $value.Comment }}
type {{ camel $value.TableName }} struct { 
{{ range .Columns }}    {{ camel .CamelName }}{{ pad (camel .ColumnName) $value.MaxColumnLength }} {{ type .DataType .IsNullAble .UdtName }}{{ pad (type .DataType .IsNullAble .UdtName) $value.MaxTypeLength }}  {{ tag .ColumnName .Comment  }}
{{ end -}} 
}
{{ end }}
`

func PGType(dataType, isNullable, udtName string) string {
	switch udtName {
	case `bigserial`:
		return `int64`
	case `bit`:
		return `[]byte`
	case `bool`:
		return `byte`
	case `box`:
		fallthrough
	case `bytea`:
		fallthrough
	case `char`:
		fallthrough
	case `cidr`:
		fallthrough
	case `circle`:
		return `string`
	case `date`:
		return `time.Time`
	case `decimal`:
		return `string`
	case `float4`:
		return `float32`
	case `float8`:
		return `float64`
	case `inet`:
		return `string`
	case `int2`:
		return `int16`
	case `int4`:
		return `int32`
	case `int8`:
		return `int64`
	case `interval`:
		return `string`
	case `json`:
		fallthrough
	case `jsonb`:
		fallthrough
	case `line`:
		fallthrough
	case `lseg`:
		fallthrough
	case `macaddr`:
		fallthrough
	case `money`:
		fallthrough
	case `numeric`:
		return `string`
	case `path`:
		fallthrough
	case `point`:
		fallthrough
	case `polygon`:
		return `string`
	case `serial`:
		return `int32`
	case `serial2`:
		return `int16`
	case `serial4`:
		return `int32`
	case `serial8`:
		return `int64`
	case `smallserial`:
		return `int16`
	case `text`:
		return `string`
	case `time`:
		fallthrough
	case `timestamp`:
		fallthrough
	case `timestamptz`:
		fallthrough
	case `timetz`:
		return `time.Time`
	case `tsquery`:
		fallthrough
	case `tsvector`:
		return `string`
	case `txid_snapshot`:
		return `string`
	case `uuid`:
		return `string`
	case `varbit`:
		return `[]byte`
	case `varchar`:
		fallthrough
	case `xml`:
		fallthrough
	default:
		return `string`
	}
}

func pgGenerate(cmd *cobra.Command, args []string) error {

	info := getPGInfo(cmd)
	orm, err := leopards.OpenOptions{
		User:     info.User,
		Password: info.Password,
		Host:     info.Host,
		Port:     info.Port,
		Database: args[0],
		Debug:    false,
		Dialect:  leopards.Postgres,
	}.Open()
	if err != nil {
		return err
	}

	schema, _ := cmd.Flags().GetString(`schema`)

	tables := make([]PgTable, 0, 20)
	t1 := orm.Table(`tables`).Schema(`information_schema`).As(`tb`)
	t2 := orm.Table(`pg_class`)
	t3 := orm.Table(`pg_description`).As(`d`)
	err = orm.Query().
		Select(
			t1.C(`table_catalog`),
			t1.C(`table_schema`),
			t1.C(`table_name`),
			t1.C(`table_type`),
			t3.C(`description`),
		).
		FromTable(t1).
		Join(t2.As(`c`)).On(t1.C(`table_name`), t2.C(`relname`)).
		LeftJoin(t3).
		On(t3.C(`objoid`), t2.C(`oid`)).OnP(leopards.EQ(t3.C(`objsubid`), 0)).
		Where(
			leopards.And(
				leopards.EQ(t1.C(`table_schema`), schema),
			),
		).
		Scan(cmd.Context(), &tables)
	if err != nil {
		return err
	}

	x1 := orm.Table(`columns`).Schema(`information_schema`).As(`col`)
	x2 := orm.Table(`pg_class`).As(`c`)
	x3 := orm.Table(`pg_description`).As(`d`)

	needImportTime := false

	for i, table := range tables {
		columns := make([]PgColumn, 0, 20)
		err = orm.Query().
			Select(
				x1.C(`table_catalog`),
				x1.C(`table_schema`),
				x1.C(`table_name`),
				x1.C(`column_name`),
				x1.C(`column_default`),
				x1.C(`is_nullable`),
				x1.C(`data_type`),
				x1.C(`character_maximum_length`),
				x1.C(`character_octet_length`),
				x1.C(`numeric_precision`),
				x1.C(`numeric_precision_radix`),
				x1.C(`numeric_scale`),
				x1.C(`udt_name`),
				x1.C(`udt_catalog`),
				x1.C(`is_updatable`),
				x3.C(`description`),
			).FromTable(x1).Join(x2).On(x1.C(`table_name`), x2.C(`relname`)).
			LeftJoin(x3).On(
			x3.C(`objoid`),
			x2.C(`oid`),
		).On(x3.C(`objsubid`), x1.C(`ordinal_position`)).
			Where(leopards.EQ(x1.C(`table_schema`), schema)).
			Where(leopards.EQ(x1.C(`table_name`), table.TableName)).
			OrderBy(x1.C(`table_name`), x1.C(`ordinal_position`)).
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

			if PGType(column.DataType, column.IsNullAble, column.UdtName) == `time.Time` {
				needImportTime = true
			}
			if length := len(camel(column.CamelName)); length > tables[i].MaxColumnLength {
				tables[i].MaxColumnLength = length
			}
			if length := len(PGType(column.DataType, column.IsNullAble, column.UdtName)); length > tables[i].MaxTypeLength {
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
		`type`:  PGType,
		`tag`:   tagPtr,
		`pad`:   pad,
		`enum`:  enum,
		`emit`: func(s *string) string {
			if s != nil {
				return *s
			}
			return ``
		},
	})
	t, err = t.Parse(TemplatePGStruct)
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

var postgresCMD = &cobra.Command{
	Use:   `postgres database table [-h]`,
	Short: `An PostgresSQL schema generate tool for leopards`,
	RunE: func(cmd *cobra.Command, args []string) error {

		return pgGenerate(cmd, args)
	},
}

func init() {
	postgresCMD.Flags().StringP(`user`, `U`, ``, `PostgresSQL database username`)
	postgresCMD.Flags().StringP(`password`, `P`, ``, `PostgresSQL database password`)
	postgresCMD.Flags().StringP(`host`, `H`, ``, `PostgresSQL database host`)
	postgresCMD.Flags().StringP(`port`, `p`, `5432`, `PostgresSQL database port`)
	postgresCMD.Flags().StringP(`schema`, `s`, `public`, `PostgresSQL schema, default public`)
	postgresCMD.Flags().StringP(`out`, `o`, ``, `output path`)
}

package cmd

import (
	bytes2 "bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
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
	MaxNameLength   int
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
{{ range .Columns }}    {{ camel .CamelName }}{{ pad (camel .ColumnName) $value.MaxColumnLength }} {{ type .DataType .IsNullAble .UdtName }}{{ pad (type .DataType .IsNullAble .UdtName) $value.MaxTypeLength }}  {{ tag .ColumnName .Comment $value.MaxNameLength }}
{{ end -}} 
}
{{ end }}
`

func PGType(dataType, isNullable, udtName string) (res string) {
	switch udtName {
	case `bigserial`:
		res = `uint64`
	case `bit`:
		res = `[]byte`
	case `bool`:
		res = `byte`
	case `box`:
		fallthrough
	case `bytea`:
		fallthrough
	case `char`:
		fallthrough
	case `cidr`:
		fallthrough
	case `circle`:
		res = `string`
	case `date`:
		res = `time.Time`
	case `decimal`:
		res = `string`
	case `float4`:
		res = `float32`
	case `float8`:
		res = `float64`
	case `inet`:
		res = `string`
	case `int2`:
		res = `int16`
	case `int4`:
		res = `int32`
	case `int8`:
		res = `int64`
	case `interval`:
		res = `string`
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
		res = `string`
	case `path`:
		fallthrough
	case `point`:
		fallthrough
	case `polygon`:
		res = `string`
	case `serial`:
		res = `uint32`
	case `serial2`:
		res = `uint16`
	case `serial4`:
		res = `uint32`
	case `serial8`:
		res = `uint64`
	case `smallserial`:
		res = `uint16`
	case `text`:
		res = `string`
	case `time`:
		fallthrough
	case `timestamp`:
		fallthrough
	case `timestamptz`:
		fallthrough
	case `timetz`:
		res = `time.Time`
	case `tsquery`:
		fallthrough
	case `tsvector`:
		res = `string`
	case `txid_snapshot`:
		res = `string`
	case `uuid`:
		res = `string`
	case `varbit`:
		res = `[]byte`
	case `varchar`:
		fallthrough
	case `xml`:
		fallthrough
	default:
		res = `string`
	}

	switch isNullable {
	case `YES`:
		res = `*` + res
	case `NO`:
	}

	return
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
		Charset:  info.Charset,
	}.Open()
	if err != nil {
		return err
	}

	schema, _ := cmd.Flags().GetString(`schema`)

	database := args[0]
	tableName := args[1]

	tables := make([]PgTable, 0, 20)
	t1 := orm.Table(`tables`).Schema(`information_schema`).As(`tb`)
	t2 := orm.Table(`pg_class`)
	t3 := orm.Table(`pg_description`).As(`d`)

	predicates := []*leopards.Predicate{
		leopards.EQ(t1.C(`table_catalog`), database),
		leopards.EQ(t1.C(`table_schema`), schema),
	}

	if tableName != `*` {
		predicates = append(predicates, leopards.EQ(t1.C(`table_name`), tableName))
	}

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
				predicates...,
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

			if strings.Contains(PGType(column.DataType, column.IsNullAble, column.UdtName), `time.Time`) {
				needImportTime = true
			}
			if length := len(camel(column.CamelName)); length > tables[i].MaxColumnLength {
				tables[i].MaxColumnLength = length
			}
			if length := len(PGType(column.DataType, column.IsNullAble, column.UdtName)); length > tables[i].MaxTypeLength {
				tables[i].MaxTypeLength = length
			}
			if length := len(*column.ColumnName); length > tables[i].MaxNameLength {
				tables[i].MaxNameLength = length
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
	Short: `A PostgreSQL schema generate tool for leopards`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return cmd.Help()
		}
		return pgGenerate(cmd, args)
	},
}

func init() {
	postgresCMD.Flags().StringP(`user`, `U`, ``, `PostgreSQL database username`)
	postgresCMD.Flags().StringP(`password`, `P`, ``, `PostgrsSQL database password`)
	postgresCMD.Flags().StringP(`host`, `H`, ``, `PostgreSQL database host`)
	postgresCMD.Flags().StringP(`port`, `p`, `5432`, `PostgreSQL database port`)
	postgresCMD.Flags().StringP(`schema`, `s`, `public`, `PostgreSQL schema, default public`)
	postgresCMD.Flags().StringP(`out`, `o`, ``, `output path`)
}

package leopards

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"reflect"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

type rowScan struct {
	types []reflect.Type
	ctype []*sql.ColumnType
	value func(v ...any) (reflect.Value, error)
}

func (rs *rowScan) values() []any {
	vs := make([]any, 0, len(rs.types))
	for _, typ := range rs.types {
		vs = append(vs, reflect.New(typ).Interface())
	}
	return vs
}

type DB struct {
	driver *sql.DB

	tx      *sql.Tx
	debug   bool
	dialect string

	beforeQuery []func(*Selector)
	afterQuery  []func(*Selector, any)

	beforeInsert []func(*InsertBuilder)
	afterInsert  []func(*InsertBuilder, any)

	beforeUpdate []func(*UpdateBuilder)
	afterUpdate  []func(*UpdateBuilder, any)

	beforeDelete []func(*DeleteBuilder)
	afterDelete  []func(*DeleteBuilder, any)
}

func (b *DB) InterceptorsQuery(iq func(*Selector)) {
	b.beforeQuery = append(b.beforeQuery, iq)
}

func (b *DB) InterceptorsAfterQuery(ia func(*Selector, any)) {
	b.afterQuery = append(b.afterQuery, ia)
}

func (b *DB) InterceptorsInsert(ii func(*InsertBuilder)) {
	b.beforeInsert = append(b.beforeInsert, ii)
}

func (b *DB) InterceptorsAfterInsert(ii func(*InsertBuilder, any)) {
	b.afterInsert = append(b.afterInsert, ii)
}

func (b *DB) InterceptorsUpdate(ii func(*UpdateBuilder)) {
	b.beforeUpdate = append(b.beforeUpdate, ii)
}

func (b *DB) InterceptorsAfterUpdate(ii func(*UpdateBuilder, any)) {
	b.afterUpdate = append(b.afterUpdate, ii)
}

func (b *DB) InterceptorsDelete(ii func(*DeleteBuilder)) {
	b.beforeDelete = append(b.beforeDelete, ii)
}

func (b *DB) InterceptorsAfterDelete(ii func(*DeleteBuilder, any)) {
	b.afterDelete = append(b.afterDelete, ii)
}

func (b *DB) columnName(f reflect.StructField) string {
	tags := []string{`leopard`, `db`, `gorm`, `sql`, `json`}
	for _, tag := range tags {
		if n, ok := f.Tag.Lookup(tag); ok {
			for _, piece := range strings.Split(n, `;`) {
				if !strings.HasPrefix(piece, `column`) {
					continue
				}
				return piece[strings.Index(piece, `:`)+1:]
			}

			if strings.Contains(n, `,`) {
				return n[0:strings.Index(n, `,`)]
			}

			return n
		}
	}
	return strings.ToLower(f.Name)
}

func (b *DB) parseEmbed(v map[string][]int, typ reflect.Type, idxs []int, index int) map[string][]int {
	idx := make([]int, index+1, index+2)
	copy(idx, idxs)

	for i := 0; i < typ.NumField(); i++ {
		f := typ.Field(i)

		if f.PkgPath != `` {
			continue
		}

		idx[index] = i

		if typ := f.Type; f.Anonymous {
			switch {
			case typ.Kind() == reflect.Struct:
				v = b.parseEmbed(v, typ, idx, index+1)
			}
			continue
		}

		newIdx := append([]int{}, idx...)
		v[f.Name] = newIdx
		v[b.columnName(f)] = newIdx
	}
	return v
}

func (b *DB) scanStruct(typ reflect.Type, columns []string, ctypes []*sql.ColumnType) (*rowScan, error) {
	names := make(map[string][]int, typ.NumField())
	rs := &rowScan{types: make([]reflect.Type, 0, typ.NumField())}

	names = b.parseEmbed(names, typ, []int{}, 0)
	for i, column := range columns {
		var idx []int
		switch name := strings.Split(column, "(")[0]; {
		case names[name] != nil:
			idx = names[name]
		case names[strings.ToLower(name)] != nil:
			idx = names[strings.ToLower(name)]
		default:
			rs.types = append(rs.types, ctypes[i].ScanType())
			continue
		}
		rtype := typ.Field(idx[0]).Type
		for _, vi := range idx[1:] {
			rtype = rtype.Field(vi).Type
		}

		rs.types = append(rs.types, rtype)
	}

	rs.value = func(vs ...any) (reflect.Value, error) {
		dest := reflect.New(typ).Elem()
		for i, v := range vs {

			name := columns[i]
			if reflect.ValueOf(v).IsNil() {
				continue
			}
			name = name[strings.Index(name, `.`)+1:]

			if rver, ok := v.(driver.Valuer); ok {
				v, _ = rver.Value()
				if v == nil {
					continue
				}
			}

			rv := reflect.Indirect(reflect.ValueOf(v))

			idx, ok := names[name]
			if !ok {
				continue
			}

			dv := dest.Field(idx[0])
			for _, vi := range idx[1:] {
				dv = dv.Field(vi)
			}

			dv.Set(rv)
		}

		return dest, nil
	}

	return rs, nil
}

func (b *DB) scanPointer(typ reflect.Type, columns []string, ctypes []*sql.ColumnType) (*rowScan, error) {
	typ = typ.Elem()
	rs, err := b.scanType(typ, columns, ctypes)
	if err != nil {
		return nil, err
	}
	w := rs.value
	rs.value = func(vs ...any) (reflect.Value, error) {
		v, err := w(vs...)
		if err != nil {
			return reflect.Value{}, err
		}
		rv := reflect.Indirect(v)
		pv := reflect.New(rv.Type())
		pv.Elem().Set(rv)
		return pv, nil
	}
	return rs, nil
}

func (b *DB) scanMap(typ reflect.Type, columns []string, ctypes []*sql.ColumnType) (*rowScan, error) {
	rs := &rowScan{types: make([]reflect.Type, 0, len(ctypes))}

	for _, ty := range ctypes {
		rs.types = append(rs.types, ty.ScanType())
	}

	rs.value = func(vs ...any) (reflect.Value, error) {
		mv := reflect.MakeMap(typ)

		for i, v := range vs {
			if rver, ok := v.(driver.Valuer); ok {
				v, _ = rver.Value()
				if v == nil {
					continue
				}
			}
			rv := reflect.Indirect(reflect.ValueOf(v))
			switch {
			case typ.Elem().Kind() == reflect.Interface || rv.Kind() == typ.Elem().Kind():
			default:
				continue
			}
			mv.SetMapIndex(reflect.ValueOf(columns[i]), rv)
		}

		return mv, nil
	}

	return rs, nil
}

func (b *DB) scanType(typ reflect.Type, columns []string, ctypes []*sql.ColumnType) (*rowScan, error) {
	switch k := typ.Kind(); {
	case k == reflect.Map:
		return b.scanMap(typ, columns, ctypes)
	case k == reflect.Interface:
		return b.scanMap(reflect.TypeOf((*map[string]any)(nil)).Elem(), columns, ctypes)
	case k == reflect.Struct:
		return b.scanStruct(typ, columns, ctypes)
	case k == reflect.Pointer:
		return b.scanPointer(typ, columns, ctypes)
	default:
		return nil, fmt.Errorf(`scanType: unsupported type ([]%s)`, k)
	}
}

func (b *DB) ScanSlice(rows *sql.Rows, dest any) error {
	v := reflect.ValueOf(dest)
	if v.Kind() != reflect.Pointer {
		return errors.New(`ScanSlice: non-pointer of dest`)
	}

	v = reflect.Indirect(v)
	if k := v.Kind(); k != reflect.Slice {
		return fmt.Errorf("ScanSlice: invalid type: %s. expect slice as artument", v)
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	types, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	scan, err := b.scanType(v.Type().Elem(), columns, types)
	if err != nil {
		return err
	}

	for rows.Next() {
		vs := scan.values()

		if err = rows.Scan(vs...); err != nil {
			return err
		}

		rv, err := scan.value(vs...)
		if err != nil {
			return err
		}

		v.Set(reflect.Append(v, rv))
	}

	return rows.Err()
}

func (b *DB) Scan(rows *sql.Rows, dest any) error {
	t := reflect.TypeOf(dest)
	if t.Kind() != reflect.Pointer {
		return errors.New(`ScanRow: non-pointer of dest`)
	}

	t = t.Elem()
	if t.Kind() == reflect.Slice {
		return b.ScanSlice(rows, dest)
	}

	v := reflect.MakeSlice(reflect.SliceOf(t), 0, 5)
	vt := reflect.NewAt(v.Type(), v.UnsafePointer())

	err := b.ScanSlice(rows, vt.Interface())
	if err != nil {
		return err
	}

	if vt.Elem().Len() == 0 {
		return nil
	}

	reflect.ValueOf(dest).Elem().Set(vt.Elem().Index(0))
	return nil
}

// OpenOptions 链接选项
type OpenOptions struct {
	User          string // 用户
	Password      string // 密码
	Host          string // 主机
	Port          string // 端口
	Database      string // 数据库
	Debug         bool   // 调试模式
	Dialect       string // 数据库类型, 可选 leopards.MySQL | leopards.SQLite | leopards.Postgres | leopards.Gremlin
	FileForSQLite string // SQLite 数据库需要配置, 其他类型忽略
}

// Open 打开链接获取一个DB操作类
func (p OpenOptions) Open() (*DB, error) {
	dri, err := sql.Open(p.Dialect, DSN(&p))
	if err != nil {
		return nil, err
	}
	b := &DB{}
	b.driver = dri
	b.dialect = p.Dialect
	b.debug = p.Debug
	return b, nil
}

func DSN(opt *OpenOptions) string {
	switch opt.Dialect {
	case MySQL:
		return opt.User + `:` + opt.Password + `@(` + opt.Host + `:` + opt.Port + `)/` + opt.Database + `?interpolateParams=true&loc=Local&parseTime=True&timeTruncate=1s`
	case Postgres: // host=<host> port=<port> user=<user> dbname=<database> password=<pass>
		return `host=` + opt.Host + ` port=` + opt.Port + ` user=` + opt.User + ` dbname=` + opt.Database + ` password=` + opt.Password
	case SQLite: //  file:ent?mode=memory&cache=shared&_fk=1
		return opt.FileForSQLite + `?mode=memory&cache=shared`
	case Gremlin: // http://localhost:8182
		return opt.Host + `:` + opt.Port
	default:
		panic(`DSN: unsupported dialect type.`)
	}
}

func Open(dialect string, dsn string) (*DB, error) {
	dri, err := sql.Open(dialect, dsn)
	if err != nil {
		return nil, err
	}
	b := &DB{}
	b.driver = dri
	b.dialect = dialect
	return b, nil
}

func OpenWithInfo(dialect, host, port, user, password, database string) (*DB, error) {
	return Open(
		dialect,
		DSN(&OpenOptions{
			User:     user,
			Password: password,
			Host:     host,
			Port:     port,
			Database: database,
			Debug:    false,
			Dialect:  dialect,
		}),
	)
}

func OpenWithDebug(dialect, dsn string) (*DB, error) {
	db, err := Open(dialect, dsn)
	if err != nil {
		return nil, err
	}
	db.debug = true

	return db, nil
}

func (b *DB) Commit(ctx context.Context) error {
	defer func() {
		b.tx = nil
	}()
	if b.tx == nil {
		return nil
	}
	err := b.tx.Commit()
	return err
}

func (b *DB) Rollback(ctx context.Context) error {
	defer func() {
		b.tx = nil
	}()
	if b.tx == nil {
		return nil
	}
	err := b.tx.Rollback()
	return err
}

func (b *DB) TX(ctx context.Context) (*DB, error) {
	tx, err := b.driver.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &DB{driver: b.driver, tx: tx, debug: b.debug}, nil
}

func (b *DB) Query() *Selector {
	return Dialect(b.dialect).Select(b)
}

func (b *DB) Update() *UpdateBuilder {
	return Dialect(b.dialect).Update(b, ``)
}

func (b *DB) Insert() *InsertBuilder {
	return Dialect(b.dialect).Insert(b, ``)
}

func (b *DB) Delete() *DeleteBuilder {
	return Dialect(b.dialect).Delete(b, ``)
}

package sqldecoder

import (
	"database/sql"
	"io"
	"reflect"
)

type fieldMap struct {
	columnName string
}
type Decoder struct {
	rows      *sql.Rows
	columnMap map[string]int
}

type unmarshalTypeError struct {
	rt reflect.Type
}

func (e unmarshalTypeError) Error() string {
	return "Cannot unmarshal into value of type " + e.rt.String()
}

func NewDecoder(rows *sql.Rows) (*Decoder, error) {
	d := &Decoder{rows, make(map[string]int)}
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	for i, key := range cols {
		d.columnMap[key] = i
	}
	return d, nil
}

// unmarshal gets the data from the row and stores it in the value pointed to by v.
func (d *Decoder) unmarshal(v interface{}) error {
	var e error

	shimmer := reflect.ValueOf(v)
	if shimmer.Type().Kind() != reflect.Ptr {
		return unmarshalTypeError{rt: shimmer.Type()}
	}
	dst := shimmer.Elem()

	switch dst.Kind() {
	case reflect.Struct:
		fields := make([]interface{}, len(d.columnMap))
		for i := 0; i < dst.NumField(); i++ {
			fv := dst.Field(i)
			if ok := dst.CanSet(); ok {
				ft := dst.Type().Field(i)
				if colName := ft.Tag.Get("db"); colName != "" {
					if idx, ok := d.columnMap[colName]; ok {
						fields[idx] = fv.Addr().Interface()
					}
				}
			}
		}
		if e = d.rows.Scan(fields...); e != nil {
			return e
		}
	default:
		return unmarshalTypeError{rt: dst.Type()}

	}
	return nil
}

// Decode the next row into v. v is expected to a struct. Returns io.EOF if there are no rows to decode.
func (d *Decoder) Decode(v interface{}) error {
	if d.rows == nil {
		return io.EOF
	}

	if ok := d.rows.Next(); ok {
		d.unmarshal(v)
	} else {
		return io.EOF
	}
	return nil
}

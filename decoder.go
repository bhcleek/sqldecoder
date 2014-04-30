package sqldecoder

import (
	"database/sql"
	"io"
	"reflect"
)

type typeMap map[reflect.Type]map[string]int
type Decoder struct {
	rows *sql.Rows
	tm   typeMap
}

type unmarshalTypeError struct {
	rt reflect.Type
}

func (e unmarshalTypeError) Error() string {
	return "Cannot unmarshal into value of type " + e.rt.String()
}

// NewDecoder returns a new decoder that reads from rows.
func NewDecoder(rows *sql.Rows) (*Decoder, error) {
	d := &Decoder{rows: rows, tm: make(typeMap)}
	return d, nil
}

// fieldMap provides a map whose keys are a column name and whose values are
// the field index. The column name for a given exported field is (in priority
// order):
// 		the value of a sql tag on on the field
//		the field name
func fieldMap(t reflect.Type) map[string]int {
	if t.Kind() != reflect.Struct {
		return nil
	}

	fm := make(map[string]int)
	tagged := make(map[string]bool)
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		colName := ft.Tag.Get("sql")
		if colName != "" {
			fm[colName] = i
			tagged[colName] = true
			continue
		}

		colName = ft.Name
		if _, ok := tagged[colName]; !ok {
			fm[colName] = i
			tagged[colName] = false
		}
	}

	return fm
}

// unmarshal gets the data from the row and stores it in the value pointed to by v.
func (d *Decoder) unmarshal(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return unmarshalTypeError{rt: rv.Type()}
	}
	dst := rv.Elem()

	switch dst.Kind() {
	case reflect.Struct:
		fm, ok := d.tm[dst.Type()]
		if !ok {
			fm = fieldMap(dst.Type())
			d.tm[dst.Type()] = fm
		}

		cols, err := d.rows.Columns()
		if err != nil {
			return err
		}

		fields := make([]interface{}, len(cols))
		for ci, v := range cols {
			if fi, ok := fm[v]; ok {
				if fv := dst.Field(fi); fv.CanSet() {
					fields[ci] = fv.Addr().Interface()
					continue
				}
			}
		}

		if err := d.rows.Scan(fields...); err != nil {
			return err
		}
	default:
		return unmarshalTypeError{rt: dst.Type()}
	}
	return nil
}

// Decode the next row into v. v is expected to be a struct.
// Returns io.EOF if there are no more rows to decode.
func (d *Decoder) Decode(v interface{}) error {
	if d.rows == nil {
		return io.EOF
	}

	if ok := d.rows.Next(); ok {
		if err := d.unmarshal(v); err != nil {
			return err
		}
	} else {
		return io.EOF
	}
	return nil
}

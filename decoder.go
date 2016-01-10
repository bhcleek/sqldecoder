package sqldecoder

import (
	"database/sql"
	"io"
	"reflect"
)

type typeMap map[reflect.Type]map[string]int

// A Decoder reads and decodes values from rows.
type Decoder struct {
	rows *sql.Rows
	d    decodeState
}

type decodeState struct {
	tm typeMap
	s  Scanner
}

type unmarshalTypeError struct {
	rt reflect.Type
}

func (e unmarshalTypeError) Error() string {
	return "Cannot unmarshal into value of type " + e.rt.String()
}

// NewDecoder returns a new decoder that reads from rows.
func NewDecoder(rows *sql.Rows) *Decoder {
	d := decodeState{tm: make(typeMap), s: rows}
	decoder := &Decoder{rows: rows, d: d}
	return decoder
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

// columnMapFromTags uses tags to provide a ColumnMap. The column name for a
// given exported field is (in priority order):
// 	the value of a sql tag on the field
// 	the field name
func (ds *decodeState) columnMapFromTags(v interface{}) (ColumnMap, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, unmarshalTypeError{rt: rv.Type()}
	}
	dst := rv.Elem()

	var cm ColumnMap
	switch dst.Kind() {
	case reflect.Struct:
		tfm, ok := ds.tm[dst.Type()]
		if !ok {
			tfm = fieldMap(dst.Type())
			ds.tm[dst.Type()] = tfm
		}

		cols, err := ds.s.Columns()
		if err != nil {
			return nil, err
		}

		cm = make(map[string]interface{}, len(cols))
		for _, col := range cols {
			if fi, ok := tfm[col]; ok {
				if fv := dst.Field(fi); fv.CanSet() {
					cm[col] = fv.Addr().Interface()
				}
			}
		}

	default:
		return nil, unmarshalTypeError{rt: dst.Type()}
	}
	return cm, nil
}

func (ds *decodeState) fields(v interface{}) ([]interface{}, error) {
	var mappedFields ColumnMap
	var err error
	if fm, ok := v.(ColumnMapper); ok {
		mappedFields = fm.ColumnMap()
	} else {
		mappedFields, err = ds.columnMapFromTags(v)
		if err != nil {
			return nil, err
		}
	}
	cols, err := ds.s.Columns()
	if err != nil {
		return nil, err
	}

	fields := make([]interface{}, len(cols))
	for i, v := range cols {
		if fieldDest, ok := mappedFields[v]; ok {
			fields[i] = fieldDest
		} else {
			fields[i] = new(interface{})
		}
	}
	return fields, nil
}

// unmarshal gets the data from the scanner and stores it in the value pointed to by v.
func (ds *decodeState) unmarshal(v interface{}) error {
	fields, err := ds.fields(v)
	if err != nil {
		return err
	}

	return ds.s.Scan(fields...)
}

// Decode the next row into v. v is expected to be a struct.
// Returns io.EOF if there are no more rows to decode.
func (d *Decoder) Decode(v interface{}) error {
	if d.rows == nil {
		return io.EOF
	}

	if ok := d.rows.Next(); ok {
		if err := d.d.unmarshal(v); err != nil {
			return err
		}
	} else {
		return io.EOF
	}
	return nil
}

// Scanner copies columns into the values pointed at by dest.
// *sql.Rows implements Scanner.
type Scanner interface {
	Scan(dest ...interface{}) error
	Columns() ([]string, error)
}

// Unmarshal gets the data from row and stores it in v.
func Unmarshal(s Scanner, v interface{}) error {
	d := decodeState{tm: make(typeMap), s: s}
	return d.unmarshal(v)

}

// ColumnMap maps column names to values into which the named column can be
// scanned. Values are expected to be pointers.
type ColumnMap map[string]interface{}

// ColumnMapper is the interface implemented by an object that provides a
// ColumnMap to resolve column names to values into which the column should be
// scanned.
type ColumnMapper interface {
	ColumnMap() ColumnMap
}

package sqldecoder

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"
	"time"

	"github.com/erikstmartin/go-testdb"
)

func TestZeroValue(t *testing.T) {
	actual := Decoder{}
	err := actual.Decode(nil)
	if err != io.EOF {
		t.Fail()
	}
}

func TestNilRows(t *testing.T) {
	actual := NewDecoder(nil)
	if err := actual.Decode(nil); err != io.EOF {
		t.Fail()
	}
}

type columnMappedContainer struct {
	id           int64
	amount       float64
	isTruth      bool
	data         []byte
	description  string
	creationTime time.Time
}

func (v *columnMappedContainer) ColumnMap() ColumnMap {
	return ColumnMap{
		"ID":           &v.id,
		"Amount":       &v.amount,
		"IsTruth":      &v.isTruth,
		"Data":         &v.data,
		"Description":  &v.description,
		"CreationTime": &v.creationTime,
	}
}

type valueContainer struct {
	ID           int64
	Amount       float64
	IsTruth      bool
	Data         []byte
	Description  string
	CreationTime time.Time
}

type taggedValueContainer struct {
	Natural      int64     `sql:"ID"`
	Amount       float64   `sql:"Amount"`
	Truth        bool      `sql:"IsTruth"`
	Blob         []byte    `sql:"Data"`
	Description  string    `sql:"Description"`
	CreationTime time.Time `sql:"CreationTime"`
}

func TestColumnMapper(t *testing.T) {
	defer testdb.Reset()

	rows, err := stubRows()
	if err != nil {
		t.Fatal(err)
	}

	target := NewDecoder(rows)

	expected := columnMappedContainer{id: 1, amount: 1.1, isTruth: false, data: []byte("blob"), description: "short and stout", creationTime: time.Date(2009, 11, 10, 23, 00, 00, 0, time.UTC)}
	actual := new(columnMappedContainer)
	if err = target.Decode(actual); err != nil {
		t.Fatalf("Decode failed: %s", err)
	}

	if len(target.d.tm) != 0 {
		t.Fatalf("decoder used type map")
	}

	if actual.id != expected.id {
		t.Errorf("got %v, expected %v", actual.id, expected.id)
	}

	if actual.amount != expected.amount {
		t.Errorf("got %v, expected %v", actual.amount, expected.amount)
	}

	if actual.isTruth != expected.isTruth {
		t.Errorf("got %v, expected %v", actual.isTruth, expected.isTruth)
	}

	if bytes.Equal(actual.data, expected.data) {
		t.Errorf("got %v, expected %v", actual.data, expected.data)
	}

	if actual.description != expected.description {
		t.Errorf("got '%v', expected '%v'", actual.description, expected.description)
	}

	if actual.creationTime != expected.creationTime {
		t.Errorf("got %v, expected %v", actual.creationTime, expected.creationTime)
	}
}

func TestTaggedStruct(t *testing.T) {
	defer testdb.Reset()

	rows, err := stubRows()
	if err != nil {
		t.Fatal(err)
	}

	target := NewDecoder(rows)

	expected := taggedValueContainer{Natural: 1, Amount: 1.1, Truth: false, Blob: []byte("blob"), Description: "short and stout", CreationTime: time.Date(2009, 11, 10, 23, 00, 00, 0, time.UTC)}
	actual := new(taggedValueContainer)
	if err = target.Decode(actual); err != nil {
		t.Fatalf("Decode failed: %s", err)
	}

	if actual.Natural != expected.Natural {
		t.Errorf("got %v, expected %v", actual.Natural, expected.Natural)
	}

	if actual.Amount != expected.Amount {
		t.Errorf("got %v, expected %v", actual.Amount, expected.Amount)
	}

	if actual.Truth != expected.Truth {
		t.Errorf("got %v, expected %v", actual.Truth, expected.Truth)
	}

	if bytes.Equal(actual.Blob, expected.Blob) {
		t.Errorf("got %v, expected %v", actual.Blob, expected.Blob)
	}

	if actual.Description != expected.Description {
		t.Errorf("got '%v', expected '%v'", actual.Description, expected.Description)
	}

	if actual.CreationTime != expected.CreationTime {
		t.Errorf("got %v, expected %v", actual.CreationTime, expected.CreationTime)
	}
}

func TestUntaggedStruct(t *testing.T) {
	defer testdb.Reset()

	rows, err := stubRows()
	if err != nil {
		t.Fatal(err)
	}

	target := NewDecoder(rows)
	expected := valueContainer{ID: 1, Amount: 1.1, IsTruth: false, Data: []byte("blob"), Description: "short and stout", CreationTime: time.Date(2009, 11, 10, 23, 00, 00, 0, time.UTC)}
	actual := new(valueContainer)
	if err = target.Decode(actual); err != nil {
		t.Fatalf("Decode failed: %s", err)
	}

	if actual.ID != expected.ID {
		t.Errorf("got %v, expected %v", actual.ID, expected.ID)
	}

	if actual.Amount != expected.Amount {
		t.Errorf("got %v, expected %v", actual.Amount, expected.Amount)
	}

	if actual.IsTruth != expected.IsTruth {
		t.Errorf("got %v, expected %v", actual.IsTruth, expected.IsTruth)
	}

	if bytes.Equal(actual.Data, expected.Data) {
		t.Errorf("got %v, expected %v", actual.Data, expected.Data)
	}

	if actual.Description != expected.Description {
		t.Errorf("got '%v', expected '%v'", actual.Description, expected.Description)
	}

	if actual.CreationTime != expected.CreationTime {
		t.Errorf("got %v, expected %v", actual.CreationTime, expected.CreationTime)
	}
}

func TestDecodeReturnsEOF(t *testing.T) {
	defer testdb.Reset()

	rows, err := stubRows()
	if err != nil {
		t.Fatal(err)
	}

	target := NewDecoder(rows)
	actual := &struct{}{}
	_ = target.Decode(actual)
	if err = target.Decode(actual); err != io.EOF {
		t.Errorf("Decode(actual), got %s, expected %s", err, io.EOF)
	}
}

func TestDecodeStructValueProvidesError(t *testing.T) {
	defer testdb.Reset()

	rows, err := stubRows()
	if err != nil {
		t.Fatal(err)
	}

	target := NewDecoder(rows)

	actual := new(taggedValueContainer)
	err = target.Decode(*actual)
	if err == nil {
		t.Fatalf("Decode(*actual), got %s", err.Error())
	}

	if _, ok := err.(unmarshalTypeError); !ok {
		t.Fatalf("Decode(*actual), got %v, expected unmarshalTypeError", err)
	}
}

func TestDecodeNonStructProvidesError(t *testing.T) {
	defer testdb.Reset()

	rows, err := stubRows()
	if err != nil {
		t.Fatal(err)
	}

	target := NewDecoder(rows)

	vc := new(int64)
	err = target.Decode(vc)
	if err == nil {
		t.Fatalf("Decode(vc), got %s", err.Error())
	}

	if _, ok := err.(unmarshalTypeError); !ok {
		t.Fatalf("Decode(*actual), got %v, expected unmarshalTypeError", err)
	}
}

// rows is a driver.Rows to be used by the testdb driver.
type rows struct {
	closed  bool
	data    [][]driver.Value
	columns []string
}

func (r *rows) Close() error {
	if !r.closed {
		r.closed = true
	}
	return nil
}

func (r *rows) Columns() []string {
	return r.columns
}

func (r *rows) Next(dest []driver.Value) error {
	if len(r.data) > 0 {
		copy(dest, r.data[0][0:])
		r.data = r.data[1:]
		return nil
	}
	return io.EOF
}

type EmptyRowsRecord struct {
}

func (rows EmptyRowsRecord) Rows() Rows {
	return nil
}

func stubRows() (Rows, error) {
	db, err := sql.Open("testdb", "")
	if err != nil {
		return nil, err
	}

	sql := "SELECT fields FROM TheTable"
	result := &rows{columns: []string{"ID", "Amount", "IsTruth", "Data", "Description", "CreationTime", "IgnoredField"},
		data: [][]driver.Value{[]driver.Value{1, 1.1, false, []byte("I am a little teapot"), []byte("short and stout"), time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC), []byte("ignored")}}}
	testdb.StubQuery(sql, result)

	return db.Query(sql)
}

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

func (rows EmptyRowsRecord) Rows() *sql.Rows {
	return nil
}

func TestZeroValue(t *testing.T) {
	actual := Decoder{}
	err := actual.Decode(nil)
	if err != io.EOF {
		t.Fail()
	}
}

func TestNilRows(t *testing.T) {
	if actual, err := NewDecoder(nil); err == nil {
		if err = actual.Decode(nil); err != io.EOF {
			t.Fail()
		}
	} else {
		t.Error("error creating decoder")
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
	Natural      int64     `sql:"id"`
	Amount       float64   `sql:"amount"`
	Truth        bool      `sql:"is_truth"`
	Blob         []byte    `sql:"data"`
	Description  string    `sql:"description"`
	CreationTime time.Time `sql:"creation_time"`
}

func TestTaggedStructPointerValues(t *testing.T) {
	defer testdb.Reset()

	db, err := sql.Open("testdb", "")
	if err != nil {
		t.Fatalf("test database did not open: %s", err)
	}

	sql := "SELECT fields FROM TheTable"
	result := &rows{columns: []string{"id", "amount", "is_truth", "data", "description", "creation_time"},
		data: [][]driver.Value{[]driver.Value{1, 1.1, false, []byte("I am a little teapot"), []byte("short and stout"), time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)}}}
	testdb.StubQuery(sql, result)

	rows, err := db.Query(sql)
	target, err := NewDecoder(rows)
	if err != nil {
		t.Fatal("error creating decoder")
	}
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

func TestUntaggedStructPointerValues(t *testing.T) {
	defer testdb.Reset()

	db, err := sql.Open("testdb", "")
	if err != nil {
		t.Fatalf("test database did not open: %s", err)
	}

	sql := "SELECT fields FROM TheTable"
	result := &rows{columns: []string{"ID", "Amount", "IsTruth", "Data", "Description", "CreationTime"},
		data: [][]driver.Value{[]driver.Value{1, 1.1, false, []byte("I am a little teapot"), []byte("short and stout"), time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)}}}
	testdb.StubQuery(sql, result)

	rows, err := db.Query(sql)
	target, err := NewDecoder(rows)
	if err != nil {
		t.Fatal("error creating decoder")
	}
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

	db, err := sql.Open("testdb", "")
	if err != nil {
		t.Fatalf("test database did not open: %s", err)
	}

	sql := "SELECT fields FROM TheTable"
	result := &rows{columns: []string{"id", "amount", "is_truth", "data", "description", "creation_time"},
		data: [][]driver.Value{[]driver.Value{1, 1.1, false, []byte("I am a little teapot"), []byte("short and stout"), time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)}}}
	testdb.StubQuery(sql, result)

	rows, err := db.Query(sql)
	target, err := NewDecoder(rows)
	if err != nil {
		t.Fatal("error creating decoder")
	}
	actual := &struct{}{}
	_ = target.Decode(actual)
	if err = target.Decode(actual); err != io.EOF {
		t.Errorf("Decode(actual), got %s, expected %s", err, io.EOF)
	}
}

func TestDecodeStructValueProvidesError(t *testing.T) {
	defer testdb.Reset()

	db, err := sql.Open("testdb", "")
	if err != nil {
		t.Fatalf("test databse did not open: %s", err)
	}
	sql := "SELECT files from table"
	result := &rows{columns: []string{"id", "amount", "is_truth", "data", "description", "creation_time"},
		data: [][]driver.Value{[]driver.Value{1, 1.1, false, []byte("I am a little teapot"), []byte("short and stout"), time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)}}}
	testdb.StubQuery(sql, result)

	rows, err := db.Query(sql)
	if err != nil {
		t.Fatalf("query error: %s", err)
	}

	target, err := NewDecoder(rows)
	if err != nil {
		t.Fatal("error creating decoder")
	}

	vc := new(valueContainer)
	if err := target.Decode(vc); err == nil {
		t.Fatalf("Decode(vc), got %s, expected nil", err.Error())
	}
}

func TestDecodeNonStructProvidesError(t *testing.T) {
	defer testdb.Reset()

	db, err := sql.Open("testdb", "")
	if err != nil {
		t.Fatalf("test databse did not open: %s", err)
	}
	sql := "SELECT files from table"
	result := &rows{columns: []string{"id", "amount", "is_truth", "data", "description", "creation_time"},
		data: [][]driver.Value{[]driver.Value{1, 1.1, false, []byte("I am a little teapot"), []byte("short and stout"), time.Date(2009, 11, 10, 23, 0, 0, 0, time.UTC)}}}
	testdb.StubQuery(sql, result)

	rows, err := db.Query(sql)
	if err != nil {
		t.Fatalf("query error: %s", err)
	}

	target, err := NewDecoder(rows)
	if err != nil {
		t.Fatal("error creating decoder")
	}

	vc := new(int64)
	if err := target.Decode(vc); err == nil {
		t.Fatalf("Decode(vc), got %s, expected nil", err.Error())
	}
}

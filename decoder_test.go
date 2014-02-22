package sqldecoder

import (
	"bytes"
	"database/sql"
	"io"
	"testing"
	"time"

	"github.com/erikstmartin/go-testdb"
)

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

type ValueContainer struct {
	natural      int64     `sql:"id"`
	amount       float64   `sql:"amount"`
	truth        bool      `sql:"isTruth"`
	blob         []byte    `sql:"data"`
	description  string    `sql:"description"`
	creationTime time.Time `sql:"creation_time"`
}

func TestValues(t *testing.T) {
	db, err := sql.Open("testdb", "")
	if err != nil {
		t.Fatalf("mock database did not open: %s", err)
	}

	sql := "SELECT fields FROM TheTable"
	//int64, float64, bool, []byte, string, time.Time

	result := `1,1.1,false,I am a little teapot,short and stout,2009-11-10 23:00:00 +0000 UTC
	`
	testdb.StubQuery(sql, testdb.RowsFromCSVString([]string{"id", "amount", "truth", "blob", "description", "creation_time"}, result))

	rows, err := db.Query(sql)
	if target, err := NewDecoder(rows); err == nil {
		expected := ValueContainer{natural: 1, amount: 1.1, truth: false, blob: []byte("blob"), description: "description", creationTime: time.Date(2009, 11, 10, 23, 00, 00, 0, time.UTC)}
		actual := new(ValueContainer)
		target.Decode(&actual)

		// todo: compare the byte slices
		if actual.natural != expected.natural {
			t.Errorf("got %+v, expected %+v", actual.natural, expected.natural)
		}

		if actual.amount != expected.amount {
			t.Errorf("got %+v, expected %+v", actual.amount, expected.amount)
		}

		if actual.truth != expected.truth {
			t.Errorf("got %+v, expected %+v", actual.truth, expected.truth)
		}

		if bytes.Equal(actual.blob, expected.blob) {
			t.Errorf("got %+v, expected %+v", actual.blob, expected.blob)
		}

		if actual.description != expected.description {
			t.Errorf("got %+v, expected %+v", actual.description, expected.description)
		}

		if actual.creationTime != expected.creationTime {
			t.Errorf("got %+v, expected %+v", actual.creationTime, expected.creationTime)
		}
	} else {
		t.Error("error creating decoder")
	}
}

/*
func TestNullsToValues(t *testing.T) {
	if target, err := NewDecoder(rows); err == nil {
		t.Fatal("not implemented")
	} else {
		t.Error("error creating decoder")
	}
}

func TestPointers(t *testing.T) {
	if target, err := NewDecoder(rows); err == nil {
		t.Fatal("not implemented")
	} else {
		t.Error("error creating decoder")
	}
}

func TestNullToPointers(t *testing.T) {
	if target, err := NewDecoder(rows); err == nil {
		t.Fatal("not implemented")
	} else {
		t.Error("error creating decoder")
	}
}
*/

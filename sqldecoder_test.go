package sqldecoder

import (
	"database/sql"
	"io"
	"testing"
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

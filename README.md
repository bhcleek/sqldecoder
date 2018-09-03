# sqldecoder

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/bhcleek/sqldecoder)

Decode `sql.Rows` into structs without having to remember ordinal positions. sqldecoder maps struct fields to SQL columns without reflection using an interface, `ColumnMapper`. Alternatively, sqldecoder will use reflection to map tagged struct fields or struct field names to SQL columns. If the struct neither implements `ColumnMapper` nor has tags, sqldecoder expects the column names and field names to match exactly.

## Quick Start

Regardless of whether a struct implements the `ColumnMapper` interface or not, using sqldecoder is the same:

```go
func GetPeople(r *sql.Rows) (people []Person){
	personDecoder, err := decoder.NewDecoder(rows)
	if err != nil {
		return nil
	}
	people := make([]Person, 4)
	for {
		someone := Person{}
		if err := decoder.Decode(&someone); err == io.EOF {
			break
		} else {
			append(people, someone)
		}
	}
	return people
}
```

### reflection-less 

Implement `ColumnMapper`

```go
type Person struct {
	firstName string
	lastName string
}

func (p *Person) ColumnMap() ColumnMap{
	return ColumnMap{ "FirstName": &p.firstName, "LastName": &p.lastName }
}

```

### with struct tags

```go
type Person struct{
	First_Name string `db:"FirstName"`
	Last_Name  string `db:"LastName"`
}
```

### synchronized column and field names

```go
type Person struct{
	FirstName string 
	LastName  string
}
```

# sqldecoder

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/bhcleek/sqldecoder)

decode sql.Rows into structs without having to remember ordinal positions. sqldecoder supports scanning into structs using either an interface for reflection-less, or you can tag the struct fields with the column names to which they map. If the struct neither implements the ColumnMap interface nor has tags, sqldecoder expects the column names and field names to match exactly.

## Quick Start

Regardless of whether a struct implements the ColumnMap interface or not, using sqldecoder is the same:

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

Implement the ColumnMap interface

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

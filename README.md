sqldecoder
==========

decode sql.Rows into structs

```go
type Person struct{
	firstName string `db:"first_name"`
	lastName string `db:"last_name"`
}

func GetPeople(r *sql.Rows) (people []Person){
	personDecoder, err := decoder.NewDecoder(rows)
	if err != nil {
		return nil
	}
	people := make([]Person, 4)
	for {
		if err := decoder.Decode(&someone); err == io.EOF {
			break
		} else {
			append(people, someone)
		}
	}
	return people
}
```

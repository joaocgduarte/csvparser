# csvparser
This package provides a fast and easy-of-use custom mapping from .csv data into Golang structs.

## Pre-requisites
Since the library uses generics, it is necessary to have `go1.18`

## Installation
```
go get github.com/plagioriginal/csvparser
```

## Examples

### Csv parsing from bytes
This will read the .csv data being sent, and will return an array of whatever you would like.

```go
type Person struct {
	Name string
	Age int
	isInSchool bool
}
    
var input = []byte(`
name,age
frank,13
anabelle,70`)

parser := csvparser.NewCsvParserFromBytes[Person](input)
parser.AddColumnParser("name", func (value string, into *Person) error {
    into.Name = strings.Trim(value, " ")
    return nil
})
parser.AddColumnParser("age", func (value string, into *Person) error {
    value = strings.Trim(value, " ")
    age, err := strconv.Atoi(value)
    if err != nil {
        return err
    }
    into.Age = age
    if age < 18 {
	    into.IsInSchool = true	
    }
    return nil
})

// res is []Person type
res, err := parser.Parse()
```

Note: as long as there is a parser for the header that you want, the order of the .csv columns will not matter

#### What if the file doesn't have headers
When instantiating the parser, you can specify the headers of the file, in order, and the parser will handle everything 
for you. Just remember that the ParserHandlers need to be added.

```go
var input = []byte(`
frank,13
anabelle,70`)

parser := csvparser.NewCsvParserFromBytes[Person](input, "name", "age")
parser.AddColumnParser("name", nameHandler)
parser.AddColumnParser("age", ageHandler)
...
```
### Csv Parsing from multipart file / anything that applies the io.Reader
If you need to directly use something like a multipart file directly, you can do something like this:
```go
func (h *OrderHandler) handlerFunc(w http.ResponseWriter, r *http.Request) {
    file, _, err := r.FormFile("file-key-in-request")
    if err != nil {
        ...
    }
	defer file.Close()
	parser := csvparser.NewCsvParserFromReader[WhateverStruct](file)
	...
}
```

### Adding a hook
You can add an hook that will run everytime something is parsed from the .csv file, so that you don't have to do another loop in the results in case you want to add more logic into it.

```go
parser := csvparser.NewCsvParserFromBytes[WhateverStruct](input)
children := make([]Person, 0)
parser.WithHook(func(parsedPerson Person) {
    if parsedPerson.IsInSchool {
        children = append(children, parsedObject)
    }
})
```
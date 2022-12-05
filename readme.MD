# csvparser
[![Go Report Card](https://goreportcard.com/badge/github.com/plagioriginal/csvparser)](https://goreportcard.com/report/github.com/plagioriginal/csvparser)

This package provides a fast and easy-of-use custom mapping from .csv data into Golang structs.

# Index
- [Pre-requisites](#pre-requisites)
- [Installation](#installation)
- [Examples](#examples)
    - [CSV Parsing from bytes](#csv-parsing-from-bytes)
    - [Csv Parsing from multipart file / anything that applies the io.Reader](#csv-parsing-from-multipart-file--anything-that-applies-the-ioreader)
    - [Adding hooks](#adding-hooks)
      - [After each successful parsing](#after-each-successful-parsing)
      - [On Error Hook](#on-error-hook)
    - [Additional Settings](#additional-settings)
      - [Terminate on row parsing error](#terminate-on-row-parsing-error)

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
translate := make([]interface{}, len(res))
	for i, s := range res {
		translate[i] = s
	}
	resultString := parseToString(translate)
	fmt.Println(resultString)
```

Note: as long as there is a parser for the header that you want, the order of the .csv columns will not matter

#### What if the file doesn't have headers
When instantiating the parser, you can specify the headers of the file, in order, and the parser will handle everything 
for you. Just remember that the ParserHandlers need to be added.

```go
var input = []byte(`
frank,13
anabelle,70`)

parser := csvparser.NewCsvParserFromBytes[Person](input, "name", "age").
	AddColumnParser("name", nameHandler).
	AddColumnParser("age", ageHandler)
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

### Adding hooks

#### After each successful parsing
You can add a hook that will run everytime something is parsed from the .csv file, 
so that you don't have to do another loop in the results in case you want to add more logic into it.
To do this, use the function `AfterEachParsingHook()`

```go
parser := csvparser.NewCsvParserFromBytes[Person](input)
children := make([]Person, 0)
parser.AfterEachParsingHook(func(person Person) {
    if parsedPerson.IsInSchool {
        children = append(children, person)
    }
})
```

#### On Error Hook
Use the `OnError()` function to handle the error of an invalid row yourself.
```go
parser := csvparser.NewCsvParserFromBytes[Person](input).
    OnError(func(row []string, err error) {
        log.Printf("row %v has thrown the error: %v", row, err)
    })
```

### Additional Settings
#### Terminate on row parsing error
You can choose if you want to throw an error on the parsing results if the input has an invalid row, or 
just continue and skip that record.
By default, the behaviour is to skip the error. 

However, you can make it stop and throw an error with the function
`TerminateOnParsingError()`:

```go
res, err := csvparser.NewCsvParserFromBytes[Person](input).
    AddColumnParser("name", nameHandler).
    AddColumnParser("age", ageHandler).
    TerminateOnParsingError().
    Parse()
```

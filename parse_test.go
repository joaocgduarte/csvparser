package csvparser

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

type person struct {
	Name   string
	Age    int
	School string
}

var impossibleAgeError = errors.New("impossible age")

func nameParser(value string, into *person) error {
	into.Name = strings.Trim(value, " ")
	return nil
}

func ageParser(value string, into *person) error {
	value = strings.Trim(value, " ")
	age, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	if age > 150 {
		return impossibleAgeError
	}
	into.Age = age
	into.School = "new school"
	if age > 20 && age < 65 {
		into.School = "middle school"
	}
	if age > 65 {
		into.School = "old school"
	}
	return nil
}

func TestCsvParserWithoutHookAndFinishingIfParsingErrorIsFound(t *testing.T) {
	tests := []struct {
		name           string
		input          []byte
		headersToAdd   []string
		parserAdder    func(parser *CsvParser[person])
		expectedResult []person
		expectedErr    error
	}{
		{
			name:           "empty input results in eof error",
			input:          []byte(""),
			headersToAdd:   []string{},
			parserAdder:    nil,
			expectedResult: []person{},
			expectedErr:    parseError{Msg: fmt.Sprintf("couldn't read headers from file: %s", io.EOF.Error())},
		},
		{
			name:           "empty input but with headers",
			input:          []byte(""),
			headersToAdd:   []string{"header one"},
			parserAdder:    nil,
			expectedResult: []person{},
			expectedErr:    parseError{Msg: fmt.Sprintf("header \"%s\" doesn't have an associated parser", "header one")},
		},
		{
			name: "header age without parser should return error",
			input: []byte(`
name,age
frank,13
anabelle,65`),
			headersToAdd: []string{},
			parserAdder: func(parser *CsvParser[person]) {
				parser.AddColumnParser("name", nameParser)
			},
			expectedResult: []person{},
			expectedErr:    parseError{Msg: fmt.Sprintf("header \"%s\" doesn't have an associated parser", "age")},
		},
		{
			name: "success with no headers added",
			input: []byte(`
name,age
frank,13
anabelle,70`),
			headersToAdd: []string{},
			parserAdder: func(parser *CsvParser[person]) {
				parser.AddColumnParser("name", nameParser).
					AddColumnParser("age", ageParser)
			},
			expectedResult: []person{
				{
					Name:   "frank",
					Age:    13,
					School: "new school",
				},
				{
					Name:   "anabelle",
					Age:    70,
					School: "old school",
				},
			},
			expectedErr: nil,
		},
		{
			name: "success with headers",
			input: []byte(`
frank,13
anabelle,70`),
			headersToAdd: []string{"name", "age"},
			parserAdder: func(parser *CsvParser[person]) {
				parser.AddColumnParser("name", nameParser).
					AddColumnParser("age", ageParser)
			},
			expectedResult: []person{
				{
					Name:   "frank",
					Age:    13,
					School: "new school",
				},
				{
					Name:   "anabelle",
					Age:    70,
					School: "old school",
				},
			},
			expectedErr: nil,
		},
		{
			name: "make sure error from custom-parser is triggered",
			input: []byte(`
name,age
frank,13
anabelle,70
rita,170`),
			headersToAdd: []string{},
			parserAdder: func(parser *CsvParser[person]) {
				parser.AddColumnParser("name", nameParser).
					AddColumnParser("age", ageParser)
			},
			expectedResult: []person{},
			expectedErr:    newparseError(impossibleAgeError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCsvParserFromBytes[person](tt.input, tt.headersToAdd...)
			parser.TerminateOnParsingError()
			if tt.parserAdder != nil {
				tt.parserAdder(parser)
			}
			res, err := parser.Parse()
			if tt.expectedErr == nil && err != nil {
				t.Errorf("wanted error \"%v\", got error \"%v\"", tt.expectedErr, err)
			}
			if tt.expectedErr != nil && !errors.Is(err, tt.expectedErr) {
				t.Errorf("wanted error \"%v\", got error \"%v\"", tt.expectedErr, err)
			}
			if !reflect.DeepEqual(tt.expectedResult, res) {
				t.Errorf("result %v, but got %v", tt.expectedResult, res)
			}
		})
	}
}

func TestCsvParserHook(t *testing.T) {
	middleAgedPeople := make([]person, 0)
	input := []byte(`
name,age
frank,13
rita, 40
robert, 25
anabelle,70`)
	parser := NewCsvParserFromBytes[person](input).
		AddColumnParser("name", nameParser).
		AddColumnParser("age", ageParser).
		AfterEachParsingHook(func(parsedObject person) {

			if parsedObject.School == "middle school" {
				middleAgedPeople = append(middleAgedPeople, parsedObject)
			}
		})
	res, err := parser.Parse()
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	expectedEndResult := []person{
		{
			Name:   "frank",
			Age:    13,
			School: "new school",
		},
		{
			Name:   "rita",
			Age:    40,
			School: "middle school",
		},
		{
			Name:   "robert",
			Age:    25,
			School: "middle school",
		},
		{
			Name:   "anabelle",
			Age:    70,
			School: "old school",
		},
	}

	expectedMiddleAgedPeople := []person{
		{
			Name:   "rita",
			Age:    40,
			School: "middle school",
		},
		{
			Name:   "robert",
			Age:    25,
			School: "middle school",
		},
	}

	if !reflect.DeepEqual(res, expectedEndResult) {
		t.Errorf("expected result %v, got result %v", expectedEndResult, res)
	}
	if !reflect.DeepEqual(middleAgedPeople, expectedMiddleAgedPeople) {
		t.Errorf("expected middle-aged people result %v, got result %v", expectedMiddleAgedPeople, middleAgedPeople)
	}
}

func TestOnParseErrorHook(t *testing.T) {
	hasOnErrorRan := false
	input := []byte(`
name,age
frank,13
rita, 40
robert, 25
anabelle,170`)
	NewCsvParserFromBytes[person](input).
		AddColumnParser("name", nameParser).
		AddColumnParser("age", ageParser).
		TerminateOnParsingError().
		OnParseError(func(row []string, err error) {
			hasOnErrorRan = true
			expectedRow := []string{"anabelle", "170"}
			expectedErr := impossibleAgeError
			if !reflect.DeepEqual(row, expectedRow) {
				t.Errorf("wanted row %v, got row %v", expectedRow, row)
			}
			if err != expectedErr {
				t.Errorf("wanted error %v, got error %v", expectedErr, err)
			}
		}).
		Parse()
	if !hasOnErrorRan {
		t.Errorf("error hook didn't start.")
	}
}

func TestOnStartAndOnFinishHooks(t *testing.T) {
	hasOnStartRan := false
	hasOnEndRan := false
	input := []byte(`
name,age
frank,13
rita, 40
robert, 25
anabelle,17`)
	NewCsvParserFromBytes[person](input).
		AddColumnParser("name", nameParser).
		AddColumnParser("age", ageParser).
		TerminateOnParsingError().
		OnStart(func() {
			hasOnStartRan = true
		}).
		OnFinish(func() {
			hasOnEndRan = true
		}).
		Parse()
	if !hasOnStartRan {
		t.Errorf("start hook didn't start.")
	}
	if !hasOnEndRan {
		t.Errorf("end hook didn't start.")
	}
}

func TestCsvParserDontFinishOnError(t *testing.T) {
	input := []byte(`
name,age
frank,13
rita, 40
robert, 25
anabelle,170`)
	results, err := NewCsvParserFromBytes[person](input).
		AddColumnParser("name", nameParser).
		AddColumnParser("age", ageParser).
		Parse()

	expectedResults := []person{
		{Name: "frank", Age: 13, School: "new school"},
		{Name: "rita", Age: 40, School: "middle school"},
		{Name: "robert", Age: 25, School: "middle school"},
	}
	if !reflect.DeepEqual(results, expectedResults) {
		t.Errorf("wanted %v, got %v", expectedResults, results)
	}
	if err != nil {
		t.Errorf("wanted nil error, got error %v", err)
	}
}

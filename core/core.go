package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/HeavyWombat/color"
	"github.com/HeavyWombat/yaml"
	"github.com/mitchellh/hashstructure"
)

// Debug log output
var Debug = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

// Info log output
var Info = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)

// Warning log output
var Warning = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)

// Error log output
var Error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

// Constants to differenciate between the different kinds of differences
const (
	ADDITION     = '+'
	REMOVAL      = '-'
	MODIFICATION = '±'
	ILLEGAL      = '✕'
	ATTENTION    = '⚠'
)

// PathElement describes a part of a path, meaning its name. In this case the "Key" string is empty. Named list entries such as "name: one" use both "Key" and "Name" to properly specify the path element.
type PathElement struct {
	Key  string
	Name string
}

// Path describes a position inside a YAML (or JSON) structure by providing a name to each hierarchy level (tree structure).
type Path []PathElement

// Detail encapsulate the actual details of a change, mainly the kind of difference and the values.
type Detail struct {
	Kind rune
	From interface{}
	To   interface{}
}

// Diff encapsulates everything noteworthy about a difference
type Diff struct {
	Path    Path
	Details []Detail
}

// Bold returns the provided string in 'bold' format
func Bold(text string) string {
	return colorEachLine(color.New(color.Bold), text)
}

// Italic returns the provided string in 'italic' format
func Italic(text string) string {
	return colorEachLine(color.New(color.Italic), text)
}

func Green(text string) string {
	return colorEachLine(color.New(color.FgGreen), text)
}

func Red(text string) string {
	return colorEachLine(color.New(color.FgRed), text)
}

func Yellow(text string) string {
	return colorEachLine(color.New(color.FgYellow), text)
}

func Color(text string, attributes ...color.Attribute) string {
	return colorEachLine(color.New(attributes...), text)
}

func colorEachLine(color *color.Color, text string) string {
	var buf bytes.Buffer

	splitted := strings.Split(text, "\n")
	length := len(splitted)
	for idx, line := range splitted {
		buf.WriteString(color.Sprint(line))

		if idx < length-1 {
			buf.WriteString("\n")
		}
	}

	return buf.String()
}

// ToDotStyle returns a path as a string in dot style separating each path element by a dot.
// Please note that path elements that are named "." will look ugly.
func ToDotStyle(path Path) string {
	result := make([]string, 0, len(path))
	for _, element := range path {
		if element.Key != "" {
			result = append(result, element.Name) // TODO make italic for human output
		} else {
			result = append(result, element.Name)
		}
	}

	return strings.Join(result, ".")
}

// ToGoPatchStyle returns a path as a string in Go-Patch (https://github.com/cppforlife/go-patch) style separating each path element by a slash. Named list entries will be shown with their respecitive identifier name such as "name", "key", or "id".
func ToGoPatchStyle(path Path) string {
	result := make([]string, 0, len(path))
	for _, element := range path {
		if element.Key != "" {
			result = append(result, fmt.Sprintf("%s=%s", element.Key, element.Name))
		} else {
			result = append(result, element.Name)
		}
	}

	return "/" + strings.Join(result, "/")
}

func (path Path) String() string {
	return ToGoPatchStyle(path)
}

// CompareDocuments is the main entry point to compare to YAML MapSlices (documents) and returns a list of differences. Each difference describes a change to comes from "from" to "to", hence the names.
func CompareDocuments(from yaml.MapSlice, to yaml.MapSlice) []Diff {
	return compareMapSlices(Path{}, from, to)
}

// CompareObjects returns a list of differences between `from` and `to`
func CompareObjects(path Path, from interface{}, to interface{}) []Diff {
	// TODO add debug check or remove output
	// Debug.Printf("compare obj %#v (%s) vs %#v (%s)", from, reflect.TypeOf(from), to, reflect.TypeOf(to))

	// Save some time and process some simple nil and type-change use cases immediately
	if from == nil && to != nil {
		return []Diff{Diff{path, []Detail{Detail{Kind: ADDITION, From: from, To: to}}}}

	} else if from != nil && to == nil {
		return []Diff{Diff{path, []Detail{Detail{Kind: REMOVAL, From: from, To: to}}}}

	} else if from == nil && to == nil {
		return []Diff{}

	} else if reflect.TypeOf(from) != reflect.TypeOf(to) {
		return []Diff{Diff{path, []Detail{Detail{Kind: MODIFICATION, From: from, To: to}}}}
	}

	result := make([]Diff, 0)

	switch from.(type) {
	case yaml.MapSlice:
		switch to.(type) {
		case yaml.MapSlice:
			result = append(result, compareMapSlices(path, from.(yaml.MapSlice), to.(yaml.MapSlice))...)

		}

	case []interface{}:
		switch to.(type) {
		case []interface{}:
			result = append(result, compareLists(path, from.([]interface{}), to.([]interface{}))...)
		}

	case string:
		switch to.(type) {
		case string:
			result = append(result, compareStrings(path, from.(string), to.(string))...)

		}

	case bool, float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
		switch to.(type) {
		case bool, float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr:
			if from != to {
				result = append(result, Diff{path, []Detail{Detail{Kind: MODIFICATION, From: from, To: to}}})
			}
		}

	default:
		panic(fmt.Sprintf("Unsupported type %s", reflect.TypeOf(from)))
	}

	return result
}

func compareMapSlices(path Path, from yaml.MapSlice, to yaml.MapSlice) []Diff {
	removals := yaml.MapSlice{}
	additions := yaml.MapSlice{}

	result := make([]Diff, 0)

	for _, fromItem := range from {
		key := fromItem.Key
		if toItem, ok := GetMapItemByKeyFromMapSlice(key, to); ok {
			// `from` and `to` contain the same `key` -> require comparison
			result = append(result, CompareObjects(newPath(path, "", key), fromItem.Value, toItem.Value)...)

		} else {
			// `from` contain the `key`, but `to` does not -> removal
			removals = append(removals, fromItem)
		}
	}

	for _, toItem := range to {
		key := toItem.Key
		if _, ok := GetMapItemByKeyFromMapSlice(key, from); !ok {
			// `to` contains a `key` that `from` does not have -> addition
			additions = append(additions, toItem)
		}
	}

	diff := Diff{Path: path, Details: []Detail{}}

	if len(removals) > 0 {
		diff.Details = append(diff.Details, Detail{Kind: REMOVAL, From: removals, To: nil})
	}

	if len(additions) > 0 {
		diff.Details = append(diff.Details, Detail{Kind: ADDITION, From: nil, To: additions})
	}

	if len(diff.Details) > 0 {
		result = append(result, diff)
	}

	return result
}

func compareLists(path Path, from []interface{}, to []interface{}) []Diff {
	if fromIdentifier := GetIdentifierFromNamedList(from); fromIdentifier != "" {
		if toIdentifier := GetIdentifierFromNamedList(to); fromIdentifier == toIdentifier {
			return compareNamedEntryLists(path, fromIdentifier, from, to)
		}
	}

	return compareSimpleLists(path, from, to)
}

func compareSimpleLists(path Path, from []interface{}, to []interface{}) []Diff {
	removals := make([]interface{}, 0)
	additions := make([]interface{}, 0)

	result := make([]Diff, 0)

	fromLength := len(from)
	toLength := len(to)

	// Back out immediately if both lists are empty
	if fromLength == 0 && fromLength == toLength {
		return result
	}

	// Special case if both lists only contain one entry: directly compare the two entries with each other
	if fromLength == 1 && fromLength == toLength {
		return CompareObjects(newPath(path, "", 0), from[0], to[0])
	}

	fromLookup := createLookUpMap(from)
	toLookup := createLookUpMap(to)

	for idxPos, fromValue := range from {
		if _, ok := toLookup[calcHash(fromValue)]; !ok {
			// `from` entry does not exist in `to` list
			removals = append(removals, from[idxPos])
		}
	}

	for idxPos, toValue := range to {
		if _, ok := fromLookup[calcHash(toValue)]; !ok {
			// `to` entry does not exist in `from` list
			additions = append(additions, to[idxPos])
		}
	}

	diff := Diff{Path: path, Details: []Detail{}}

	if len(removals) > 0 {
		diff.Details = append(diff.Details, Detail{Kind: REMOVAL, From: removals, To: nil})
	}

	if len(additions) > 0 {
		diff.Details = append(diff.Details, Detail{Kind: ADDITION, From: nil, To: additions})
	}

	if len(diff.Details) > 0 {
		result = append(result, diff)
	}

	return result
}

func compareNamedEntryLists(path Path, identifier string, from []interface{}, to []interface{}) []Diff {
	removals := make([]interface{}, 0)
	additions := make([]interface{}, 0)

	result := make([]Diff, 0)

	for _, fromEntry := range from {
		name := GetKeyValue(fromEntry.(yaml.MapSlice), identifier)
		if toEntry, ok := GetEntryFromNamedList(to, identifier, name); ok {
			// `from` and `to` have the same entry idenfified by identifier and name -> require comparison
			result = append(result, CompareObjects(newPath(path, identifier, name), fromEntry, toEntry)...)

		} else {
			// `from` has an entry (identified by identifier and name), but `to` does not -> removal
			removals = append(removals, fromEntry)
		}
	}

	for _, toEntry := range to {
		name := GetKeyValue(toEntry.(yaml.MapSlice), identifier)
		if _, ok := GetEntryFromNamedList(from, identifier, name); !ok {
			// `to` has an entry (identified by identifier and name), but `from` does not -> addition
			additions = append(additions, toEntry)
		}
	}

	diff := Diff{Path: path, Details: []Detail{}}

	if len(removals) > 0 {
		diff.Details = append(diff.Details, Detail{Kind: REMOVAL, From: removals, To: nil})
	}

	if len(additions) > 0 {
		diff.Details = append(diff.Details, Detail{Kind: ADDITION, From: nil, To: additions})
	}

	if len(diff.Details) > 0 {
		result = append(result, diff)
	}

	return result
}

func compareStrings(path Path, from string, to string) []Diff {
	result := make([]Diff, 0)
	if strings.Compare(from, to) != 0 {
		// distance := levenshtein.DistanceForStrings([]rune(from), []rune(to), levenshtein.DefaultOptions)
		result = append(result, Diff{path, []Detail{Detail{Kind: MODIFICATION, From: from, To: to}}})
	}

	return result
}

func newPath(path Path, key interface{}, name interface{}) Path {
	result := make(Path, len(path))
	copy(result, path)

	return append(result, PathElement{Key: fmt.Sprintf("%v", key),
		Name: fmt.Sprintf("%v", name)})
}

// GetMapItemByKeyFromMapSlice returns the MapItem (tuple of key/value) where the MapItem key matches the provided key. It will return an empty MapItem and bool false if the given MapSlice does not contain a suitable MapItem.
func GetMapItemByKeyFromMapSlice(key interface{}, mapslice yaml.MapSlice) (yaml.MapItem, bool) {
	for _, mapitem := range mapslice {
		if mapitem.Key == key {
			return mapitem, true
		}
	}

	return yaml.MapItem{}, false
}

// GetKeyValue returns the value for a given key in a provided MapSlice. This is comparable to getting a value from a map with `foobar[key]`. Function will panic if there is no such key. This is only intended to be used in scenarios where you know a key has to be present.
func GetKeyValue(mapslice yaml.MapSlice, key string) interface{} {
	for _, element := range mapslice {
		if element.Key == key {
			return element.Value
		}
	}

	panic(fmt.Sprintf("There is no key `%s` in MapSlice %v", key, mapslice))
}

// GetEntryFromNamedList returns the entry that is identified by the identifier key and a name, for example: `name: one` where name is the identifier key and one the name. Function will return nil with bool false if there is no such entry.
func GetEntryFromNamedList(list []interface{}, identifier string, name interface{}) (interface{}, bool) {
	for _, listEntry := range list {
		mapslice := listEntry.(yaml.MapSlice)

		for _, element := range mapslice {
			if element.Key == identifier && element.Value == name {
				return mapslice, true
			}
		}
	}

	return nil, false
}

// GetIdentifierFromNamedList returns the identifier key used in the provided list, or an empty string if there is none. The identifier key is either 'name', 'key', or 'id'.
func GetIdentifierFromNamedList(list []interface{}) string {
	counters := map[interface{}]int{}

	for _, sliceEntry := range list {
		switch sliceEntry.(type) {
		case yaml.MapSlice:
			for _, mapSliceEntry := range sliceEntry.(yaml.MapSlice) {
				if _, ok := counters[mapSliceEntry.Key]; !ok {
					counters[mapSliceEntry.Key] = 0
				}

				counters[mapSliceEntry.Key]++
			}
		}
	}

	sliceLength := len(list)
	for _, identifier := range []string{"name", "key", "id"} {
		if count, ok := counters[identifier]; ok && count == sliceLength {
			return identifier
		}
	}

	return ""
}

func createLookUpMap(list []interface{}) map[uint64]int {
	result := make(map[uint64]int, len(list))
	for idx, entry := range list {
		result[calcHash(entry)] = idx
	}

	return result
}

func calcHash(obj interface{}) uint64 {
	var hash uint64
	var err error
	if hash, err = hashstructure.Hash(obj, nil); err != nil {
		panic(err)
	}

	return hash
}

// LoadYAMLs loads two YAMLs from the provided locations concurrently
func LoadYAMLs(locationA string, locationB string) (yaml.MapSlice, yaml.MapSlice, error) {
	type resultPair struct {
		mapslice yaml.MapSlice
		err      error
	}

	fromChan := make(chan resultPair, 1)
	toChan := make(chan resultPair, 1)

	go func() {
		mapslice, err := LoadYAMLFromLocation(locationA)
		fromChan <- resultPair{mapslice, err}
	}()

	go func() {
		mapslice, err := LoadYAMLFromLocation(locationB)
		toChan <- resultPair{mapslice, err}
	}()

	from := <-fromChan
	if from.err != nil {
		return nil, nil, from.err
	}

	to := <-toChan
	if to.err != nil {
		return nil, nil, to.err
	}

	return from.mapslice, to.mapslice, nil
}

// LoadYAMLFromLocation processes the provided input location to load a YAML (or JSON) into a yaml.MapSlice
func LoadYAMLFromLocation(location string) (yaml.MapSlice, error) {
	// TODO Support URIs as loaction
	// TODO Generate error if file contains more than one document

	var data []byte
	var content = yaml.MapSlice{}
	var err error

	if location == "-" {
		if data, err = ioutil.ReadAll(os.Stdin); err != nil {
			return nil, err
		}
	} else {
		if data, err = ioutil.ReadFile(location); err != nil {
			return nil, err
		}
	}

	if err = yaml.UnmarshalStrict([]byte(data), &content); err != nil {
		return nil, err
	}

	return content, nil
}

// ToJSONString converts the provided object into a human readable JSON string.
func ToJSONString(obj interface{}) (string, error) {
	switch v := obj.(type) {

	case []interface{}:
		result := make([]string, 0)
		for _, i := range v {
			value, err := ToJSONString(i)
			if err != nil {
				return "", err
			}
			result = append(result, value)
		}

		return fmt.Sprintf("[%s]", strings.Join(result, ", ")), nil

	case yaml.MapSlice:
		result := make([]string, 0)
		for _, i := range v {
			value, err := ToJSONString(i)
			if err != nil {
				return "", err
			}
			result = append(result, value)
		}

		return fmt.Sprintf("{%s}", strings.Join(result, ", ")), nil

	case yaml.MapItem:
		key, keyError := ToJSONString(v.Key)
		if keyError != nil {
			return "", keyError
		}

		value, valueError := ToJSONString(v.Value)
		if valueError != nil {
			return "", valueError
		}

		return fmt.Sprintf("%s: %s", key, value), nil

	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}

		return fmt.Sprintf("%s", string(bytes)), nil
	}
}

// ToYAMLString converts the provided YAML MapSlice into a human readable YAML string.
func ToYAMLString(content yaml.MapSlice) (string, error) {
	out, err := yaml.Marshal(content)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("---\n%s\n", string(out)), nil
}

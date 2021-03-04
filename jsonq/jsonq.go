package jsonq

import (
	"fmt"
	"strconv"
)

// JSONQuery is an object that enables querying of a Go map with a simple
// positional query language.
type JSONQuery struct {
	blob map[string]interface{}
}

// stringFromInterface converts an interface{} to a string and returns an error if types don't match.
func stringFromInterface(val interface{}) (string, error) {
	switch val.(type) {
	case string:
		return val.(string), nil
	}
	return "", fmt.Errorf("Expected string value for String, got \"%v\"", val)
}

// getstringFromInterface converts an interface{} to a string and returns an error if types don't match.
func getstringFromInterface(val interface{}) string {
	switch val.(type) {
	case string:
		return val.(string)
	}
	return ""
}

// boolFromInterface converts an interface{} to a bool and returns an error if types don't match.
func boolFromInterface(val interface{}) (bool, error) {
	switch val.(type) {
	case bool:
		return val.(bool), nil
	}
	return false, fmt.Errorf("Expected boolean value for Bool, got \"%v\"", val)
}

// floatFromInterface converts an interface{} to a float64 and returns an error if types don't match.
func floatFromInterface(val interface{}) (float64, error) {
	switch val.(type) {
	case float64:
		return val.(float64), nil
	case int:
		return float64(val.(int)), nil
	case string:
		fval, err := strconv.ParseFloat(val.(string), 64)
		if err == nil {
			return fval, nil
		}
	}
	return 0.0, fmt.Errorf("Expected numeric value for Float, got \"%v\"", val)
}

// intFromInterface converts an interface{} to an int and returns an error if types don't match.
func intFromInterface(val interface{}) (int, error) {
	switch val.(type) {
	case float64:
		return int(val.(float64)), nil
	case string:
		ival, err := strconv.ParseFloat(val.(string), 64)
		if err == nil {
			return int(ival), nil
		}
	case int:
		return val.(int), nil
	}
	return 0, fmt.Errorf("Expected numeric value for Int, got \"%v\"", val)
}

// GetintFromInterface converts an interface{} to an int and returns an error if types don't match.
func GetintFromInterface(val interface{}) int {
	switch val.(type) {
	case float64:
		return int(val.(float64))
	case string:
		ival, err := strconv.ParseFloat(val.(string), 64)
		if err == nil {
			return int(ival)
		}
	case int:
		return val.(int)
	}
	return 0
}

// objectFromInterface converts an interface{} to a map[string]interface{} and returns an error if types don't match.
func objectFromInterface(val interface{}) (map[string]interface{}, error) {
	switch val.(type) {
	case map[string]interface{}:
		return val.(map[string]interface{}), nil
	}
	return map[string]interface{}{}, fmt.Errorf("Expected json object for Object, got \"%v\"", val)
}

// arrayFromInterface converts an interface{} to an []interface{} and returns an error if types don't match.
func arrayFromInterface(val interface{}) ([]interface{}, error) {
	switch val.(type) {
	case []interface{}:
		return val.([]interface{}), nil
	}
	return []interface{}{}, fmt.Errorf("Expected json array for Array, got \"%v\"", val)
}

// NewQuery creates a new JsonQuery obj from an interface{}.
func NewQuery(data interface{}) *JSONQuery {
	j := new(JSONQuery)
	j.blob = data.(map[string]interface{})
	return j
}

// Bool extracts a bool the JsonQuery
func (j *JSONQuery) Bool(s ...string) (bool, error) {
	val, err := rquery(j.blob, s...)
	if err != nil {
		return false, err
	}
	return boolFromInterface(val)
}

// Float extracts a float from the JsonQuery
func (j *JSONQuery) Float(s ...string) (float64, error) {
	val, err := rquery(j.blob, s...)
	if err != nil {
		return 0.0, err
	}
	return floatFromInterface(val)
}

// Int extracts an int from the JsonQuery
func (j *JSONQuery) Int(s ...string) (int, error) {
	val, err := rquery(j.blob, s...)
	if err != nil {
		return 0, err
	}
	return intFromInterface(val)
}

//GetInt extracts an int from the JsonQuery
func (j *JSONQuery) GetInt(s ...string) int {
	val, err := rquery(j.blob, s...)
	if err != nil {
		return 0
	}
	return GetintFromInterface(val)
}

// String extracts a string from the JsonQuery
func (j *JSONQuery) String(s ...string) (string, error) {
	val, err := rquery(j.blob, s...)
	if err != nil {
		return "", err
	}
	return stringFromInterface(val)
}

// GetString extracts a string from the JsonQuery
func (j *JSONQuery) GetString(s ...string) string {
	val, err := rquery(j.blob, s...)
	if err != nil {
		return ""
	}
	return getstringFromInterface(val)
}

// Object extracts a json object from the JsonQuery
func (j *JSONQuery) Object(s ...string) (map[string]interface{}, error) {
	val, err := rquery(j.blob, s...)
	if err != nil {
		return map[string]interface{}{}, err
	}
	return objectFromInterface(val)
}

// Array extracts a []interface{} from the JsonQuery
func (j *JSONQuery) Array(s ...string) ([]interface{}, error) {
	val, err := rquery(j.blob, s...)
	if err != nil {
		return []interface{}{}, err
	}
	return arrayFromInterface(val)
}

// Interface extracts an interface{} from the JsonQuery
func (j *JSONQuery) Interface(s ...string) (interface{}, error) {
	val, err := rquery(j.blob, s...)
	if err != nil {
		return nil, err
	}
	return val, nil
}

// ArrayOfStrings extracts an array of strings from some json
func (j *JSONQuery) ArrayOfStrings(s ...string) ([]string, error) {
	array, err := j.Array(s...)
	if err != nil {
		return []string{}, err
	}
	toReturn := make([]string, len(array))
	for index, val := range array {
		toReturn[index], err = stringFromInterface(val)
		if err != nil {
			return toReturn, err
		}
	}
	return toReturn, nil
}

// ArrayOfInts extracts an array of ints from some json
func (j *JSONQuery) ArrayOfInts(s ...string) ([]int, error) {
	array, err := j.Array(s...)
	if err != nil {
		return []int{}, err
	}
	toReturn := make([]int, len(array))
	for index, val := range array {
		toReturn[index], err = intFromInterface(val)
		if err != nil {
			return toReturn, err
		}
	}
	return toReturn, nil
}

// ArrayOfFloats extracts an array of float64s from some json
func (j *JSONQuery) ArrayOfFloats(s ...string) ([]float64, error) {
	array, err := j.Array(s...)
	if err != nil {
		return []float64{}, err
	}
	toReturn := make([]float64, len(array))
	for index, val := range array {
		toReturn[index], err = floatFromInterface(val)
		if err != nil {
			return toReturn, err
		}
	}
	return toReturn, nil
}

// ArrayOfBools extracts an array of bools from some json
func (j *JSONQuery) ArrayOfBools(s ...string) ([]bool, error) {
	array, err := j.Array(s...)
	if err != nil {
		return []bool{}, err
	}
	toReturn := make([]bool, len(array))
	for index, val := range array {
		toReturn[index], err = boolFromInterface(val)
		if err != nil {
			return toReturn, err
		}
	}
	return toReturn, nil
}

// ArrayOfObjects extracts an array of map[string]interface{} (objects) from some json
func (j *JSONQuery) ArrayOfObjects(s ...string) ([]map[string]interface{}, error) {
	array, err := j.Array(s...)
	if err != nil {
		return []map[string]interface{}{}, err
	}
	toReturn := make([]map[string]interface{}, len(array))
	for index, val := range array {
		toReturn[index], err = objectFromInterface(val)
		if err != nil {
			return toReturn, err
		}
	}
	return toReturn, nil
}

// ArrayOfArrays extracts an array of []interface{} (arrays) from some json
func (j *JSONQuery) ArrayOfArrays(s ...string) ([][]interface{}, error) {
	array, err := j.Array(s...)
	if err != nil {
		return [][]interface{}{}, err
	}
	toReturn := make([][]interface{}, len(array))
	for index, val := range array {
		toReturn[index], err = arrayFromInterface(val)
		if err != nil {
			return toReturn, err
		}
	}
	return toReturn, nil
}

// Matrix2D is an alias for ArrayOfArrays
func (j *JSONQuery) Matrix2D(s ...string) ([][]interface{}, error) {
	return j.ArrayOfArrays(s...)
}

// Recursively query a decoded json blob
func rquery(blob interface{}, s ...string) (interface{}, error) {
	var (
		val interface{}
		err error
	)
	val = blob
	for _, q := range s {
		val, err = query(val, q)
		if err != nil {
			return nil, err
		}
	}
	switch val.(type) {
	case nil:
		return nil, fmt.Errorf("Nil value found at %s", s[len(s)-1])
	}
	return val, nil
}

// query a json blob for a single field or index.  If query is a string, then
// the blob is treated as a json object (map[string]interface{}).  If query is
// an integer, the blob is treated as a json array ([]interface{}).  Any kind
// of key or index error will result in a nil return value with an error set.
func query(blob interface{}, query string) (interface{}, error) {
	index, err := strconv.Atoi(query)
	// if it's an integer, then we treat the current interface as an array
	if err == nil {
		switch blob.(type) {
		case []interface{}:
		default:
			return nil, fmt.Errorf("Array index on non-array %v", blob)
		}
		if len(blob.([]interface{})) > index {
			return blob.([]interface{})[index], nil
		}
		return nil, fmt.Errorf("Array index %d on array %v out of bounds", index, blob)
	}

	// blob is likely an object, but verify first
	switch blob.(type) {
	case map[string]interface{}:
	default:
		return nil, fmt.Errorf("Object lookup \"%s\" on non-object %v", query, blob)
	}

	val, ok := blob.(map[string]interface{})[query]
	if !ok {
		return nil, fmt.Errorf("Object %v does not contain field %s", blob, query)
	}
	return val, nil
}

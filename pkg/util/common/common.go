package common

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
)

func StringExistsInSlice(target string, slice []string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

func CapitalizeString(input string) string {
	return strings.ToUpper(string(input[0])) + input[1:]
}

func ToQueryParams(s interface{}) string {
	values := url.Values{}
	v := reflect.ValueOf(s)
	t := reflect.TypeOf(s)

	// Iterate over all struct fields
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		typeField := t.Field(i)

		// Get the field's JSON tag if present, otherwise use the field name
		tag := typeField.Tag.Get("json")
		if tag == "" {
			tag = typeField.Name
		} else {
			// Strip out any options like 'omitempty'
			tag = strings.Split(tag, ",")[0]
		}
		if tag == "-" {
			continue
		}

		// Handle pointer fields
		if field.Kind() == reflect.Ptr && field.IsNil() {
			continue
		}

		// Handle different field types
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			values.Set(tag, strconv.FormatInt(field.Int(), 10))
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			values.Set(tag, strconv.FormatUint(field.Uint(), 10))
		case reflect.Float32, reflect.Float64:
			values.Set(tag, strconv.FormatFloat(field.Float(), 'f', -1, 64))
		case reflect.Bool:
			values.Set(tag, strconv.FormatBool(field.Bool()))
		case reflect.String:
			values.Set(tag, field.String())
		case reflect.Slice:
			for j := 0; j < field.Len(); j++ {
				values.Add(tag, fmt.Sprintf("%v", field.Index(j)))
			}
		case reflect.Map:
			for _, key := range field.MapKeys() {
				values.Set(fmt.Sprintf("%s[%v]", tag, key), fmt.Sprintf("%v", field.MapIndex(key)))
			}
		case reflect.Ptr:
			if !field.IsNil() {
				values.Set(tag, fmt.Sprintf("%v", field.Elem()))
			}
		default:
			// Handle nested structs
			if field.Kind() == reflect.Struct {
				nestedParams := ToQueryParams(field.Interface())
				var parsed, _ = url.ParseQuery(nestedParams)
				for key, val := range parsed {
					for _, v := range val {
						values.Add(fmt.Sprintf("%s.%s", tag, key), v)
					}
				}
			}
		}
	}

	return values.Encode()
}

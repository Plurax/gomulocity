package generic

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

func JsonFromObject(o interface{}) ([]byte, error) {
	// is it a pointer of struct?
	structValue, ok := pointerOfStruct(&o)
	if ok == false {
		return nil, errors.New("input is not a pointer of struct")
	}

	// Convert the struct to a map
	mapValue, err := mapFromStruct(structValue)
	if err != nil {
		return nil, err
	}

	// Marshal the map to a json string
	j, err := json.Marshal(mapValue)
	if err != nil {
		return nil, err
	}

	return j, nil
}

/*
Maps a given struct to a map.
Handles `json:...` tags and `jsonc:...` tags for flattening.
Returns the map or an error
*/
func mapFromStruct(structValue *reflect.Value) (*map[string]interface{}, error) {
	targetMap := make(map[string]interface{})
	structType := structValue.Type()

	// Iterate over the struct fields
	for i := 0; i < structValue.NumField(); i++ {
		fieldType := structType.Field(i)
		fieldValue := structValue.Field(i)
		jsonCTag := getJsonTag(&fieldType, "jsonc")

		// If no `jsonc` tag: Just add the field into the map - maybe has `json`-Tags
		if jsonCTag == nil {
			insertTaggedFieldIntoMap(&targetMap, &fieldType, &fieldValue)
		} else {
			switch jsonCTag.Name {
			// `jsonc:"flat"` -> Must be a map. Flatten it to the target Map.
			case "flat":
				err := insertReflectMapIntoMap(&targetMap, &fieldValue)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("error: on collection %s: %s", fieldType.Name, err.Error()))
				}
				break
			// `jsonc:"collection"` -> Must be a slice. Handle each element recursive with `mapFromStruct`
			case "collection":
				err := handleCollection(&targetMap, &fieldType, &fieldValue)
				if err != nil {
					return nil, errors.New(fmt.Sprintf("error: on collection %s: %s", fieldType.Name, err.Error()))
				}
				break
			}
		}
	}

	return &targetMap, nil
}

/*
 * Handles `json:"collection"`
 * Must be a slice. Then handle every slice element as a struct and call it with `mapFromStruct`. Then insert this new
 * map into the `targetMap`.
 */
func handleCollection(targetMapPtr *map[string]interface{}, fieldType *reflect.StructField, fieldValue *reflect.Value) error {
	if fieldValue.Kind() != reflect.Slice {
		return errors.New("is not a slice")
	}

	slice := make([]map[string]interface{}, fieldValue.Len())
	for i := 0; i < fieldValue.Len(); i++ {
		structItem := fieldValue.Index(i)
		mapItem, err := mapFromStruct(&structItem)
		if err != nil {
			return errors.New(fmt.Sprintf("error: Can not convert item %d: %s", i, err.Error()))
		}

		slice[i] = *mapItem
	}

	v := reflect.ValueOf(slice)
	insertTaggedFieldIntoMap(targetMapPtr, fieldType, &v)

	return nil
}

// `fieldValue` must be a Map. Then insert every element into the target map.
func insertReflectMapIntoMap(targetMapPtr *map[string]interface{}, fieldValue *reflect.Value) error {
	targetMap := *targetMapPtr
	if fieldValue.Kind() != reflect.Map {
		return errors.New("is not a map")
	}

	// flat process
	iter := fieldValue.MapRange()
	for iter.Next() {
		targetMap[iter.Key().String()] = iter.Value().Interface()
	}

	return nil
}

/*
 * Handles a field and its value based on the given tags and insert it into the given map.
 * eg: A field "A -> "foo" without any tag, will result in map["A"] -> "foo"
       A field "A -> "foo" `json:customA` will result in map["customA"] -> "foo"
*/
func insertTaggedFieldIntoMap(targetMapPtr *map[string]interface{}, fieldType *reflect.StructField, fieldValue *reflect.Value) {
	tag := getJsonTag(fieldType, "json")
	targetMap := *targetMapPtr

	// no tag -> original name
	if tag == nil {
		targetMap[fieldType.Name] = fieldValue.Interface()
		return
	}

	// - -> omit value
	if tag.Name == "-" {
		return
	}

	// OmitEmpty and is empty -> omit value
	if tag.OmitEmpty && isEmptyValue(fieldValue) {
		return
	} else {
		targetMap[tag.Name] = fieldValue.Interface()
	}
}

func isEmptyValue(v *reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	default:
		return !v.IsValid() || v.IsZero()
	}
}

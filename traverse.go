package reflect

import (
	"fmt"
	"reflect"
	"strconv"
)

// State
type State struct {
	Path  string
	Depth int
	// Custom value.
	Value interface{}
}

func (s *State) next(path string) *State {
	return &State{Path: path, Depth: s.Depth + 1, Value: s.Value}
}

// ProcessValue is type of callback function for Traverse function.
type ProcessValue func(value reflect.Value, state *State, field *reflect.StructField) error

// Traverse iterates through all the nested elements of the passed variable.
func Traverse(v interface{}, process ProcessValue) error {
	return TraverseValue(reflect.ValueOf(v), process)
}

// TraverseValue iterates through all the nested elements of the passed variable.
func TraverseValue(v reflect.Value, process ProcessValue) error {
	return traverseValue(v, &State{}, nil, process)
}

func addFieldName(path, field string) string {
	if path == "" {
		return field
	}
	return path + "." + field
}

func traverseValue(v reflect.Value, state *State, field *reflect.StructField, process ProcessValue) error {
	//v = reflect.Indirect(v)
	if err := process(v, state, field); err != nil {
		return err
	}
	//depth++

	switch v.Kind() {
	case reflect.Struct:
		structType := v.Type()
		for i := 0; i < structType.NumField(); i++ {
			structField := structType.Field(i)
			fieldValue := v.Field(i)
			if err := traverseValue(fieldValue, state.next(addFieldName(state.Path, structField.Name)), &structField, process); err != nil {
				return err
			}
		}
	case reflect.Slice, reflect.Array:
		length := v.Len()
		for i := 0; i < length; i++ {
			if err := traverseValue(v.Index(i), state.next(state.Path+"["+strconv.Itoa(i)+"]"), nil, process); err != nil {
				return err
			}
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			keyStr := "[]"
			if key.CanInterface() {
				keyStr = fmt.Sprintf("[%v]", key.Interface())
			}
			if err := traverseValue(v.MapIndex(key), state.next(state.Path+keyStr), nil, process); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if !v.IsNil() {
			if err := traverseValue(v.Elem(), state.next("*("+state.Path+")"), nil, process); err != nil {
				return err
			}
		}
	}

	return nil
}

// TraverseFields iterates through all structs's fields of the passed variable.
func TraverseFields(v interface{}, processField ProcessValue) error {
	return TraverseValueFields(reflect.ValueOf(v), processField)
}

// TraverseValueFields iterates through all structs's fields of the passed variable.
func TraverseValueFields(v reflect.Value, processField ProcessValue) error {
	process := func(value reflect.Value, state *State, field *reflect.StructField) error {
		if field != nil {
			return processField(value, state, field)
		}
		return nil
	}
	return traverseValue(v, &State{}, nil, process)
}

// Clear variable
func Clear(v interface{}) {
	p := reflect.ValueOf(v).Elem()
	p.Set(reflect.Zero(p.Type()))
}

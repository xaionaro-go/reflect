package reflect

import (
	"encoding"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

var (
	errValueIsNotAssignable = errors.New("Invalid value, should be assignable")
)

// type TextMarshalers map[reflect.Type]func() (text []byte, err error)
// type TextUnmarshalers map[reflect.Type]func(text []byte) error

func getUnmarshaler(v reflect.Value) encoding.TextUnmarshaler {
	if v.CanInterface() {
		if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
			return u
		}
	}
	return nil
}

func findUnmarshaler(v reflect.Value) encoding.TextUnmarshaler {
	if u := getUnmarshaler(v); u != nil {
		return u
	}
	if v.CanAddr() {
		if u := getUnmarshaler(v.Addr()); u != nil {
			return u
		}
	}
	if u := getUnmarshaler(reflect.Indirect(v)); u != nil {
		return u
	}
	return nil
}

// AssignStringToValue tries to convert the string to the appropriate type and assign it to the destination variable.
func AssignStringToValue(dst reflect.Value, src string) (err error) {
	if !dst.CanSet() {
		return errValueIsNotAssignable
	}

	v := dst
	var ptr reflect.Value
	// Allocate if pointer is nil
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			ptr = reflect.New(v.Type().Elem())
			v = ptr.Elem()
		} else {
			v = v.Elem()
		}
	}
	if u := findUnmarshaler(v); u != nil {
		return u.UnmarshalText([]byte(src))
	}
	if v.CanInterface() {
		if _, ok := v.Interface().(time.Duration); ok {
			duration, err := time.ParseDuration(src)
			if err == nil {
				v.Set(reflect.ValueOf(duration))
			}
			return err
		}
	}
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(src, 0, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		ui, err := strconv.ParseUint(src, 0, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetUint(ui)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(src, v.Type().Bits())
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(src)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.String:
		v.SetString(src)
	default:
		err = fmt.Errorf("Unable to convert string \"%s\" to type \"%s\"", src, v.Type().Name())
	}
	if ptr.Kind() == reflect.Ptr {
		dst.Set(ptr)
	}
	return
}

// AssignString tries to convert the string to the appropriate type and assign it to the destination variable.
func AssignString(dst interface{}, src string) error {
	return AssignStringToValue(reflect.Indirect(reflect.ValueOf(dst)), src)
}

// AssignValue tries to convert source to destination. If possible, it converts the string to the destination type.
func AssignValue(dst, src reflect.Value) (err error) {
	//dst = reflect.Indirect(dst)
	if !dst.CanSet() {
		return errValueIsNotAssignable
	}
	src = reflect.Indirect(src)

	if src.Kind() == reflect.String {
		return AssignStringToValue(dst, src.String())
	}

	dstValue, dstNew := dst, dst
	for dstValue.Kind() == reflect.Ptr {
		if dstValue.IsNil() {
			dstValue.Set(reflect.New(dstValue.Type().Elem()))
		}
		dstValue = dstValue.Elem()
	}
	if src.Type().ConvertibleTo(dstValue.Type()) {
		dstValue.Set(src.Convert(dstValue.Type()))
		dst.Set(dstNew)
		return nil
	}
	if dstType := dstValue.Type(); src.Kind() == reflect.Slice && dstType.Kind() == reflect.Slice {
		i := src.Interface()
		println(i)
		s := reflect.MakeSlice(dstType, src.Len(), src.Len())
		for i := 0; i < src.Len(); i++ {
			if err := AssignValue(s.Index(i), src.Index(i)); err != nil {
				return err
			}
		}
		dstValue.Set(s)
		dst.Set(dstNew)
		return nil
	}
	return fmt.Errorf("value of type \"%s\" cannot be converted to type \"%s\"", src.Type().Name(), dstValue.Type().Name())
}

// Assign tries to convert source to destination. If possible, it converts the string to the destination type.
func Assign(dst, src interface{}) error {
	return AssignValue(reflect.Indirect(reflect.ValueOf(dst)), reflect.ValueOf(src))
}

package util

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// UnmarshalConfig unmarshals a map of strings into the Config struct using dot notation.
func UnmarshalConfig(config any, data map[string]string) error {
	for key, value := range data {
		if err := setField(config, key, value); err != nil {
			return err
		}
	}
	return nil
}

// setField sets the value of a field in the Config struct based on the dot notation key.
func setField(config any, key string, value string) error {
	keys := strings.Split(key, ".")
	v := reflect.ValueOf(config).Elem()

	for _, k := range keys {
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			// ignoring unknown fields
			continue
		}
		v = v.FieldByNameFunc(func(name string) bool {
			return strings.EqualFold(name, k)
		})
		if !v.IsValid() {
			continue
		}
	}

	if v.CanSet() {
		switch v.Kind() {
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			i, err := strconv.ParseUint(value, 10, 64)
			if err != nil {
				return err
			}
			v.SetUint(i)
		case reflect.Float32, reflect.Float64:
			f, err := strconv.ParseFloat(value, 64)
			if err != nil {
				return err
			}
			v.SetFloat(f)
		case reflect.Int64:
			switch v.Type().String() {
			case "time.Duration":
				d, err := time.ParseDuration(value)
				if err != nil {
					return err
				}
				v.Set(reflect.ValueOf(d))
			case "int64":
				i, err := strconv.Atoi(value)
				if err != nil {
					return err
				}
				v.SetInt(int64(i))
			}
		case reflect.String:
			v.SetString(value)
		case reflect.Bool:
			switch strings.ToLower(value) {
			case "true":
				v.SetBool(true)
			case "false":
				v.SetBool(false)
			default:
				return errors.New("invalid boolean value for field: " + key)
			}
		case reflect.Int:
			// Handle integer parsing if needed
			i, err := strconv.Atoi(value)
			v.SetInt(int64(i))
			if err != nil {
				return err
			}
		// Add cases for other types as needed
		default:
			return errors.New("unsupported field type for field: " + key)
		}
	}
	return nil
}

package json_field

import "reflect"

func Valuer(field reflect.Value) interface{} {
	if field.Kind() == reflect.Struct {
		presentField := field.FieldByName("Present")

		if presentField.IsValid() && presentField.Bool() {
			valueField := field.FieldByName("Value")

			if valueField.IsValid() && !valueField.IsNil() {
				return valueField.Elem().Interface()
			}
		}

		return nil
	}
	return nil
}

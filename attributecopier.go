package apikit

import (
	"reflect"
	"errors"
)

// Copies attributes that should be immutable from the source RESTObject to the dest RESTObject
type ImmutableAttributeCopier interface {
	RESTObject
	CopyImmutableAttributesTo(dest interface{}) error
}

func CopyImmutableAttributes(source, dest interface{}) error {
	if dest == nil {
		return errors.New("Given a nil destination object")
	}
	if source == nil {
		return errors.New("Given a nil source object")
	}
	if copier, ok := source.(ImmutableAttributeCopier); ok {
		// use the custom implementation if exists
		return copier.CopyImmutableAttributesTo(dest)
	} else {
		// copy immutable attributes based on struct tags
		var vOld reflect.Value = reflect.ValueOf(source)
		if vOld.Type().Kind() == reflect.Ptr {
			if vOld.Elem().Type().Kind() == reflect.Struct {
				vOld = vOld.Elem()
			} else {
				return errors.New("Source is not a pointer to a struct")
			}
		} else {
			return errors.New("Source is not a pointer to a struct")
		}

		var vNew reflect.Value = reflect.ValueOf(dest)
		if vNew.Type().Kind() == reflect.Ptr {
			if vNew.Elem().Type().Kind() == reflect.Struct {
				vNew = vNew.Elem()
			} else {
				return errors.New("Destination is not a pointer to a struct")
			}
		} else {
			return errors.New("Destination is not a pointer to a struct")
		}
		const tagKeyName = "apikit"
		const immutableKeyValue = "immutable"

		for i := 0; i < vOld.NumField(); i++ {
			oldFieldType := vOld.Type().Field(i)
			oldFieldVal := vOld.Field(i)
			fieldName := oldFieldType.Name
			if oldFieldType.Tag.Get(tagKeyName) == immutableKeyValue {
				// this field was marked as immutable
				if newField := vNew.FieldByName(fieldName); newField.IsValid() && newField.CanSet() {
					newField.Set(vOld.FieldByName(fieldName))
				}
			} else {
				// check to see if it is a struct
				if oldFieldType.Anonymous {
					newFieldType := vNew.Type().Field(i)
					if newFieldVal := vNew.FieldByName(fieldName); newFieldVal.IsValid() && newFieldType.Anonymous {
						o := oldFieldVal.Addr().Interface()
						n := newFieldVal.Addr().Interface()
						if err := CopyImmutableAttributes(o, n); err != nil {
							return err
						}
					}
				}
			}
		}
		return nil
	}
}

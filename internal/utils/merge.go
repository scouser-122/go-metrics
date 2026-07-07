package utils

import "reflect"

func MergeStructs(dst, src interface{}) {
	dstVal := reflect.ValueOf(dst).Elem()
	srcVal := reflect.ValueOf(src).Elem()

	for i := 0; i < dstVal.NumField(); i++ {
		dstField := dstVal.Field(i)
		srcField := srcVal.Field(i)

		// Check if source field is not zero value
		if !srcField.IsZero() {
			// Check if destination field is zero value
			if dstField.IsZero() {
				dstField.Set(srcField)
			}
		}
	}
}

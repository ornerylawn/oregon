package knob

import (
	"fmt"
	"reflect"
)

func PrintKnobs(e interface{}) error {
	// e must be a pointer to struct.
	ptr := reflect.ValueOf(e)
	if ptr.Kind() != reflect.Ptr {
		return fmt.Errorf("knob: expected a pointer to struct but was %v", reflect.TypeOf(e))
	}
	v := ptr.Elem()
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("knob: expected a pointer to struct but was %v", reflect.TypeOf(e))
	}
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		vf := reflect.Indirect(v.Field(i))
		tf := t.Field(i)
		if tf.PkgPath != "" {
			continue
		}
		switch vf.Kind() {
		case reflect.Struct:
			err := PrintKnobs(vf.Addr().Interface())
			if err != nil {
				return err
			}
		case reflect.Slice:
			fallthrough
		case reflect.Array:
			for j := 0; j < vf.Len(); j++ {
				err := PrintKnobs(reflect.Indirect(vf.Index(j)).Addr().Interface())
				if err != nil {
					return err
				}
			}
		default:
			if tf.Tag.Get("knob") != "" {
				fmt.Println(tf.Name, tf.Type, vf.Kind(), tf.Tag.Get("knob"))
			}
		}
	}
	return nil
}

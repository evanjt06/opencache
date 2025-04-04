package internal

import (
	"fmt"
	"reflect"
)

func ValidateKey(key interface{}) error {
	if key == nil {
		return fmt.Errorf("key cannot be nil")
	}
	if !reflect.TypeOf(key).Comparable() {
		return fmt.Errorf("invalid key type: %T is not comparable", key)
	}

	return nil
}

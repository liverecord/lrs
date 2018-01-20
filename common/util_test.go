package common

import (
	"testing"
	"os"
)

func TestEnv_ExistingValue(t *testing.T) {
	key := "TEST_KEY"
	value := "TEST_VALUE"

	os.Setenv(key, value)
	defer os.Unsetenv(key)

	res := Env(key, "default")
	if res != value {
		t.Errorf("I expected to get \"%s\" but got \"%s\"", value, res)
	}
}

func TestEnv_NotExistingValue(t *testing.T) {
	key := "TEST_KEY"
	def := "DEFAULT"

	res := Env(key, def)
	if res != def {
		t.Errorf("I expected to get \"%s\" but got \"%s\"", def, res)
	}
}

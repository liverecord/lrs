package common

import (
	"os"
	"testing"
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

func TestBoolEnv_ExistingValue(t *testing.T) {
	testCases := map[string]bool{
		"TRUE":  true,
		"true":  true,
		"FALSE": false,
		"false": false,
	}

	key := "TEST_KEY"

	for envValue, expectedValue := range testCases {
		os.Setenv(key, envValue)

		res := BoolEnv(key, !expectedValue)
		if res != expectedValue {
			t.Errorf("I expected to get %v but got %v", expectedValue, res)
		}

		os.Unsetenv(key)
	}
}

func TestBoolEnv_NotExistingValue(t *testing.T) {
	testCases := map[bool]bool{
		true:  true,
		false: false,
	}

	key := "TEST_KEY"

	for defValue, expectedValue := range testCases {
		res := BoolEnv(key, defValue)
		if res != expectedValue {
			t.Errorf("I expected to get %v but got %v", expectedValue, res)
		}
	}
}

package enverrors

import "errors"

var ErrUntaggedEnvStructField = errors.New("untagged env struct field")

type EnvVariableNotFound struct {
	VariableName string
}

func (e *EnvVariableNotFound) Error() string {
	return "env variable not found: " + e.VariableName
}

package env

import (
	"os"
	"reflect"
	"strconv"

	"github.com/alexkalak/whatever-system/src/shared/tools/env/enverrors"
	"github.com/joho/godotenv"
)

type Env struct {
	IsDev            bool   `env:"IS_DEV"`
	BscRPCURL        string `env:"BSC_RPC_URL"`
	PostgresHost     string `env:"POSTGRES_HOST"`
	PostgresUser     string `env:"POSTGRES_USER"`
	PostgresPassword string `env:"POSTGRES_PASSWORD"`
	PostgresDBName   string `env:"POSTGRES_DBNAME"`
	PostgresPort     int    `env:"POSTGRES_PORT"`
}

var env Env
var isLoaded bool

func GetEnv() (Env, error) {
	if !isLoaded {
		err := godotenv.Load()
		if err != nil {
			return env, err
		}
		isLoaded = true
	}

	err := fillEnv()

	return env, err
}

func fillEnv() error {
	envReflectValue := reflect.ValueOf(&env).Elem()
	envReflectType := envReflectValue.Type()

	for i := 0; i < envReflectValue.NumField(); i++ {
		field := envReflectValue.Field(i)
		structField := envReflectType.Field(i)

		tag := structField.Tag.Get("env")

		if tag == "" {
			return enverrors.ErrUntaggedEnvStructField
		}

		strValue, err := loadVariable(tag)
		if err != nil {
			return err
		}

		switch field.Kind() {
		case reflect.String:
			field.SetString(strValue)
		case reflect.Int:
			n, err := strconv.Atoi(strValue)
			if err != nil {
				return err
			}
			field.SetInt(int64(n))
		case reflect.Bool:
			b, err := strconv.ParseBool(strValue)
			if err != nil {
				return err
			}
			field.SetBool(b)
		}
	}

	return nil
}

func loadVariable(name string) (string, error) {
	value := os.Getenv(name)
	if value == "" {
		return "", &enverrors.EnvVariableNotFound{VariableName: name}
	}

	return value, nil
}

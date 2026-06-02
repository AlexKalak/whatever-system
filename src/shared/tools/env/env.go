package env

import (
	"os"
	"reflect"
	"strconv"

	"github.com/alexkalak/whatever-system/src/shared/tools/env/enverrors"
	"github.com/joho/godotenv"
)

type Env struct {
	IsDev                        bool   `env:"IS_DEV"`
	BscRPCWsURL                  string `env:"BSC_RPC_WS_URL"`
	BscRPCHTTPSURL               string `env:"BSC_RPC_HTTPS_URL"`
	EthRPCWsURL                  string `env:"ETH_RPC_WS_URL"`
	EthRPCHTTPSURL               string `env:"ETH_RPC_HTTPS_URL"`
	BscMulticall3Address         string `env:"BSC_MULTICALL3_ADDRESS"`
	EthMulticall3Address         string `env:"ETH_MULTICALL3_ADDRESS"`
	PostgresHost                 string `env:"POSTGRES_HOST"`
	PostgresUser                 string `env:"POSTGRES_USER"`
	PostgresPassword             string `env:"POSTGRES_PASSWORD"`
	PostgresDBName               string `env:"POSTGRES_DBNAME"`
	PostgresPort                 int    `env:"POSTGRES_PORT"`
	KafkaBrokers                 string `env:"KAFKA_BROKERS"`
	KafkaDexConsumerGroup        string `env:"KAFKA_DEX_CONSUMER_GROUP"`
	KafkaDexActionsConsumerGroup string `env:"KAFKA_DEX_ACTIONS_CONSUMER_GROUP"`
	KafkaMempoolConsumerGroup    string `env:"KAFKA_MEMPOOL_CONSUMER_GROUP"`
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

package psql

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type DB struct {
	URL    string
	Driver string
}

type Config struct {
	StateMachine DB
	Templates    DB
}

func NewConfig() Config {
	config := Config{
		StateMachine: DB{
			URL:    parseDbURL("db.stateMachine"),
			Driver: viper.GetString("db.stateMachine.driver"),
		},

		Templates: DB{
			URL:    parseDbURL("db.stateMachine"),
			Driver: viper.GetString("db.stateMachine.driver"),
		},
	}

	return config
}

func parseDbURL(dbConfigName string) string {
	host := viper.GetString(fmt.Sprintf("%s.host", dbConfigName))
	port := viper.GetString(fmt.Sprintf("%s.port", dbConfigName))
	dbname := viper.GetString(fmt.Sprintf("%s.dbname", dbConfigName))
	user := viper.GetString(fmt.Sprintf("%s.user", dbConfigName))
	password := viper.GetString(fmt.Sprintf("%s.password", dbConfigName))

	urlTemplate := "host=%s port=%s dbname=%s user=%s password=%s  sslmode=disable"

	result := fmt.Sprintf(urlTemplate, host, port, dbname, user, password)

	log.Println(result)

	return result
}

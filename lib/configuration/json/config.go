package json

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

const configFileName = "config.json"

type Config struct {
	DBType string
	DBName string
}

func ConstructDefaultConfig() []byte {
	defaultConfig := Config{
		DBType: "json",
		DBName: "tasks.json"}

	defaultConfigJson, _ := json.Marshal(defaultConfig)
	return defaultConfigJson
}

func InitializeDefaultConfig() {
	SaveConfig(ConstructDefaultConfig())
}

func SaveConfig(jsonContent []byte) {
	err := ioutil.WriteFile(configFileName, jsonContent, 0660)
	if err != nil {
		fmt.Println(err)
	}
}

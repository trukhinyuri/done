package json

import (
	"encoding/json"
	"testing"
)

func Test_ConstructDefaultConfig(t *testing.T) {
	var config Config
	var jsonContent = ConstructDefaultConfig()
	json.Unmarshal(jsonContent, &config)
	println(config.DBName)
}

func Test_SaveConfig(t *testing.T) {

}

func Test_InitializeDefaultConfig(t *testing.T) {
	InitializeDefaultConfig()
}

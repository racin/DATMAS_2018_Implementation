package configuration

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

func TestLoadAppConfig(t *testing.T) {
	conf, err := LoadAppConfig("test/appConfig")
	if err != nil {
		t.Fatal("Error loading app config: " + err.Error())
	}

	assert.NotEmpty(t, conf, "App config is empty")
	oldVal := (*conf).RpcType
	(*conf).RpcType = "Testing__"

	conf2 := AppConfig()

	assert.Equal(t, "Testing__", conf2.RpcType, "Could not set attribute in config")
	conf.RpcType = oldVal

	assert.Equal(t, &conf, &conf2, "Pointer not equal")
	fmt.Printf("%+v", *AppConfig())
}

func TestLoadClientConfig(t *testing.T) {
	conf, err := LoadClientConfig("test/clientConfig")
	if err != nil {
		t.Fatal("Error loading app config: " + err.Error())
	}

	assert.NotEmpty(t, conf, "App config is empty")
	oldVal := (*conf).RemoteAddr
	(*conf).RemoteAddr = "Testing__"

	conf2 := ClientConfig()

	assert.Equal(t, "Testing__", conf2.RemoteAddr, "Could not set attribute in config")
	conf.RemoteAddr = oldVal

	assert.Equal(t, &conf, &conf2, "Pointer not equal")
	fmt.Printf("%+v", *ClientConfig())
}
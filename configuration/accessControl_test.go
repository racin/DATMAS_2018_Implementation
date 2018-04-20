package configuration

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAccessControl (t *testing.T){
	appConf, err := LoadAppConfig("../configuration/test/appConfig");
	if err != nil {
		t.Fatal("Error loading app config: " + err.Error())
	}

	acl := GetAccessList(ListPathTest)
	assert.NotEmpty(t, acl.Identities, "Access list empty")

	acl2 := GetAccessList(appConf.BasePath + appConf.AccessList)
	assert.NotEqual(t, acl, acl2, "Test data is not isolated.")
}
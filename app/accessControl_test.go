package app

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/racin/DATMAS_2018_Implementation/configuration"
)

func TestAccessControl (t *testing.T){
	if _, err := configuration.LoadAppConfig("../configuration/test/appConfig"); err != nil {
		t.Fatal("Error loading app config: " + err.Error())
	}

	acl := GetAccessList("test")
	assert.NotEmpty(t, acl.Identities, "Access list empty")

	acl2 := GetAccessList()
	assert.NotEqual(t, acl, acl2, "Test data is not isolated.")
}
package app

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestAccessControl (t *testing.T){
	acl := GetAccessList(true, true)
	assert.NotEmpty(t, acl.Identities, "Access list empty")

	acl2 := GetAccessList()
	assert.NotEqual(t, acl, acl2, "Test data is not isolated.")
}
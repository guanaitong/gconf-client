package gconf_mongo

import (
	"testing"

	"github.com/guanaitong/gconf-go-client"
)

func TestMongoConfig_NewClient(t *testing.T) {
	gconf.Init("gce-api-saas")
	c := GetDefaultMongoConfig()
	if c == nil {
		t.Error("c is nil")
	}

	client := c.NewClient()
	if client == nil {
		t.Error("client is nil")
	}
}

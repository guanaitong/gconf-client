package gconf_redis

import (
	"context"
	"testing"

	"github.com/guanaitong/gconf-go-client"
)

const (
	key      = "key"
	value    = "value"
	expected = "value"
)

func TestGetRedisConfig(t *testing.T) {
	gconf.Init("for-test-java")
	d := GetDefaultRedisConfig()
	d2 := GetRedisConfig("redis-config.json")
	if d == nil {
		t.Error("d is not nil")
	}
	if d2 == nil {
		t.Error("d2 is not nil")
	}
	client := d.NewClient()
	ctx := context.Background()
	err := client.Set(ctx, key, value, 0).Err()
	if err != nil {
		t.Error(err)
	}
	actual, err := client.Get(ctx, "key").Result()
	if err != nil {
		t.Error(err)
	}
	if actual != value {
		t.Error("not expect value")
	}

}

package gconf_mysql

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/guanaitong/gconf-go-client"
)

func TestGetDataSourceConfig(t *testing.T) {
	gconf.Init("gce-api-ucenter")
	d := GetDefaultMySQLDataSourceConfig()
	d2 := GetMySQLDataSourceConfig("datasource.json")
	if d == nil {
		t.Error("d is not nil")
	}
	if d2 == nil {
		t.Error("d2 is not nil")
	}

	m := d.MasterDataSourceName()

	s := d.SlaveDataSourceName()

	if m == "" || s == "" {
		t.Error("m or s is not nil")
	}
	db, err := d.OpenMaster()
	if err != nil {
		t.Error(err)
	}
	t.Log(db)

	err = db.Ping()
	if err != nil {
		t.Error(err)
	}
	//db.Exec()
}

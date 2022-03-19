package gconf

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func init() {
	Init("userdoor")
}

type es struct {
	ClusterName string
	hosts       string
	Port        int32
}

type impower struct {
	Path   string `config:"path"`
	Cmd    string `config:"cmd"`
	Period int    `config:"period"`
	HasX   bool   `config:"has"`
}

func TestGetConfigCollection(t *testing.T) {
	t.Log("[impower]-------------------------------")
	d1 := GetConfigCollection("impower")
	t.Log(d1, d1.AsMap())

	dm1 := d1.GetValue("deny.properties")
	t.Log(dm1)

	imp := new(impower)
	if err := d1.GetValue("deny.properties").Register(imp); err != nil {
		t.Error(err)
	} else if imp.Cmd != "date122" || imp.Path != "/tmp/impower" {
		t.Error(errors.New("not same from server"))
	}
	t.Log("[golang]------------------------------")
	if GetGlobalConfigCollection() == nil {
		t.Fail()
	}
	c := GetConfigCollection("userdoor")
	c1 := GetCurrentConfigCollection()
	if c != c1 {
		t.FailNow()
	}
}

type testBean struct {
	a string `config:"a"`
	b int
	c uint64
}

func TestReflectBase(t *testing.T) {

	p := new(testBean)
	v := reflect.ValueOf(p).Elem()

	for i := 0; i < v.NumField(); i++ {
		fieldInfo := v.Type().Field(i) // a reflect.StructField
		fieldInfo1 := v.Field(i)
		fmt.Println(i)
		fmt.Println(fieldInfo.Name)
		fmt.Println(fieldInfo.Type)
		fmt.Println(fieldInfo.Tag)
		fmt.Println(fieldInfo1.Type().Name())

	}
}

func TestDecrypt(t *testing.T) {
	epwd := "E+lDullrAU/qV1MVqR7L0GrbBkFHWaftsKTVni3ooL90/PZyH/VpcKF/HqJqzAyzoHI8vR+tawW/kE5sgRcpVkYivugNhWhEtnQpbRNjvnkCd8OcyuhjEVnrzDg4iNtJ4+RWKq37vb4aXU1/skmXDLd1Jf2ZNYndzTgHM1EbP6Ac0KqWzpeS4o2QxtX4E1nzdrxCOtEYtTewtXiaxA4kHdVb6fIkLa/OvY2xDNOQZKhlw9IU6LC3Ypq8qqQPq1dCW+Y/TzktZcbKVmQ0aHchPLuWpiO2VNwojHu7hiD7ZiNsELiDvose8iNNSwwpfTKbIODqjtgBrRWD/VLjCbMcxg=="
	pwd := Decrypt(epwd)
	if pwd == "" {
		t.Error("error")
	}
	fmt.Println(pwd)
}

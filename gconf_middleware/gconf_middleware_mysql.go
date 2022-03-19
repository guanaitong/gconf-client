package gconf_middleware

import (
	"crypto/rsa"
	"crypto/x509"
	"database/sql"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/guanaitong/gconf-go-client"
)

const defaultMySQLConfigKey = "datasource.json"

type MysqlServer struct {
	Name      string `json:"name"`
	Domain    string `json:"domain"`
	Ip        string `json:"ip"`
	Port      string `json:"port"`
	Version   string `json:"version"`
	GroupName string `json:"groupName"`
	Role      string `json:"role"`
}

type MySQLDataSourceConfig struct {
	DbName            string            `json:"dbName"`
	Username          string            `json:"username"`
	EncryptedPassword string            `json:"encryptedPassword"`
	Password          string            `json:"password"`
	GroupName         string            `json:"groupName"`
	MysqlServers      []*MysqlServer    `json:"mysqlServers"`
	Params            map[string]string `json:"params"`
}

func (dataSourceConfig *MySQLDataSourceConfig) OpenMaster() (db *sql.DB, err error) {
	return dataSourceConfig.open(false)
}

func (dataSourceConfig *MySQLDataSourceConfig) OpenSlave() (db *sql.DB, err error) {
	return dataSourceConfig.open(true)
}

func (dataSourceConfig *MySQLDataSourceConfig) open(preferSlave bool) (db *sql.DB, err error) {
	db, err = sql.Open("mysql", dataSourceConfig.dataSourceName(preferSlave))
	if err == nil {
		db.SetMaxOpenConns(dataSourceConfig.getParamValue("maxOpenConns", 20))
		db.SetMaxIdleConns(dataSourceConfig.getParamValue("maxIdleConns", 3))
		db.SetConnMaxLifetime(time.Second * time.Duration(dataSourceConfig.getParamValue("maxIdleConns", 1200)))
	}
	return
}

func (dataSourceConfig *MySQLDataSourceConfig) MasterDataSourceName() string {
	return dataSourceConfig.dataSourceName(false)
}

func (dataSourceConfig *MySQLDataSourceConfig) SlaveDataSourceName() string {
	return dataSourceConfig.dataSourceName(true)
}

func (dataSourceConfig *MySQLDataSourceConfig) dataSourceName(preferSlave bool) string {
	var pwd = decrypt(dataSourceConfig.EncryptedPassword)
	if pwd == "" {
		pwd = dataSourceConfig.Password
	}
	mysqlServer := dataSourceConfig.getMysqlServer(preferSlave)
	if mysqlServer == nil {
		panic("there is no mysql server")
	}
	var host = mysqlServer.Domain
	if host == "" {
		host = mysqlServer.Ip
	}
	timezone := "'+8:00'"
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		dataSourceConfig.Username,
		pwd,
		host,
		mysqlServer.Port,
		dataSourceConfig.DbName,
	) + "?charset=utf8mb4&parseTime=true&loc=Local&time_zone=" + url.QueryEscape(timezone)
}

func (dataSourceConfig *MySQLDataSourceConfig) getParamValue(key string, defaultValue int) int {
	v, ok := dataSourceConfig.Params[key]
	if ok {
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}
	return defaultValue
}

func (dataSourceConfig *MySQLDataSourceConfig) getMysqlServer(preferSlave bool) *MysqlServer {
	size := len(dataSourceConfig.MysqlServers)
	if size == 1 {
		return dataSourceConfig.MysqlServers[0]
	} else if size > 1 {
		var master *MysqlServer
		var slave *MysqlServer
		for _, ms := range dataSourceConfig.MysqlServers {
			if "master" == strings.ToLower(ms.Role) {
				master = ms
			} else if "slave" == strings.ToLower(ms.Role) {
				slave = ms
			}
		}

		if preferSlave && slave != nil {
			return slave
		}
		return master
	}
	return nil
}

func decrypt(encryptedPassword string) string {
	if encryptedPassword == "" {
		return ""
	}
	encryptedDecodeBytes, err := base64.StdEncoding.DecodeString(encryptedPassword)
	if err != nil {
		return ""
	}
	publicKey := gconf.GetGlobalConfigCollection().GetValue("publicKey").Raw()
	key, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return ""
	}
	pubKey, err := x509.ParsePKIXPublicKey(key)
	if err != nil {
		return ""
	}
	pub := pubKey.(*rsa.PublicKey)
	return string(rsaPublicDecrypt(pub, encryptedDecodeBytes))
}

func rsaPublicDecrypt(pubKey *rsa.PublicKey, data []byte) []byte {
	c := new(big.Int)
	m := new(big.Int)
	m.SetBytes(data)
	e := big.NewInt(int64(pubKey.E))
	c.Exp(m, e, pubKey.N)
	out := c.Bytes()
	skip := 0
	for i := 2; i < len(out); i++ {
		if i+1 >= len(out) {
			break
		}
		if out[i] == 0xff && out[i+1] == 0 {
			skip = i + 2
			break
		}
	}
	return out[skip:]
}

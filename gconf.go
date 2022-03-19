package gconf

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

var currentAppId string
var clientId string
var ds *dataStore

func Init(appName string) {
	appNameFromEnv := os.Getenv("APP_NAME")
	if appNameFromEnv != "" && appNameFromEnv != appName {
		panic(fmt.Sprintf("appName[%s]与环境变量中值[%s]不一致", appName, appNameFromEnv))
	}
	currentAppId = appName
	inK8s := len(os.Getenv("KUBERNETES_SERVICE_HOST")) > 0

	//在k8s里，使用HOSTNAME，VM里使用APP_INSTANCE_NAME

	getEnv := func(key, defaultValue string) string {
		value := os.Getenv(key)
		if value == "" {
			return defaultValue
		}
		return value
	}
	workEnv, workIdc := getEnv("WORK_ENV", "dev"), getEnv("WORK_IDC", "ofc")

	domainSuffix := ""
	if inK8s {
		// in k8s,gconf deploy to kube-system namespace
		domainSuffix = "kube-system"
	} else if workEnv == "prepare" && workIdc == "sh" {
		domainSuffix = "services.product.sh"
	} else {
		domainSuffix = "services." + workEnv + "." + workIdc
	}
	var appInstance string
	if inK8s {
		appInstance = getEnv("HOSTNAME", "unknown")
	} else {
		appInstance = getEnv("APP_INSTANCE_NAME", "unknown")
	}
	clientId = appName + "-->" + appInstance + "-->" + fmt.Sprint(rand.Int63n(time.Now().UnixNano()))

	ds = &dataStore{
		dataCache: map[string]*ConfigCollection{},
		client: &gConfHttpClient{
			baseUrl:  "http://gconf." + domainSuffix + "/api",
			clientId: clientId,
		},
		mux: sync.Mutex{},
	}
	ds.startBackgroundTask()
}

type dataStore struct {
	dataCache map[string]*ConfigCollection
	client    *gConfHttpClient
	mux       sync.Mutex
}

func (ds *dataStore) startBackgroundTask() {
	go func() {
		for {
			if len(ds.dataCache) == 0 {
				time.Sleep(time.Second * 2)
				continue
			}
			var appIdList []string
			for k := range ds.dataCache {
				appIdList = append(appIdList, k)
			}
			needChangeAppIdList := ds.client.watch(appIdList)

			for _, appId := range needChangeAppIdList {
				ds.dataCache[appId].refreshData(ds.client)
			}
		}
	}()
}

func (ds *dataStore) getConfigCollection(appId string) *ConfigCollection {
	res, ok := ds.dataCache[appId]
	if ok {
		return res
	}

	ds.mux.Lock()
	defer ds.mux.Unlock()

	//double check
	res, ok = ds.dataCache[appId]
	if ok {
		return res
	}

	configApp := ds.client.getConfigApp(appId)
	if configApp == nil {
		return nil
	}

	res = &ConfigCollection{
		appId:     appId,
		name:      configApp.Name,
		data:      map[string]*Value{},
		listeners: map[string][]ConfigChangeListener{},
	}
	res.refreshData(ds.client)
	ds.dataCache[appId] = res
	return res
}

// 获取当前应用的配置集合
func GetCurrentConfigCollection() *ConfigCollection {
	return GetConfigCollection(currentAppId)
}

// 获取全局的配置配置集合，此方法用于框架的统一配置。
// 应用不需要调用此方法
func GetGlobalConfigCollection() *ConfigCollection {
	return GetConfigCollection("golang")
}

// 获取某个appId的配置集合
func GetConfigCollection(appId string) *ConfigCollection {
	return ds.getConfigCollection(appId)
}

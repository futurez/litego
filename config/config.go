package config

import (
	"fmt"
)

const (
	IniProtocol = "ini"
)

type ConfigAdapter interface {
	ParseFile(filename string) (Configurer, error)
	ParseData(data []byte) (Configurer, error)
}

type Configurer interface {
	GetBool(key string, v ...bool) (bool, error)
	GetFloat(key string, v ...float64) (float64, error)
	GetInt(key string, v ...int) (int, error)
	GetInt64(key string, v ...int64) (int64, error)
	GetString(key string, v ...string) string
	GetStrings(key string, v ...string) []string
}

var adapters = make(map[string]ConfigAdapter)

func Register(name string, adapter ConfigAdapter) {
	if adapter == nil {
		panic("config.Register adapter is nil")
	}
	if _, ok := adapters[name]; ok {
		panic("config.Register called 2 adapter " + name)
	}
	adapters[name] = adapter
}

// adapterName is ini, filename is the config file path.
func NewConfig(adapterName, filename string) (Configurer, error) {
	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("config: unknown adaptername %q", adapterName)
	}
	return adapter.ParseFile(filename)
}

// adapterName is ini, data is the config data.
func NewConfigData(adapterName string, data []byte) (Configurer, error) {
	adapter, ok := adapters[adapterName]
	if !ok {
		return nil, fmt.Errorf("config: unknown adaptername %q", adapterName)
	}
	return adapter.ParseData(data)
}

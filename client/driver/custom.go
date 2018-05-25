package driver

import (
	"fmt"
	"io/ioutil"
	"os"
	"plugin"
	"strings"
)

const customDriversDir = "drivers"

func init() {
	if _, err := os.Stat(customDriversDir); os.IsNotExist(err) {
		return
	}

	files, err := ioutil.ReadDir()
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".so") {
			pluginName := strings.StripSuffix(file.Name(), ".so")

			plug, err := plugin.Open(file.Name())
			if err != nil {
				fmt.Println(err)
				continue
			}

			constructorLookup, err := plug.Lookup("NewDriver")
			if err != nil {
				fmt.Println(err)
				continue
			}

			var factory Factory
			factory, ok := constructorLookup.(Factory)
			if !ok {
				fmt.Println("unexpected type from module symbol ", factory)
				continue
			}

			BuiltinDrivers[pluginName] = factory
		}
	}
}

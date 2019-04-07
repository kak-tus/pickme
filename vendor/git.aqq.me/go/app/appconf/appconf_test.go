package appconf_test

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"git.aqq.me/go/app"
	"git.aqq.me/go/app/appconf"
	"github.com/iph0/conf"
)

type mapLoader struct {
	layers map[string]interface{}
}

func init() {
	mapLdr := NewLoader()
	appconf.RegisterLoader("test", mapLdr)

	appconf.Require(
		map[string]interface{}{
			"paramA": "default:valA",
			"paramZ": "default:valZ",
		},

		"test:foo",
		"test:bar",
	)
}

func TestBasic(t *testing.T) {
	err := app.Init()

	if err != nil {
		t.Error(err)
		return
	}

	tConfig := appconf.GetConfig()

	eConfig := map[string]interface{}{
		"paramA": "foo:valA",
		"paramB": "bar:valB",
		"paramC": "bar:valC",
		"paramZ": "default:valZ",
	}

	if !reflect.DeepEqual(tConfig, eConfig) {
		t.Errorf("unexpected configuration returned: %#v", tConfig)
	}

	err = app.Reload()

	if err != nil {
		t.Error(err)
		return
	}

	err = app.Stop()

	if err != nil {
		t.Error(err)
	}
}

func TestPanic(t *testing.T) {
	t.Run("invalid_locator",
		func(t *testing.T) {
			defer func() {
				err := recover()
				errStr := fmt.Sprintf("%v", err)

				if err == nil {
					t.Error("no error happened")
				} else if strings.Index(errStr, "locator has invalid type") == -1 {
					t.Error("other error happened:", err)
				}
			}()

			appconf.Require(42)
		},
	)

	t.Run("not_initialized",
		func(t *testing.T) {
			defer func() {
				err := recover()
				errStr := fmt.Sprintf("%v", err)

				if err == nil {
					t.Error("no error happened")
				} else if strings.Index(errStr, "must be initialized first") == -1 {
					t.Error("other error happened")
				}
			}()

			appconf.GetConfig()
		},
	)
}

func NewLoader() conf.Loader {
	return &mapLoader{
		map[string]interface{}{
			"foo": map[string]interface{}{
				"paramA": "foo:valA",
				"paramB": "foo:valB",
			},

			"bar": map[string]interface{}{
				"paramB": "bar:valB",
				"paramC": "bar:valC",
			},
		},
	}
}

func (p *mapLoader) Load(loc *conf.Locator) (interface{}, error) {
	key := loc.BareLocator
	layer, _ := p.layers[key]

	return layer, nil
}

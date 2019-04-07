package appmetrics_test

import (
	"testing"

	"git.aqq.me/go/app"
	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/appmetrics"
)

var config = map[string]interface{}{
	"metrics": map[string]interface{}{
		"host":         "localhost",
		"port":         8125,
		"metricPrefix": "test.",
		"tagFormat":    "inname",
		"metricFilter": map[string]interface{}{
			"foo": true,
			"bar": true,
			"moo": true,
			"jar": true,
			"zoo": false,
		},
	},
}

func TestBasic(t *testing.T) {
	appconf.Require(config)
	err := app.Init()

	if err != nil {
		t.Error(err)
		return
	}

	client := appmetrics.GetClient()

	t.Run("incr",
		func(t *testing.T) {
			client.Incr("foo",
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("decr",
		func(t *testing.T) {
			client.Decr("foo",
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("incr_by",
		func(t *testing.T) {
			client.IncrBy("foo", 3,
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("decr_by",
		func(t *testing.T) {
			client.DecrBy("foo", 3,
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("timing",
		func(t *testing.T) {
			client.Timing("bar", 20,
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("precision_timing",
		func(t *testing.T) {
			client.PrecisionTiming("bar", 20,
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("gauge",
		func(t *testing.T) {
			client.Gauge("moo", 300,
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("gauge_float",
		func(t *testing.T) {
			client.GaugeFloat("moo", 300.15,
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("gauge_delta",
		func(t *testing.T) {
			client.GaugeDelta("moo", 10,
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("gauge_delta_float",
		func(t *testing.T) {
			client.GaugeDeltaFloat("moo", 10.15,
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("set",
		func(t *testing.T) {
			client.Set("jar", "uniqueValue",
				"host", "test.com",
				"dc", "dtln",
			)
		},
	)

	t.Run("disabled",
		func(t *testing.T) {
			client.Incr("zoo")
		},
	)

	err = app.Stop()

	if err != nil {
		t.Error(err)
	}
}

func BenchmarkBasic(b *testing.B) {
	appconf.Require(config)
	err := app.Init()

	if err != nil {
		b.Error(err)
		return
	}

	client := appmetrics.GetClient()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		client.Incr("foo",
			"host", "test.com",
			"dc", "dtln",
		)
	}
}

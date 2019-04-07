package appmetrics

import (
	"fmt"
	"strings"
	"time"

	"git.aqq.me/go/app/appconf"
	"github.com/smira/go-statsd"
)

const (
	configBranchName = "metrics"

	tagFormatInName   = "inname"
	tagFormatInfluxDB = "influxdb"
	tagFormatDatadog  = "datadog"

	tagSep = "."
)

var defaultConfig = map[string]interface{}{
	"metrics": map[string]interface{}{
		"host":              "localhost",
		"port":              8125,
		"tagFormat":         tagFormatInName,
		"maxPacketSize":     1432,
		"flushInterval":     300,
		"reconnectInterval": 0,
		"retryTimeout":      5000,
		"bufPoolCapacity":   20,
		"sendQueueCapacity": 10,
		"sendLoopCount":     1,
	},
}

// Client sends metric values to the StatsD daemon.
type Client struct {
	config      clientConfig
	statsd      *statsd.Client
	tagReplacer *strings.Replacer
}

type clientConfig struct {
	Host              string
	Port              int
	MetricPrefix      string
	TagFormat         string
	MaxPacketSize     int
	FlushInterval     time.Duration
	ReconnectInterval time.Duration
	RetryTimeout      time.Duration
	BufPoolCapacity   int
	SendQueueCapacity int
	SendLoopCount     int
	MetricAliases     map[string]string
	MetricFilter      map[string]bool
	Disable           bool
}

func init() {
	appconf.Require(defaultConfig)
}

// NewClient method creates new StatsD client instance.
func NewClient() (*Client, error) {
	var config clientConfig
	configRaw := appconf.GetConfig()
	err := appconf.Decode(configRaw[configBranchName], &config)

	if err != nil {
		return nil, fmt.Errorf("%s: invalid configuration: %s", errPref, err)
	}

	config.FlushInterval *= time.Millisecond
	config.ReconnectInterval *= time.Millisecond
	config.RetryTimeout *= time.Millisecond

	var tagFormat *statsd.TagFormat

	if config.TagFormat == tagFormatInName ||
		config.TagFormat == tagFormatInfluxDB {
		tagFormat = statsd.TagFormatInfluxDB
	} else if config.TagFormat == tagFormatDatadog {
		tagFormat = statsd.TagFormatDatadog
	} else {
		return nil, fmt.Errorf("%s: unknown tag format: %s", errPref,
			config.TagFormat)
	}

	if config.Disable {
		return &Client{
			config: config,
		}, nil
	}

	hostport := fmt.Sprintf("%s:%d", config.Host, config.Port)

	statsdCli := statsd.NewClient(hostport,
		statsd.MetricPrefix(config.MetricPrefix),
		statsd.TagStyle(tagFormat),
		statsd.MaxPacketSize(config.MaxPacketSize),
		statsd.FlushInterval(config.FlushInterval),
		statsd.ReconnectInterval(config.ReconnectInterval),
		statsd.RetryTimeout(config.RetryTimeout),
		statsd.BufPoolCapacity(config.BufPoolCapacity),
		statsd.SendQueueCapacity(config.SendQueueCapacity),
		statsd.SendLoopCount(config.SendLoopCount),
		statsd.ReportInterval(0),
	)

	return &Client{
		config:      config,
		statsd:      statsdCli,
		tagReplacer: strings.NewReplacer(tagSep, "_"),
	}, nil
}

// Incr method increments metric value by 1.
func (c *Client) Incr(name string, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.Incr(name, 1, ptags...)
}

// Decr method decrements metric value by 1.
func (c *Client) Decr(name string, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.Decr(name, 1, ptags...)
}

// IncrBy method increments metric value by n.
func (c *Client) IncrBy(name string, n int64, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.Incr(name, n, ptags...)
}

// DecrBy method decrements metric value by n.
func (c *Client) DecrBy(name string, n int64, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.Decr(name, n, ptags...)
}

// Timing method sends value to the timing metric.
func (c *Client) Timing(name string, delta int64, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.Timing(name, delta, ptags...)
}

// PrecisionTiming method sends value to the timing metric, where value has type
// time.Duration.
func (c *Client) PrecisionTiming(name string, delta time.Duration, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.PrecisionTiming(name, delta, ptags...)
}

// Gauge method sends value to the gauge metric.
func (c *Client) Gauge(name string, value int64, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.Gauge(name, value, ptags...)
}

// GaugeFloat method sends value to the gauge metric, where value is a float number.
func (c *Client) GaugeFloat(name string, value float64, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.FGauge(name, value, ptags...)
}

// GaugeDelta method sends delta to the gauge metric.
func (c *Client) GaugeDelta(name string, delta int64, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.GaugeDelta(name, delta, ptags...)
}

// GaugeDeltaFloat method sends delta to the gauge metric, where delta is a float number.
func (c *Client) GaugeDeltaFloat(name string, delta float64, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.FGaugeDelta(name, delta, ptags...)
}

// Set method sends value to the set metric.
func (c *Client) Set(name string, value string, tags ...string) {
	if !c.isEnabled(name) {
		return
	}

	name, ptags := c.prepare(name, tags)
	c.statsd.SetAdd(name, value, ptags...)
}

// Close method performs correct closure of the client.
func (c *Client) Close() {
	if c.statsd != nil {
		c.statsd.Close()
	}
}

func (c *Client) isEnabled(name string) bool {
	if c.config.Disable {
		return false
	}

	if len(c.config.MetricFilter) > 0 {
		enabled, ok := c.config.MetricFilter[name]

		if !ok || !enabled {
			return false
		}
	}

	return true
}

func (c *Client) prepare(name string, tags []string) (string, []statsd.Tag) {
	if realName, ok := c.config.MetricAliases[name]; ok {
		name = realName
	}

	var ptags []statsd.Tag

	if c.config.TagFormat == tagFormatInName {
		name = c.appendTags(name, tags)
	} else {
		ptags = prepareTags(tags)
	}

	return name, ptags
}

func (c *Client) appendTags(name string, tags []string) string {
	tagsLen := len(tags)

	if tagsLen == 0 {
		return name
	} else if tagsLen%2 > 0 {
		tags = tags[:tagsLen-1]
	}

	for i := 1; i < tagsLen; i += 2 {
		if strings.Index(tags[i], tagSep) >= 0 {
			tags[i] = c.tagReplacer.Replace(tags[i])
		}

		name += tagSep + tags[i]
	}

	return name
}

func prepareTags(tags []string) []statsd.Tag {
	tagsLen := len(tags)
	ptags := make([]statsd.Tag, 0, tagsLen)

	for i := 1; i < tagsLen; i += 2 {
		ptags = append(ptags,
			statsd.StringTag(tags[i-1], tags[i]))
	}

	return ptags
}

package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConf(t *testing.T) {
	os.Setenv("PICKME_REDISADDRS", "1")
	os.Setenv("PICKME_TELEGRAM_PATH", "2")

	cnf, err := newConf()
	require.NoError(t, err)
	require.NotNil(t, cnf)
	require.Equal(t, "1", cnf.RedisAddrs)
	require.Equal(t, "2", cnf.Telegram.Path)
}

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

const key = "pickme{%s}"

type stored struct {
	Done    []int
	doneIdx map[int]bool
	Items   []string
	Names   []string
	uniq    string
}

func (o *instanceObj) store(ctx context.Context, value stored) error {
	enc, err := o.enc.MarshalToString(value)
	if err != nil {
		return err
	}

	k := fmt.Sprintf(key, value.uniq)

	_, err = o.rdb.Set(ctx, k, enc, time.Hour*24*7).Result()
	if err != nil && err != redis.Nil {
		return err
	}

	return nil
}

func (o *instanceObj) get(ctx context.Context, id string) (*stored, error) {
	k := fmt.Sprintf(key, id)

	res, err := o.rdb.Get(ctx, k).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	if res == "" {
		return nil, nil
	}

	st := stored{}

	err = o.enc.UnmarshalFromString(res, &st)
	if err != nil {
		return nil, err
	}

	st.uniq = id
	st.doneIdx = make(map[int]bool)

	for _, v := range st.Done {
		st.doneIdx[v] = true
	}

	return &st, nil
}

package main

import (
	"context"
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func (o *instanceObj) formAndStoreKB(ctx context.Context, st stored) (*tgbotapi.InlineKeyboardMarkup, error) {
	uniq := st.uniq

	if uniq == "" {
		uniq = o.gen.Generate()

		st.uniq = uniq
	}

	if len(st.Items)-len(st.Done) == 0 {
		return nil, nil
	}

	btns := make([][]tgbotapi.InlineKeyboardButton, len(st.Items)-len(st.Done))
	pos := 0

	for i, v := range st.Items {
		if st.doneIdx[i] {
			continue
		}

		id := fmt.Sprintf("%s_%d", uniq, i)

		btns[pos] = []tgbotapi.InlineKeyboardButton{
			{
				CallbackData: &id,
				Text:         v,
			},
		}
		pos++
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(btns...)

	err := o.store(ctx, st)
	if err != nil {
		return nil, err
	}

	return &kb, nil
}

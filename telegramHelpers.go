package main

import (
	"fmt"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/zach-klippenstein/goregen"
)

func (o *instanceObj) formAndStoreKB(st stored) (*tgbotapi.InlineKeyboardMarkup, error) {
	uniq := st.uniq

	if uniq == "" {
		var err error

		// 64 bytes limit, so 58 + 1 (_) + 2 (0..99) + reserve
		uniq, err = regen.Generate("[A-Za-z0-9]{58}")
		if err != nil {
			return nil, err
		}

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
			tgbotapi.InlineKeyboardButton{
				CallbackData: &id,
				Text:         v,
			},
		}
		pos++
	}

	kb := tgbotapi.NewInlineKeyboardMarkup(btns...)

	err := o.store(st)
	if err != nil {
		return nil, err
	}

	return &kb, nil
}

package main

import (
	"fmt"
	"strings"
)

func parseMsg(msg string) (string, string, []string) {
	items := make([]string, 0)

	rows := strings.Split(msg, "\n")

	for _, r := range rows {
		cols := strings.Split(r, ",")

		for _, c := range cols {
			v := strings.Trim(c, " ")

			if v == "" {
				continue
			}

			items = append(items, v)
		}
	}

	title := "Shopping list or TODO list"
	if len(items) != 0 {
		title = items[0]
		items = items[1:]
	}

	if len(items) == 0 {
		return "Send", title, items
	}

	send := fmt.Sprintf("Send (%s)", strings.Join(items, ","))

	return send, title, items
}

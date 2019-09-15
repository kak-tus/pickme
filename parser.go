package main

import (
	"fmt"
	"strings"
)

func parseMsg(msg string) (string, []string) {
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

	if len(items) == 0 {
		return "Send", items
	}

	send := fmt.Sprintf("Send (%s)", strings.Join(items, ","))

	return send, items
}

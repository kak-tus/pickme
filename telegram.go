package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"git.aqq.me/go/app/appconf"
	"git.aqq.me/go/app/applog"
	"git.aqq.me/go/app/event"
	"github.com/go-redis/redis"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/iph0/conf"
	jsoniter "github.com/json-iterator/go"
	"golang.org/x/net/proxy"
)

var inst *instanceObj

// Count of buttons is limited by telegram API or client
const maxButtons = 100

func init() {
	event.Init.AddHandler(
		func() error {
			cnfMap := appconf.GetConfig()["pickme"]

			var cnf instanceConf
			err := conf.Decode(cnfMap, &cnf)
			if err != nil {
				return err
			}

			addrs := strings.Split(cnf.RedisAddrs, ",")

			ropt := &redis.ClusterOptions{
				Addrs:        addrs,
				ReadTimeout:  time.Minute,
				WriteTimeout: time.Minute,
			}

			rdb := redis.NewClusterClient(ropt)

			log := applog.GetLogger().Sugar()

			httpTransport := &http.Transport{}

			if cnf.Telegram.Proxy != "" {
				dialer, err := proxy.SOCKS5("tcp", cnf.Telegram.Proxy, nil, proxy.Direct)
				if err != nil {
					return err
				}

				httpTransport.Dial = dialer.Dial
			}

			httpClient := &http.Client{Transport: httpTransport, Timeout: time.Minute}

			bot, err := tgbotapi.NewBotAPIWithClient(cnf.Telegram.Token, httpClient)
			if err != nil {
				return err
			}

			srv := &http.Server{Addr: cnf.Listen}

			inst = &instanceObj{
				bot: bot,
				cnf: cnf,
				enc: jsoniter.Config{UseNumber: true}.Froze(),
				log: log,
				rdb: rdb,
				srv: srv,
			}

			return nil
		},
	)

	event.Stop.AddHandler(
		func() error {
			err := inst.srv.Shutdown(nil)
			if err != nil {
				return err
			}

			return nil
		},
	)
}

func (o *instanceObj) Start() error {
	res, err := o.bot.SetWebhook(tgbotapi.NewWebhook(o.cnf.Telegram.URL + o.cnf.Telegram.Path))
	if err != nil {
		return err
	}

	o.log.Debug(res.Description)

	updates := o.bot.ListenForWebhook("/" + o.cnf.Telegram.Path)

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	go func() {
		err := o.srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			o.log.Panic(err)
		}
	}()

	go func() {
		for {
			msg := <-updates

			go func() {
				err := o.process(msg)
				if err != nil {
					o.log.Error(err)
				}
			}()
		}
	}()

	return nil
}

func (o *instanceObj) process(msg tgbotapi.Update) error {
	if msg.Message != nil {
		return o.processMessage(msg.Message)
	} else if msg.InlineQuery != nil {
		return o.processInline(msg.InlineQuery)
	} else if msg.CallbackQuery != nil {
		return o.processCallback(msg.CallbackQuery)
	}

	return nil
}

func (o *instanceObj) processMessage(msg *tgbotapi.Message) error {
	if msg.Command() == "start" {
		repl := tgbotapi.NewMessage(msg.Chat.ID, "Use this bot in inline mode: @pickmebot in any chat, even in this or send me message directly.")

		_, err := o.bot.Send(repl)
		if err != nil {
			return err
		}
	}

	_, items := parseMsg(msg.Text)

	// Hacky detection of message from self thru user
	// If no items found
	if len(items) == 0 {
		return nil
	}

	repl := tgbotapi.NewMessage(msg.Chat.ID, "pickme")

	// SliceTricks
	batchSize := maxButtons
	var batches [][]string

	for batchSize < len(items) {
		items, batches = items[batchSize:], append(batches, items[0:batchSize:batchSize])
	}
	batches = append(batches, items)

	for _, b := range batches {
		kb, err := o.formAndStoreKB(stored{Items: b})
		if err != nil {
			return err
		}

		repl.ReplyMarkup = kb

		_, err = o.bot.Send(repl)
		if err != nil {
			return err
		}
	}

	return nil
}

func (o *instanceObj) processInline(msg *tgbotapi.InlineQuery) error {
	subj, items := parseMsg(msg.Query)

	// Telegram API limits
	if len(msg.Query) >= 255 || len(items) > maxButtons {
		subj = "Limit reached. Send me direct message."
		items = []string{}
	}

	repl := tgbotapi.NewInlineQueryResultArticle("0", subj, subj)

	if len(items) != 0 {
		kb, err := o.formAndStoreKB(stored{Items: items})
		if err != nil {
			return err
		}

		repl.ReplyMarkup = kb
	}

	_, err := o.bot.AnswerInlineQuery(tgbotapi.InlineConfig{
		InlineQueryID: msg.ID,
		IsPersonal:    true,
		Results:       []interface{}{repl},
	})
	if err != nil {
		return err
	}

	return nil
}

func (o *instanceObj) processCallback(msg *tgbotapi.CallbackQuery) error {
	if msg.Data == "" {
		return nil
	}

	keys := strings.Split(msg.Data, "_")
	if len(keys) == 0 {
		return nil
	}

	uniq := keys[0]

	st, err := o.get(uniq)
	if err != nil {
		return err
	}

	if st == nil && msg.Message != nil {
		// Only with direct message
		repl := tgbotapi.NewMessage(msg.Message.Chat.ID, "List already expired")

		_, err := o.bot.Send(repl)
		if err != nil {
			return err
		}

		return nil
	}

	done := make([]string, 0)

	for i, v := range st.Done {
		done = append(done, fmt.Sprintf("%s (%s)", st.Items[v], st.Names[i]))
	}

	for i, v := range st.Items {
		if st.doneIdx[i] {
			continue
		}

		id := fmt.Sprintf("%s_%d", uniq, i)

		if msg.Data != id {
			continue
		}

		name := msg.From.FirstName
		if msg.From.LastName != "" {
			name += " " + msg.From.LastName
		}

		done = append(done, fmt.Sprintf("%s (%s)", v, name))

		st.Done = append(st.Done, i)
		st.Names = append(st.Names, name)
		st.doneIdx[i] = true
	}

	txt := strings.Join(done, "\n")

	repl := tgbotapi.EditMessageTextConfig{Text: txt}

	if msg.Message != nil {
		// Direct message
		repl.ChatID = msg.Message.Chat.ID
		repl.MessageID = msg.Message.MessageID
	} else {
		// Inline message
		repl.InlineMessageID = msg.InlineMessageID
	}

	kb, err := o.formAndStoreKB(*st)
	if err != nil {
		return err
	}

	if kb != nil {
		repl.ReplyMarkup = kb
	}

	_, err = o.bot.Send(repl)
	if err != nil {
		return err
	}

	return nil
}

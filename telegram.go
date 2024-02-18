package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	regen "github.com/zach-klippenstein/goregen"
	"go.uber.org/zap"
	"golang.org/x/net/proxy"
)

// Count of buttons is limited by telegram API or client
// Correct limit is unknown
// May be it limited by message size, not by buttons count?
const maxButtons = 50

func newTg(log zerolog.Logger) (*instanceObj, error) {
	cnf, err := newConf()
	if err != nil {
		return nil, err
	}

	addrs := strings.Split(cnf.RedisAddrs, ",")

	ropt := &redis.ClusterOptions{
		Addrs:        addrs,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}

	rdb := redis.NewClusterClient(ropt)

	loggerOld, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}

	logOld := loggerOld.Sugar()

	httpClient, err := getClient(cnf)
	if err != nil {
		return nil, err
	}

	bot, err := tgbotapi.NewBotAPIWithClient(cnf.Telegram.Token, tgbotapi.APIEndpoint, httpClient)
	if err != nil {
		return nil, err
	}

	srv := &http.Server{Addr: cnf.Listen}

	enc := jsoniter.Config{UseNumber: true}.Froze()

	arg := &regen.GeneratorArgs{
		RngSource: rand.NewSource(time.Now().UnixNano()),
	}

	// 64 bytes limit, so 58 + 1 (_) + 2 (0..99) + reserve
	gen, err := regen.NewGenerator("[A-Za-z0-9]{58}", arg)
	if err != nil {
		return nil, err
	}

	inst := &instanceObj{
		bot:    bot,
		cnf:    cnf,
		enc:    enc,
		gen:    gen,
		log:    log,
		logOld: logOld,
		rdb:    rdb,
		srv:    srv,
	}

	return inst, nil
}

func getClient(cnf *instanceConf) (*http.Client, error) {
	httpTransport := &http.Transport{}

	if cnf.Telegram.Proxy != "" {
		dialer, err := proxy.SOCKS5("tcp", cnf.Telegram.Proxy, nil, proxy.Direct)
		if err != nil {
			return nil, err
		}

		httpTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			done := make(chan bool)

			var (
				con net.Conn
				err error
			)

			go func() {
				con, err = dialer.Dial(network, addr)
				done <- true
			}()

			select {
			case <-ctx.Done():
				return nil, errors.New("dial timeout")
			case <-done:
				return con, err
			}
		}
	}

	httpClient := &http.Client{Transport: httpTransport, Timeout: time.Minute}

	return httpClient, nil
}

func (o *instanceObj) start() error {
	webhookCnf, err := tgbotapi.NewWebhook(o.cnf.Telegram.URL + o.cnf.Telegram.Path)
	if err != nil {
		return err
	}

	resp, err := o.bot.Request(webhookCnf)
	if err != nil {
		return err
	}

	o.logOld.Info(resp.Description)

	updates := o.bot.ListenForWebhook("/" + o.cnf.Telegram.Path)

	http.HandleFunc("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, "ok")
	})

	go func() {
		for {
			msg := <-updates

			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)

				if err := o.process(ctx, msg); err != nil {
					o.log.Error().Err(err).Msg("process fail")
				}

				cancel()
			}()
		}
	}()

	err = o.srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (o *instanceObj) stop() error {
	_ = o.logOld.Sync()

	if err := o.srv.Shutdown(context.TODO()); err != nil {
		return err
	}

	return nil
}

func (o *instanceObj) process(ctx context.Context, msg tgbotapi.Update) error {
	switch {
	case msg.Message != nil:
		if err := o.processMessage(ctx, msg.Message); err != nil {
			return errors.Wrap(err, "message process fail")
		}
	case msg.InlineQuery != nil:
		if err := o.processInline(ctx, msg.InlineQuery); err != nil {
			return errors.Wrap(err, "inline process fail")
		}
	case msg.CallbackQuery != nil:
		if err := o.processCallback(ctx, msg.CallbackQuery); err != nil {
			return errors.Wrap(err, "callback process fail")
		}
	}

	return nil
}

func (o *instanceObj) processMessage(ctx context.Context, msg *tgbotapi.Message) error {
	if msg.Command() == "start" {
		repl := tgbotapi.NewMessage(
			msg.Chat.ID,
			"Use this bot in inline mode: @pickmebot in any chat, even in this or send me message directly.",
		)

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
		kb, err := o.formAndStoreKB(ctx, stored{Items: b})
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

func (o *instanceObj) processInline(ctx context.Context, msg *tgbotapi.InlineQuery) error {
	subj, items := parseMsg(msg.Query)

	// Telegram API limits
	if len(msg.Query) >= 255 || len(items) > maxButtons {
		subj = "Limit reached. Send me direct message."
		items = []string{}
	}

	repl := tgbotapi.NewInlineQueryResultArticle("0", subj, subj)

	if len(items) != 0 {
		kb, err := o.formAndStoreKB(ctx, stored{Items: items})
		if err != nil {
			return err
		}

		repl.ReplyMarkup = kb
	}

	_, err := o.bot.Send(tgbotapi.InlineConfig{
		InlineQueryID: msg.ID,
		IsPersonal:    true,
		Results:       []any{repl},
	})
	if err != nil {
		return err
	}

	return nil
}

func (o *instanceObj) processCallback(ctx context.Context, msg *tgbotapi.CallbackQuery) error {
	if msg.Data == "" {
		return nil
	}

	keys := strings.Split(msg.Data, "_")
	if len(keys) == 0 {
		return nil
	}

	uniq := keys[0]

	st, err := o.get(ctx, uniq)
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

	kb, err := o.formAndStoreKB(ctx, *st)
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

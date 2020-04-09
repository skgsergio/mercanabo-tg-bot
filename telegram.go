// Copyright (c) 2020 Sergio Conde skgsergio@gmail.com
//
// This program is free software: you can redistribute it and/or modify it under
// the terms of the GNU General Public License as published by the Free Software
// Foundation, version 3.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY
// WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
// PARTICULAR PURPOSE. See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with
// this program. If not, see <https://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: GPL-3.0-only

package main

import (
	"fmt"
	"time"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/rs/zerolog/log"
)

// Telegram represents the telegram bot
type Telegram struct {
	bot                *tb.Bot
	handlersRegistered bool
}

// NewBot returns a Telegram bot
func NewBot(token string) (*Telegram, error) {
	bot, err := tb.NewBot(tb.Settings{
		Token:  token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
		Reporter: func(err error) {
			log.Error().Str("module", "telegram").Err(err).Msg("telebot internal error")
		},
	})

	if err != nil {
		return nil, err
	}

	log.Info().Str("module", "telegram").Int("id", bot.Me.ID).Str("name", bot.Me.FirstName).Str("username", bot.Me.Username).Msg("connected to telegram")

	return &Telegram{bot: bot}, nil
}

// Start starts polling for telegram updates
func (t *Telegram) Start() {
	t.registerHandlers()

	log.Info().Str("module", "telegram").Msg("start polling")
	t.bot.Start()
}

// RegisterHandlers registers all the handlers
func (t *Telegram) registerHandlers() {
	if t.handlersRegistered {
		return
	}

	log.Info().Str("module", "telegram").Msg("registering handlers")

	t.bot.Handle("/start", t.handleStart)
	t.bot.Handle(tb.OnAddedToGroup, t.handleAddedToGroup)
	t.bot.Handle(fmt.Sprintf("/%s", texts.Help.Cmd), t.handleHelpCmd)
	t.bot.Handle(fmt.Sprintf("/%s", texts.Buy.Cmd), t.handleBuyCmd)
	t.bot.Handle(fmt.Sprintf("/%s", texts.Sell.Cmd), t.handleSellCmd)
	t.bot.Handle(fmt.Sprintf("/%s", texts.List.Cmd), t.handleListCmd)

	t.handlersRegistered = true
}

// send sends a message with error logging and retries
func (t *Telegram) send(to tb.Recipient, what interface{}, options ...interface{}) *tb.Message {
	var (
		msg *tb.Message = nil
		err error       = nil
		try int         = 1
	)

	hasParseMode := false
	for _, opt := range options {
		if _, hasParseMode = opt.(tb.ParseMode); hasParseMode {
			break
		}
	}

	if !hasParseMode {
		options = append(options, tb.ModeHTML)
	}

	for {
		msg, err = t.bot.Send(to, what, options...)

		if err == nil {
			break
		}

		if try > 5 {
			log.Error().Err(err).Msg("send aborted, retry limit exceeded")
			break
		}

		backoff := time.Second * 5 * time.Duration(try)
		log.Warn().Err(err).Str("sleep", backoff.String()).Msg("send failed, sleeping and retrying")
		time.Sleep(backoff)
		try++
	}

	return msg
}

// reply replies a message with error logging and retries
func (t *Telegram) reply(to *tb.Message, what interface{}, options ...interface{}) *tb.Message {
	var (
		msg *tb.Message = nil
		err error       = nil
		try int         = 1
	)

	hasParseMode := false
	for _, opt := range options {
		if _, hasParseMode = opt.(tb.ParseMode); hasParseMode {
			break
		}
	}

	if !hasParseMode {
		options = append(options, tb.ModeHTML)
	}

	for {
		_, err = t.bot.Reply(to, what, options...)

		if err == nil {
			break
		}

		if try > 5 {
			log.Error().Err(err).Msg("reply aborted, retry limit exceeded")
			break
		}

		backoff := time.Second * 5 * time.Duration(try)
		log.Warn().Err(err).Str("sleep", backoff.String()).Msg("reply failed, sleeping and retrying")
		time.Sleep(backoff)
		try++
	}

	return msg
}

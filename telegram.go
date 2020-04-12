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
	t.bot.Handle(fmt.Sprintf("/%s", texts.Admin.Cmd), t.handleAdminCmd)
	t.bot.Handle(fmt.Sprintf("/%s", texts.Buy.Cmd), t.handleBuyCmd)
	t.bot.Handle(fmt.Sprintf("/%s", texts.IslandPrice.Cmd), t.handleIslandPriceCmd)
	t.bot.Handle(fmt.Sprintf("/%s", texts.Sell.Cmd), t.handleSellCmd)
	t.bot.Handle(fmt.Sprintf("/%s", texts.List.Cmd), t.handleListCmd)
	t.bot.Handle(fmt.Sprintf("/%s", texts.Chart.Cmd), t.handleChartCmd)
	t.bot.Handle(fmt.Sprintf("/%s", texts.Delete.Cmd), t.handleDeleteCmd)

	t.handlersRegistered = true
}

// send sends a message with error logging and retries
func (t *Telegram) send(to tb.Recipient, what interface{}, options ...interface{}) *tb.Message {
	hasParseMode := false
	for _, opt := range options {
		if _, hasParseMode = opt.(tb.ParseMode); hasParseMode {
			break
		}
	}

	if !hasParseMode {
		options = append(options, tb.ModeHTML)
	}

	try := 1
	for {
		msg, err := t.bot.Send(to, what, options...)

		if err == nil {
			return msg
		}

		if try > 5 {
			log.Error().Str("module", "telegram").Err(err).Msg("send aborted, retry limit exceeded")
			return nil
		}

		backoff := time.Second * 5 * time.Duration(try)
		log.Warn().Str("module", "telegram").Err(err).Str("sleep", backoff.String()).Msg("send failed, sleeping and retrying")
		time.Sleep(backoff)
		try++
	}
}

// reply replies a message with error logging and retries
func (t *Telegram) reply(to *tb.Message, what interface{}, options ...interface{}) *tb.Message {
	hasParseMode := false
	for _, opt := range options {
		if _, hasParseMode = opt.(tb.ParseMode); hasParseMode {
			break
		}
	}

	if !hasParseMode {
		options = append(options, tb.ModeHTML)
	}

	try := 1
	for {
		msg, err := t.bot.Reply(to, what, options...)

		if err == nil {
			return msg
		}

		if try > 5 {
			log.Error().Str("module", "telegram").Err(err).Msg("reply aborted, retry limit exceeded")
			return nil
		}

		backoff := time.Second * 5 * time.Duration(try)
		log.Warn().Str("module", "telegram").Err(err).Str("sleep", backoff.String()).Msg("reply failed, sleeping and retrying")
		time.Sleep(backoff)
		try++
	}
}

func (t *Telegram) cleanupChatMsgs(chat *tb.Chat, msgs []*tb.Message) {
	var err error = nil

	// Check if the group requires message deletion
	group, err := db.GetGroup(chat)
	if err != nil {
		log.Error().Str("module", "telegram").Err(err).Msg("failed getting group delete seconds")
		return
	}

	if group.DeleteSeconds == 0 {
		return
	}

	// Grab bot permissions over the chat
	cm, err := t.bot.ChatMemberOf(chat, t.bot.Me)

	if err != nil {
		log.Error().Str("module", "telegram").Err(err).Msg("failed getting bot membership in chat")
		return
	}

	// Sleep
	time.Sleep(time.Duration(group.DeleteSeconds) * time.Second)

	for _, m := range msgs {
		if m == nil {
			log.Error().Str("module", "telegram").Msg("message to delete is nil")
			continue
		}

		// Chech the message belongs to the chat
		if m.Chat.ID != chat.ID {
			log.Error().Str("module", "telegram").Int64("chat_id", chat.ID).Int64("m_chat_id", m.Chat.ID).Msg("message to delete doesn't belong the chat")
			continue
		}

		// If the message is from the bot we can just delete it
		if m.Sender.ID == t.bot.Me.ID || cm.CanDeleteMessages {
			err = t.bot.Delete(m)
			if err != nil {
				log.Error().Str("module", "telegram").Err(err).Msg("failed deleting message")
			}
		}
	}
}

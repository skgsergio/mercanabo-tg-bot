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
	"strings"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/rs/zerolog/log"
)

// handleStart triggers when /start is sent on private
func (t *Telegram) handleStart(m *tb.Message) {
	if !m.Private() {
		return
	}

	t.send(m.Chat, texts.GroupOnly)
}

// handleAddedToGroup triggers when the bot is added to a group
func (t *Telegram) handleAddedToGroup(m *tb.Message) {
	log.Info().Str("module", "telegram").Int64("chat_id", m.Chat.ID).Str("chat_title", m.Chat.Title).Msg("added to group")

	t.send(m.Chat, texts.JoinText)

	// Register the group in the DB
	_, err := db.GetGroup(m.Chat)
	if err != nil {
		log.Error().Str("module", "telegram").Err(err).Msg("error getting or creating group")
	}
}

// handleHelpCmd triggers when the help cmd is sent to a group
func (t *Telegram) handleHelpCmd(m *tb.Message) {
	if m.Private() {
		t.send(m.Chat, texts.GroupOnly)
		return
	}

	log.Info().
		Str("module", "telegram").
		Int64("chat_id", m.Chat.ID).Str("chat_title", m.Chat.Title).
		Int("user_id", m.Sender.ID).Str("user_first_name", m.Sender.FirstName).
		Str("user_last_name", m.Sender.LastName).Str("user_username", m.Sender.Username).
		Msg(m.Text)

	helpLines := []string{
		"Estos son los comandos disponibles:",
		fmt.Sprintf("\n<code>/%s</code>\n%s", texts.Help.Cmd, texts.Help.Desc),
		fmt.Sprintf("\n<code>/%s</code>\n%s", texts.List.Cmd, texts.List.Desc),
		fmt.Sprintf("\n<code>/%s %s</code>\n%s", texts.Buy.Cmd, texts.Buy.Params, texts.Buy.Desc),
		fmt.Sprintf("\n<code>/%s %s</code>\n%s", texts.Sell.Cmd, texts.Sell.Params, texts.Sell.Desc),
		fmt.Sprintf("\n<code>/%s %s</code>\n%s", texts.ChangeTZ.Cmd, texts.ChangeTZ.Params, texts.ChangeTZ.Desc),
	}

	t.reply(m, strings.Join(helpLines, "\n"), &tb.SendOptions{
		DisableWebPagePreview: true,
		ParseMode:             tb.ModeHTML,
	})
}

// handleBuyCmd triggers when the buy cmd is sent to a group, if sent in private the user will be warned
func (t *Telegram) handleBuyCmd(m *tb.Message) {
	if m.Private() {
		t.send(m.Chat, texts.GroupOnly)
		return
	}

	log.Info().
		Str("module", "telegram").
		Int64("chat_id", m.Chat.ID).Str("chat_title", m.Chat.Title).
		Int("user_id", m.Sender.ID).Str("user_first_name", m.Sender.FirstName).
		Str("user_last_name", m.Sender.LastName).Str("user_username", m.Sender.Username).
		Msg(m.Text)

	// Validate the parameters
	parameters := strings.Fields(m.Payload)
	if len(parameters) != 2 {
		t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Buy.Params))
		return
	}

	units, erru := parseUint32(parameters[0])
	bells, errb := parseUint32(parameters[1])
	if erru != nil || errb != nil {
		t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Buy.Params))
		return
	}

	// Store user turnips
	new, oldUnits, oldBells, err := db.SaveThisWeekOwned(m.Sender, m.Chat, units, bells)
	if err != nil {
		t.reply(m, texts.InternalError)
		return
	}

	if new {
		t.reply(m, fmt.Sprintf(texts.Buy.Saved, units, bells))
	} else {
		t.reply(m, fmt.Sprintf(texts.Buy.Changed, units, bells, oldUnits, oldBells))
	}
}

// handleSellCmd triggers when the sell cmd is sent to a group, if sent in private the user will be warned
func (t *Telegram) handleSellCmd(m *tb.Message) {
	if m.Private() {
		t.send(m.Chat, texts.GroupOnly)
		return
	}

	log.Info().
		Str("module", "telegram").
		Int64("chat_id", m.Chat.ID).Str("chat_title", m.Chat.Title).
		Int("user_id", m.Sender.ID).Str("user_first_name", m.Sender.FirstName).
		Str("user_last_name", m.Sender.LastName).Str("user_username", m.Sender.Username).
		Msg(m.Text)

	// Validate the parameters
	parameters := strings.Fields(m.Payload)
	if len(parameters) != 1 && len(parameters) != 3 {
		t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Sell.Params))
		return
	}

	bells, err := parseUint32(parameters[0])
	if err != nil {
		t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Sell.Params))
		return
	}

	// Save the price
	var (
		new      bool
		oldBells uint32
		date     string
	)

	if len(parameters) == 1 {
		new, oldBells, date, err = db.SaveCurrentSellPrice(m.Sender, m.Chat, bells)
		if err != nil {
			t.reply(m, texts.InternalError)
			return
		}
	} else {
		new, oldBells, date, err = db.SaveSellPrice(m.Sender, m.Chat, bells, strings.Join(parameters[1:], " "))
		if err != nil {
			if err == ErrDateParse {
				t.reply(m, fmt.Sprintf(texts.Sell.InvalidDate, strings.Join(parameters[1:], " ")))
				return
			}

			t.reply(m, texts.InternalError)
			return
		}
	}

	if new {
		t.reply(m, fmt.Sprintf(texts.Sell.Saved, bells, date))
	} else {
		t.reply(m, fmt.Sprintf(texts.Sell.Changed, bells, date, oldBells))
	}
}

// handleListCmd triggers when the list cmd is sent to a group, if sent in private the user will be warned
func (t *Telegram) handleListCmd(m *tb.Message) {
	if m.Private() {
		t.send(m.Chat, texts.GroupOnly)
		return
	}

	log.Info().
		Str("module", "telegram").
		Int64("chat_id", m.Chat.ID).Str("chat_title", m.Chat.Title).
		Int("user_id", m.Sender.ID).Str("user_first_name", m.Sender.FirstName).
		Str("user_last_name", m.Sender.LastName).Str("user_username", m.Sender.Username).
		Msg(m.Text)

	owned, err := db.GetThisWeekOwned(m.Sender, m.Chat)
	if err != nil {
		t.reply(m, texts.InternalError)
		return
	}

	cost := int64(owned.Units * owned.Bells)

	prices, err := db.GetCurrentSellPrices(m.Chat)
	if err != nil {
		t.reply(m, texts.InternalError)
		return
	}

	var reply string

	if cost > 0 {
		reply += fmt.Sprintf(texts.List.ReplyOwned, owned.Units, owned.Bells)
		reply += "\n\n"
	}

	reply += texts.List.ReplyPrices + "\n"
	for _, price := range prices {
		reply += "\n" + price.User.Name()

		if price.User.Username != "" {
			reply += fmt.Sprintf(" (<code>@%s</code>)", price.User.Username)
		}

		reply += fmt.Sprintf(": <b>%v</b> %s", price.Bells, texts.BellsName)

		if cost > 0 {
			var profits int64 = int64(owned.Units*price.Bells) - cost

			if profits > 0 {
				reply += " 📈 "
			} else {
				reply += " 📉 "
			}

			reply += fmt.Sprintf("%v %s", profits, texts.BellsName)
		}
	}

	t.reply(m, reply, &tb.SendOptions{
		ParseMode: tb.ModeHTML,
	})
}

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
	"time"

	tb "gopkg.in/tucnak/telebot.v2"

	"github.com/rs/zerolog/log"
)

const (
	tzListURL = "https://en.wikipedia.org/wiki/List_of_tz_database_time_zones"
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

	// Register the group in the DB
	group, err := db.GetGroup(m.Chat)
	if err != nil {
		log.Error().Str("module", "telegram").Err(err).Msg("error getting or creating group")
		return
	}

	// Send welcome text
	t.send(m.Chat, fmt.Sprintf(texts.JoinText, group.TZ, texts.Help.Cmd))
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
		texts.Help.AvailableCmds,
		fmt.Sprintf("\n<code>/%s</code>\n%s", texts.Help.Cmd, texts.Help.Desc),
		fmt.Sprintf("\n<code>/%s</code>\n%s", texts.Admin.Cmd, texts.Admin.Desc),
		fmt.Sprintf("\n<code>/%s</code>\n%s", texts.List.Cmd, texts.List.Desc),
		fmt.Sprintf("\n<code>/%s</code>\n%s", texts.Chart.Cmd, texts.Chart.Desc),
		fmt.Sprintf("\n<code>/%s %s</code>\n%s", texts.Buy.Cmd, texts.Buy.Params, texts.Buy.Desc),
		fmt.Sprintf("\n<code>/%s %s</code>\n%s", texts.IslandPrice.Cmd, texts.IslandPrice.Params, fmt.Sprintf(texts.IslandPrice.Desc, texts.Buy.Cmd)),
		fmt.Sprintf("\n<code>/%s %s</code>\n%s", texts.Sell.Cmd, texts.Sell.Params, texts.Sell.Desc),
	}

	t.send(m.Chat, strings.Join(helpLines, "\n"), tb.NoPreview)
	t.cleanupChatMsgs(m.Chat, []*tb.Message{m})
}

// handleAdminCmd triggers when the admin cmd is sent to a group
func (t *Telegram) handleAdminCmd(m *tb.Message) {
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
		texts.Admin.AvailableCmds,
		fmt.Sprintf("\n<code>/%s %s</code>\n%s", texts.Delete.Cmd, texts.Delete.Params, texts.Delete.Desc),
		fmt.Sprintf("\n<code>/%s %s</code>\n%s", texts.ChangeTZ.Cmd, texts.ChangeTZ.Params, fmt.Sprintf(texts.ChangeTZ.Desc, tzListURL)),
	}

	t.send(m.Chat, strings.Join(helpLines, "\n"), tb.NoPreview)
	t.cleanupChatMsgs(m.Chat, []*tb.Message{m})
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
	if len(parameters) != 2 && len(parameters) != 3 {
		rm := t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Buy.Params))
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	units, err := parseUint32(parameters[0])
	bells, err2 := parseUint32(parameters[1])
	if err != nil || err2 != nil {
		rm := t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Buy.Params))
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	islandPrice := bells
	if len(parameters) == 3 {
		islandPrice, err = parseUint32(parameters[2])
		if err != nil {
			rm := t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Buy.Params))
			t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
			return
		}
	}

	// Store user turnips
	newO, oldUnits, oldBells, err := db.SaveThisWeekOwned(m.Sender, m.Chat, units, bells)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	var msgTxt1 string
	if newO || (oldUnits == units && oldBells == bells) {
		msgTxt1 = fmt.Sprintf(texts.Buy.Saved, units, bells)
	} else {
		msgTxt1 = fmt.Sprintf(texts.Buy.Changed, units, bells, oldUnits, oldBells)
	}

	// Store island price
	newIP, oldIslandPrice, err := db.SaveUserIslandPrice(m.Sender, m.Chat, islandPrice)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	var msgTxt2 string
	if newIP || (oldIslandPrice == islandPrice) {
		msgTxt2 = fmt.Sprintf(texts.IslandPrice.Saved, islandPrice)
	} else {
		msgTxt2 = fmt.Sprintf(texts.IslandPrice.Changed, islandPrice, oldIslandPrice)
	}

	// Send reply
	rm := t.reply(m, fmt.Sprintf("%s\n\n%s", msgTxt1, msgTxt2))
	t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
}

// handleIslandPriceCmd triggers when the islandprice cmd is sent to a group, if sent in private the user will be warned
func (t *Telegram) handleIslandPriceCmd(m *tb.Message) {
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
	if len(parameters) != 1 {
		rm := t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Buy.Params))
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	islandPrice, err := parseUint32(parameters[0])
	if err != nil {
		rm := t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Buy.Params))
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	// Store island price
	newIP, oldIslandPrice, err := db.SaveUserIslandPrice(m.Sender, m.Chat, islandPrice)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	var msgTxt string
	if newIP || (oldIslandPrice == islandPrice) {
		msgTxt = fmt.Sprintf(texts.IslandPrice.Saved, islandPrice)
	} else {
		msgTxt = fmt.Sprintf(texts.IslandPrice.Changed, islandPrice, oldIslandPrice)
	}

	rm := t.reply(m, msgTxt)
	t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
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
		rm := t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Sell.Params))
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	bells, err := parseUint32(parameters[0])
	if err != nil {
		rm := t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Sell.Params))
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	// Save the price
	var (
		new      bool
		oldBells uint32
		date     string
	)

	if len(parameters) == 1 {
		new, oldBells, date, err = db.SaveUserCurrentPrice(m.Sender, m.Chat, bells)
		if err != nil {
			if err == ErrBuyDay {
				rm := t.reply(m, fmt.Sprintf(texts.Sell.NoMarketToday, date, texts.Days[turnipSellDay]))
				t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
				return
			}

			rm := t.reply(m, texts.InternalError)
			t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
			return
		}
	} else {
		new, oldBells, date, err = db.SaveUserPrice(m.Sender, m.Chat, bells, strings.Join(parameters[1:], " "))
		if err != nil {
			if err == ErrDateParse {
				rm := t.reply(m, fmt.Sprintf(texts.Sell.InvalidDate, strings.Join(parameters[1:], " ")))
				t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
				return
			}

			if err == ErrBuyDay {
				rm := t.reply(m, fmt.Sprintf(texts.Sell.NoMarketToday, date, texts.Days[turnipSellDay]))
				t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
				return
			}

			rm := t.reply(m, texts.InternalError)
			t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
			return
		}
	}

	var rm *tb.Message
	if new {
		rm = t.reply(m, fmt.Sprintf(texts.Sell.Saved, bells, date))
	} else {
		rm = t.reply(m, fmt.Sprintf(texts.Sell.Changed, bells, date, oldBells))
	}
	t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
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

	owned, err := db.GetUserWeekOwned(m.Sender, m.Chat)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	cost := int64(owned.Units * owned.Bells)

	prices, date, err := db.GetCurrentPrices(m.Chat)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	var reply string

	if cost > 0 {
		reply += fmt.Sprintf(texts.List.Owned, owned.Units, owned.Bells) + "\n\n"
	}

	if len(prices) == 0 {
		reply += fmt.Sprintf(texts.List.NoPrices, date)
	} else {
		reply += fmt.Sprintf(texts.List.Prices, date) + "\n"

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

				reply += fmt.Sprintf("<b>%v</b> %s", profits, texts.BellsName)
			}
		}
	}

	t.send(m.Chat, reply)
	t.cleanupChatMsgs(m.Chat, []*tb.Message{m})
}

// handleChartCmd triggers when the chart cmd is sent to a group, if sent in private the user will be warned
func (t *Telegram) handleChartCmd(m *tb.Message) {
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

	// Get group timezone
	user, group, err := db.GetUserAndGroup(m.Sender, m.Chat)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	groupNow, err := group.NowConfig()
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	// Get prices
	prices, err := db.GetUserWeekPrices(m.Sender, m.Chat)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	if len(prices) == 0 {
		rm := t.reply(m, texts.Chart.NoPrices)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	// Get owned
	owned, err := db.GetUserWeekOwned(m.Sender, m.Chat)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	// Craft data to have a good looking graph when data is missing
	xValues := make([]time.Time, 12)
	yValues := make([]float64, 12)

	initDate := groupNow.With(time.Now().In(groupNow.TimeLocation)).BeginningOfWeek().Add(time.Hour * 24)

	for i := 0; i < 12; i++ {
		xValues[i] = initDate.Add(time.Hour * 12 * time.Duration(i))
	}

	for _, price := range prices {
		for i := range xValues {
			if price.Date.Equal(xValues[i]) {
				yValues[i] = float64(price.Bells)
				break
			}
		}
	}

	// Generate chart
	chart, err := TimeSeriesChart(user.String(), xValues, yValues, float64(owned.Bells), groupNow.TimeLocation, true)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	t.send(m.Chat, &tb.Photo{File: tb.FromReader(chart)})
	t.cleanupChatMsgs(m.Chat, []*tb.Message{m})
}

// handleDeleteCmd triggers when the delete cmd is sent to a group, if sent in private the user will be warned
func (t *Telegram) handleDeleteCmd(m *tb.Message) {
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

	// Check if it is a super admin
	isSuperAdmin := false

	for _, uid := range superAdmins {
		if int64(m.Sender.ID) == uid {
			isSuperAdmin = true
			break
		}
	}

	// Check if it is an admin
	cm, err := t.bot.ChatMemberOf(m.Chat, t.bot.Me)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	if !isSuperAdmin && cm.Role != tb.Creator && cm.Role != tb.Administrator {
		rm := t.reply(m, texts.Unprivileged)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	// Parse payload
	seconds, err := parseUint8(m.Payload)
	if err != nil || seconds > 30 {
		rm := t.reply(m, fmt.Sprintf("%v %v", texts.InvalidParams, texts.Delete.Params))
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	err = db.ChangeGroupDeleteSeconds(m.Chat, seconds)
	if err != nil {
		rm := t.reply(m, texts.InternalError)
		t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
		return
	}

	var rm *tb.Message
	if seconds > 0 {
		rm = t.reply(m, fmt.Sprintf(texts.Delete.Done, seconds))
	} else {
		rm = t.reply(m, texts.Delete.Disabled)
	}
	t.cleanupChatMsgs(m.Chat, []*tb.Message{m, rm})
}

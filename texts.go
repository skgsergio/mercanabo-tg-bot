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
	"encoding/json"
	"fmt"
	"os"

	"github.com/rs/zerolog/log"
)

// Texts represent the texts used in user parts of the bot
type Texts struct {
	GroupOnly     string   `json:"group_only"`
	JoinText      string   `json:"join_text"`
	InternalError string   `json:"internal_error"`
	InvalidParams string   `json:"invalid_parameters"`
	BellsName     string   `json:"bells_name"`
	Days          []string `json:"days"`
	DaysShort     []string `json:"days_short"`
	Unprivileged  string   `json:"unprivileged"`

	Help struct {
		Cmd           string `json:"cmd"`
		Desc          string `json:"desc"`
		AvailableCmds string `json:"available_cmds"`
		AdminCmds     string `json:"admin_cmds"`
	} `json:"help"`

	Buy struct {
		Cmd     string `json:"cmd"`
		Params  string `json:"params"`
		Desc    string `json:"desc"`
		Saved   string `json:"saved"`
		Changed string `json:"changed"`
	} `json:"buy"`

	Sell struct {
		Cmd           string `json:"cmd"`
		Params        string `json:"params"`
		Desc          string `json:"desc"`
		Saved         string `json:"saved"`
		Changed       string `json:"changed"`
		InvalidDate   string `json:"invalid_date"`
		NoMarketToday string `json:"no_market_today"`
	} `json:"sell"`

	List struct {
		Cmd         string `json:"cmd"`
		Desc        string `json:"desc"`
		ReplyOwned  string `json:"reply_owned"`
		ReplyPrices string `json:"reply_prices"`
	} `json:"list"`

	Chart struct {
		Cmd  string `json:"cmd"`
		Desc string `json:"desc"`
	} `json:"chart"`

	Delete struct {
		Cmd      string `json:"cmd"`
		Params   string `json:"params"`
		Desc     string `json:"desc"`
		Done     string `json:"done"`
		Disabled string `json:"disabled"`
	} `json:"delete"`

	ChangeTZ struct {
		Cmd    string `json:"cmd"`
		Params string `json:"params"`
		Desc   string `json:"desc"`
	} `json:"changetz"`
}

// LoadTexts load a language texts json file and returns it as Texts
func LoadTexts(lang string) (*Texts, error) {
	txtFile, err := os.Open(fmt.Sprintf("texts/%s.json", lang))
	if err != nil {
		return nil, err
	}
	defer func() {
		err = txtFile.Close()
		if err != nil {
			log.Error().Err(err).Msg("texts file close")
		}
	}()

	var txt = Texts{}
	decoder := json.NewDecoder(txtFile)
	if decoder.Decode(&txt) != nil {
		return nil, err
	}

	return &txt, nil
}

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
	GroupOnly     string `json:"group_only"`
	JoinText      string `json:"join_text"`
	InternalError string `json:"internal_error"`
	InvalidParams string `json:"invalid_parameters"`
	BuyCmd        string `json:"buy_cmd"`
	BuyDesc       string `json:"buy_desc"`
	BuySaved      string `json:"buy_saved"`
	BuyChanged    string `json:"buy_changed"`
	SellCmd       string `json:"sell_cmd"`
	SellDesc      string `json:"sell_desc"`
	SellSaved     string `json:"sell_saved"`
	SellChanged   string `json:"sell_changed"`
	TZCmd         string `json:"tz_cmd"`
	TZDesc        string `json:"tz_desc"`
	BellsName     string `json:"bells_name"`
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

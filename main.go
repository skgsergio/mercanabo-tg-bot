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
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	turnipSellDay  = time.Sunday
	timeFormatAMPM = "2006-01-02 PM"
)

var (
	defaultTZ   string    = "UTC"
	bot         *Telegram = nil
	db          *Database = nil
	texts       *Texts    = nil
	superAdmins []int64   = []int64{}
)

func main() {
	var (
		lang string = "default"
		err  error  = nil
	)

	// Configure logger
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.SetGlobalLevel(zerolog.InfoLevel)

	if os.Getenv("MERCANABO_DEBUG") == "true" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	log.Logger = log.With().Caller().Logger()

	// Check required env vars
	for _, envVar := range []string{
		"MERCANABO_TOKEN",
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"POSTGRES_SSLMODE",
	} {
		if os.Getenv(envVar) == "" {
			log.Fatal().Str("module", "main").Str("envvar", envVar).Msg("missing environment variable")
		}
	}

	// Load bot texts
	if envl := os.Getenv("MERCANABO_LANG"); envl != "" {
		lang = envl
	}

	texts, err = LoadTexts(lang)
	if err != nil {
		log.Fatal().Str("module", "main").Err(err).Msg("failed loading texts file")
	}

	log.Info().Str("module", "main").Str("lang", lang).Msg("loaded texts")

	// Load default time zone
	if envtz := os.Getenv("MERCANABO_DEFAULT_TZ"); envtz != "" {
		defaultTZ = envtz

		if _, err = time.LoadLocation(defaultTZ); err != nil {
			log.Fatal().Str("module", "main").Str("timezone", defaultTZ).Err(err).Msg("failed loading time zone")
		}
	}

	log.Info().Str("module", "main").Str("timezone", defaultTZ).Msg("loaded default timezone for new groups")

	// Load superadmin list
	if envsa := os.Getenv("MERCANABO_SUPERADMINS"); envsa != "" {
		for _, uidStr := range strings.Split(envsa, ",") {
			uid, errp := parseInt64(uidStr)

			if errp != nil {
				log.Fatal().Str("module", "main").Err(err).Msg("failed parsing superadmins list")
			}

			superAdmins = append(superAdmins, uid)
		}
	}

	log.Info().Str("module", "main").Ints64("user_ids", superAdmins).Msg("loaded superadmins")

	// Connecto to the DB
	db, err = OpenDB(
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
		os.Getenv("POSTGRES_SSLMODE"),
		os.Getenv("MERCANABO_DEBUG") == "true",
	)

	if err != nil {
		log.Fatal().Str("module", "main").Err(err).Msg("failed opening database")
	}

	db.SetupDB()

	// Create bot
	bot, err = NewBot(os.Getenv("MERCANABO_TOKEN"))

	if err != nil {
		log.Fatal().Str("module", "telegram").Err(err).Msg("failed bot instantiaion")
	}

	// Start the bot
	bot.Start()
}

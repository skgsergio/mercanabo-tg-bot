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

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"

	"github.com/rs/zerolog/log"
)

// Database represents the database with some basic queries
type Database struct {
	DB *gorm.DB
}

// OpenDB opens the database and sets the logger
func OpenDB(host string, port string, user string, password string, dbname string, sslmode string, debug bool) (*Database, error) {
	db, err := gorm.Open("postgres", fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	))

	if err != nil {
		log.Error().Str("module", "database").Err(err).Msg("failed opening database")
		return nil, err
	}

	db.SetLogger(ZerologGorm{})
	db.LogMode(debug)

	return &Database{DB: db}, nil
}

// ZerologGorm is a simple custom logger using Zerolog for GORM
type ZerologGorm struct{}

// Print a GORM log entry
func (ZerologGorm) Print(v ...interface{}) {
	switch v[0] {
	case "sql":
		log.Debug().
			Str("module", "gorm").
			Fields(map[string]interface{}{
				"type":   v[0],
				"rows":   v[5],
				"src":    v[1],
				"values": v[4],
			}).
			Msg(fmt.Sprintf("%v", v[3]))

	case "log":
		log.Debug().
			Str("module", "gorm").
			Fields(map[string]interface{}{
				"type": v[0],
			}).
			Msg(fmt.Sprintf("%v", v[2]))
	}
}

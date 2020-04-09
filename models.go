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
	"time"

	"github.com/rs/zerolog/log"
)

// Group represents a Telegram group
type Group struct {
	ID    int64  `gorm:"PRIMARY_KEY;NOT NULL"`
	Title string `gorm:"NOT NULL"`
	TZ    string `gorm:"NOT NULL"`
}

// User represents a Telegram user
type User struct {
	ID        int64  `gorm:"PRIMARY_KEY;NOT NULL"`
	FirstName string `gorm:"NOT NULL"`
	LastName  string `gorm:"NOT NULL"`
	Username  string `gorm:"NOT NULL"`
}

// Price is a price that a User recorded in a Group
type Price struct {
	ID      uint64    `gorm:"PRIMARY_KEY;AUTO_INCREMENT;NOT NULL"`
	GroupID int64     `gorm:"INDEX;NOT NULL"`
	Group   Group     `gorm:"FOREIGNKEY:GroupID"`
	UserID  int64     `gorm:"INDEX;NOT NULL"`
	User    User      `gorm:"FOREIGNKEY:UserID"`
	Bells   uint32    `gorm:"NOT NULL"`
	Date    time.Time `gorm:"INDEX;NOT NULL"`
}

// Turnip represents how many turnips owns an User in a Group in a given date
// An User has to record in each Group how many turnips owns to handle correctly Groups with differnt time zones.
type Turnip struct {
	ID      uint64    `gorm:"PRIMARY_KEY;AUTO_INCREMENT;NOT NULL"`
	GroupID int64     `gorm:"INDEX;NOT NULL"`
	Group   Group     `gorm:"FOREIGNKEY:GroupID"`
	UserID  int64     `gorm:"INDEX;NOT NULL"`
	User    User      `gorm:"FOREIGNKEY:UserID"`
	Units   uint32    `gorm:"NOT NULL"`
	Bells   uint32    `gorm:"NOT NULL"`
	Date    time.Time `gorm:"INDEX;NOT NULL"`
}

// SetupDB runs database migrations
func (d *Database) SetupDB() {
	log.Info().Str("module", "database").Msg("running database migrations")
	// Run migrations
	d.DB.AutoMigrate(
		&Group{},
		&User{},
		&Price{},
		&Turnip{},
	)

	// Add the FKs
	priceModel := d.DB.Model(&Price{})
	priceModel.AddForeignKey("group_id", "groups(id)", "CASCADE", "CASCADE")
	priceModel.AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")

	turnipModel := d.DB.Model(&Turnip{})
	turnipModel.AddForeignKey("group_id", "groups(id)", "CASCADE", "CASCADE")
	turnipModel.AddForeignKey("user_id", "users(id)", "CASCADE", "CASCADE")
}

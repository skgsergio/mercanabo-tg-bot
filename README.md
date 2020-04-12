# Mercanabo Telegram Bot
A simple Telegram bot that allows users of a Telegram Group to track Turnip sell
prices and store how much they bought in oder to maximize profits.

## Build

```sh
go build
```

## Running it

### Requirements

- [PostgreSQL](https://www.postgresql.org/)

### Configuration environment variables

- `MERCANABO_TOKEN`: Telegram token. See https://core.telegram.org/bots
- `MERCANABO_SUPERADMINS`: Comma separated list of Telegram user ids,
  superadmins will have power over the bot in any channel as well as accesing
  private administration commands.
- `MERCANABO_DEFAULT_TZ` (default: `UTC`): Default Time Zone for new groups, group admins can
  change the timezone for their group. See: https://en.wikipedia.org/wiki/List_of_tz_database_time_zones
- `MERCANABO_LANG` (default: `default`): Bot language, this is global and can't be changed per
  group. See: [texts](texts)
- `MERCANABO_DEBUG`: If `true` then sets the log level to `debug`, changes the
  log output to a colorful mode and enables `gorm` debug log.
- `POSTGRES_HOST`: PostgreSQL hostname.
- `POSTGRES_PORT`: PostgreSQL port.
- `POSTGRES_SSLMODE`: PostgreSQL sslmode. See: https://pkg.go.dev/github.com/lib/pq
- `POSTGRES_USER`: PostgreSQL user.
- `POSTGRES_PASSWORD`: PostgreSQL password.
- `POSTGRES_DB`: PostgreSQL database name.

### Docker
To simplify the deployment there is a [Docker Compose](docker-compose.yml) file
ready to use. You can just use it and it will deploy a postgresql container, with
a volume for persistence, and a bot container, that is build from local source.

## TODO

- Implement the forecaster ([resources](forecast/README.md))

## License
This project is licensed under the GPL 3.0 License. See the [LICENSE](LICENSE)
file for details.

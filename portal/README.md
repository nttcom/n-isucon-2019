## Deployment

TBD

## Development
### requirements

* Ruby 2.6
* mysql

### how to run
1. `bundle install --path .bundle`
2. copy `.env.example` to `.env` and edit secret params
3. create database `dashboard`
3. `bundle exec rackup`
4. access localhost:9292

### daemonize
1. put `nisucon-portal.service` to systemd config dir & `systemd daemon-reload`
2. put `nisucon-portal.nginx` to nginx setting dir & restart nginx
3. access to https://portal.ncom.dev/

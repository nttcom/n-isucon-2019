#!/bin/bash

# initialize db for scoring

ROOT_DIR=$(cd $(dirname $0)/..; pwd)
DB_DIR="$ROOT_DIR/db"
BENCH_DIR="$ROOT_DIR/bench"

MYSQL_DATABASE=app
MYSQL_USER=username
MYSQL_PASSWORD=password
MYSQL_ROOT_PASSWORD=rootpassword

# delete except initial data
mysql -h db -u$MYSQL_USER -p$MYSQL_PASSWORD -e "DELETE FROM users WHERE id > 10000"
mysql -h db -u$MYSQL_USER -p$MYSQL_PASSWORD -e "DELETE FROM icon WHERE id > 10000"
mysql -h db -u$MYSQL_USER -p$MYSQL_PASSWORD -e "DELETE FROM items WHERE id > 20000"
mysql -h db -u$MYSQL_USER -p$MYSQL_PASSWORD -e "DELETE FROM comments WHERE id > 10081"

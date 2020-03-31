#!/bin/bash

# force DB initialization

ROOT_DIR=$(cd $(dirname $0)/..; pwd)
DB_DIR="$ROOT_DIR/db"
BENCH_DIR="$ROOT_DIR/bench"

MYSQL_DATABASE=app
MYSQL_USER=username
MYSQL_PASSWORD=password
MYSQL_ROOT_PASSWORD=rootpassword

mysql -h db -u$MYSQL_USER -p$MYSQL_PASSWORD -e "DROP DATABASE IF EXISTS $MYSQL_DATABASE; CREATE DATABASE $MYSQL_DATABASE;"
mysql -h db -u$MYSQL_USER -p$MYSQL_PASSWORD $MYSQL_DATABASE < "$DB_DIR/init.sql"
if [[ -f "$DB_DIR/seed.sql" ]]; then
        mysql -h $MYSQL_HOST -u$MYSQL_USER -p$MYSQL_PASSWORD $MYSQL_DATABASE < "$DB_DIR/seed.sql"
fi

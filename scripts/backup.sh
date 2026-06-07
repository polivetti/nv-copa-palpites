#!/bin/sh
set -eu

DB_PATH="${DB_PATH:-${DATABASE_PATH:-/data/copa.db}}"
BACKUP_DIR="${BACKUP_DIR:-/backups}"

mkdir -p "$BACKUP_DIR"

if [ ! -f "$DB_PATH" ]; then
  echo "database not found: $DB_PATH"
  exit 0
fi

stamp="$(date +%Y%m%d-%H%M%S)"
target="$BACKUP_DIR/copa-$stamp.db"

sqlite3 "$DB_PATH" ".backup '$target'"

find "$BACKUP_DIR" -name "copa-*.db" -type f -mtime +14 -delete
echo "backup written: $target"

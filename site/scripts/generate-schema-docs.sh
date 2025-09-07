#!/usr/bin/env bash
set -e

# Paths
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
SCHEMASPY_DIR="$ROOT_DIR/docs/schemaspy"
OUTPUT_DIR="$ROOT_DIR/docs/schema"
JAR_FILE="$SCHEMASPY_DIR/schemaspy.jar"
DRIVER_FILE="$SCHEMASPY_DIR/postgresql.jar"

# Config
DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="ledger"
DB_USER="admin"
DB_PASS="admin"
DB_SCHEMA="public"

# Ensure dependencies
command -v java >/dev/null 2>&1 || {
  echo "❌ Java not found. Install OpenJDK."
  exit 1
}
command -v dot >/dev/null 2>&1 || {
  echo "❌ Graphviz not found. Install with: pkg install graphviz"
  exit 1
}

# Ensure jars exist
if [[ ! -f "$JAR_FILE" ]]; then
  echo "❌ schemaspy.jar not found in $JAR_FILE"
  echo "   Download from: https://github.com/schemaspy/schemaspy/releases"
  exit 1
fi

if [[ ! -f "$DRIVER_FILE" ]]; then
  echo "❌ postgresql.jar not found in $DRIVER_FILE"
  echo "   Download from: https://jdbc.postgresql.org/download/"
  exit 1
fi

# Run SchemaSpy
echo "🚀 Generating database documentation for schema '$DB_SCHEMA'..."
cd "$SCHEMASPY_DIR"

java -jar "$JAR_FILE" \
  -t pgsql \
  -host "$DB_HOST" \
  -port "$DB_PORT" \
  -db "$DB_NAME" \
  -s "$DB_SCHEMA" \
  -u "$DB_USER" \
  -p "$DB_PASS" \
  -dp "$DRIVER_FILE" \
  -o "$OUTPUT_DIR" \
  -vizjs \
  -showComments \
  -noCatalog \
  -noroutines

cd "$ROOT_DIR"
echo "✅ Documentation generated at: docs/schema/index.html"

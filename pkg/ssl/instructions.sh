#!/bin/bash
set -euo pipefail

# Default values
CN="localhost"
DAYS=3650
PASSWORD="$(openssl rand -base64 32)" # generate a strong random password
FORCE=0

usage() {
  echo "Usage: $0 [--cn <common name>] [--days <days>] [--password <pass>] [--force]"
  exit 1
}

while [[ $# -gt 0 ]]; do
  case $1 in
  --cn)
    CN="$2"
    shift 2
    ;;
  --days)
    DAYS="$2"
    shift 2
    ;;
  --password)
    PASSWORD="$2"
    shift 2
    ;;
  --force)
    FORCE=1
    shift
    ;;
  *) usage ;;
  esac
done

# Check for openssl
command -v openssl >/dev/null 2>&1 || {
  echo "openssl not found"
  exit 1
}

# Protect existing files if not forced
for f in ca.key ca.crt ...; do
  if [[ -e $f ]]; then
    read -p "File $f exists. Overwrite? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then exit 1; fi
  fi
done

echo "Generating CA key and certificate..."
openssl genrsa -aes256 -passout "pass:${PASSWORD}" -out ca.key 4096
openssl req -new -x509 -days "${DAYS}" -passin "pass:${PASSWORD}" \
  -key ca.key -out ca.crt -subj "/CN=${CN}"

echo "Generating server private key..."
openssl genrsa -aes256 -passout "pass:${PASSWORD}" -out server.key 4096

echo "Creating server certificate signing request..."
openssl req -new -passin "pass:${PASSWORD}" \
  -key server.key -out server.csr -subj "/CN=${CN}"

echo "Signing server certificate with CA..."
openssl x509 -req -days "${DAYS}" -passin "pass:${PASSWORD}" \
  -in server.csr -CA ca.crt -CAkey ca.key -set_serial "0x$(openssl rand -hex 16)" \
  -out server.crt -extfile <(printf "subjectAltName=DNS:${CN}")

echo "Converting server key to PKCS#8 unencrypted format for gRPC..."
openssl pkcs8 -topk8 -nocrypt -passin "pass:${PASSWORD}" \
  -in server.key -out server.pem

# Secure private key permissions
chmod 600 ca.key server.key server.pem

echo "Done."
echo "Private files: ca.key server.key server.pem"
echo "Public files:  ca.crt server.crt"
echo "CSR (can be discarded now): server.csr"

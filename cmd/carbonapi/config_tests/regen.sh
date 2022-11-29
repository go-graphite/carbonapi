#!/usr/bin/env bash

echo "THIS IS INSECURE AND IS ONLY USED FOR Integration test purposes"
echo "Never use it in production, please read how to do it properly!!!"

# Create CA
openssl genpkey -algorithm Ed25519 -out mTLS-server.key
openssl req -new -x509 -sha256 -key mTLS-server.key -out mTLS-server.crt -days 3650 -subj '/CN=localhost' -addext "subjectAltName = DNS:localhost"

echo "mTLS Test Server certificate has been created:" 
openssl x509 -noout -text -in mTLS-server.crt

# Client cert
openssl genpkey -algorithm Ed25519 -out mTLS-client.key
openssl req -new -key mTLS-client.key -out mTLS-client.csr -subj '/CN=test-uuid'

# Sign our client certificate with our CA
echo "00" > file.srl
openssl x509 -req -in mTLS-client.csr -CA mTLS-server.crt -CAkey mTLS-server.key -CAserial file.srl -out mTLS-client.crt

echo "mTLS Test Client certificate has been created:" 
openssl x509 -noout -text -in mTLS-client.crt


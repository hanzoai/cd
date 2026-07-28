# Testing certificates

This directory contains all TLS certificates used for testing Hanzo CD, including
the E2E tests. It also contains the CA certificate and key used for signing the
certificates.

## Certificate Files

- `cd-test-ca.crt` and `cd-test-ca.key` - The CA certificate and key
- `cd-test-client.crt` and `cd-test-client.key` - Client certificate/key pair for mTLS testing
- `cd-test-server.crt` and `cd-test-server.key` - Server certificate/key pair (CN=localhost, SANs: localhost, cd-e2e-server, 127.0.0.1)
- `cd-e2e-server.crt` and `cd-e2e-server.key` - Server certificate for remote E2E tests (CN=cd-e2e-server)

All keys have no passphrase. All certs are valid for 100 years.

**Do not use these certs for anything other than Hanzo CD tests.**

## Regenerating Certificates

If you need to regenerate the certificates (e.g., for compliance with newer TLS standards), run the following commands from this directory:

### 1. Generate CA Certificate

```bash
openssl genrsa -out cd-test-ca.key 4096

openssl req -new -x509 -days 36500 -key cd-test-ca.key \
    -out cd-test-ca.crt \
    -subj "/CN=Hanzo CD Test CA" \
    -addext "basicConstraints=critical,CA:TRUE" \
    -addext "keyUsage=critical,keyCertSign,cRLSign"
```

### 2. Generate Server Certificate (for localhost testing)

```bash
openssl genrsa -out cd-test-server.key 4096

openssl req -new -key cd-test-server.key \
    -out cd-test-server.csr \
    -subj "/CN=localhost"

cat > server_ext.cnf << 'EOF'
basicConstraints=CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:localhost,DNS:cd-e2e-server,IP:127.0.0.1
EOF

openssl x509 -req -in cd-test-server.csr \
    -CA cd-test-ca.crt -CAkey cd-test-ca.key \
    -CAcreateserial -out cd-test-server.crt \
    -days 36500 -sha256 \
    -extfile server_ext.cnf

rm -f server_ext.cnf cd-test-server.csr cd-test-ca.srl
```

### 3. Generate Client Certificate (for mTLS testing)

```bash
openssl genrsa -out cd-test-client.key 4096

openssl req -new -key cd-test-client.key \
    -out cd-test-client.csr \
    -subj "/CN=Hanzo CD Test Client"

cat > client_ext.cnf << 'EOF'
basicConstraints=CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=clientAuth
EOF

openssl x509 -req -in cd-test-client.csr \
    -CA cd-test-ca.crt -CAkey cd-test-ca.key \
    -CAcreateserial -out cd-test-client.crt \
    -days 36500 -sha256 \
    -extfile client_ext.cnf

rm -f client_ext.cnf cd-test-client.csr cd-test-ca.srl
```

### 4. Generate E2E Server Certificate (for remote testing)

```bash
openssl genrsa -out cd-e2e-server.key 4096

openssl req -new -key cd-e2e-server.key \
    -out cd-e2e-server.csr \
    -subj "/CN=cd-e2e-server"

cat > e2e_server_ext.cnf << 'EOF'
basicConstraints=CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth
subjectAltName=DNS:cd-e2e-server,DNS:localhost,IP:127.0.0.1
EOF

openssl x509 -req -in cd-e2e-server.csr \
    -CA cd-test-ca.crt -CAkey cd-test-ca.key \
    -CAcreateserial -out cd-e2e-server.crt \
    -days 36500 -sha256 \
    -extfile e2e_server_ext.cnf

rm -f e2e_server_ext.cnf cd-e2e-server.csr cd-test-ca.srl
```

### Verify Certificates

After regenerating, verify the certificate chain:

```bash
openssl verify -CAfile cd-test-ca.crt cd-test-server.crt
openssl verify -CAfile cd-test-ca.crt cd-test-client.crt
openssl verify -CAfile cd-test-ca.crt cd-e2e-server.crt
```

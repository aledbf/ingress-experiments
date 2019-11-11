
test-certificate:
	openssl genrsa -des3 -out rootCA.key 4096
	openssl req -x509 -new -nodes -key rootCA.key -sha256 -days 1024 -out rootCA.crt

	openssl genrsa -out localhost.key 2048

	openssl req -new -sha256 \
		-key localhost.key \
		-subj "/O=Acme Co/OU=Fake SSL Certificate/CN=127.0.0.1" \
		-out localhost.csr

	openssl x509 -req \
		-in localhost.csr -CA rootCA.crt -CAkey rootCA.key -CAcreateserial \
		-out localhost.crt -days 500 -sha256

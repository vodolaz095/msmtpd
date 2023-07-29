deps:
	go mod download
	go mod verify
	go mod tidy

test: check

lint:
	gofmt  -w=true -s=true -l=true ./
	golint ./...
	go vet ./...

check: lint
	go test -v ./...

clean:
	go clean

start_simple:
	go run example/simple/main.go

check_simple:
	swaks --to recipient@example.org \
          --from sender@example.org \
          --server localhost --port 1025 \
          --timeout 600

start_smtp_proxy:
	go run example/smtp_proxy/main.go

check_smtp_proxy:
	swaks --to recipient@example.org \
          --from sender@yandex.ru \
          --server localhost --port 1025 \
          --tls --timeout 600

start_dovecot_inbound:
	go run example/dovecot_inbound/main.go

check_dovecot_inbound:
	swaks --to recipient@example.org \
          --from sender@yandex.ru \
          --helo localhost \
          --server localhost --port 1025 \
          --tls --timeout 600

start_dovecot_outbound:
	go run example/dovecot_outbound/main.go

check_dovecot_outbound:
	swaks --to recipient@example.org \
          --from sender@yandex.ru \
          --helo localhost \
          --auth-user vodolaz095 --auth-password thisIsNotAPassword \
          --server localhost --port 1587 \
          --tls-on-connect --timeout 600


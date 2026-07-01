.PHONY: run build tidy init-db

run:
	go run ./cmd/api -config configs/config.yaml

build:
	GOTMPDIR=.tmp go build -o bin/usercore ./cmd/api

tidy:
	GOTMPDIR=.tmp go mod tidy

init-db:
	@test -n "$(APP_PASSWORD)" || (echo "用法: make init-db APP_PASSWORD=你的密码 [SUDO=1]"; exit 1)
	chmod +x deploy/setup_db.sh
ifeq ($(SUDO),1)
	./deploy/setup_db.sh "$(APP_PASSWORD)" --sudo
else
	./deploy/setup_db.sh "$(APP_PASSWORD)" "$${PG_SUPERUSER:-postgres}"
endif

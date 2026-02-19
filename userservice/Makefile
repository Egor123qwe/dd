DB_DRIVER=mysql
DB_STRING="user:password@tcp(localhost:3306)/farm?parseTime=true"
MIGRATIONS_DIR=./migration/mysql/farm

# go mod tidy: при GOSUMDB=off или кастомном GOPROXY используйте make tidy
tidy:
	GOSUMDB=sum.golang.org GOPROXY=https://proxy.golang.org,direct go mod tidy

pull:
	git pull

migrate-create:
	@goose -dir $(MIGRATIONS_DIR) create $(name) sql

migrate-up:
	@goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) $(DB_STRING) up

migrate-down:
	@goose -dir $(MIGRATIONS_DIR) $(DB_DRIVER) $(DB_STRING) down


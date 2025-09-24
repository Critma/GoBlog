docs:
	swag init -g ./api/main.go -d cmd,internal && swag fmt
.PHONY: docs
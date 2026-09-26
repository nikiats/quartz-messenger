.PHONY: gen

# Генерация Go-кода из OpenAPI-спецификаций (api/)
gen:
	go generate ./...

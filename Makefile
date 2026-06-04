.PHONY: swagger
swagger:
	swag init -g main.go -o ./docs --parseDependency --parseInternal

.PHONY: swagger-clean
swagger-clean:
	rm -rf docs/

.PHONY: swagger-serve
swagger-serve:
	swag init -g main.go -o ./docs && \
	docker run -p 8081:8080 -e SWAGGER_JSON=/docs/swagger.json -v $(PWD)/docs:/docs swaggerapi/swagger-ui
API_PORT=4041
PREF_ENV=development #preferred environment values production or development
API_BINARY_NAME=alpha-api
BINARY_DIR=bin
DOMAIN=http://127.0.0.1

build_api:
	@echo "building prototype api backend..."
	@go build -o ${BINARY_DIR}/${API_BINARY_NAME} ./cmd/*.go
	@echo "prototype api backend built"

run_api:
	./${BINARY_DIR}/${API_BINARY_NAME} -environment ${PREF_ENV} -port ${API_PORT} -apiUrl ${DOMAIN}:${API_PORT}/

test_db_package:
	 go test -v ./internal/db

start_api: clean_api build_api run_api

clean_api:
	@echo "cleaning api binary files..."
	@go clean
	@- rm -f ${BINARY_DIR}/${API_BINARY_NAME}
	@echo "api binary files cleaned"

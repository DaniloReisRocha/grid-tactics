mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(dir $(mkfile_path))
frontend_api := $(abspath $(current_dir)../frontend/src/api/zodios/apiClient.ts)

zod:
	go run $(current_dir)tools/zodios/zodios.go $(frontend_api)

dev:
	go run $(current_dir)cmd/server/main.go $(current_dir)envs
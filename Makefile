build:
	docker compose -f docker-compose-go.yaml run --rm go_plugin_compile

run:
	docker compose down && docker compose up

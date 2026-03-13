.PHONY: clean
clean:
	go clean

.PHONY: scheduler
scheduler:
	go build -race -ldflags "-s -w" -o ./bin ./scheduler

.PHONY: server
server:
	go build -race -ldflags "-s -w" -o ./bin ./server/

templ:
	go tool templ generate --watch --proxy="http://localhost:8090" --open-browser=false

tailwind-clean:
	npx @tailwindcss/cli -i ./assets/css/input.css -o ./assets/css/output.css --clean

tailwind-watch:
	npx @tailwindcss/cli -i ./assets/css/input.css -o ./assets/css/output.css --watch

air:
	air \
    --build.cmd "go build -race -ldflags '-s -w' -o ./bin ./server/" \
    --build.entrypoint "./bin/server" \
    --build.delay "50" \
    --build.exclude_dir "node_modules" \
    --build.include_ext "go" \
    --build.stop_on_error "false" \
    --misc.clean_on_exit true

devenv:
	make tailwind-clean
	make scheduler
	make -j3 tailwind-watch templ air

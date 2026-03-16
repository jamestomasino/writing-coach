SHELL := /bin/bash

LT_HOME ?= /opt/languagetool
LT_PORT ?= 8081
LT_PID_FILE ?= .writing-coach/languagetool.pid
LT_LOG_FILE ?= .writing-coach/languagetool.log
LT_URL ?= http://localhost:$(LT_PORT)
LT_SERVER_JAR ?= $(firstword $(wildcard $(LT_HOME)/languagetool-server.jar $(LT_HOME)/LanguageTool-*/languagetool-server.jar))
LT_HEALTHCHECK = curl -fsS -d 'language=en-US' -d 'text=Health check.' "$(LT_URL)/v2/check"

.PHONY: help init build prompt submit review coach-review history progress vale-install languagetool-start languagetool-stop languagetool-status

help: ## show available targets and what they do
	@echo "targets:"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
	| sed -n 's/^\(.*\): \(.*\)##\(.*\)/  \1|\3/p' \
	| column -t -s '|'

init: ## initialize config, schema, and seeded curriculum state
	go run ./cmd/writing-coach init

build: ## compile the Go project
	go build ./...

prompt: ## generate and store the next writing exercise
	go run ./cmd/writing-coach prompt next

submit: ## save a draft file as a submission; requires EXERCISE=<id> FILE=<path>
	@test -n "$(EXERCISE)" || (echo "EXERCISE is required"; exit 1)
	@test -n "$(FILE)" || (echo "FILE is required"; exit 1)
	go run ./cmd/writing-coach submit --exercise $(EXERCISE) --file $(FILE)

review: ## review a submission with current analyzers/models; requires SUBMISSION=<id>
	@test -n "$(SUBMISSION)" || (echo "SUBMISSION is required"; exit 1)
	LANGUAGETOOL_URL=$(LT_URL) go run ./cmd/writing-coach review --submission $(SUBMISSION)

coach-review: ## start LanguageTool if needed, then run a full review; requires SUBMISSION=<id>
	@test -n "$(SUBMISSION)" || (echo "SUBMISSION is required"; exit 1)
	@$(MAKE) languagetool-start
	LANGUAGETOOL_URL=$(LT_URL) go run ./cmd/writing-coach review --submission $(SUBMISSION)

history: ## show recent exercises, submissions, and reviews
	go run ./cmd/writing-coach history

progress: ## show current focus and recent per-skill trends
	go run ./cmd/writing-coach progress

vale-install: ## install a repo-local Vale binary at .writing-coach/bin/vale
	@mkdir -p .writing-coach/bin
	GOBIN="$(CURDIR)/.writing-coach/bin" go install github.com/errata-ai/vale/v3/cmd/vale@latest
	@echo "Vale installed at .writing-coach/bin/vale"

languagetool-start: ## launch the LanguageTool HTTP server in the background
	@mkdir -p .writing-coach
	@if [ -z "$(LT_SERVER_JAR)" ]; then \
		echo "LanguageTool server jar not found under $(LT_HOME)"; \
		exit 1; \
	fi
	@if $(LT_HEALTHCHECK) >/dev/null 2>&1; then \
		test -n "$$(lsof -tiTCP:$(LT_PORT) -sTCP:LISTEN 2>/dev/null | head -n 1)" && echo "$$(lsof -tiTCP:$(LT_PORT) -sTCP:LISTEN 2>/dev/null | head -n 1)" >"$(LT_PID_FILE)"; \
		echo "LanguageTool already running"; \
		test -f "$(LT_PID_FILE)" && echo "PID: $$(cat $(LT_PID_FILE))"; \
		echo "URL: $(LT_URL)"; \
		exit 0; \
	fi
	@setsid -f bash -lc 'cd "$(dir $(LT_SERVER_JAR))" && exec java -cp "$(notdir $(LT_SERVER_JAR)):libs/*" org.languagetool.server.HTTPServer --port $(LT_PORT) --allow-origin' >"$(LT_LOG_FILE)" 2>&1
	@for i in 1 2 3 4 5 6 7 8 9 10; do \
		if $(LT_HEALTHCHECK) >/dev/null 2>&1; then \
			lsof -tiTCP:$(LT_PORT) -sTCP:LISTEN 2>/dev/null | head -n 1 >"$(LT_PID_FILE)"; \
			break; \
		fi; \
		sleep 1; \
	done
	@if $(LT_HEALTHCHECK) >/dev/null 2>&1; then \
		echo "LanguageTool started"; \
		test -f "$(LT_PID_FILE)" && echo "PID: $$(cat $(LT_PID_FILE))"; \
		echo "URL: $(LT_URL)"; \
		echo "Log: $(LT_LOG_FILE)"; \
	else \
		echo "LanguageTool failed to start"; \
		test -f "$(LT_LOG_FILE)" && tail -n 40 "$(LT_LOG_FILE)"; \
		rm -f "$(LT_PID_FILE)"; \
		exit 1; \
	fi

languagetool-stop: ## stop the LanguageTool server listening on LT_PORT
	@if pids="$$(lsof -tiTCP:$(LT_PORT) -sTCP:LISTEN 2>/dev/null)"; [ -n "$$pids" ]; then \
		kill $$pids; \
		echo "Stopped LanguageTool PID(s): $$pids"; \
	else \
		echo "LanguageTool is not running"; \
	fi
	@rm -f "$(LT_PID_FILE)"

languagetool-status: ## check whether the LanguageTool server is reachable
	@if $(LT_HEALTHCHECK) >/dev/null 2>&1; then \
		lsof -tiTCP:$(LT_PORT) -sTCP:LISTEN 2>/dev/null | head -n 1 >"$(LT_PID_FILE)"; \
		echo "LanguageTool running"; \
		test -f "$(LT_PID_FILE)" && echo "PID: $$(cat $(LT_PID_FILE))"; \
		echo "URL: $(LT_URL)"; \
	else \
		echo "LanguageTool not running"; \
		exit 1; \
	fi

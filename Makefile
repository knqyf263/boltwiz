.PHONY: ui-install ui-build

UI_DIR := ui
NODE_MODULES := $(UI_DIR)/node_modules
DIST := $(UI_DIR)/dist

$(NODE_MODULES):
	cd $(UI_DIR) && npm install

$(DIST): $(NODE_MODULES)
	cd $(UI_DIR) && npm run build

ui-install: $(NODE_MODULES)

ui-build: $(DIST)
# Makefile for rex

THEMES = light dark blue green brown
CSS_FILES = $(patsubst %,pkg/app-%.css,$(THEMES))

# Default target
all: test build

# Build the rex binary
build: themes
	go build ./pkg/...
	go build -o rex .

test:
	go test ./pkg/...

# Build CSS themes from LESS templates
themes: $(CSS_FILES)
	@rm -f pkg/variables.less
	@echo "All themes built successfully!"

pkg/app-%.css: pkg/variables-%.less pkg/app.less
	@echo "Building $* theme..."
	@cp $< pkg/variables.less
	lessc pkg/app.less $@
	@echo "✓ $* theme generated: $@"

# Remove the built binary and themes
clean: clean-themes
	rm -f rex

clean-themes:
	@echo "Cleaning up themes..."
	@rm -f $(CSS_FILES) pkg/variables.less
	@echo "✓ Themes cleaned."

.PHONY: all build clean test themes clean-themes test-js

test-js:
	npm test


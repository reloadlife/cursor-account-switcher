BINARY := cursor-switch
INSTALL_DIR := $(HOME)/.local/bin

.PHONY: build install clean

build:
	go build -o $(BINARY) .

install: build
	@mkdir -p $(INSTALL_DIR)
	@install -m 755 $(BINARY) $(INSTALL_DIR)/$(BINARY)
	@echo "installed $(INSTALL_DIR)/$(BINARY)"

clean:
	rm -f $(BINARY)

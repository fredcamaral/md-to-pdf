PLUGIN_DIR = ../../plugins

.PHONY: all clean toc mermaid

all: toc mermaid

# Create plugins directory if it doesn't exist
$(PLUGIN_DIR):
	mkdir -p $(PLUGIN_DIR)

toc: $(PLUGIN_DIR)
	cd toc && go build -buildmode=plugin -o $(PLUGIN_DIR)/toc.so toc.go

mermaid: $(PLUGIN_DIR)
	cd mermaid && go build -buildmode=plugin -o $(PLUGIN_DIR)/mermaid.so mermaid.go

clean:
	rm -f $(PLUGIN_DIR)/*.so
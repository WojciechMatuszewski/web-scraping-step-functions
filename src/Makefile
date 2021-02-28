FUNCTIONS_DIR = ./functions

build: $(FUNCTIONS_DIR)/*
			for file in $^; do \
				echo $${file}; \
				cd ${PWD}/$${file} && GOARCH=amd64 GOOS=linux go build -o ../../dist/$${file}/main .; \
			done

# Provide your own values here!
queue:
	HTTP post API_URL url="WEBSITE_URL"

.PHONY: update build build-linux

NAME := CalculationResultGovernanceContest

# build-linux: 
# 	@echo Building for linux...
# 	@env GOOS=linux GOARCH=amd64 go build -o ./bin/$(NAME) -v
# 	@echo Done file in "bin" folder.
#don't have in binding linux lib

build:
	@echo Building for MacOs...
	@env go build -o ./bin/$(NAME) -v
	@echo Done file in "bin" folder.

CMD=license-checker

all: build run

build:
	go build -o ${CMD}

run:
	./${CMD}

# Compiler settings
CC = gcc
CFLAGS =-std=gnu23 -Wall -Wextra -g -I${HOMEBREW_PREFIX}/include/
LDFLAGS=-L${HOMEBREW_PREFIX}/lib -lrdkafka

# Project name
TARGET = jr_kafka

# Source files
SRC = jr_kafka.c

# Object files
OBJ = $(SRC:.c=.o)

# Default target
all: compile

# Compilation
compile:
	$(CC) $(CFLAGS) $(LDFLAGS) $(SRC) -o $(TARGET)

# Clean build files
clean:
	rm -f $(TARGET) $(OBJ)

# Create kafka config directory
config:
	mkdir -p kafka

# Install dependencies (for Ubuntu/Debian)
deps:
	sudo apt-get update
	sudo apt-get install -y librdkafka-dev

# For macOS users (using Homebrew)
deps-mac:
	brew install librdkafka

.PHONY: all clean config deps deps-mac

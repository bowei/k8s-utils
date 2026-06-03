CC      = gcc
CFLAGS  = -O2 -Wall -Wextra -std=c11

.PHONY: all run clean

all: swap_test

swap_test: swap_test.c
	$(CC) $(CFLAGS) -o $@ $<

run: all
	bash run.sh

clean:
	rm -f swap_test

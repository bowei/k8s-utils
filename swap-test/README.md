# swap test

A simple program to implement a test of swap behavior on Linux:

Test program

- Allocate 1 Gb of data. Make sure to touch all of the pages with random values
  to force the page to be marked as used.
- Call madvise pageout to swap the data out. Measure the precise time this takes.
- Call madvise to page the data in. Measure the precise time this takes.

Tracing harness

Measure relevant I/O statistics about the swap such as page / sec, IOPs, bandwidth.

Run script

Run script should start the tracing harness, run the test program and report the results.

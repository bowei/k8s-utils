# Godoc Multi-Column

## Design doc

The main design doc is here in doc/design.md. Please refer to this to understand
the overall design of the tool.

Important files:

* Makefile
* main.go -- entry point for the executable.
* pkg/ -- code for generating the site
 
## Example execution

```bash
./rex                      \
  -type k8s.io/api/core/v1.Pod \
  -output /tmp/core.html       \
  $(find . -type d \( -name 'v1' -o -name 'v1beta*' -o -name 'v1alpha*' \) )
```

## Project practices

* Make sure you run unit tests to check that your code works. `make test`.

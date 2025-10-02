# Resource Explorer

Generates documentation for Go plain-old-data style APIs (e.g. Kubernetes) in a
multi-column format. An example of the output can be [found here][example].

![screenshot](screenshot.png "Screenshot")

## Building

```bash
$ make
```

## Usage

```bash
$ rex -output=OUTPUT_FILE -type=STARTING_TYPE [package dirs...]
```

Run the documentation generator, writing to `OUTPUT_FILE`. `STARTING_TYPE` will
be the first type shown on the page. `package dirs` is a list of packages to
scan for APIs.

You may need to `go get` missing package sources to get a complete picture of
all of the types used.


### Examples

```bash
# Generate documentation for the resources in core/v1 and networking/v1.
$ ./rex -output=/tmp/index.html -type k8s.io/api/core/v1.Pod \
    ~/work/api/core/v1 ~/work/api/networking/v1
```

```bash
# Generate documentation for v?, v?alpha*, v?beta* packages under the apis/ directory.
$ ./rex -output=/tmp/index.html -type k8s.io/api/core/v1.Pod \
    $(find apis/ -type d \( -name 'v?' -o -name 'v?beta*' -o -name 'v?alpha*' \) )
```

[example]: https://bowei.github.io/k8s/core.html#k8s.io/api/core/v1.Pod/Spec/Containers/WorkingDir

# Snifferbeat

Transport and process your Iot device logs.

More about Beats, See also [Beats Platform Reference](https://www.elastic.co/guide/en/beats/libbeat/current/index.html)

## Quick Start


```yml

snifferbeat:
  # Defines how often an event is sent to the output
  period: 1s

  serial:
    # Port, a number or a device name
    name: /dev/ttyUSB1

    # Set baud rate, default=115200
    baud: 952


# Name
name: "Store-1"
fields_under_root: true
fields:
  # Addr and other mark
  mark: "Store No.0. descript"

  # Geopoint of this device
  location: 
    lat: -71.34
    lon: 41.12

  # ... more Custom fields
```


### Q&A

```shell
$ ./snifferbeat
bash: ./snifferbeat: Permission denied
```

```
$ chmod -R snifferbeat
```

```shell
$ sudo ./snifferbeat
snifferbeat2017/05/02 01:50:49.440981 beat.go:339: CRIT Exiting: error loading config file: config file ("snifferbeat.yml") must be owned by the beat user (uid=0) or root
Exiting: error loading config file: config file ("snifferbeat.yml") must be owned by the beat user (uid=0) or root
```

```shell
$ chmod go-w snifferbeat.yml
```

## Getting Started with Snifferbeat

Ensure that this folder is at the following location:
`${GOPATH}/github.com/gitaiqaq/snifferbeat`

### Requirements

* [Golang](https://golang.org/dl/) 1.7

### Init Project
To get running with Snifferbeat and also install the
dependencies, run the following command:

```
make setup
```

It will create a clean git history for each major step. Note that you can always rewrite the history if you wish before pushing your changes.

To push Snifferbeat in the git repository, run the following commands:

```
git remote set-url origin https://github.com/gitaiqaq/snifferbeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Snifferbeat run the command below. This will generate a binary
in the same directory with the name snifferbeat.

```
make
```


### Run

To run Snifferbeat with debugging output enabled, run:

```
./snifferbeat -c snifferbeat.yml -e -d "*"
```


### Test

To test Snifferbeat, run the following command:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `etc/fields.yml`.
To generate etc/snifferbeat.template.json and etc/snifferbeat.asciidoc

```
make update
```


### Cleanup

To clean  Snifferbeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Snifferbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/github.com/gitaiqaq/snifferbeat
cd ${GOPATH}/github.com/gitaiqaq/snifferbeat
git clone https://github.com/gitaiqaq/snifferbeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).


## Packaging

The beat frameworks provides tools to crosscompile and package your beat for different platforms. This requires [docker](https://www.docker.com/) and vendoring as described above. To build packages of your beat, run the following command:

```
make package
```

This will fetch and create all images required for the build process. The hole process to finish can take several minutes.

# EdgeTX CloudBuild

EdgeTX CloudBuild is an open-source EdgeTX firmware build service

## CLI tool

It is possible to use this tool from command line as well.

### Prerequisites

Unix based operating system, git & podman installed.

Install git: https://git-scm.com/download/linux
Install podman: https://podman.io/getting-started/installation.html

### Setup



### Example

```
go run cmd/cli/main.go -commit 55b3f91d0cf1d0130371343aef458bee1bfccbdf -build-flags ./tx16s-internal-elrs.json
```

Where `./tx16s-internal-elhrs.json` is in this format:

```
[
    {
        "key": "DISABLE_COMPANION",
        "value": "YES"
    },
    {
        "key": "CMAKE_BUILD_TYPE",
        "value": "Release"
    },
    {
        "key": "TRACE_SIMPGMSPACE",
        "value": "NO"
    },
    {
        "key": "VERBOSE_CMAKELISTS",
        "value": "YES"
    },
    {
        "key": "CMAKE_RULE_MESSAGES",
        "value": "OFF"
    },
    {
        "key": "PCB",
        "value": "X10"
    },
    {
        "key": "PCBREV",
        "value": "T16"
    },
    {
        "key": "INTERNAL_MODULE_MULTI",
        "value": "ON"
    }
]
```

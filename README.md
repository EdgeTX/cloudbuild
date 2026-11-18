# EdgeTX CloudBuild

EdgeTX CloudBuild is an open-source EdgeTX firmware build service.

It is not meant to be run directly connected to the internet, but instead be behind a
proxy that will only allow the following endpoint prefixes toward the public interface:
- `/api/*`: public and authenticated endpoints.

If you are using the local file system storage (not recommended in production), you will need
to allow the download URL as well:
- `/firmwares/*`

The administrative UI is accessible from the root endpoint `/`. The static content endpoints are
**not** authenticated.

## Public API

The public API is documented [here](doc/PublicAPI.md).


## Build and run locally

It is possible to use this tool from command line as well.

### Prerequisites

Linux operating system with:
- `docker` (including `docker-compose` plugin).
- `git` (required only to checkout this repository).

### Setup

First clone this repository:

```shell
git clone https://github.com/EdgeTX/cloudbuild.git
```

Then you will need to configure some things:

```shell
# Create config from template
cp api.env.example api.env
```

For a quick test, all you need to edit is external URL (as it will be seen by web client)
and the directory in which the firmwares are stored (which should be shared among all 
containers as a volume).

```env
EBUILD_STORAGE_PATH=/home/rootless/src/static/firmwares
EBUILD_DOWNLOAD_URL=http://localhost:3000/firmwares
```

### Build container image

The runtime container image can be built with:

```shell
docker compose build api
```

### Run the containers

#### On the very first start, or after a database schema change

Start the database first:

```shell
docker compose up -d db
```

Then start the API:

```shell
docker compose up -d api
```

And setup the database schema with:

```shell
docker exec -it cloudbuild-api-1 ./ebuild db migrate
```

Once the database is setup, proceed to the next step to start the workers as well.

#### After the first time

Then the rest of the stack can be run all together:

```shell
docker compose up -d --scale worker=2
```

You can increase the number of workers if you want to build more firmwares in parallel.


## Using S3 compatible storage

To offload serving the firmware files to a S3 compatible storage, a few configuration
parameters need to be set additionally in the environment file (`api.env`):

```env
# Common settings for all provider
EBUILD_STORAGE_TYPE: S3
EBUILD_S3_ACCESS_KEY: myAccessKey
EBUILD_S3_SECRET_KEY: super-secret-key

# These settings depend on the provider:
EBUILD_S3_URL: https://s3.super-provider.com
EBUILD_S3_URL_IMMUTABLE: true

# Public download URL prefix (object key is appended to create download link)
EBUILD_DOWNLOAD_URL: https://bucket.s3.super-provider.com
```

## Generating a token to access the UI

To be able to use the administrative UI, a token must be generated for every user:

```shell
docker exec -it cloudbuild-api-1 ./ebuild auth create [some name]
AccessKey: [some access key]
SecretKey: [very secret key]
```

The token can be later removed:

```shell
docker exec -it cloudbuild-api-1 ./ebuild auth remove [Access Key]
token [Access Key] removed
```

## Using a local Git mirror

To speed up builds, it is possible to use a local mirror by changing
the repository URL:
```env
EBUILD_SRC_REPO=http://local-mirror:8080/EdgeTX/edgetx
```

[cloudbuild-mirror](https://github.com/EdgeTX/cloudbuild-mirror) can be
used as a local mirror. It also supports automatic refreshing of the
targets (see next section).

With this configuration only, the submodules will still be fetched from `http://github.com`,
as the full URL is encoded into the respective `.gitmodules` files. Fortunately, Git offers
a simple to override this with [insteadOf](https://git-scm.com/docs/git-config#Documentation/git-config.txt-urlbaseinsteadOf).

Just mount this file into the workers at `/etc/gitconfig`:
```
[url "http://local-mirror:8080"]
  insteadOf = https://github.com
```

## Refresh targets automatically from URL

It is also possible to fetch `targets.json` from a URL instead of a local
file. In that case, the targets are refreshed automatically (defaults to 5 minutes interval).

```env
# Use cloudbuild mirrored repository as the source for targets.json
EBUILD_TARGETS=http://local-mirror:8080/gitweb.cgi?p=EdgeTX/cloudbuild.git;a=blob_plain;f=targets.json;hb=HEAD
# This is the default refresh interval in seconds
EBUILD_TARGETS_REFRESH_INTERVAL=300
```

## Limiting targets and flag values to releases

Some radios, options and option values only exist in some EdgeTX releases.
In `targets.json` they can be limited with
[semver constraints](https://github.com/Masterminds/semver#checking-version-constraints),
checked against each release's version (`sem_ver` if set, e.g. `nightly`):

| Field | On | Limits |
|---|---|---|
| `version_supported` | a target | the target |
| `version_supported` | a flag | the whole flag, i.e. every value in that definition |
| `values_version_supported` | a flag | individual values, as a map from value to constraint |

```json
"targets": {
  "newradio": { "description": "Example New Radio", "version_supported": ">= 3.0.0-0" }
},
"flags": {
  "power_led_blue": {
    "build_flag": "POWER_LED_BLUE",
    "values": ["YES", "NO"],
    "version_supported": "< 3.0.0-0"
  }
},
"tags": {
  "colorlcd": {
    "flags": {
      "language": {
        "values": ["AR", "CN"],
        "values_version_supported": { "AR": ">= 3.0.0-0" }
      }
    }
  }
}
```

- A flag's `version_supported` and `values_version_supported` can be
  combined: a value is supported only if both allow the release. Keys of
  `values_version_supported` must be listed in `values`.
- Flags are defined globally and per tag. A constraint only applies to the
  definition it's in, so a flag defined in several places is limited in
  each.
- Build requests that use an unsupported target, flag or value for the
  requested release are rejected.
- The API exposes the result per release as `exclude_targets` and
  `exclude_flags`, so clients can hide what a release can't build. A flag
  with all its values excluded is dropped by clients.

**Use `-0` on the version in constraints** (`>= 3.0.0-0`, `< 3.0.0-0`)
when pre-releases should follow the same rule as releases. A constraint
without a pre-release part never matches pre-releases, so if pre-releases
are listed in `releases`:

- `>= 3.0.0` excludes `v3.0.0-rc1`; `>= 3.0.0-0` includes it.
- `< 3.0.0` also excludes `v2.12.6-rc1`; `< 3.0.0-0` includes 2.x
  pre-releases and excludes 3.0 ones.

`-0` is the lowest possible pre-release, so `3.0.0-0` sits just before every
3.0 pre-release. This only matters for releases with a pre-release version:
`nightly` uses `sem_ver` (e.g. `v3.0.0`), which isn't one.

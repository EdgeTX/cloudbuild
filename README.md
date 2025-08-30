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

# EdgeTX CloudBuild

EdgeTX CloudBuild is an open-source EdgeTX firmware build service

## Build and run locally

It is possible to use this tool from command line as well.

### Prerequisites

Linux operating system with:
- `docker` (including `docker-compose` plugin).
- `git` (required only to checkout this repository).

### Setup

First clone this repository:

``` shell
git clone https://github.com/EdgeTX/cloudbuild.git
```

Then you will need to configure some things:

``` shell
# Create config from template
cp api.env.example api.env
```

For a quick test, all you need to edit is external URL (as it will be seen by web client)
and the directory in which the firmwares are stored (which should be shared among all 
containers as a volume).

```
EBUILD_STORAGE_PATH=/home/rootless/src/static/firmwares
EBUILD_DOWNLOAD_URL=http://localhost:3000/firmwares
```

### Build container image

The runtime container image can be built with:

``` shell
docker compose build api
```

### Run the containers

#### On the very first start, or after a database schema change

Start the database first:

``` shell
docker compose up -d db
```

Then start the API:

``` shell
docker compose up -d api
```

And setup the database schema with:

``` shell
docker exec -it cloudbuild-api-1 ./ebuild db migrate
```

Once the database is setup, proceed to the next step to start the workers as well.

#### After the first time

Then the rest of the stack can be run all together:

``` shell
docker compose up -d --scale worker=2
```

You can increase the number of workers if you want to build more firmwares in parallel.


## Using S3 compatible storage

To offload serving the firmware files to a S3 compatible storage, a few configuration
parameters need to be set additionally in the environment file (`api.env`):

```
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

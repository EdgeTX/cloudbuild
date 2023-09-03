# Public Cloudbuild API

## Supported Targets

### **GET** - /api/targets

This endpoint allows for retrieving the supported targets and build options.
It returns the configured `targets.json` (see [here](../targets.json)).

#### CURL

```sh
curl -X GET "https://cloudbuild.edgetx.org/api/targets"
```

## Build Job Requests

### **POST** - /api/jobs

This request allows for **creating** build jobs. In case this build job
(uniquely identified by `[commit hash, target, flags]`) already exists,
its current status is returned. If the build job does not exist, it will
be created.

#### CURL

```sh
cat << EOF > body.json
{
  "release": "v2.9.0",
  "target": "x9e",
  "flags": [
	{
	  "name": "sticks",
	  "value": "HORUS"
	}
  ]
}
EOF

curl -X POST "https://cloudbuild.edgetx.org/api/jobs" \
    --data-raw @body.json
```

#### Body Parameters

- **body** should respect the following schema:

```json
{
  "release": "v2.9.0",
  "target": "x9e",
  "flags": [
	{
	  "name": "sticks",
	  "value": "HORUS"
	}
  ]
}
```

### **POST** - /api/status

This request allows for **fetching the status** of an existing build jobs.
Build jobs are uniquely identified by `[commit hash, target, flags]`.

#### CURL

```sh
cat << EOF > body.json
{
  "release": "v2.9.0",
  "target": "x9e",
  "flags": [
	{
	  "name": "sticks",
	  "value": "HORUS"
	}
  ]
}
EOF

curl -X POST "https://cloudbuild.edgetx.org/api/status" \
    --data-raw "$body"
```

#### Body Parameters

- **body** should respect the following schema:

```json
{
  "release": "v2.9.0",
  "target": "x9e",
  "flags": [
	{
	  "name": "sticks",
	  "value": "HORUS"
	}
  ]
}
```

## Error Format

All requests will return a JSON formated reply on errors.

Example:
``` json
{
  "error": "option flag not supported: ppm_unit=USE"
}
```

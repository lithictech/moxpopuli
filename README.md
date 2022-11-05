[![Status](https://github.com/lithictech/sequel-state-machine/actions/workflows/pr-checks.yml/badge.svg)](https://github.com/lithictech/moxpopuli/actions/workflows/pr-checks.yml)
[![GoDoc](https://godoc.org/github.com/lithictech/moxpopuli?status.svg)](http://godoc.org/github.com/lithictech/moxpopuli)
[![license](http://img.shields.io/badge/license-MIT-orange.svg)](https://raw.githubusercontent.com/lithictech/moxpopuli/main/LICENSE)


# Mox Populi

**See the presentation we gave at AsyncAPI Conf 2022 about Mox Populi:
_[Building AsyncAPI Specs from Real-World API Data](https://github.com/lithictech/moxpopuli/blob/main/docs/asyncapi-conf-2022-notes.pdf)_**

**And once you have an idea what it is, 
_[try Mox Populi from your browser](https://moxpopuli.webhookdb.com/swaggerui/index.html#/default/postV1SchemagenQuickstart)_.**

Mox Populi is a tool to build comprehensive JSON Schemas from real-world events and payloads.
That is, instead of the centralized, top-down approach of using a specification to drive development,
you can generate a specification from actual running services.
Mox Populi represents a democratic (or anarchic) approach to specification generation
and API governance.

In pratical terms, Mox Populi is useful for any project concerned with ingesting and processing data
from external systems where schema are unreliable or unpublished.
It can be used for exploratory purposes, to better understand the data coming into a system,
for building powerful regression testing systems, or as the basis for a hand-built specification.

Examples of where Mox Populi has been useful include:

- Integrating products that do not have easy-to-use staging environments,
  so you cannot easily generate test data or load to send into your systems.
- Products that do not have programmatic ways to generate load or test data (the API is read-only or read-mostly).
- Generating enough load to test a backend in a way that would be impractical otherwise.
- Fixture data for something similar to property-based testing,
  where the Mox Populi-generated schema can be used to fixture data used
  for unit and integration testing.

Mox Populi is currently used by [WebhookDB](https://webhookdb.com)
to power its extensive unit and integration regression testing systems,
and to make AsyncAPI specifications available for supported services.

## Examples

Generate a JSONSchema using each object in a JSON Lines file:

```shell
moxpopuli schemagen --payload-loader=file://./inputs.jsonl
```

Load and save the JSONSchema to a file:

```shell
moxpopuli schemagen \
    --payload-loader=file://./testdata/complexpayloads.jsonl \
    --loader=file://./schema-from-inputs.json \
    --saver=file://./schema-from-inputs.json
```

Generate a JSONSchema for each row in a JSONB column in a Postgres database:

```shell
moxpopuli schemagen \
    --payload-loader=postgres://webhookdb:webhookdb@localhost:18005/webhookdb \
    --payload-loader-arg="SELECT request_body FROM logged_webhooks WHERE truncated_at IS NULL LIMIT 100"
```

Generate a full AsyncAPI spec from some events stored in a file:

```shell
moxpopuli specgen \
    --loader='file://./testdata/whdbspecseed.json' \
    --event-loader='file://./testdata/requests.jsonl'
```

## Limitations

Schema generation is based on a supported subset of JSON Schema.

Schema generation does not support:

- References. Do not include references in your schemas. `moxpopuli` will never write out references.
  Because we cannot be sure about the schemas we see, using references could result in too much
  diffing and confusion.
- Non-object payloads. It's extremely rare for this to be a problem;
  if it is, support can be added.
- YAML. Easy enough to add support but for now we're JSON-only.

Spec generation is limited only to what can be figured out from events,
which is usually channels, subscription operations, and their data
like bindings, headers, and payloads.
Usually you will supply a 'seed' specification
that includes an `info` section and whatever else is needed.

## Mox Populi Server

Running `moxpopuli server` starts a server on `$PORT`, or 22021 by default.
It exposes four endpoints, which roughly correspond to the `schemagen`, `specgen`, and `datagen` CLI commands.

To see the OpenAPI document describing the API, and test it out yourself,
go to https://moxpopuli.webhookdb.com/swaggerui/index.html#/default/postV1SchemagenQuickstart

The OpenAPI specification is built mostly from the actual runtime Go structs
using the [Sashay](https://github.com/rgalanakis/sashay) tool.
It's really neat since it uses actual runtime code that your server uses,
not separate documentation!

You can run `moxpopuli server openapi` to generate the OpenAPI spec,
or use `moxpopuli server swaggerui` to open a browser to the SwaggerUI page against localhost.

## How it Works

The [linked presentation](https://github.com/lithictech/moxpopuli/blob/main/docs/asyncapi-conf-2022-notes.pdf)
covers this at a high level. But here are some lower level and implementation notes.

### Building Schemas from Payloads

The core of `moxpoopuli` is generating JSON Schemas from real-world production payloads,
such as generating a message payload from a request body,
or message headers from request headers.

The gist of it is:

- Schema calculation is always based on an existing schema (which can be empty),
  and a real-world payload.
- `moxpopuli` generates a new schema from the real-world payload.
- `moxpopuli` "merges" the two schemas such that the resulting schema is accurate
  for all seen payloads. For example, if the first payload has a property with
  the value of 1,000, we'd assume it's an `int32`.
  But if the next payload has the same property with a value of 15 billion,
  the new and resulting schema would be an `int64`.
- `moxpopuli` tries to be as descriptive as possible about the payloads it analyzes.
  For example, any values that are not accomodated by a schema get added to `examples`,
  it uses additional data types for values (`uuid`, `email`, etc),
  and uses `x-` extension fields to store information it needs for future analysis
  or generating meaningful sample data.

### Loaders and Savers

`moxpopuli` uses a system of loaders and savers to load information like specifications
or events, and then save them after processing.
It does this via "loaders" and "savers".

Loaders and savers are usually URLs.
Each loader/saver can also take an additional argument,
which can be used to configure the saver/loader.
One example is using a `postgres://` URL to connect to a database,
with an argument that is the SQL query to select or update the required data.

#### Single Objects Load and Save

When `moxpopuli` is working with things like JSONSchema or an AsyncAPI specification,
it loads and saves them as a single JSON object.
This is always done with the `--saver/-s` and `--loader/-l` arguments,
plus `--saver-arg/-sa` and `--loader-arg/-la` to provide saver and loader arguments.

Examples of valid loader options are:

- `moxpopuli schemagen -l file://./myschema.json` would read `myschema.json` from the current directory
  and write it to STDOUT.
- `printf '{"x":1}' | moxpopuli schemagen -l -` would read STDIN as a schema.
- `moxpopuli schemagen -l=_ -la='{"x":1}'` would parse `{"x":1}` as JSON as the schema.
- `moxpopuli schemagen -l=postgres://u:p@localhost:5432/myapp -la='SELECT schema FROM asyncapischemas WHERE id=1'`
  would parse the selected row/column as a schema.
  Note that for single objects, only one column must be returned.
- If no loader is given, or '.' is used, do not load anything.

Examples of valid saver options are:

- `moxpopuli schemagen -s=-` would write to stdout.
- `moxpopuli schemagen -s=file://./myschema.json` would save to `myschema.json`.
- `moxpopuli schemagen -s=postgres://u:p@localhost:5432/myapp -sa='UPDATE asyncapischemas SET schema=$1 WHERE id=1`
  would run that query with the updated schema as the argument.
  Note that for single objects, there should be only one positional argument.

#### Iterator Loaders

When generating a schema or specification, `moxpopuli` runs over many events/payloads.
In these cases, we use the same loader system, but expect the data to be a collection.

Examples of iterator loaders would be:

- `-pl=file://./requests.jsonl` would treat each line in the JSONLines file as a separate object.
- `-pl=file://./requests.json` would expect the file to be a JSON array.
- `-pl=_ -pla='{"x":1}\n{"x":2}'` would expect each line in the loader argument to be a JSON object, like JSONLines.
- `-pl=-` would expect each line from STDIN to be a JSON object, like JSONLines.
- `-pl=postgres://u:p@localhost:5432/myapp -pla='SELECT body FROM requests WHERE service=stripe LIMIT 10'`
  would use select rows from Postgres.

There are two types of iterator loaders: **Payloads** and **Events**.
The only difference is that:

- Payloads are freeform. The entire JSON object is the payload. If using a `postgres` loader,
  the query must return a single column, which is parsed as JSON to get the payload.
- Event loads expect a certain set of keys, based on the binding.
  If using a `postgres` loader, the select/column names should match the keys;
  if loading JSON directly through other loaders, the loaded JSON should match the keys.
- Event loader keys are:
  - `http` binding: `path` (string), `method` (string), `headers` ({string:string} map), `body` ({string:any} map)

## Development

Check out the Makefile,
it has everything you will need.

If it doesn't, we should add to the Makefile 😬

## Supported by

<p>
  <a href="https://webhookdb.com/">
    <img src="https://webhookdb.com/content/brand/blueonblack.png" />
  </a>
</p>

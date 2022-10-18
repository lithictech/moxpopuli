This is the supporting code for the Mox Populi presentation at AsyncAPI Conf 2022.

Internal note: This was rendered for the slides by using deck-deck-go with
a carbon terminal, dracula theme, and `#08090a` background so element screenshots
have dark corners.


## Generating Payload Schemas

Here's an example payload you may be dealing with:

```json
{
"currentUserHasAccessToDetails": true,
"DeletedDate": null,
"end_date": "06/10/2022 12:30",
"start_date": "06/10/2022 11:30",
"LastKeptAppointmentDateTimeMDY": "05/02/2022 10:00 AM",
"Age": "15",
"NextAppointmentDateTimeMDY": "-",
"status": {
  "AbsenceReason": null,
  "Status": "Upcoming"
},
"statusBadge": "",
"statusJson": "{\"Status\":\"Upcoming\",\"AbsenceReason\":null}",
"teletherapySessionId": null,
"teletherapyTitle": null,
"type": 1
}
```

Go code:

```go
func Sniff(t jsontype.JsonType, value interface{}) JsonFormat {
	if t == jsontype.T_STRING {
		s := value.(string)
		if sniffEmail(s) {
			return F_EMAIL
		} else if sniffUrl(s) {
			return F_URI
		} else if sniffIPv4(s) {
			return F_IPV4
		} else if sniffIPv6(s) {
			return F_IPV6
		} else if sniffCountry(s) {
			return F_COUNTRY
		} else if sniffCurrency(s) {
			return F_CURRENCY
		} else if sniffUuid4(s) {
			return F_UUID4
		} else if sniffNumericalString(s) {
			return F_NUMERICAL
		} else if sniffDateTimeNoTZ(s) {
			return F_DATETIME_NOTZ
		} else if sniffDateTimeTZ(s) {
			// MUST go after 'noTz' search
			return F_DATETIME
		} else if sniffDate(s) {
			return F_DATE
		} else if sniffTime(s) {
			return F_TIME
		} else if sniffDuration(s) {
			return F_DURATION
		} else if sniffBinary(s) {
			return F_BINARY
		} else if sniffBase64(s) {
			return F_BYTE
		}
		return F_NOFORMAT
	} else if t == jsontype.T_INTEGER {
		// We can fit any integer into a float, and we don't need the *actual* value,
		// so use floats for sniff functions that also need floats.
		i := value.(int)
		f := float64(i)
		if sniffTimestamp(f) {
			return F_TIMESTAMP
		} else if sniffTimestampMS(f) {
			return F_TIMESTAMP_MS
		} else if sniffInt32(i) {
			return F_INT32
		}
		return F_INT64
	} else if t == jsontype.T_NUMBER {
		f := value.(float64)
		if sniffTimestamp(f) {
			return F_TIMESTAMP
		} else if sniffTimestampMS(f) {
			return F_TIMESTAMP_MS
		} else if sniffFloat32(f) {
			return F_FLOAT
		}
		return F_DOUBLE
	}
	return F_NOFORMAT
}
```

Simple payload:

```json5
{
  "x": 0,
  "z": 1664818615
}
```

Naive JSONSchema:

```json5
{
  "x": {
    "type": "integer"
  },
  "z": {
    "type": "integer"
  }
}
```

Described via Mox Pop:

```json5
// moxpopuli schemagen -p=_ -pa='{"x":0,"z":1664818615}'
{
  "properties": {
    "x": {
      "format": "int32",
      "type": "integer",
      "x-seenMaximum": 0,
      "x-seenMinimum": 0
    },
    "z": {
      "format": "timestamp",
      "type": "integer",
      "x-seenMaximum": 1664818615,
      "x-seenMinimum": 1664818615
    }
  },
  "type": "object",
  "x-samples": 1
}
```

```json5
// moxpopuli schemagen -p=file://./testdata/simplepayloads.jsonl
{
  "properties": {
    "x": {
      "format": "float",
      "type": "number",
      "x-samples": 4,
      "x-seenMaximum": 1243340323232320.5,
      "x-seenMinimum": 0
    },
    "z": {
      "oneOf": [
        {
          "format": "timestamp",
          "type": "integer",
          "x-samples": 3,
          "x-seenMaximum": 1664818617,
          "x-seenMinimum": 1664818615
        },
        {
          "type": "string",
          "x-samples": 1,
          "x-seenMaxLength": 1,
          "x-seenMinLength": 1
        }
      ]
    }
  }, "type": "object", "x-samples": 4}
```

How we regain precision:

```json5
// moxpopuli schemagen -p=_ -pa='{"x": "abc123"}'
{
  "properties": {
    "x": {
      "type": "string",
      "x-seenMaxLength": 6,
      "x-seenMinLength": 6,
      "x-seenStrings": ["abc123"]
    }
  }, "type": "object", "x-samples": 1}
```

Regain precision with identifiers

```json5
// moxpopuli schemagen -p=file://./testdata/ids.jsonl
{
  "properties": {
    "x": {
      "type": "string",
      "x-identifier": true,
      "x-samples": 6,
      "x-seenMaxLength": 28,
      "x-seenMinLength": 28,
      "x-sensitive": true
    }
  }, "type": "object", "x-samples": 6}
```

Regain precision with enums

```json5
// moxpopuli schemagen -p=file://./testdata/enums.jsonl
{
  "properties": {
    "x": {
      "enum": [
        "VALUE1",
        "VALUE2",
        "VALUE3",
        "VALUE4"
      ],
      "type": "string",
      "x-samples": 34,
      "x-seenMaxLength": 6,
      "x-seenMinLength": 6
    }
  }, "type": "object", "x-samples": 34}
```

How sensitive strings work:

```json5
// moxpopuli schemagen -p=_ -pa='{"x":"94flkfw03qpf89"}'
{
  "properties": {
    "x": {
      "type": "string",
      "x-seenMaxLength": 14,
      "x-seenMinLength": 14,
      "x-seenStrings": [
        "LlLWDYu9TqVohj"
      ],
      "x-sensitive": true
    }
  }, "type": "object", "x-samples": 1}
```

```
$ moxpopuli schemagen -p=_ -pa='{"x":"94flkfw03qpf89"}' | grep 'x-seenStrings' -A2
      "x-seenStrings": [
        "cTdEARltNsyMPR"
      ],
$ moxpopuli schemagen -p=_ -pa='{"x":"94flkfw03qpf89"}' | grep 'x-seenStrings' -A2
      "x-seenStrings": [
        "t0CXQuMftuQaCl"
      ],
$ MOXPOPULI_SALT=abcd moxpopuli schemagen -p=_ -pa='{"x":"94flkfw03qpf89"}' | grep 'x-seenStrings' -A2
      "x-seenStrings": [
        "t7b0a9UtZWgadr"
      ],
$ MOXPOPULI_SALT=abcd moxpopuli schemagen -p=_ -pa='{"x":"94flkfw03qpf89"}' | grep 'x-seenStrings' -A2
      "x-seenStrings": [
        "t7b0a9UtZWgadr"
      ],
```

```json5
// printf '{"id":"3bfa"}\n{"id":5}' | moxpopuli schemagen -p=- --examples=1
{
  "examples": [
    {"id": "3bfa"},
    {"id": 5}],
  "properties": {
    "id": {
      "oneOf": [
        {
          "type": "string",
          "x-seenMaxLength": 4,
          "x-seenMinLength": 4,
          "x-seenStrings": ["3bfa"]
        },
        {
          "format": "int32",
          "type": "integer",
          "x-samples": 1,
          "x-seenMaximum": 5,
          "x-seenMinimum": 5
        }
      ]
    }
  }, "type": "object", "x-samples": 2}
```

Full example

```json5
// moxpopuli fixturegen --lines --count=10 | moxpopuli schemagen -p=-
{
  "properties": {
    "date-time": {
      "format": "date-time",
      "type": "string",
      "x-seenMaximum": "2032-02-01T10:04:39-08:00",
      "x-seenMinimum": "2018-02-23T21:20:42-08:00"
    },
    "date-time-notz": {
      "format": "date-time-notz",
      "type": "string",
      "x-seenMaximum": "2032-05-18T21:21:30",
      "x-seenMinimum": "2013-06-25T18:49:31"
    },
    "duration": {
      "format": "duration",
      "type": "string",
      "x-seenMaximum": "P4Y8M20DT17H13M45S",
      "x-seenMinimum": "P2M12DT16H7M24S"
    },
    "iso-country": {
      "format": "iso-country",
      "maxLength": 3,
      "minLength": 3,
      "type": "string",
    },
    "iso-currency": {
      "format": "iso-currency",
      "maxLength": 2,
      "minLength": 2,
      "type": "string",
    },
    "numerical": {
      "format": "numerical",
      "type": "string",
      "x-seenMaximum": "762129038907",
      "x-seenMinimum": "-935736330713"
    },
    "timestamp": {
      "format": "timestamp",
      "type": "integer",
      "x-seenMaximum": 1925840647,
      "x-seenMinimum": 110039324
    },
    "timestamp-ms": {
      "format": "timestamp-ms",
      "type": "integer",
      "x-seenMaximum": 1847512946466,
      "x-seenMinimum": 95515572188
    },
    "uuid4": {
      "format": "uuid4",
      "maxLength": 36,
      "minLength": 36,
      "type": "string",
      "x-sensitive": true
    },
    "zero-one": {
      "enum": [0, 1],
      "format": "zero-one",
      "type": "integer",
      "x-seenMaximum": 1,
      "x-seenMinimum": 0
    }
  },
  "type": "object"
}
```

Protocol headers:

```json
{
  "message": {
    "bindings": {
      "http": {
        "Accept-Encoding": {
          "type": "string",
          "x-lastValue": "gzip;q=1.0,deflate;q=0.6,identity;q=0.3"
        },
        "Host": {
          "type": "string",
          "x-lastValue": "localhost:18001"
        },
        "User-Agent": {
          "type": "string",
          "x-lastValue": "WebhookDB/v1 webhookdb.com 2022-10-01T00:00:00Z"
        },
        "Version": {
          "type": "string",
          "x-lastValue": "HTTP/1.1"
        }}}}}
```

Request headers:

```json5
{
  "channels": {
    "/v1/service_integrations/svi_81f5em7skqagk7pstse7b4j1r": {
      "subscribe": {
        message: {
          "headers": {
            "properties": {
              "Whdb-Secret": {
                "type": "string",
                "x-identifier": true,
                "x-samples": 7,
                "x-seenMaxLength": 25,
                "x-seenMinLength": 25,
                "x-sensitive": true
              }}}}}}}}
```

Example spec:

```json
{
  "asyncapi": "2.4.0",
  "info": {
    "contact": {
      "email": "hello@webhookdb.com",
      "name": "Hello"
    },
    "description": "These are the WebhookDB endpoints available for a demo org.",
    "termsOfService": "https://webhookdb.com/terms/",
    "title": "WebhookDB Integrations for Demo Org",
    "version": "1.0.0"
  },
  "servers": {
    "localhost:18001": {
      "protocol": "http",
      "protocolVersion": "1.1",
      "url": "localhost:18001"
    }
  },
  "channels": {
    "/v1/service_integrations/svi_81f5em7skqagk7pstse7b4j1r": {
      "subscribe": {
        "bindings": {
          "http": {
            "method": "POST",
            "type": "request"
          }
        },
        "message": {
          "bindings": {
            "http": {
              "headers": {
                "properties": {
                  "Accept": {
                    "type": "string",
                    "x-lastValue": "*/*"
                  }
                },
                "type": "object"
              }
            }
          },
          "contentType": "application/json",
          "correlationId": {
            "location": "$message.header#/Trace-Id"
          },
          "headers": {
            "properties": {
              "Whdb-Secret": {
                "type": "string",
                "x-samples": 8,
                "x-seenMaxLength": 25,
                "x-seenMinLength": 24,
                "x-sensitive": true
              }
            },
            "type": "object",
            "x-samples": 8
          },
          "payload": {
            "properties": {
              "created_at": {
                "format": "date-time",
                "type": "string",
                "x-samples": 8,
                "x-seenMaximum": "2022-09-15T03:14:39.885+00:00",
                "x-seenMinimum": "2022-09-15T03:14:17.222+00:00"
              }
            },
            "type": "object",
            "x-samples": 8
          }
        }
      }
    }
  }
}
```

Datagen:

```json5
// moxpopuli datagen --l=file://./testdata/fixturedemo.schema.json
{
  "SessoinIP": "30.34.254.115",
  "array-of-ids": [
    "14f35dd7-ddbd-5238-dc14-1fb7ec03b5bc",
    "b4fa9ab3-4c59-629c-9019-f87272617e90"
  ],
  "arrayofobjects": [{"myid": "7f968714-d8af-55fd-9e44-c9e710bb7c8f"}],
  "base64bytes": "d3c06f1859fd1a9e82d5c8",
  "currency": "XBB",
  "databaseid": "937",
  "email": "xiMWdMU@luaDPwc.com",
  "ended_at": "2024-08-08T02:00:00-07:00",
  "homepage": "https://PQdXgBt.info/xuZlXoA",
  "started_on": "2021-09-30"
}
```

Vox:

```
// moxpopuli vox -l=file://./testdata/whdbspec.json --count=100 --print

REQUEST /v1/service_integrations/svi_ct14kxb4ngg3auyrysjwzjlk5-0
POST /v1/service_integrations/svi_ct14kxb4ngg3auyrysjwzjlk5 HTTP/1.1
Host: localhost:18001
User-Agent: my awesome app
Content-Length: 284
Accept: */*
Accept-Encoding: gzip
Connection: close
Content-Type: application/json
Trace-Id: d274eb7f-d4a2-e1f8-17f7-cdd7c97587f7
Version: HTTP/1.1
Whdb-Secret: c13cc9c79c6fb8347c7aba14

{"created_at":"2022-09-14T20:14:39-07:00","email":"oGQtORV@YsWSHXp.ru","id":20,"name":"55e3a8408bc5f8d6e961","note":"","opaque_id":"593b44dd4b4603876d8d71c8d8ea","password_digest":"aaa15e368f2fb1df23dd141e86ebccb7b4ad6a21988a878c6f0266bba8a2","soft_deleted_at":null,"updated_at":null}

RESPONSE /v1/service_integrations/svi_ct14kxb4ngg3auyrysjwzjlk5-0
HTTP/1.0 200 OK
Connection: close
Content-Type: application/json
Vary: Origin

{"o":"k"}
```

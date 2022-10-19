package httpmerge

import (
	"context"
	"fmt"
	"github.com/lithictech/moxpopuli/asyncapispec"
	"github.com/lithictech/moxpopuli/asyncapispecmerge/internal"
	moxinternal "github.com/lithictech/moxpopuli/internal"
	"github.com/lithictech/moxpopuli/schema"
	"github.com/lithictech/moxpopuli/schemamerge"
	"github.com/pkg/errors"
	"net/url"
	"strings"
)

func MergeHttp(ctx context.Context, in internal.MergeInput) error {
	channels := in.Spec.GetOrAddChannels()
	servers := in.Spec.GetOrAddServers()
	for in.EventIterator.Next() {
		event, err := in.EventIterator.Read(ctx)
		if err != nil {
			return errors.Wrap(err, "reading events")
		}
		var hevent HttpEvent
		if hev, ok := event.(HttpEvent); ok {
			hevent = hev
		} else if mapev, ok := event.(map[string]interface{}); ok {
			hevent, err = NewHttpEvent(mapev)
			if err != nil {
				return err
			}
		} else {
			return errors.New("event must be an HttpEvent or map[string]interface{}")
		}
		eventUrl, err := url.ParseRequestURI(hevent.Path)
		if err != nil {
			return errors.Wrap(err, "could not parse event path")
		}
		chanItem := channels.GetOrAddItem(eventUrl.Path)
		subscribe := chanItem.GetOrAddSubscribe()
		bindings := subscribe.GetOrAddBindings()
		httpBinding := bindings.GetOrAddHttp()
		httpBinding["type"] = "request"
		httpBinding["method"] = hevent.Method
		q := httpBinding.GetOrAddOrTypeQuery()
		queryMergeResult, err := schemamerge.MergeOne(ctx, schemamerge.MergeOneInput{Schema: q, Payload: moxinternal.UrlValuesToMap(eventUrl.Query())})
		if err != nil {
			return errors.Wrap(err, "merging query")
		}
		if len(queryMergeResult.Schema.MustObject().Properties()) > 0 {
			httpBinding["query"] = queryMergeResult.Schema
		} else {
			delete(httpBinding, "query")
		}
		message := subscribe.GetOrAddMessage()
		if err := mergeHttpMessage(ctx, message, hevent); err != nil {
			return err
		}
		if host, ok := hevent.CanonicalHeaders["host"]; ok {
			srv := servers.GetOrAddServer(host)
			srv["url"] = host
			srv["protocol"] = "http"
			if version, ok := hevent.CanonicalHeaders["version"]; ok {
				srv["protocolVersion"] = moxinternal.LastString(strings.Split(version, "/"))
			}
		}
	}
	return nil
}

func mergeHttpMessage(ctx context.Context, message asyncapispec.Message, event HttpEvent) error {
	appHeaders := make(map[string]interface{}, 8)
	protoHeaders := make(map[string]interface{}, 8)
	for headerName, headervalue := range event.Headers {
		canonicalHeader := internal.CanonicalHeader(headerName)
		if _, ok := ignoreHeaders[canonicalHeader]; ok {
			continue
		} else if canonicalHeader == "content-type" {
			message["contentType"] = headervalue
		} else if isCorrellationId(canonicalHeader) {
			message["correlationId"] = map[string]interface{}{
				"location": fmt.Sprintf("$message.header#/%s", headerName),
			}
		} else if _, ok := protocolHeaders[canonicalHeader]; ok {
			protoHeaders[headerName] = headervalue
		} else {
			appHeaders[headerName] = headervalue
		}
	}
	setProtocolHeaders(message.GetOrAddBindings().GetOrAddHttp(), protoHeaders)
	if _, ok := message["contentType"]; !ok {
		message["contentType"] = "application/json"
	}

	headerMergeResult, err := schemamerge.MergeOne(ctx, schemamerge.MergeOneInput{Schema: message.GetOrAddHeaders(), Payload: appHeaders})
	if err != nil {
		return errors.Wrap(err, "merging message headers")
	}
	message["headers"] = headerMergeResult.Schema

	payloadMergeResult, err := schemamerge.MergeOne(ctx, schemamerge.MergeOneInput{Schema: message.GetOrAddPayload(), Payload: event.Body})
	if err != nil {
		return errors.Wrap(err, "merging payload headers")
	}
	message["payload"] = payloadMergeResult.Schema
	return nil
}

func setProtocolHeaders(http asyncapispec.HttpMessageBinding, headers map[string]interface{}) {
	headerProps := http.GetOrAddHeaders().Properties()
	for k, v := range headers {
		headerProps[k] = schema.Schema{
			schema.P_TYPE:        "string",
			schema.PX_LAST_VALUE: v,
		}
	}
}

type HttpEvent struct {
	Path             string                 `json:"path" description:"Path of the HTTP request."`
	Method           string                 `json:"method" description:"HTTP method, like 'GET' or 'POST'."`
	Headers          map[string]string      `json:"headers" description:"All headers for the HTTP request."`
	CanonicalHeaders map[string]string      `json:"-"`
	Body             map[string]interface{} `json:"body" description:"HTTP body. Only objects are supported for now."`
}

func (h *HttpEvent) CanonizeHeaders() {
	h.CanonicalHeaders = make(map[string]string, len(h.Headers))
	for k, v := range h.Headers {
		h.CanonicalHeaders[internal.CanonicalHeader(k)] = v
	}
}

func NewHttpEvent(e map[string]interface{}) (HttpEvent, error) {
	h := HttpEvent{}
	if v, ok := e["path"]; ok {
		if vt, ok := v.(string); ok {
			h.Path = vt
		} else {
			return h, errors.New("event path must be a string")
		}
	} else {
		return h, errors.New("event requires 'path' key")
	}
	if v, ok := e["method"]; ok {
		if vt, ok := v.(string); ok {
			h.Method = vt
		} else {
			return h, errors.New("event method must be a string")
		}
	} else {
		return h, errors.New("event requires 'method' key")
	}
	if v, ok := e["headers"]; ok {
		if vt, ok := v.(map[string]interface{}); !ok {
			return h, errors.New("event headers must be a map[string]interface{}")
		} else {
			h.Headers = make(map[string]string, len(vt))
			h.CanonicalHeaders = make(map[string]string, len(vt))
			for k, v := range vt {
				hs := fmt.Sprintf("%v", v)
				h.Headers[k] = hs
				h.CanonicalHeaders[internal.CanonicalHeader(k)] = hs
			}
		}
	} else {
		return h, errors.New("event requires 'headers' key")
	}
	if v, ok := e["body"]; ok {
		if vt, ok := v.(map[string]interface{}); !ok {
			return h, errors.New("event body must be a map[string]interface{}")
		} else {
			h.Body = vt
		}
	} else {
		return h, errors.New("event requires 'body' key")
	}
	return h, nil
}

func isCorrellationId(s string) bool {
	return strings.Contains(s, "requestid") ||
		strings.Contains(s, "request-id") ||
		strings.Contains(s, "traceid") ||
		strings.Contains(s, "trace-id") ||
		strings.Contains(s, "correlationid") ||
		strings.Contains(s, "correlation-id")
}

var protocolHeaders = internal.LinesToHeaderNames(`A-IM
Accept
Accept-Charset
Accept-Datetime
Accept-Encoding
Accept-Language
Access-Control-Request-Method
Access-Control-Request-Headers
Authorization
Cache-Control
Connection
Permanent
Content-Encoding
Content-Length
Content-MD5
Content-Type
Cookie
Date
Expect
ForwardedPermanent
From
Host
HTTP2-Settings
If-Match
If-Modified-Since
If-None-Match
If-Range
If-Unmodified-Since
Max-Forwards
Origin
Pragma
Prefer
Proxy-Authorization
Range
Referer
TE
Trailer
Transfer-Encoding
User-Agent
Upgrade
Version
Via
Warning`)

var ignoreHeaders = internal.LinesToHeaderNames(`Upgrade-Insecure-Requests
X-Requested-With
DNT
X-Forwarded-For
X-Forwarded-Host
X-Forwarded-Proto
X-Http-Method-Override
Proxy-Connection
X-UIDH
X-Csrf-Token
Save-Data`)

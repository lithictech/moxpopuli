package moxvox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/lithictech/moxpopuli/asyncapispec"
	"github.com/lithictech/moxpopuli/datagen"
	"github.com/lithictech/moxpopuli/faker"
	"github.com/lithictech/moxpopuli/fp"
	"github.com/lithictech/moxpopuli/schema"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"sync"
)

type Vox func(context.Context, VoxInput) error

type VoxInput struct {
	Spec           asyncapispec.Specification
	Count          int
	ChannelMatcher *regexp.Regexp
	Printer        io.Writer
}

func collectEventFixturesForApiSpec(_ context.Context, in VoxInput, binding string) []EventFixture {
	servers := fp.Values(in.Spec.GetOrAddServers())
	channels := in.Spec.GetOrAddChannels()
	eventSpecs := make([]EventFixture, 0, len(channels)*in.Count)
	for chanName := range channels {
		if !in.ChannelMatcher.MatchString(chanName) {
			continue
		}
		channel := channels.GetOrAddItem(chanName)
		subscribe := channel.GetOrAddSubscribe()
		if _, ok := subscribe.GetOrAddBindings()["http"]; !ok {
			// No http bindings for this subscription
			continue
		}
		opBinding := subscribe.GetOrAddBindings().GetOrAdd(binding)
		msg := subscribe.GetOrAddMessage()
		// This is gross but we need to 'prime' this here so we don't hit a race condition later,
		// since this mutates the receiver in place.
		headerSchema := msg.GetOrAddHeaders()
		payloadSchema := msg.GetOrAddPayload()
		for i := 0; i < in.Count; i++ {
			eventSpecs = append(eventSpecs, EventFixture{
				Id:               fmt.Sprintf("%s-%d", chanName, i),
				Server:           fp.Sample(servers).(map[string]interface{}),
				ChannelName:      chanName,
				Channel:          channel,
				Operation:        subscribe,
				OperationBinding: opBinding,
				Message:          msg,
				Headers:          datagen.Generate("", headerSchema).(map[string]interface{}),
				Payload:          datagen.Generate("", payloadSchema),
			})
		}
	}
	return eventSpecs
}

func playEvents(ctx context.Context, events []EventFixture, playEvent func(context.Context, EventFixture) error) error {
	// This is the only place concurrency is used so we keep it inline,
	// we should use more sophisticated tools if we need more concurrency
	concurrency := 4
	sem := make(chan struct{}, concurrency)
	errs := make([]error, len(events))
	wg := sync.WaitGroup{}
	wg.Add(len(events))
	for ii, eiter := range events {
		go func(i int, e EventFixture) {
			defer wg.Done()
			sem <- struct{}{}
			errs[i] = playEvent(ctx, e)
			<-sem
		}(ii, eiter)
	}
	wg.Wait()
	errLines := make([]string, 0, 4)
	for _, err := range errs {
		if err != nil {
			errLines = append(errLines, err.Error())
		}
	}
	if len(errLines) > 0 {
		return errors.Errorf("%d errors playing events: %s", len(errLines), strings.Join(errLines, ", "))
	}
	return nil
}

func HttpVox(ctx context.Context, in VoxInput) error {
	eventSpecs := collectEventFixturesForApiSpec(ctx, in, "http")
	for _, e := range eventSpecs {
		// Gross but we need to prime these to avoid mutations during concurrent requests.
		// We replace instances of map[string]interface into map[string]Schema
		e.Message.GetOrAddBindings().GetOrAddHttp().GetOrAddHeaders().Properties()
	}
	mux := sync.Mutex{}
	return playEvents(ctx, eventSpecs, func(ctx context.Context, e EventFixture) error {
		method := asyncapispec.HttpOperationBinding(e.OperationBinding).Method()
		url := fmt.Sprintf("%s://%s%s", e.Server.Protocol(), e.Server.Url(), e.ChannelName)
		body, err := e.MarshalBody()
		if err != nil {
			return err
		}
		req, err := http.NewRequest(method, url, bytes.NewReader(body))
		if err != nil {
			return err
		}
		for k, v := range e.Message.GetOrAddBindings().GetOrAddHttp().GetOrAddHeaders().Properties() {
			req.Header.Set(k, v[schema.PX_LAST_VALUE].(string))
		}
		for k, v := range e.Headers {
			req.Header.Set(k, fmt.Sprintf("%v", v))
		}
		req.Header.Set("Content-Type", e.Message.ContentType())
		if cidheader, ok := e.Message.CorrelationIdHeaderKey(); ok {
			req.Header.Set(cidheader, faker.UUID4())
		}
		var reqDump []byte
		if in.Printer != nil {
			reqDump, _ = httputil.DumpRequestOut(req, true)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrap(err, "making request")
		}
		if in.Printer != nil {
			mux.Lock()
			_, _ = fmt.Fprintf(in.Printer, "REQUEST %s\n%s\n\n", e.Id, reqDump)
			respDump, _ := httputil.DumpResponse(resp, true)
			_, _ = fmt.Fprintf(in.Printer, "RESPONSE %s\n%s\n\n", e.Id, respDump)
			mux.Unlock()
		}
		if resp.StatusCode >= 300 {
			rbod, _ := io.ReadAll(resp.Body)
			return errors.Errorf("Status %d calling %s: %s", resp.StatusCode, e.Server.Url(), string(rbod))
		}
		return nil
	})
}

type EventFixture struct {
	Id               string
	Server           asyncapispec.Server
	ChannelName      string
	Channel          asyncapispec.ChannelItem
	Operation        asyncapispec.Operation
	OperationBinding asyncapispec.OperationBinding
	Message          asyncapispec.Message
	Headers          map[string]interface{}
	Payload          interface{}
}

func (e EventFixture) MarshalBody() ([]byte, error) {
	ct := e.Message.ContentType()
	if ct == "" || strings.Contains(ct, "application/json") {
		return json.Marshal(e.Payload)
	}
	return nil, errors.New("unsupported content type for fixturing: " + e.Message.ContentType())
}

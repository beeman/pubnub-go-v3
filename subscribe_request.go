package pubnub

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/pubnub/go/utils"
)

const subscribePath = "/v2/subscribe/%s/%s/0"

type SubscribeResponse struct {
}

type subscribeOpts struct {
	pubnub *PubNub

	Channels      []string
	ChannelGroups []string

	Heartbeat        int
	Region           string
	Timetoken        int64
	FilterExpression string
	WithPresence     bool
	State            map[string]interface{}
	stringState      string

	ctx Context
}

type subscribeBuilder struct {
	opts      *subscribeOpts
	operation *SubscribeOperation
}

func newSubscribeBuilder(pubnub *PubNub) *subscribeBuilder {
	builder := subscribeBuilder{
		opts: &subscribeOpts{
			pubnub: pubnub,
		},
		operation: &SubscribeOperation{},
	}

	return &builder
}

func (b *subscribeBuilder) Channels(channels []string) *subscribeBuilder {
	b.operation.Channels = channels

	return b
}

func (b *subscribeBuilder) ChannelGroups(groups []string) *subscribeBuilder {
	b.operation.ChannelGroups = groups

	return b
}

func (b *subscribeBuilder) Timetoken(tt int64) *subscribeBuilder {
	b.operation.Timetoken = tt

	return b
}

func (b *subscribeBuilder) FilterExpression(expr string) *subscribeBuilder {
	b.operation.FilterExpression = expr

	return b
}

func (b *subscribeBuilder) WithPresence(pres bool) *subscribeBuilder {
	b.operation.PresenceEnabled = pres

	return b
}

func (b *subscribeBuilder) Execute() {
	b.opts.pubnub.subscriptionManager.adaptSubscribe(b.operation)
}

func (o *subscribeOpts) config() Config {
	return *o.pubnub.Config
}

func (o *subscribeOpts) client() *http.Client {
	return o.pubnub.GetSubscribeClient()
}

func (o *subscribeOpts) context() Context {
	return o.ctx
}

func (o *subscribeOpts) validate() error {
	if o.config().PublishKey == "" {
		return newValidationError(o, StrMissingPubKey)
	}

	if o.config().SubscribeKey == "" {
		return newValidationError(o, StrMissingSubKey)
	}

	if len(o.Channels) == 0 && len(o.ChannelGroups) == 0 {
		return newValidationError(o, StrMissingChannel)
	}

	if o.State != nil {
		state, err := json.Marshal(o.State)
		if err != nil {
			return newValidationError(o, err.Error())
		}
		o.stringState = string(state)
	}
	return nil
}

func (o *subscribeOpts) buildPath() (string, error) {
	channels := utils.JoinChannels(o.Channels)

	return fmt.Sprintf(subscribePath,
		o.pubnub.Config.SubscribeKey,
		channels,
	), nil
}

func (o *subscribeOpts) buildQuery() (*url.Values, error) {
	q := defaultQuery(o.pubnub.Config.Uuid, o.pubnub.telemetryManager)

	if len(o.ChannelGroups) > 0 {
		channelGroup := utils.JoinChannels(o.ChannelGroups)
		q.Set("channel-group", string(channelGroup))
	}

	if o.Timetoken != 0 {
		q.Set("tt", strconv.FormatInt(o.Timetoken, 10))
	}

	if o.Region != "" {
		q.Set("tr", o.Region)
	}

	if o.FilterExpression != "" {
		q.Set("filter-expr", utils.UrlEncode(o.FilterExpression))
	}

	// hb timeout should be at least 4 seconds
	if o.Heartbeat >= 4 {
		q.Set("heartbeat", fmt.Sprintf("%d", o.Heartbeat))
	}

	/*state, err = json.Marshal(o.State)
	if err != nil {
		return nil, err
	}*/

	if o.State != nil {
		q.Set("state", o.stringState)
	}

	return q, nil
}

func (o *subscribeOpts) buildBody() ([]byte, error) {
	return []byte{}, nil
}

func (o *subscribeOpts) httpMethod() string {
	return "GET"
}

func (o *subscribeOpts) isAuthRequired() bool {
	return true
}

func (o *subscribeOpts) requestTimeout() int {
	return o.pubnub.Config.SubscribeRequestTimeout
}

func (o *subscribeOpts) connectTimeout() int {
	return o.pubnub.Config.ConnectTimeout
}

func (o *subscribeOpts) operationType() OperationType {
	return PNSubscribeOperation
}

func (o *subscribeOpts) telemetryManager() *TelemetryManager {
	return o.pubnub.telemetryManager
}

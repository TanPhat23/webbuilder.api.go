package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"my-go-app/internal/agentproto"
	wsentity "my-go-app/internal/realtime/delivery/ws/entity"
	domainEntity "my-go-app/internal/realtime/domain/entity"
	domainErrors "my-go-app/internal/realtime/domain/errors"
	"my-go-app/internal/realtime/utils/logger"
)

// agentTarget is a connected editor tab. *wsentity.Client implements it.
type agentTarget interface {
	Send(domainEntity.Message) error
}

type agentLocator interface {
	MostRecentClient(projectID, userID string) agentTarget
}

// hubLocator adapts the hub (which returns a concrete *Client, and a nil
// *Client must not become a non-nil interface).
type hubLocator struct{ hub *wsentity.Hub }

func (l hubLocator) MostRecentClient(projectID, userID string) agentTarget {
	if c := l.hub.MostRecentClient(projectID, userID); c != nil {
		return c
	}
	return nil
}

type claimer interface {
	Claim(ctx context.Context, requestID string) (bool, error)
}

// redisClaimer lets exactly one API instance deliver a given request.
type redisClaimer struct{ rdb *redis.Client }

func (c redisClaimer) Claim(ctx context.Context, requestID string) (bool, error) {
	return c.rdb.SetNX(ctx, "agent:claim:"+requestID, 1, 30*time.Second).Result()
}

// agentBridge connects the MCP relay (Redis) to editor tabs (WebSocket).
type agentBridge struct {
	loc    agentLocator
	bus    agentproto.Bus
	claim  claimer
	logger logger.Logger

	mu     sync.Mutex
	issued map[string]agentTarget // requestId -> the session it was sent to
}

func newAgentBridge(loc agentLocator, bus agentproto.Bus, claim claimer, lg logger.Logger) *agentBridge {
	return &agentBridge{loc: loc, bus: bus, claim: claim, logger: lg, issued: map[string]agentTarget{}}
}

// Run delivers requests until ctx ends. It subscribes before returning control
// to the loop, so call it in its own goroutine.
func (b *agentBridge) Run(ctx context.Context) {
	reqs, stop, err := b.bus.PSubscribe(ctx, agentproto.RequestChannelPrefix+"*")
	if err != nil {
		b.logger.Errorf("agent bridge: subscribe failed: %v", err)
		return
	}
	defer stop()
	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-reqs:
			if !ok {
				return
			}
			var req agentproto.Request
			if err := json.Unmarshal(m.Payload, &req); err != nil || req.RequestID == "" || req.UserID == "" {
				b.logger.Errorf("agent bridge: bad request payload: %v", err)
				continue
			}
			b.deliver(ctx, req)
		}
	}
}

func (b *agentBridge) deliver(ctx context.Context, req agentproto.Request) {
	target := b.loc.MostRecentClient(req.ProjectID, req.UserID)
	if target == nil {
		return // another instance may hold the tab; with none anywhere the relay times out on the missing ack
	}
	// Every API instance hears every request. Only the instance that wins the
	// claim delivers, so a user with tabs on two instances still gets one edit.
	if won, err := b.claim.Claim(ctx, req.RequestID); err != nil || !won {
		return
	}
	payload, _ := json.Marshal(map[string]any{"tool": req.Tool, "args": req.Args, "pageId": req.PageID})
	b.mu.Lock()
	b.issued[req.RequestID] = target
	b.mu.Unlock()
	msg := domainEntity.Message{
		Type: domainEntity.TypeAgentRequest, ProjectID: req.ProjectID, PageID: req.PageID,
		RequestID: req.RequestID, Payload: payload,
	}
	if err := target.Send(msg); err != nil {
		b.mu.Lock()
		delete(b.issued, req.RequestID)
		b.mu.Unlock()
		return
	}
	ack, _ := json.Marshal(struct{}{})
	_ = b.bus.Publish(ctx, agentproto.AckChannelPrefix+req.RequestID, ack)
}

// handleResponse accepts the tab's answer only from the session the request
// was sent to, and only once.
func (b *agentBridge) handleResponse(ctx context.Context, from agentTarget, msg domainEntity.Message) error {
	b.mu.Lock()
	want, ok := b.issued[msg.RequestID]
	if ok && want == from {
		delete(b.issued, msg.RequestID)
	}
	b.mu.Unlock()
	if !ok || want != from {
		return domainErrors.New(domainErrors.CodeValidation, "unknown agent request", errors.New("response from a session that was not addressed"))
	}
	var resp agentproto.Response
	if err := json.Unmarshal(msg.Payload, &resp); err != nil {
		resp = agentproto.Response{Error: &agentproto.Error{Code: agentproto.CodeInternal, Message: "the editor sent an unreadable response"}}
	}
	resp.RequestID = msg.RequestID
	out, _ := json.Marshal(resp)
	return b.bus.Publish(ctx, agentproto.ResponseChannelPrefix+msg.RequestID, out)
}

package http_server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"hmans.de/chatto/internal/core"
	"hmans.de/chatto/internal/events"
	graphauth "hmans.de/chatto/internal/graph/auth"
	apiv1 "hmans.de/chatto/internal/pb/chatto/api/v1"
	corev1 "hmans.de/chatto/internal/pb/chatto/core/v1"
	wirev1 "hmans.de/chatto/internal/pb/chatto/wire/v1"
)

const (
	wireProtocolVersion = "chatto-wire-v1"
	wireMaxFrameBytes   = 1 << 20
	wireWriteTimeout    = 10 * time.Second
	wireHelloTimeout    = 10 * time.Second
)

const (
	wireMethodGetViewer           = "chatto.api.v1.ChattoApiService/GetViewer"
	wireMethodListMyRooms         = "chatto.api.v1.ChattoApiService/ListMyRooms"
	wireMethodGetRoomTimeline     = "chatto.api.v1.ChattoApiService/GetRoomTimeline"
	wireMethodPostMessage         = "chatto.api.v1.ChattoApiService/PostMessage"
	wireMethodSendTypingIndicator = "chatto.api.v1.ChattoApiService/SendTypingIndicator"
)

var errWireInvalidArgument = errors.New("invalid wire request")

var wireMethods = []string{
	"/" + wireMethodGetViewer,
	"/" + wireMethodListMyRooms,
	"/" + wireMethodGetRoomTimeline,
	"/" + wireMethodPostMessage,
	"/" + wireMethodSendTypingIndicator,
}

type wireConn struct {
	server *HTTPServer
	conn   *websocket.Conn
	out    chan *wirev1.ServerFrame

	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.Mutex
	user     *corev1.User
	hello    bool
	requests map[string]context.CancelFunc
}

func (s *HTTPServer) setupWireAPI(allowedOrigins []string) {
	upgrader := websocket.Upgrader{
		EnableCompression: s.config.Webserver.WebSocketCompressionEnabled(),
		CheckOrigin: func(r *http.Request) bool {
			return s.checkWireWebSocketOrigin(r, allowedOrigins)
		},
	}

	s.router.GET("/api/wire", func(c *gin.Context) {
		s.requestContextWithAuditMetadata(c)
		authenticatedRequest := s.injectUserIntoContext(c)

		conn, err := upgrader.Upgrade(c.Writer, authenticatedRequest, nil)
		if err != nil {
			s.logger.Warn("Wire WebSocket upgrade failed", "error", err)
			return
		}

		ctx, cancel := context.WithCancel(authenticatedRequest.Context())
		wc := &wireConn{
			server:   s,
			conn:     conn,
			out:      make(chan *wirev1.ServerFrame, 256),
			ctx:      ctx,
			cancel:   cancel,
			user:     graphauth.ForContext(authenticatedRequest.Context()),
			requests: make(map[string]context.CancelFunc),
		}
		wc.run()
	})
}

func (s *HTTPServer) checkWireWebSocketOrigin(r *http.Request, allowedOrigins []string) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if s.matchOrigin(origin, allowedOrigins) != originNotAllowed {
		return true
	}

	host := r.Host
	if forwarded := r.Header.Get("X-Forwarded-Host"); forwarded != "" {
		host = forwarded
	}
	if parsedOrigin, err := url.Parse(origin); err == nil && strings.EqualFold(parsedOrigin.Host, host) {
		return true
	}

	s.logger.Warn("Wire WebSocket connection rejected: origin mismatch",
		"origin", origin, "host", host, "allowed", allowedOrigins)
	return false
}

func (c *wireConn) run() {
	defer c.cancel()
	defer c.conn.Close()

	c.conn.SetReadLimit(wireMaxFrameBytes)
	_ = c.conn.SetReadDeadline(time.Now().Add(wireHelloTimeout))

	var writerDone sync.WaitGroup
	writerDone.Add(1)
	go func() {
		defer writerDone.Done()
		c.writeLoop()
	}()

	c.readLoop()
	c.cancel()
	_ = c.conn.Close()
	c.cancelInflight()
	writerDone.Wait()
}

func (c *wireConn) writeLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case frame, ok := <-c.out:
			if !ok {
				return
			}
			data, err := proto.Marshal(frame)
			if err != nil {
				c.server.logger.Warn("Failed to marshal wire frame", "error", err)
				return
			}
			if err := c.conn.SetWriteDeadline(time.Now().Add(wireWriteTimeout)); err != nil {
				return
			}
			if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
				return
			}
		}
	}
}

func (c *wireConn) readLoop() {
	for {
		messageType, data, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		if messageType != websocket.BinaryMessage {
			c.sendError("", "", wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "wire frames must be binary protobuf messages", false)
			continue
		}

		var frame wirev1.ClientFrame
		if err := proto.Unmarshal(data, &frame); err != nil {
			c.sendError("", "", wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid protobuf frame", false)
			continue
		}

		switch kind := frame.GetKind().(type) {
		case *wirev1.ClientFrame_Hello:
			c.handleHello(frame.GetFrameId(), kind.Hello)
		case *wirev1.ClientFrame_Request:
			c.handleRequestFrame(frame.GetFrameId(), kind.Request)
		case *wirev1.ClientFrame_Cancel:
			c.handleCancel(kind.Cancel)
		case *wirev1.ClientFrame_Ack:
			// Event acknowledgements are intentionally advisory in the prototype.
		default:
			c.sendError(frame.GetFrameId(), "", wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "frame kind is required", false)
		}
	}
}

func (c *wireConn) handleHello(frameID string, hello *wirev1.ClientHello) {
	if hello == nil {
		c.sendError(frameID, "", wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "hello is required", false)
		return
	}

	c.mu.Lock()
	if c.hello {
		c.mu.Unlock()
		c.sendError(frameID, "", wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "hello already received", false)
		return
	}
	c.mu.Unlock()

	user, err := c.authenticateHello(hello)
	if err != nil {
		c.sendError(frameID, "", wirev1.ErrorCode_ERROR_CODE_UNAUTHENTICATED, "authentication required", false)
		return
	}

	_ = c.conn.SetReadDeadline(time.Time{})

	c.mu.Lock()
	c.user = user
	c.hello = true
	c.mu.Unlock()

	c.send(&wirev1.ServerFrame{
		FrameId: frameID,
		Kind: &wirev1.ServerFrame_Hello{Hello: &wirev1.ServerHello{
			ProtocolVersion: wireProtocolVersion,
			ServerVersion:   c.server.version,
			Methods:         append([]string(nil), wireMethods...),
			Features:        []string{"binary-protobuf", "requests", "my-events"},
		}},
	})

	events, err := c.server.core.StreamMyEvents(c.ctx, user.GetId())
	if err != nil {
		c.sendError(frameID, "", wirev1.ErrorCode_ERROR_CODE_INTERNAL, "failed to subscribe to events", true)
		return
	}
	go c.forwardEvents(events)
}

func (c *wireConn) authenticateHello(hello *wirev1.ClientHello) (*corev1.User, error) {
	c.mu.Lock()
	user := c.user
	c.mu.Unlock()
	if user != nil {
		return user, nil
	}

	token := strings.TrimSpace(hello.GetBearerToken())
	if token == "" {
		return nil, core.ErrNotAuthenticated
	}
	userID, err := c.server.core.ValidateAuthToken(c.ctx, token)
	if err != nil {
		return nil, err
	}
	return c.server.core.GetUser(c.ctx, userID)
}

func (c *wireConn) handleRequestFrame(frameID string, req *wirev1.Request) {
	if req == nil {
		c.sendError(frameID, "", wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "request is required", false)
		return
	}
	user := c.currentUser()
	if user == nil {
		c.sendError(frameID, req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_UNAUTHENTICATED, "authentication required", false)
		return
	}
	if req.GetRequestId() == "" {
		c.sendError(frameID, "", wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "request_id is required", false)
		return
	}

	reqCtx, cancel := context.WithCancel(c.ctx)
	c.mu.Lock()
	if _, exists := c.requests[req.GetRequestId()]; exists {
		c.mu.Unlock()
		cancel()
		c.sendError(frameID, req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "request_id is already in flight", false)
		return
	}
	c.requests[req.GetRequestId()] = cancel
	c.mu.Unlock()

	go func() {
		defer func() {
			c.mu.Lock()
			delete(c.requests, req.GetRequestId())
			c.mu.Unlock()
			cancel()
		}()

		resp, wireErr := c.handleRequest(reqCtx, user, req)
		if wireErr != nil {
			c.send(&wirev1.ServerFrame{
				FrameId: frameID,
				Kind:    &wirev1.ServerFrame_Error{Error: wireErr},
			})
			return
		}
		data, err := proto.Marshal(resp)
		if err != nil {
			c.sendError(frameID, req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_INTERNAL, "failed to marshal response", false)
			return
		}
		c.send(&wirev1.ServerFrame{
			FrameId: frameID,
			Kind: &wirev1.ServerFrame_Response{Response: &wirev1.Response{
				RequestId: req.GetRequestId(),
				Body:      data,
			}},
		})
	}()
}

func (c *wireConn) handleRequest(ctx context.Context, user *corev1.User, req *wirev1.Request) (proto.Message, *wirev1.WireError) {
	method := normalizeWireMethod(req.GetMethod())
	switch method {
	case wireMethodGetViewer:
		var body apiv1.GetViewerRequest
		if err := proto.Unmarshal(req.GetBody(), &body); err != nil {
			return nil, wireError(req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid GetViewerRequest", false)
		}
		return &apiv1.GetViewerResponse{
			Viewer: &apiv1.Viewer{User: cloneUser(user)},
		}, nil

	case wireMethodListMyRooms:
		var body apiv1.ListMyRoomsRequest
		if err := proto.Unmarshal(req.GetBody(), &body); err != nil {
			return nil, wireError(req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid ListMyRoomsRequest", false)
		}
		kind := roomKindFromProto(body.GetKind())
		rooms, err := c.server.core.ListMemberRooms(ctx, kind, user.GetId(), core.MemberRoomListOptions{})
		if err != nil {
			return nil, c.errorFromRequestErr(req.GetRequestId(), err)
		}
		return &apiv1.ListMyRoomsResponse{Rooms: cloneRooms(rooms)}, nil

	case wireMethodGetRoomTimeline:
		var body apiv1.GetRoomTimelineRequest
		if err := proto.Unmarshal(req.GetBody(), &body); err != nil {
			return nil, wireError(req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid GetRoomTimelineRequest", false)
		}
		room, kind, err := c.authorizedRoom(ctx, user.GetId(), body.GetRoomId())
		if err != nil {
			return nil, c.errorFromRequestErr(req.GetRequestId(), err)
		}
		_ = room
		var before *uint64
		if body.GetBeforeSequence() > 0 {
			seq := body.GetBeforeSequence()
			before = &seq
		}
		result, err := c.server.core.GetRoomEvents(ctx, kind, body.GetRoomId(), int(body.GetLimit()), before)
		if err != nil {
			return nil, c.errorFromRequestErr(req.GetRequestId(), err)
		}
		return timelineResponse(result), nil

	case wireMethodPostMessage:
		var body apiv1.PostMessageRequest
		if err := proto.Unmarshal(req.GetBody(), &body); err != nil {
			return nil, wireError(req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid PostMessageRequest", false)
		}
		_, kind, err := c.authorizedRoom(ctx, user.GetId(), body.GetRoomId())
		if err != nil {
			return nil, c.errorFromRequestErr(req.GetRequestId(), err)
		}
		canPost, err := c.server.core.CanPostMessage(ctx, user.GetId(), kind, body.GetRoomId())
		if err != nil {
			return nil, c.errorFromRequestErr(req.GetRequestId(), err)
		}
		if !canPost {
			return nil, wireError(req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "permission denied", false)
		}
		opts := []core.PostMessageOption{}
		if body.GetLargeMentionConfirmed() {
			opts = append(opts, core.WithLargeMentionConfirmed())
		}
		event, err := c.server.core.PostMessage(ctx, kind, body.GetRoomId(), user.GetId(), body.GetBody(), nil, body.GetThreadRootEventId(), body.GetInReplyToEventId(), nil, body.GetAlsoSendToChannel(), opts...)
		if err != nil {
			return nil, c.errorFromRequestErr(req.GetRequestId(), err)
		}
		seq, err := c.server.core.GetEventSequence(ctx, kind, body.GetRoomId(), event.GetId())
		if err != nil {
			return nil, c.errorFromRequestErr(req.GetRequestId(), err)
		}
		return &apiv1.PostMessageResponse{Event: cloneEvent(event), Sequence: seq}, nil

	case wireMethodSendTypingIndicator:
		var body apiv1.SendTypingIndicatorRequest
		if err := proto.Unmarshal(req.GetBody(), &body); err != nil {
			return nil, wireError(req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, "invalid SendTypingIndicatorRequest", false)
		}
		_, kind, err := c.authorizedRoom(ctx, user.GetId(), body.GetRoomId())
		if err != nil {
			return nil, c.errorFromRequestErr(req.GetRequestId(), err)
		}
		var threadRoot *string
		if body.GetThreadRootEventId() != "" {
			value := body.GetThreadRootEventId()
			threadRoot = &value
		}
		if err := c.server.core.PublishTypingIndicator(ctx, user.GetId(), kind, body.GetRoomId(), threadRoot); err != nil {
			return nil, c.errorFromRequestErr(req.GetRequestId(), err)
		}
		return &apiv1.SendTypingIndicatorResponse{}, nil

	default:
		return nil, wireError(req.GetRequestId(), wirev1.ErrorCode_ERROR_CODE_UNIMPLEMENTED, "unknown method", false)
	}
}

func (c *wireConn) handleCancel(cancel *wirev1.CancelRequest) {
	if cancel == nil || cancel.GetRequestId() == "" {
		return
	}
	c.mu.Lock()
	cancelFunc := c.requests[cancel.GetRequestId()]
	c.mu.Unlock()
	if cancelFunc != nil {
		cancelFunc()
	}
}

func (c *wireConn) forwardEvents(events <-chan core.EventEnvelope) {
	for {
		select {
		case <-c.ctx.Done():
			return
		case event, ok := <-events:
			if !ok {
				return
			}
			streamEvent := c.streamEvent(event)
			if streamEvent == nil {
				continue
			}
			if !c.send(&wirev1.ServerFrame{
				FrameId: "",
				Kind:    &wirev1.ServerFrame_Event{Event: streamEvent},
			}) {
				return
			}
		}
	}
}

func (c *wireConn) streamEvent(event core.EventEnvelope) *wirev1.StreamEvent {
	if event == nil {
		return nil
	}
	streamEvent := &wirev1.StreamEvent{
		EventId:     event.ID(),
		EventType:   wireEventType(event),
		Invalidates: wireInvalidationHints(event),
	}
	switch {
	case event.EVTEvent() != nil:
		streamEvent.Payload = &wirev1.StreamEvent_DurableEvent{DurableEvent: cloneEvent(event.EVTEvent())}
	case event.LiveEvent() != nil:
		streamEvent.Payload = &wirev1.StreamEvent_LiveEvent{LiveEvent: cloneLiveEvent(event.LiveEvent())}
	case event.HeartbeatEvent() != nil:
		streamEvent.Payload = &wirev1.StreamEvent_Heartbeat{Heartbeat: &corev1.HeartbeatEvent{}}
	}
	return streamEvent
}

func (c *wireConn) currentUser() *corev1.User {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.hello {
		return nil
	}
	return c.user
}

func (c *wireConn) authorizedRoom(ctx context.Context, userID, roomID string) (*corev1.Room, core.RoomKind, error) {
	if roomID == "" {
		return nil, "", fmt.Errorf("%w: room_id is required", errWireInvalidArgument)
	}
	room, err := c.server.core.FindRoomByID(ctx, roomID)
	if err != nil {
		return nil, "", err
	}
	kind := core.KindOfRoom(room)
	member, err := c.server.core.RoomMembershipExists(ctx, kind, userID, roomID)
	if err != nil {
		return nil, "", err
	}
	if !member {
		return nil, "", core.ErrPermissionDenied
	}
	return room, kind, nil
}

func (c *wireConn) sendError(frameID, requestID string, code wirev1.ErrorCode, message string, retryable bool) bool {
	return c.send(&wirev1.ServerFrame{
		FrameId: frameID,
		Kind:    &wirev1.ServerFrame_Error{Error: wireError(requestID, code, message, retryable)},
	})
}

func (c *wireConn) send(frame *wirev1.ServerFrame) bool {
	select {
	case <-c.ctx.Done():
		return false
	case c.out <- frame:
		return true
	}
}

func (c *wireConn) cancelInflight() {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, cancel := range c.requests {
		cancel()
	}
	clear(c.requests)
}

func (c *wireConn) errorFromRequestErr(requestID string, err error) *wirev1.WireError {
	switch {
	case errors.Is(err, context.Canceled):
		return wireError(requestID, wirev1.ErrorCode_ERROR_CODE_CANCELLED, "request cancelled", false)
	case errors.Is(err, core.ErrNotAuthenticated):
		return wireError(requestID, wirev1.ErrorCode_ERROR_CODE_UNAUTHENTICATED, "authentication required", false)
	case errors.Is(err, core.ErrPermissionDenied):
		return wireError(requestID, wirev1.ErrorCode_ERROR_CODE_PERMISSION_DENIED, "permission denied", false)
	case errors.Is(err, core.ErrNotFound), errors.Is(err, core.ErrMessageNotFound), errors.Is(err, jetstream.ErrKeyNotFound):
		return wireError(requestID, wirev1.ErrorCode_ERROR_CODE_NOT_FOUND, "not found", false)
	case errors.Is(err, errWireInvalidArgument), errors.Is(err, core.ErrMessageTooLong):
		return wireError(requestID, wirev1.ErrorCode_ERROR_CODE_INVALID_ARGUMENT, err.Error(), false)
	default:
		c.server.logger.Warn("Wire request failed", "error", err)
		return wireError(requestID, wirev1.ErrorCode_ERROR_CODE_INTERNAL, "internal error", true)
	}
}

func wireError(requestID string, code wirev1.ErrorCode, message string, retryable bool) *wirev1.WireError {
	return &wirev1.WireError{
		RequestId: requestID,
		Code:      code,
		Message:   message,
		Retryable: retryable,
	}
}

func normalizeWireMethod(method string) string {
	return strings.TrimPrefix(strings.TrimSpace(method), "/")
}

func roomKindFromProto(kind corev1.RoomKind) core.RoomKind {
	if kind == corev1.RoomKind_ROOM_KIND_DM {
		return core.KindDM
	}
	return core.KindChannel
}

func timelineResponse(result *core.RoomEventsResult) *apiv1.GetRoomTimelineResponse {
	resp := &apiv1.GetRoomTimelineResponse{}
	if result == nil {
		return resp
	}
	resp.HasOlder = result.HasOlder
	resp.HasNewer = result.HasNewer
	resp.StartSequence = result.StartCursorSeq
	resp.EndSequence = result.EndCursorSeq
	resp.Events = make([]*apiv1.TimelineEvent, 0, len(result.Events))
	for _, event := range result.Events {
		if event == nil || event.Event == nil {
			continue
		}
		resp.Events = append(resp.Events, &apiv1.TimelineEvent{
			Event:    cloneEvent(event.Event),
			Sequence: event.Sequence,
		})
	}
	return resp
}

func cloneUser(user *corev1.User) *corev1.User {
	if user == nil {
		return nil
	}
	return proto.Clone(user).(*corev1.User)
}

func cloneRooms(rooms []*corev1.Room) []*corev1.Room {
	cloned := make([]*corev1.Room, 0, len(rooms))
	for _, room := range rooms {
		if room != nil {
			cloned = append(cloned, proto.Clone(room).(*corev1.Room))
		}
	}
	return cloned
}

func cloneEvent(event *corev1.Event) *corev1.Event {
	if event == nil {
		return nil
	}
	return proto.Clone(event).(*corev1.Event)
}

func cloneLiveEvent(event *corev1.LiveEvent) *corev1.LiveEvent {
	if event == nil {
		return nil
	}
	return proto.Clone(event).(*corev1.LiveEvent)
}

func wireEventType(event core.EventEnvelope) string {
	switch {
	case event == nil:
		return ""
	case event.EVTEvent() != nil:
		return events.EventTypeOf(event.EVTEvent())
	case event.LiveEvent() != nil:
		return protoOneofFieldName(event.LiveEvent().ProtoReflect(), "event")
	case event.HeartbeatEvent() != nil:
		return "heartbeat"
	default:
		return ""
	}
}

func protoOneofFieldName(msg protoreflect.Message, oneofName protoreflect.Name) string {
	oneof := msg.Descriptor().Oneofs().ByName(oneofName)
	if oneof == nil {
		return ""
	}
	field := msg.WhichOneof(oneof)
	if field == nil {
		return ""
	}
	return string(field.Name())
}

func wireInvalidationHints(event core.EventEnvelope) []*wirev1.InvalidationHint {
	if event == nil {
		return nil
	}
	if roomID := wireRoomIDOfEnvelope(event); roomID != "" {
		return []*wirev1.InvalidationHint{
			{Kind: wirev1.InvalidationKind_INVALIDATION_KIND_ROOM, Id: roomID},
			{Kind: wirev1.InvalidationKind_INVALIDATION_KIND_ROOM_TIMELINE, Id: roomID},
		}
	}
	if live := event.LiveEvent(); live != nil {
		switch e := live.GetEvent().(type) {
		case *corev1.LiveEvent_UserProfileUpdated:
			return []*wirev1.InvalidationHint{{Kind: wirev1.InvalidationKind_INVALIDATION_KIND_USER, Id: e.UserProfileUpdated.GetUserId()}}
		case *corev1.LiveEvent_ServerUpdated, *corev1.LiveEvent_RoomGroupsUpdated:
			return []*wirev1.InvalidationHint{{Kind: wirev1.InvalidationKind_INVALIDATION_KIND_SERVER, Id: "server"}}
		case *corev1.LiveEvent_ServerUserPreferencesUpdated,
			*corev1.LiveEvent_NotificationLevelChanged,
			*corev1.LiveEvent_ThreadFollowChanged,
			*corev1.LiveEvent_NotificationCreated,
			*corev1.LiveEvent_NotificationDismissed,
			*corev1.LiveEvent_RoomMarkedAsRead,
			*corev1.LiveEvent_MentionStatusCleared:
			return []*wirev1.InvalidationHint{{Kind: wirev1.InvalidationKind_INVALIDATION_KIND_VIEWER, Id: live.GetActorId()}}
		}
	}
	return nil
}

func wireRoomIDOfEnvelope(event core.EventEnvelope) string {
	if event == nil {
		return ""
	}
	if evt := event.EVTEvent(); evt != nil {
		return wireRoomIDOfEvent(evt)
	}
	if live := event.LiveEvent(); live != nil {
		switch e := live.GetEvent().(type) {
		case *corev1.LiveEvent_UserTyping:
			return e.UserTyping.GetRoomId()
		case *corev1.LiveEvent_CallParticipantJoined:
			return e.CallParticipantJoined.GetRoomId()
		case *corev1.LiveEvent_CallParticipantLeft:
			return e.CallParticipantLeft.GetRoomId()
		}
	}
	return ""
}

func wireRoomIDOfEvent(event *corev1.Event) string {
	if event == nil {
		return ""
	}
	switch e := event.GetEvent().(type) {
	case *corev1.Event_RoomCreated:
		return e.RoomCreated.GetRoomId()
	case *corev1.Event_RoomUpdated:
		return e.RoomUpdated.GetRoomId()
	case *corev1.Event_RoomDeleted:
		return e.RoomDeleted.GetRoomId()
	case *corev1.Event_RoomArchived:
		return e.RoomArchived.GetRoomId()
	case *corev1.Event_RoomUnarchived:
		return e.RoomUnarchived.GetRoomId()
	case *corev1.Event_UserJoinedRoom:
		return e.UserJoinedRoom.GetRoomId()
	case *corev1.Event_UserLeftRoom:
		return e.UserLeftRoom.GetRoomId()
	case *corev1.Event_RoomMemberBanned:
		return e.RoomMemberBanned.GetRoomId()
	case *corev1.Event_RoomMemberUnbanned:
		return e.RoomMemberUnbanned.GetRoomId()
	case *corev1.Event_MessagePosted:
		return e.MessagePosted.GetRoomId()
	case *corev1.Event_MessageEdited:
		return e.MessageEdited.GetRoomId()
	case *corev1.Event_MessageRetracted:
		return e.MessageRetracted.GetRoomId()
	case *corev1.Event_ThreadCreated:
		return e.ThreadCreated.GetRoomId()
	case *corev1.Event_ReactionAdded:
		return e.ReactionAdded.GetRoomId()
	case *corev1.Event_ReactionRemoved:
		return e.ReactionRemoved.GetRoomId()
	case *corev1.Event_VoiceCallParticipantJoined:
		return e.VoiceCallParticipantJoined.GetRoomId()
	case *corev1.Event_VoiceCallParticipantLeft:
		return e.VoiceCallParticipantLeft.GetRoomId()
	case *corev1.Event_VoiceCallStarted:
		return e.VoiceCallStarted.GetRoomId()
	case *corev1.Event_VoiceCallEnded:
		return e.VoiceCallEnded.GetRoomId()
	}
	return ""
}

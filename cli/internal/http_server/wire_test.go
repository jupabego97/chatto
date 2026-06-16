package http_server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"

	"hmans.de/chatto/internal/core"
	apiv1 "hmans.de/chatto/internal/pb/chatto/api/v1"
	wirev1 "hmans.de/chatto/internal/pb/chatto/wire/v1"
)

func (env *wsTestEnv) connectWire(t *testing.T) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(env.server.URL, "http") + "/api/wire"
	header := http.Header{}
	for _, c := range env.cookieJar.Cookies(mustParseURL(env.server.URL)) {
		header.Add("Cookie", c.String())
	}

	conn, resp, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		if resp != nil {
			t.Fatalf("wire WebSocket dial failed with status %d: %v", resp.StatusCode, err)
		}
		t.Fatalf("wire WebSocket dial failed: %v", err)
	}

	t.Cleanup(func() { conn.Close() })
	return conn
}

func sendWireFrame(t *testing.T, conn *websocket.Conn, frame *wirev1.ClientFrame) {
	t.Helper()
	data, err := proto.Marshal(frame)
	if err != nil {
		t.Fatalf("marshal wire frame: %v", err)
	}
	if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		t.Fatalf("send wire frame: %v", err)
	}
}

func readWireFrame(t *testing.T, conn *websocket.Conn, timeout time.Duration) *wirev1.ServerFrame {
	t.Helper()
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		t.Fatalf("set read deadline: %v", err)
	}
	messageType, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read wire frame: %v", err)
	}
	if messageType != websocket.BinaryMessage {
		t.Fatalf("wire message type = %d, want binary", messageType)
	}
	var frame wirev1.ServerFrame
	if err := proto.Unmarshal(data, &frame); err != nil {
		t.Fatalf("unmarshal wire frame: %v", err)
	}
	return &frame
}

func mustProtoBytes(t *testing.T, msg proto.Message) []byte {
	t.Helper()
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal proto body: %v", err)
	}
	return data
}

func sendWireHello(t *testing.T, conn *websocket.Conn, resumeAfter string) {
	t.Helper()
	sendWireFrame(t, conn, &wirev1.ClientFrame{
		FrameId: "hello",
		Kind: &wirev1.ClientFrame_Hello{Hello: &wirev1.ClientHello{
			ProtocolVersion: wireProtocolVersion,
			ResumeAfter:     resumeAfter,
		}},
	})
	frame := readWireFrame(t, conn, 5*time.Second)
	if frame.GetHello() == nil {
		t.Fatalf("first wire frame = %T, want ServerHello", frame.GetKind())
	}
	if frame.GetHello().GetProtocolVersion() != wireProtocolVersion {
		t.Fatalf("protocol version = %q, want %q", frame.GetHello().GetProtocolVersion(), wireProtocolVersion)
	}
}

func sendWireRequest(t *testing.T, conn *websocket.Conn, frameID, requestID, method string, body proto.Message) {
	t.Helper()
	sendWireFrame(t, conn, &wirev1.ClientFrame{
		FrameId: frameID,
		Kind: &wirev1.ClientFrame_Request{Request: &wirev1.Request{
			RequestId: requestID,
			Method:    method,
			Body:      mustProtoBytes(t, body),
		}},
	})
}

func TestWireAPI_RequestResponseAndLiveEvent(t *testing.T) {
	env := setupWebSocketTestServer(t)

	user, err := env.core.CreateUser(env.ctx, "system", "wireuser", "Wire User", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	room, err := env.core.CreateRoom(env.ctx, user.Id, core.KindChannel, "", "wire-room", "")
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if _, err := env.core.JoinRoom(env.ctx, user.Id, core.KindChannel, user.Id, room.Id); err != nil {
		t.Fatalf("JoinRoom: %v", err)
	}

	env.login(t, "wireuser", "password123")
	conn := env.connectWire(t)
	sendWireHello(t, conn, "")

	sendWireRequest(t, conn, "viewer-frame", "viewer", "/chatto.api.v1.ChattoApiService/GetViewer", &apiv1.GetViewerRequest{})
	viewerResp := readWireResponse(t, conn, "viewer", 5*time.Second)
	var viewer apiv1.GetViewerResponse
	if err := proto.Unmarshal(viewerResp.GetBody(), &viewer); err != nil {
		t.Fatalf("unmarshal viewer response: %v", err)
	}
	if viewer.GetViewer().GetUser().GetId() != user.Id {
		t.Fatalf("viewer user id = %q, want %q", viewer.GetViewer().GetUser().GetId(), user.Id)
	}

	sendWireRequest(t, conn, "post-frame", "post", "/chatto.api.v1.ChattoApiService/PostMessage", &apiv1.PostMessageRequest{
		RoomId: room.Id,
		Body:   "hello over wire",
	})

	var postResp *apiv1.PostMessageResponse
	var pushed *wirev1.StreamEvent
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) && (postResp == nil || pushed == nil) {
		frame := readWireFrame(t, conn, time.Until(deadline))
		if resp := frame.GetResponse(); resp != nil && resp.GetRequestId() == "post" {
			var decoded apiv1.PostMessageResponse
			if err := proto.Unmarshal(resp.GetBody(), &decoded); err != nil {
				t.Fatalf("unmarshal post response: %v", err)
			}
			postResp = &decoded
			continue
		}
		if event := frame.GetEvent(); event != nil {
			if event.GetDurableEvent().GetMessagePosted().GetRoomId() == room.Id {
				pushed = event
			}
		}
		if errFrame := frame.GetError(); errFrame != nil {
			t.Fatalf("unexpected wire error: %s", errFrame.GetMessage())
		}
	}
	if postResp == nil {
		t.Fatal("did not receive PostMessage response")
	}
	if pushed == nil {
		t.Fatal("did not receive pushed MessagePosted event")
	}
	if postResp.GetEvent().GetId() == "" {
		t.Fatal("PostMessage response event id is empty")
	}
	if pushed.GetDurableEvent().GetId() != postResp.GetEvent().GetId() {
		t.Fatalf("pushed event id = %q, want response event id %q", pushed.GetDurableEvent().GetId(), postResp.GetEvent().GetId())
	}
	if pushed.GetDeliveryCursor() == "" {
		t.Fatal("pushed durable event has no delivery cursor")
	}
	if pushed.GetEventType() != "message_posted" {
		t.Fatalf("pushed event type = %q, want message_posted", pushed.GetEventType())
	}
	if !hasInvalidation(pushed, wirev1.InvalidationKind_INVALIDATION_KIND_ROOM_TIMELINE, room.Id) {
		t.Fatal("pushed event did not include room timeline invalidation")
	}
}

func TestWireAPI_UnknownMethodReturnsStructuredError(t *testing.T) {
	env := setupWebSocketTestServer(t)

	if _, err := env.core.CreateUser(env.ctx, "system", "wireerror", "Wire Error", "password123"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	env.login(t, "wireerror", "password123")

	conn := env.connectWire(t)
	sendWireHello(t, conn, "")

	sendWireRequest(t, conn, "bad-frame", "bad", "/chatto.api.v1.ChattoApiService/Nope", &apiv1.GetViewerRequest{})
	errFrame := readWireError(t, conn, "bad", 5*time.Second)
	if errFrame == nil {
		t.Fatal("expected wire error")
	}
	if errFrame.GetRequestId() != "bad" {
		t.Fatalf("error request id = %q, want bad", errFrame.GetRequestId())
	}
	if errFrame.GetCode() != wirev1.ErrorCode_ERROR_CODE_UNIMPLEMENTED {
		t.Fatalf("error code = %v, want UNIMPLEMENTED", errFrame.GetCode())
	}
}

func TestWireAPI_ReplaysDurableEventsAfterCursor(t *testing.T) {
	env := setupWebSocketTestServer(t)

	user, err := env.core.CreateUser(env.ctx, "system", "wirereplay", "Wire Replay", "password123")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	room, err := env.core.CreateRoom(env.ctx, user.Id, core.KindChannel, "", "wire-replay", "")
	if err != nil {
		t.Fatalf("CreateRoom: %v", err)
	}
	if _, err := env.core.JoinRoom(env.ctx, user.Id, core.KindChannel, user.Id, room.Id); err != nil {
		t.Fatalf("JoinRoom: %v", err)
	}

	env.login(t, "wirereplay", "password123")
	conn := env.connectWire(t)
	sendWireHello(t, conn, "")

	first, err := env.core.PostMessage(env.ctx, core.KindChannel, room.Id, user.Id, "first replay marker", nil, "", "", nil, false)
	if err != nil {
		t.Fatalf("PostMessage first: %v", err)
	}
	firstEvent := readWireDurableEvent(t, conn, first.GetId(), 5*time.Second)
	if firstEvent.GetDeliveryCursor() == "" {
		t.Fatal("first event has no delivery cursor")
	}
	_ = conn.Close()

	second, err := env.core.PostMessage(env.ctx, core.KindChannel, room.Id, user.Id, "second replay marker", nil, "", "", nil, false)
	if err != nil {
		t.Fatalf("PostMessage second: %v", err)
	}

	reconnected := env.connectWire(t)
	sendWireHello(t, reconnected, firstEvent.GetDeliveryCursor())
	replayed := readWireDurableEvent(t, reconnected, second.GetId(), 5*time.Second)
	if replayed.GetDeliveryCursor() == "" {
		t.Fatal("replayed event has no delivery cursor")
	}
}

func readWireResponse(t *testing.T, conn *websocket.Conn, requestID string, timeout time.Duration) *wirev1.Response {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		frame := readWireFrame(t, conn, time.Until(deadline))
		if resp := frame.GetResponse(); resp != nil && resp.GetRequestId() == requestID {
			return resp
		}
		if errFrame := frame.GetError(); errFrame != nil && errFrame.GetRequestId() == requestID {
			t.Fatalf("wire request %s failed: %s", requestID, errFrame.GetMessage())
		}
	}
	t.Fatalf("did not receive response for request %s", requestID)
	return nil
}

func readWireError(t *testing.T, conn *websocket.Conn, requestID string, timeout time.Duration) *wirev1.WireError {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		frame := readWireFrame(t, conn, time.Until(deadline))
		if errFrame := frame.GetError(); errFrame != nil && errFrame.GetRequestId() == requestID {
			return errFrame
		}
	}
	t.Fatalf("did not receive error for request %s", requestID)
	return nil
}

func readWireDurableEvent(t *testing.T, conn *websocket.Conn, eventID string, timeout time.Duration) *wirev1.StreamEvent {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		frame := readWireFrame(t, conn, time.Until(deadline))
		if event := frame.GetEvent(); event != nil && event.GetDurableEvent().GetId() == eventID {
			return event
		}
		if errFrame := frame.GetError(); errFrame != nil {
			t.Fatalf("unexpected wire error: %s", errFrame.GetMessage())
		}
	}
	t.Fatalf("did not receive durable event %s", eventID)
	return nil
}

func hasInvalidation(event *wirev1.StreamEvent, kind wirev1.InvalidationKind, id string) bool {
	for _, hint := range event.GetInvalidates() {
		if hint.GetKind() == kind && hint.GetId() == id {
			return true
		}
	}
	return false
}

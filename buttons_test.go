package buttons

import (
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/polymorfa/hypermeow/proto/waE2E"
)

func TestMessageAppendsTrailingEmptyButton(t *testing.T) {
	msg := Message("body", "footer", QuickReply("A", "a"))
	btns := msg.GetInteractiveMessage().GetNativeFlowMessage().GetButtons()
	if len(btns) != 2 {
		t.Fatalf("expected 2 buttons (quick_reply + trailing empty), got %d", len(btns))
	}
	if btns[1].GetName() != "" {
		t.Fatalf("trailing button name = %q, want empty", btns[1].GetName())
	}
	if btns[0].GetName() != "quick_reply" {
		t.Fatalf("first button name = %q, want quick_reply", btns[0].GetName())
	}
}

func TestLocationRequestHasNoTrailingButton(t *testing.T) {
	msg := LocationRequestMessage("body", "")
	btns := msg.GetInteractiveMessage().GetNativeFlowMessage().GetButtons()
	if len(btns) != 1 {
		t.Fatalf("expected exactly 1 button, got %d", len(btns))
	}
	if btns[0].GetName() != "send_location" || btns[0].GetButtonParamsJSON() != "" {
		t.Fatalf("location button = name %q params %q, want send_location with empty params",
			btns[0].GetName(), btns[0].GetButtonParamsJSON())
	}
}

func TestParseTap(t *testing.T) {
	m := &waE2E.Message{InteractiveResponseMessage: &waE2E.InteractiveResponseMessage{
		InteractiveResponseMessage: &waE2E.InteractiveResponseMessage_NativeFlowResponseMessage_{
			NativeFlowResponseMessage: &waE2E.InteractiveResponseMessage_NativeFlowResponseMessage{
				Name:       proto.String("quick_reply"),
				ParamsJSON: proto.String(`{"id":"a","display_text":"Option A"}`),
			},
		},
	}}
	tap, ok := ParseTap(m)
	if !ok {
		t.Fatal("ParseTap returned ok=false")
	}
	if tap.Name != "quick_reply" || tap.ID != "a" {
		t.Fatalf("tap = %+v, want name=quick_reply id=a", tap)
	}
	if _, ok := ParseTap(&waE2E.Message{Conversation: proto.String("hi")}); ok {
		t.Fatal("ParseTap on a plain text message returned ok=true")
	}
}

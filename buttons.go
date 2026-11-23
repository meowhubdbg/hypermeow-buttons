// Package buttons sends WhatsApp native-flow buttons with HyperMeow.
//
// Only button types that were LIVE-TESTED on a real phone (2026-10-21) are
// included: quick_reply, cta_url, cta_url as an in-app webview, cta_copy,
// cta_call, single_select dropdowns (including multi-menu panels), and
// send_location requests.
//
// Classic ListMessage IS included (classic.go) — it renders on Android
// AND iOS once HyperMeow gets a one-line patch (see ClassicMessage and
// patch-hypermeow.sh). Legacy ButtonsMessage and hydrated template
// buttons are NOT included: the server accepts them but current clients
// render nothing.
package buttons

import (
	"context"
	"encoding/json"

	"google.golang.org/protobuf/proto"

	whatsmeow "github.com/polymorfa/hypermeow"
	"github.com/polymorfa/hypermeow/proto/waE2E"
	"github.com/polymorfa/hypermeow/types"
)

// Button is a single native-flow button.
type Button = waE2E.InteractiveMessage_NativeFlowMessage_NativeFlowButton

func rawButton(name string, params any) *Button {
	encoded, _ := json.Marshal(params)
	return &Button{
		Name:             proto.String(name),
		ButtonParamsJSON: proto.String(string(encoded)),
	}
}

// QuickReply is a simple tappable reply button. Tapping it sends a
// native-flow response carrying the given id (see ParseTap).
func QuickReply(displayText, id string) *Button {
	return rawButton("quick_reply", map[string]string{
		"display_text": displayText,
		"id":           id,
	})
}

// CTAURL opens a link in the external browser when tapped.
func CTAURL(displayText, url string) *Button {
	return rawButton("cta_url", map[string]string{
		"display_text": displayText,
		"url":          url,
	})
}

// CTAWebview opens a link inside WhatsApp's in-app browser (full screen,
// behind a "Securely browse this business website" interstitial). It is a
// cta_url button carrying merchant_url + webview hints.
func CTAWebview(displayText, url string) *Button {
	return rawButton("cta_url", map[string]any{
		"display_text":         displayText,
		"url":                  url,
		"merchant_url":         url,
		"webview_presentation": "full",
		"webview_interaction":  true,
		"payment_link_preview": false,
	})
}

// CTACopy copies copyCode to the clipboard when tapped. It produces no
// callback — the bot never learns that it was tapped.
func CTACopy(displayText, copyCode string) *Button {
	return rawButton("cta_copy", map[string]string{
		"display_text": displayText,
		"copy_code":    copyCode,
	})
}

// CTACall dials phoneNumber (international format, e.g. "+972501234567")
// when tapped.
func CTACall(displayText, phoneNumber string) *Button {
	return rawButton("cta_call", map[string]string{
		"display_text": displayText,
		"phone_number": phoneNumber,
	})
}

// Row is one selectable row in a SingleSelect dropdown.
type Row struct {
	ID          string
	Title       string
	Description string
}

// Section groups rows under a title in a SingleSelect dropdown.
type Section struct {
	Title string
	Rows  []Row
}

// SingleSelect is a dropdown list inside a button. Tapping a row sends a
// native-flow response carrying the row id (see ParseTap). Pass several
// SingleSelect buttons to Send to build a settings-panel style message
// with multiple menus. Renders on Android; like all native-flow lists it
// may not render on iOS.
func SingleSelect(title string, sections []Section) *Button {
	secs := make([]map[string]any, 0, len(sections))
	for _, s := range sections {
		rows := make([]map[string]string, 0, len(s.Rows))
		for _, r := range s.Rows {
			rows = append(rows, map[string]string{
				"id":          r.ID,
				"title":       r.Title,
				"description": r.Description,
			})
		}
		secs = append(secs, map[string]any{"title": s.Title, "rows": rows})
	}
	return rawButton("single_select", map[string]any{
		"title":    title,
		"sections": secs,
	})
}

// Message builds an interactive native-flow message. A nameless trailing
// button is appended automatically: it makes HyperMeow tag the biz node
// "mixed", which avoids a 405 error on some accounts when a message mixes
// button types.
func Message(body, footer string, btns ...*Button) *waE2E.Message {
	btns = append(btns, &Button{Name: proto.String("")})
	msg := &waE2E.InteractiveMessage{
		Body: &waE2E.InteractiveMessage_Body{Text: proto.String(body)},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{Buttons: btns},
		},
	}
	if footer != "" {
		msg.Footer = &waE2E.InteractiveMessage_Footer{Text: proto.String(footer)}
	}
	return &waE2E.Message{InteractiveMessage: msg}
}

// Send builds and sends a native-flow button message.
func Send(ctx context.Context, client *whatsmeow.Client, to types.JID, body, footer string, btns ...*Button) (whatsmeow.SendResponse, error) {
	return client.SendMessage(ctx, to, Message(body, footer, btns...))
}

// LocationRequestMessage builds a send_location request message. Two
// details make or break it: the button params must be EMPTY, and the
// message must NOT carry the trailing empty button — that button would
// rename the biz node to "mixed", while a real location request keeps
// the name "send_location". The shared location returns as a normal
// location message.
func LocationRequestMessage(body, footer string) *waE2E.Message {
	msg := &waE2E.InteractiveMessage{
		Body: &waE2E.InteractiveMessage_Body{Text: proto.String(body)},
		InteractiveMessage: &waE2E.InteractiveMessage_NativeFlowMessage_{
			NativeFlowMessage: &waE2E.InteractiveMessage_NativeFlowMessage{Buttons: []*Button{{
				Name:             proto.String("send_location"),
				ButtonParamsJSON: proto.String(""),
			}}},
		},
	}
	if footer != "" {
		msg.Footer = &waE2E.InteractiveMessage_Footer{Text: proto.String(footer)}
	}
	return &waE2E.Message{InteractiveMessage: msg}
}

// SendLocationRequest builds and sends a send_location request message.
func SendLocationRequest(ctx context.Context, client *whatsmeow.Client, to types.JID, body, footer string) (whatsmeow.SendResponse, error) {
	return client.SendMessage(ctx, to, LocationRequestMessage(body, footer))
}

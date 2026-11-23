package buttons

import (
	"context"

	"google.golang.org/protobuf/proto"

	whatsmeow "github.com/polymorfa/hypermeow"
	"github.com/polymorfa/hypermeow/proto/waE2E"
	"github.com/polymorfa/hypermeow/types"
)

// ClassicMessage builds a classic ListMessage — the 2021-era list that is
// still the ONLY dropdown menu WhatsApp's iOS client renders. It reuses
// the Section/Row types from SingleSelect.
//
// REQUIREMENT: the message only renders when the client sends it with a
// <biz><list type="product_list"/> node while the proto itself stays
// SINGLE_SELECT — real WhatsApp/Baileys payloads decouple the two, but
// HyperMeow derives the node type from the proto, so a stock build sends
// type="single_select" and the phone shows nothing. Run patch-hypermeow.sh
// (in this repo) once to apply the one-line library fix, then rebuild.
// Verified live on a real phone 2026-10-21, and the framing was confirmed
// by capturing a real Baileys list stanza (RawNodeHandler): stanza
// type="media", enc mediatype="list", biz list type="product_list",
// proto SINGLE_SELECT.
//
// Size limits (enforced by WhatsApp, counted in BYTES): buttonText 20,
// section titles and row titles 24, row descriptions 72, description
// 1024, footer 60, at most 10 rows in total.
func ClassicMessage(title, description, buttonText, footer string, sections []Section) *waE2E.Message {
	secs := make([]*waE2E.ListMessage_Section, 0, len(sections))
	for _, s := range sections {
		rows := make([]*waE2E.ListMessage_Row, 0, len(s.Rows))
		for _, r := range s.Rows {
			rows = append(rows, &waE2E.ListMessage_Row{
				RowID:       proto.String(r.ID),
				Title:       proto.String(r.Title),
				Description: proto.String(r.Description),
			})
		}
		secs = append(secs, &waE2E.ListMessage_Section{
			Title: proto.String(s.Title),
			Rows:  rows,
		})
	}
	return &waE2E.Message{
		ListMessage: &waE2E.ListMessage{
			Title:       proto.String(title),
			Description: proto.String(description),
			ButtonText:  proto.String(buttonText),
			FooterText:  proto.String(footer),
			ListType:    waE2E.ListMessage_SINGLE_SELECT.Enum(),
			Sections:    secs,
		},
	}
}

// SendClassic sends a classic list built by ClassicMessage. The client
// must run against the patched HyperMeow (see ClassicMessage).
func SendClassic(ctx context.Context, client *whatsmeow.Client, to types.JID, title, description, buttonText, footer string, sections []Section) (whatsmeow.SendResponse, error) {
	return client.SendMessage(ctx, to, ClassicMessage(title, description, buttonText, footer, sections))
}

// ListTapID returns the selected row id from an incoming classic-list
// reply (a ListResponseMessage), or "" if the message is not one.
func ListTapID(m *waE2E.Message) string {
	return m.GetListResponseMessage().GetSingleSelectReply().GetSelectedRowID()
}

package buttons

import (
	"testing"

	"github.com/polymorfa/hypermeow/proto/waE2E"
)

func TestClassicMessageShape(t *testing.T) {
	msg := ClassicMessage("Title", "Desc", "Open", "Foot", []Section{
		{Title: "S1", Rows: []Row{
			{ID: "r1", Title: "Row 1", Description: "first"},
			{ID: "r2", Title: "Row 2", Description: "second"},
		}},
	})
	lm := msg.GetListMessage()
	if lm == nil {
		t.Fatal("message has no ListMessage")
	}
	// The proto MUST stay SINGLE_SELECT; product_list lives only in the
	// biz node (applied by the library patch), never in the proto.
	if lm.GetListType() != waE2E.ListMessage_SINGLE_SELECT {
		t.Fatalf("listType = %v, want SINGLE_SELECT", lm.GetListType())
	}
	if lm.GetTitle() != "Title" || lm.GetButtonText() != "Open" || lm.GetFooterText() != "Foot" {
		t.Fatalf("header fields wrong: %+v", lm)
	}
	secs := lm.GetSections()
	if len(secs) != 1 || len(secs[0].GetRows()) != 2 {
		t.Fatalf("sections/rows wrong: %+v", secs)
	}
	if secs[0].GetRows()[1].GetRowID() != "r2" {
		t.Fatalf("row id = %q, want r2", secs[0].GetRows()[1].GetRowID())
	}
}

func TestListTapID(t *testing.T) {
	m := &waE2E.Message{ListResponseMessage: &waE2E.ListResponseMessage{
		SingleSelectReply: &waE2E.ListResponseMessage_SingleSelectReply{
			SelectedRowID: strPtr("row42"),
		},
	}}
	if got := ListTapID(m); got != "row42" {
		t.Fatalf("ListTapID = %q, want row42", got)
	}
	if got := ListTapID(&waE2E.Message{}); got != "" {
		t.Fatalf("ListTapID on empty message = %q, want empty", got)
	}
}

func strPtr(s string) *string { return &s }

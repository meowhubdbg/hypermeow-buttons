package buttons

import (
	"encoding/json"

	"github.com/polymorfa/hypermeow/proto/waE2E"
)

// Tap is a parsed native-flow button tap: the button name that was tapped
// (e.g. "quick_reply", "single_select"), the raw paramsJSON the client
// sent back, and the selected id extracted from it when present.
type Tap struct {
	Name       string
	ParamsJSON string
	ID         string
}

// ParseTap extracts a native-flow tap from an incoming message. The
// second return value is false when the message is not a native-flow
// response.
func ParseTap(m *waE2E.Message) (Tap, bool) {
	nf := m.GetInteractiveResponseMessage().GetNativeFlowResponseMessage()
	if nf == nil {
		return Tap{}, false
	}
	tap := Tap{Name: nf.GetName(), ParamsJSON: nf.GetParamsJSON()}
	tap.ID = findID(tap.ParamsJSON)
	return tap, true
}

// TapID is a convenience wrapper: just the selected id, or "".
func TapID(m *waE2E.Message) string {
	tap, ok := ParseTap(m)
	if !ok {
		return ""
	}
	return tap.ID
}

// findID digs the selected id out of a tap's paramsJSON. quick_reply taps
// carry a top-level "id"; dropdown selections nest it in different shapes
// across clients, so a few known key names are searched recursively.
func findID(paramsJSON string) string {
	var root any
	if json.Unmarshal([]byte(paramsJSON), &root) != nil {
		return ""
	}
	for _, key := range []string{"id", "selected_row_id", "row_id", "selected_id"} {
		if id := searchKey(root, key); id != "" {
			return id
		}
	}
	return ""
}

func searchKey(v any, key string) string {
	switch node := v.(type) {
	case map[string]any:
		if raw, ok := node[key]; ok {
			if s, ok := raw.(string); ok && s != "" {
				return s
			}
		}
		for _, child := range node {
			if s := searchKey(child, key); s != "" {
				return s
			}
		}
	case []any:
		for _, child := range node {
			if s := searchKey(child, key); s != "" {
				return s
			}
		}
	}
	return ""
}

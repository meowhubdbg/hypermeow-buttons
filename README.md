# hypermeow-buttons

Native-flow buttons for [HyperMeow](https://github.com/polymorfa/hypermeow) bots — quick replies, link buttons, in-app webview, copy/call buttons, dropdown menus, location requests, and classic lists (the iOS-safe menu, with a one-line library patch).

**Only button types that actually work are included.** Everything here was live-tested on a real phone (October 2026): sending, rendering and tap callbacks.

```bash
go get github.com/meowhubdbg/hypermeow-buttons
```

## Quick start

```go
import buttons "github.com/meowhubdbg/hypermeow-buttons"

// Send a message with buttons
_, err := buttons.Send(ctx, client, chatJID,
    "Pick one:", "footer text",
    buttons.QuickReply("Option A", "a"),
    buttons.QuickReply("Option B", "b"),
    buttons.SingleSelect("More options", []buttons.Section{{
        Title: "More options",
        Rows:  []buttons.Row{{ID: "c", Title: "Option C", Description: "desc"}},
    }}),
    buttons.CTAWebview("Open site", "https://example.com"),
    buttons.CTACopy("Copy code", "ABC123"),
)

// Read taps back in your message handler
if tap, ok := buttons.ParseTap(evt.Message); ok {
    fmt.Println("tapped:", tap.Name, "id:", tap.ID, "raw:", tap.ParamsJSON)
}
```

A runnable bot using all of it lives in [`example/`](example/).

## API

| Function | What it does |
|---|---|
| `QuickReply(displayText, id)` | Reply button; tapping sends the `id` back |
| `CTAURL(displayText, url)` | Opens the link in the external browser |
| `CTAWebview(displayText, url)` | Opens the link in WhatsApp's in-app browser, full screen |
| `CTACopy(displayText, copyCode)` | Copies to clipboard (no callback) |
| `CTACall(displayText, phoneNumber)` | Dials the number |
| `SingleSelect(title, sections)` | Dropdown list in a button; pass several to build multi-menu panels |
| `Send(ctx, client, to, body, footer, btns...)` | Builds and sends a button message |
| `Message(body, footer, btns...)` | Builds the raw message without sending |
| `SendLocationRequest(ctx, client, to, body, footer)` | Asks the user to share their location |
| `ParseTap(msg)` / `TapID(msg)` | Extracts button name + selected id from an incoming tap |

## What works (live-tested)

- ✅ `quick_reply` — taps return the button `id`
- ✅ `cta_url` — external browser
- ✅ `cta_url` webview — in-app browser (`merchant_url` + `webview_presentation: "full"`)
- ✅ `cta_copy`, `cta_call`
- ✅ `single_select` dropdowns, including several menus in one message (settings-panel style, where each row id is a bot command)
- ✅ `send_location` requests — the shared location returns as a normal location message
- ✅ All of the above mixed in a single message

Two details that make or break it, both handled for you:

- **The trailing empty button.** `Message`/`Send` append a nameless button, which makes HyperMeow tag the biz node `"mixed"` and avoids a 405 error on some accounts when button types are mixed.
- **Location requests are the exception.** A `send_location` message must carry *empty* button params and *no* trailing empty button, or the client ignores it. `SendLocationRequest` does exactly that.

## Classic lists — the iOS-safe menu (needs a one-line library patch)

The classic `ListMessage` is the only dropdown menu WhatsApp's **iOS** client renders. It works from HyperMeow too — live-verified on a real phone — but only with a one-line patch to the library:

```bash
./patch-hypermeow.sh /path/to/your/project   # then: go mod tidy && go build .
```

```go
_, err := buttons.SendClassic(ctx, client, chatJID,
    \"Title\", \"Pick one:\", \"Open menu\", \"footer\",
    []buttons.Section{{Title: \"Options\", Rows: []buttons.Row{
        {ID: \"a\", Title: \"Option A\", Description: \"first\"},
        {ID: \"b\", Title: \"Option B\", Description: \"second\"},
    }}})

// Taps come back as ListResponseMessage:
if id := buttons.ListTapID(evt.Message); id != \"\" {
    fmt.Println(\"picked:\", id)
}
```

**Why the patch:** real WhatsApp/Baileys payloads decouple two values — the proto says `SINGLE_SELECT` while the stanza's `<biz><list>` node says `type=\"product_list\"`. HyperMeow derives the node type from the proto, so stock builds send `single_select` and phones show nothing. This was proven by capturing a real Baileys list stanza with a RawNodeHandler (stanza `type=\"media\"`, enc `mediatype=\"list\"`, biz node `product_list`, proto `SINGLE_SELECT`) and diffing it against HyperMeow's output: exactly one value differed. The patch hardcodes that one value in `getButtonAttributes`. Approaches that did NOT work: the library builder as-is, raw proto with `ListType=PRODUCT_LIST` (poisons the proto), injecting a second biz node (server error 479), changing the stanza type/mediatype (they were already correct).

Size limits in bytes: button text 20, section/row titles 24, row descriptions 72, description 1024, footer 60, max 10 rows total.

## Not included — they don't work

Legacy `ButtonsMessage` and hydrated template buttons were tested too: the WhatsApp server accepts them, but current clients render nothing.

Note: native-flow dropdowns (`SingleSelect`) render on Android but not on iOS — for iOS menus use the classic list above.

## License

MIT

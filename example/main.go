// Minimal demo bot for hypermeow-buttons: pair it, then send "buttons"
// in any 1:1 chat. It replies with one message using every button type,
// plus a location request, and echoes back whatever gets tapped.
package main

import (
	"bufio"
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"google.golang.org/protobuf/proto"

	whatsmeow "github.com/polymorfa/hypermeow"
	"github.com/polymorfa/hypermeow/proto/waE2E"
	"github.com/polymorfa/hypermeow/store/sqlstore"
	"github.com/polymorfa/hypermeow/types"
	"github.com/polymorfa/hypermeow/types/events"
	waLog "github.com/polymorfa/hypermeow/util/log"
	_ "modernc.org/sqlite"

	buttons "github.com/meowhubdbg/hypermeow-buttons"
)

var cli *whatsmeow.Client

func sendDemo(ctx context.Context, to types.JID) {
	_, err := buttons.Send(ctx, cli, to,
		"Pick one:", "hypermeow-buttons demo",
		buttons.QuickReply("Option A", "a"),
		buttons.QuickReply("Option B", "b"),
		buttons.SingleSelect("More options", []buttons.Section{{
			Title: "More options",
			Rows: []buttons.Row{
				{ID: "c", Title: "Option C", Description: "from the dropdown"},
			},
		}}),
		buttons.CTAWebview("Open in WhatsApp", "https://example.com"),
		buttons.CTAURL("Open externally", "https://example.com"),
		buttons.CTACopy("Copy code", "ABC123"),
	)
	if err != nil {
		fmt.Println("send demo failed:", err)
		return
	}
	_, err = buttons.SendLocationRequest(ctx, cli, to, "Or share your location:", "hypermeow-buttons demo")
	if err != nil {
		fmt.Println("send location request failed:", err)
	}
}

func handleMessage(ctx context.Context, evt *events.Message) {
	m := evt.Message
	if m == nil || evt.Info.Chat.Server == types.NewsletterServer {
		return
	}
	chat := evt.Info.Chat
	if tap, ok := buttons.ParseTap(m); ok {
		fmt.Printf("tap: name=%q id=%q params=%s\n", tap.Name, tap.ID, tap.ParamsJSON)
		_, _ = cli.SendMessage(ctx, chat, &waE2E.Message{Conversation: proto.String(fmt.Sprintf("You tapped: %s (id: %s)", tap.Name, tap.ID))})
		return
	}
	text := m.GetConversation()
	if text == "" {
		text = m.GetExtendedTextMessage().GetText()
	}
	if strings.TrimSpace(text) == "buttons" {
		sendDemo(ctx, chat)
	}
}

func main() {
	ctx := context.Background()
	_ = os.MkdirAll("auth", 0755)
	sqlDB, err := sql.Open("sqlite", "file:"+filepath.Join("auth", "hypermeow.db")+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(10000)")
	if err != nil {
		panic(err)
	}
	sqlDB.SetMaxOpenConns(1)
	container := sqlstore.NewWithDB(sqlDB, "sqlite", waLog.Stdout("DB", "ERROR", true))
	if err := container.Upgrade(ctx); err != nil {
		panic(err)
	}
	deviceStore, err := container.GetFirstDevice(ctx)
	if err != nil {
		panic(err)
	}
	cli = whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "ERROR", true))
	cli.AddEventHandler(func(evt any) {
		switch v := evt.(type) {
		case *events.Message:
			go handleMessage(ctx, v)
		case *events.Connected:
			fmt.Println("connected — send \"buttons\" in a chat")
		case *events.LoggedOut:
			fmt.Println("logged out — delete the auth dir and pair again")
			os.Exit(1)
		}
	})

	if cli.Store.ID == nil {
		qrChan, _ := cli.GetQRChannel(ctx)
		if err := cli.Connect(); err != nil {
			panic(err)
		}
		select {
		case <-qrChan:
		case <-time.After(15 * time.Second):
		}
		fmt.Print("Phone number (international format, e.g. 9725XXXXXXXX): ")
		phone, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		phone = strings.Map(func(r rune) rune {
			if r >= '0' && r <= '9' {
				return r
			}
			return -1
		}, phone)
		time.Sleep(3 * time.Second)
		code, err := cli.PairPhone(ctx, phone, true, whatsmeow.PairClientChrome, "Chrome (Linux)")
		if err != nil {
			panic(err)
		}
		fmt.Println("Pairing code:", code)
	} else if err := cli.Connect(); err != nil {
		panic(err)
	}
	select {}
}

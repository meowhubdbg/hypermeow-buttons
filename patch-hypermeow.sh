#!/usr/bin/env bash
# One-line patch that makes classic ListMessages render with HyperMeow,
# including on iOS (verified live on a real phone, 2026-10-21).
#
# Why: real WhatsApp/Baileys payloads send a ListMessage whose proto says
# SINGLE_SELECT while the stanza's <biz><list> node says type="product_list"
# — the two are decoupled. HyperMeow derives the node type from the proto,
# so it sends type="single_select" and phones render nothing. (Confirmed by
# capturing a real Baileys stanza: type="media", enc mediatype="list",
# biz list type="product_list", proto SINGLE_SELECT.)
#
# This script copies hypermeow out of your Go module cache, changes that
# ONE line in send.go, and points your project at the copy with a replace
# directive. The module cache itself is never touched.
# Undo: go mod edit -dropreplace github.com/polymorfa/hypermeow
#
# Usage: ./patch-hypermeow.sh [path-to-your-go-project]   (default: $PWD)
set -e

PROJ="${1:-$PWD}"
if [ ! -f "$PROJ/go.mod" ]; then
  echo "ERROR: no go.mod in $PROJ — pass your project dir as an argument"
  exit 1
fi

GOMODCACHE=$(go env GOMODCACHE)
SRC=$(ls -d "$GOMODCACHE"/github.com/polymorfa/hypermeow@* | tail -1)
DST="$HOME/hypermeow-patched"
echo "library source: $SRC"
echo "patched copy:   $DST"

rm -rf "$DST"
cp -r "$SRC" "$DST"
chmod -R u+w "$DST"
SEND="$DST/send.go"

sed -i 's|"type": strings.ToLower(waE2E.ListMessage_ListType_name\[int32(msg.ListMessage.GetListType())\]),|"type": "product_list",|' "$SEND"

if ! grep -q '"type": "product_list",' "$SEND"; then
  echo "ERROR: patch did not land — your hypermeow version differs; the line lives in getButtonAttributes in send.go"
  exit 1
fi
DIFFLINES=$(diff "$SRC/send.go" "$SEND" | grep -c '^[<>]' || true)
if [ "$DIFFLINES" != "2" ]; then
  echo "ERROR: expected exactly 1 changed line, diff shows $DIFFLINES lines"
  exit 1
fi
echo "patch verified: exactly 1 line changed ✓"

cd "$PROJ"
go mod edit -replace github.com/polymorfa/hypermeow="$DST"
echo "done ✓ — now run: go mod tidy && go build ."

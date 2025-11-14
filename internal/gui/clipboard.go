package gui

import (
	"fmt"
	"io"
	"strings"

	"gioui.org/app"
	"gioui.org/io/clipboard"
	"gioui.org/io/event"
	"gioui.org/io/transfer"
	"gioui.org/layout"
)

// readClipboard initiates a clipboard read request
// The actual content will be received as a transfer.DataEvent in the event queue
// Caller should handle transfer.DataEvent to get the actual clipboard content
func readClipboard(w *app.Window, gtx layout.Context) {
	// Request clipboard read operation
	// The tag is used to route the clipboard data back to the handler
	event.Op(gtx.Ops, w)
	gtx.Execute(clipboard.ReadCmd{Tag: w})
}

// writeClipboard writes text content to the system clipboard
// Returns an error if the text is empty, otherwise queues the write operation
func writeClipboard(w *app.Window, gtx layout.Context, text string) error {
	if text == "" {
		return fmt.Errorf("cannot write empty text to clipboard")
	}

	// Create a ReadCloser from the text string
	reader := io.NopCloser(strings.NewReader(text))

	// Perform clipboard write operation with application/text MIME type
	// (matching the pattern used in Gio's Editor widget)
	gtx.Execute(clipboard.WriteCmd{
		Type: "application/text",
		Data: reader,
	})

	return nil
}

// handleClipboardEvent processes transfer.DataEvent and returns the clipboard text if available
// Returns empty string if no clipboard data is available or if an error occurs
func handleClipboardEvent(events []event.Event) string {
	for _, e := range events {
		if de, ok := e.(transfer.DataEvent); ok {
			// Check if it's text data (support both text/plain and application/text)
			if de.Type == "text/plain" || de.Type == "text/plain;charset=utf-8" || de.Type == "application/text" {
				// Open the data reader
				reader := de.Open()
				if reader == nil {
					continue
				}
				defer reader.Close()

				// Read the clipboard text
				data, err := io.ReadAll(reader)
				if err != nil {
					// Error reading clipboard, return empty string
					return ""
				}

				return string(data)
			}
		}
	}
	return ""
}

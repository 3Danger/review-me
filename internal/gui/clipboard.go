package gui

import (
	"fmt"
	"io"
	"strings"

	"gioui.org/app"
	"gioui.org/io/clipboard"
	"gioui.org/io/transfer"
	"gioui.org/layout"
)

// ExtractClipboardText extracts text from a transfer.DataEvent
// Supports MIME types: text/plain, text/plain;charset=utf-8, application/text
func ExtractClipboardText(de transfer.DataEvent) string {
	// Check if it's text data
	if de.Type == "text/plain" || de.Type == "text/plain;charset=utf-8" || de.Type == "application/text" {
		reader := de.Open()
		if reader == nil {
			return ""
		}
		defer reader.Close()

		// Read the clipboard text
		data, err := io.ReadAll(reader)
		if err != nil {
			return ""
		}

		return string(data)
	}
	return ""
}

// readClipboard initiates a clipboard read request
// The actual content will be received as a transfer.DataEvent in the event queue
// Caller should handle transfer.DataEvent to get the actual clipboard content
func readClipboard(gtx layout.Context, w *app.Window) {
	// Request clipboard read operation with window as target
	gtx.Execute(clipboard.ReadCmd{Tag: w})
}

// writeClipboard writes text content to the system clipboard
// Returns an error if the text is empty, otherwise queues the write operation
func writeClipboard(gtx layout.Context, text string) error {
	if text == "" {
		return fmt.Errorf("невозможно записать пустой текст в буфер обмена")
	}

	// Create a ReadCloser from the text string
	reader := io.NopCloser(strings.NewReader(text))

	// Perform clipboard write operation with application/text MIME type
	// (matching the pattern used in Gio's Editor widget)
	gtx.Execute(clipboard.WriteCmd{
		Type: "text/plain",
		Data: reader,
	})

	return nil
}

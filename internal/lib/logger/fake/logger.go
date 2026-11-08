package fake

import (
	"io"
	"log/slog"
)

var Logger = slog.New(
	slog.NewTextHandler(
		io.Discard,
		nil,
	),
)

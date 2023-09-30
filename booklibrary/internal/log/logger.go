package log

import (
	"io"
	"log/slog"
)

func New(w io.Writer, debug bool) *slog.Logger {
	opts := slog.HandlerOptions{
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Time(a.Key, a.Value.Time().UTC())
			}
			return a
		},
	}
	if debug {
		opts.Level = slog.LevelDebug
	}

	return slog.New(slog.NewTextHandler(w, &opts))
}

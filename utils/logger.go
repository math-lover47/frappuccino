package utils

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func NewLogger() (*slog.Logger, *os.File) {
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o755)
	if err != nil {
		fmt.Println(err)

		return nil, nil
	}

	out := io.MultiWriter(file, os.Stdout)
	logger := slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{}))
	return logger, file
}

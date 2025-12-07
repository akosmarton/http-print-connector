package main

import (
	"bufio"
	"io"
	"log/slog"
	"net"
	"net/url"
	"os"
	"time"

	"resty.dev/v3"
)

func main() {
	slog.Info("HTTP Print Connector started")

	apiUrl := os.Getenv("API_URL")
	if apiUrl == "" {
		slog.Error("API_URL environment variable is not set")
		os.Exit(1)
	}

	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		slog.Error("API_KEY environment variable is not set")
		os.Exit(1)
	}

	printer_url := os.Getenv("PRINTER_URL")
	if printer_url == "" {
		slog.Error("PRINTER_URL environment variable is not set")
		os.Exit(1)
	}

	purl, err := url.ParseRequestURI(printer_url)
	if err != nil {
		slog.Error("Invalid PRINTER_URL", "err", err)
		os.Exit(1)
	}

	if purl.Scheme != "file" && purl.Scheme != "tcp" {
		slog.Error("Unsupported PRINTER_URL", "scheme", purl.Scheme)
		os.Exit(1)
	}

	client := resty.New()
	defer client.Close()
	client.SetAuthToken(apiKey)

	for {
		slog.Info("Waiting for new print job...")
		res, err := client.R().Get(apiUrl)
		if err != nil {
			slog.Error("Request failed", "err", err)
			time.Sleep(time.Second * 5)
			continue
		}
		if res.IsSuccess() {
			slog.Info("Print job received", "size", res.Size())
			if err := print(purl, res.Body); err != nil {
				slog.Error("Print failed", "err", err)
			} else {
				slog.Info("Print successful", "size", res.Size())
			}
		} else {
			slog.Info("Request failed", "status", res.Status())
			time.Sleep(time.Second * 5)
			continue
		}
	}
}

func print(purl *url.URL, payload io.ReadCloser) error {
	switch purl.Scheme {
	case "file":
		f, err := os.OpenFile(purl.Path, os.O_WRONLY, 0)
		if err != nil {
			slog.Error("File open", "err", err)
		}
		defer f.Close()
		w := bufio.NewWriter(f)
		defer w.Flush()
		if _, err := io.Copy(w, payload); err != nil {
			return err
		}
	case "tcp":
		conn, err := net.Dial("tcp", purl.Host)
		if err != nil {
			return err
		}
		defer conn.Close()
		if _, err := io.Copy(conn, payload); err != nil {
			return err
		}
	}
	return nil
}

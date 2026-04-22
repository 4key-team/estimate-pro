// Copyright 2026 Daniil Vdovin. All rights reserved.
// SPDX-License-Identifier: AGPL-3.0-only

package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/VDV001/estimate-pro/backend/internal/modules/bot/domain"
	"golang.org/x/net/proxy"
)

// Client is a Telegram Bot API client.
type Client struct {
	token      string
	httpClient *http.Client
	baseURL    string
}

// NewClient creates a new Telegram Bot API client.
func NewClient(token string) *Client {
	client := &Client{
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    fmt.Sprintf("https://api.telegram.org/bot%s", token),
	}
	logClientInitialization(nil)
	return client
}

// NewClientWithProxy creates a new Telegram Bot API client with optional SOCKS5 proxy support.
func NewClientWithProxy(token, proxyURL string) (*Client, error) {
	httpClient, err := newHTTPClient(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("telegram.NewClientWithProxy: %w", err)
	}

	var proxyConfig *url.URL
	if proxyURL != "" {
		proxyConfig, err = url.Parse(proxyURL)
		if err != nil {
			return nil, fmt.Errorf("telegram.NewClientWithProxy: parse proxy URL for logging: %w", err)
		}
	}

	client := &Client{
		token:      token,
		httpClient: httpClient,
		baseURL:    fmt.Sprintf("https://api.telegram.org/bot%s", token),
	}
	logClientInitialization(proxyConfig)
	return client, nil
}

func newHTTPClient(proxyURL string) (*http.Client, error) {
	if proxyURL == "" {
		return &http.Client{Timeout: 30 * time.Second}, nil
	}

	u, err := url.Parse(proxyURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy URL: %w", err)
	}
	if u.Host == "" {
		return nil, fmt.Errorf("proxy URL host is required")
	}
	if u.Scheme != "socks5" && u.Scheme != "socks5h" {
		return nil, fmt.Errorf("unsupported proxy scheme %q", u.Scheme)
	}

	var auth *proxy.Auth
	if u.User != nil {
		password, _ := u.User.Password()
		auth = &proxy.Auth{
			User:     u.User.Username(),
			Password: password,
		}
	}

	forward := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	dialer, err := proxy.SOCKS5("tcp", u.Host, auth, forward)
	if err != nil {
		return nil, fmt.Errorf("create SOCKS5 dialer: %w", err)
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.Proxy = nil
	transport.Dial = nil
	if contextDialer, ok := dialer.(proxy.ContextDialer); ok {
		transport.DialContext = contextDialer.DialContext
	} else {
		transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}
	}

	return &http.Client{
		Timeout:   30 * time.Second,
		Transport: transport,
	}, nil
}

func logClientInitialization(proxyConfig *url.URL) {
	attrs := []any{
		slog.Duration("timeout", 30*time.Second),
	}

	if proxyConfig == nil {
		attrs = append(attrs, slog.Bool("proxy_enabled", false))
		slog.Info("telegram.Client initialized", attrs...)
		return
	}

	username := ""
	hasPassword := false
	if proxyConfig.User != nil {
		username = proxyConfig.User.Username()
		_, hasPassword = proxyConfig.User.Password()
	}

	attrs = append(attrs,
		slog.Bool("proxy_enabled", true),
		slog.String("proxy_scheme", proxyConfig.Scheme),
		slog.String("proxy_host", proxyConfig.Host),
		slog.String("proxy_username", username),
		slog.Bool("proxy_has_password", hasPassword),
	)
	slog.Info("telegram.Client initialized", attrs...)
}

// SendMessage sends a plain text message to a chat.
func (c *Client) SendMessage(ctx context.Context, chatID, text string) error {
	slog.DebugContext(ctx, "telegram.SendMessage", slog.String("chat_id", chatID), slog.Int("text_len", len(text)))
	req := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
	}
	_, err := c.doRequest(ctx, "sendMessage", req)
	if err != nil {
		slog.ErrorContext(ctx, "telegram.SendMessage failed", slog.String("chat_id", chatID), slog.String("error", err.Error()))
		return fmt.Errorf("telegram.Client.SendMessage: %w", err)
	}
	return nil
}

// SendMarkdown sends a Markdown-formatted message to a chat.
func (c *Client) SendMarkdown(ctx context.Context, chatID, text string) error {
	slog.DebugContext(ctx, "telegram.SendMarkdown", slog.String("chat_id", chatID), slog.Int("text_len", len(text)))
	req := SendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	}
	_, err := c.doRequest(ctx, "sendMessage", req)
	if err != nil {
		slog.ErrorContext(ctx, "telegram.SendMarkdown failed", slog.String("chat_id", chatID), slog.String("error", err.Error()))
		return fmt.Errorf("telegram.Client.SendMarkdown: %w", err)
	}
	return nil
}

// SendInlineKeyboard sends a message with an inline keyboard to a chat.
func (c *Client) SendInlineKeyboard(ctx context.Context, chatID, text string, keyboard [][]domain.InlineKeyboardButton) error {
	slog.DebugContext(ctx, "telegram.SendInlineKeyboard", slog.String("chat_id", chatID), slog.Int("rows", len(keyboard)))
	// Convert domain buttons to telegram wire format.
	tgKeyboard := make([][]InlineKeyboardButton, len(keyboard))
	for i, row := range keyboard {
		tgKeyboard[i] = make([]InlineKeyboardButton, len(row))
		for j, btn := range row {
			tgKeyboard[i][j] = InlineKeyboardButton{
				Text:         btn.Text,
				CallbackData: btn.CallbackData,
			}
		}
	}
	req := SendMessageRequest{
		ChatID: chatID,
		Text:   text,
		ReplyMarkup: &InlineKeyboardMarkup{
			InlineKeyboard: tgKeyboard,
		},
	}
	_, err := c.doRequest(ctx, "sendMessage", req)
	if err != nil {
		slog.ErrorContext(ctx, "telegram.SendInlineKeyboard failed", slog.String("chat_id", chatID), slog.String("error", err.Error()))
		return fmt.Errorf("telegram.Client.SendInlineKeyboard: %w", err)
	}
	return nil
}

// AnswerCallbackQuery answers an incoming callback query.
func (c *Client) AnswerCallbackQuery(ctx context.Context, callbackQueryID, text string) error {
	slog.DebugContext(ctx, "telegram.AnswerCallbackQuery", slog.String("callback_query_id", callbackQueryID))
	payload := struct {
		CallbackQueryID string `json:"callback_query_id"`
		Text            string `json:"text,omitempty"`
	}{
		CallbackQueryID: callbackQueryID,
		Text:            text,
	}
	_, err := c.doRequest(ctx, "answerCallbackQuery", payload)
	if err != nil {
		slog.ErrorContext(ctx, "telegram.AnswerCallbackQuery failed", slog.String("callback_query_id", callbackQueryID), slog.String("error", err.Error()))
		return fmt.Errorf("telegram.Client.AnswerCallbackQuery: %w", err)
	}
	return nil
}

// SetReaction sets an emoji reaction on a message.
func (c *Client) SetReaction(ctx context.Context, chatID string, messageID int64, emoji string) error {
	slog.DebugContext(ctx, "telegram.SetReaction", slog.String("chat_id", chatID), slog.Int64("message_id", messageID), slog.String("emoji", emoji))
	payload := struct {
		ChatID    string `json:"chat_id"`
		MessageID int64  `json:"message_id"`
		Reaction  []struct {
			Type  string `json:"type"`
			Emoji string `json:"emoji"`
		} `json:"reaction"`
	}{
		ChatID:    chatID,
		MessageID: messageID,
		Reaction: []struct {
			Type  string `json:"type"`
			Emoji string `json:"emoji"`
		}{{Type: "emoji", Emoji: emoji}},
	}
	_, err := c.doRequest(ctx, "setMessageReaction", payload)
	if err != nil {
		slog.WarnContext(ctx, "telegram.SetReaction failed", slog.String("chat_id", chatID), slog.String("error", err.Error()))
		return fmt.Errorf("telegram.Client.SetReaction: %w", err)
	}
	return nil
}

// GetFileURL retrieves the download URL for a file by its file ID.
func (c *Client) GetFileURL(ctx context.Context, fileID string) (string, error) {
	slog.DebugContext(ctx, "telegram.GetFileURL", slog.String("file_id", fileID))
	payload := struct {
		FileID string `json:"file_id"`
	}{
		FileID: fileID,
	}
	resp, err := c.doRequest(ctx, "getFile", payload)
	if err != nil {
		return "", fmt.Errorf("telegram.Client.GetFileURL: %w", err)
	}

	var fileResp struct {
		FilePath string `json:"file_path"`
	}
	if err := json.Unmarshal(resp.Result, &fileResp); err != nil {
		return "", fmt.Errorf("telegram.Client.GetFileURL: unmarshal result: %w", err)
	}

	downloadURL := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", c.token, fileResp.FilePath)
	slog.DebugContext(ctx, "telegram.GetFileURL resolved", slog.String("file_id", fileID), slog.String("file_path", fileResp.FilePath))
	return downloadURL, nil
}

// DownloadFile downloads a file from the given URL and returns its content.
func (c *Client) DownloadFile(ctx context.Context, url string) ([]byte, error) {
	slog.DebugContext(ctx, "telegram.DownloadFile", slog.String("url", url))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("telegram.Client.DownloadFile: create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("telegram.Client.DownloadFile: execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram.Client.DownloadFile: unexpected status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("telegram.Client.DownloadFile: read body: %w", err)
	}
	slog.DebugContext(ctx, "telegram.DownloadFile complete", slog.Int("bytes", len(data)))
	return data, nil
}

// doRequest marshals the payload, sends a POST request to the Telegram Bot API,
// and returns the parsed API response.
func (c *Client) doRequest(ctx context.Context, method string, payload any) (*APIResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/"+method, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if !apiResp.OK {
		slog.WarnContext(ctx, "telegram.doRequest: API error", slog.String("method", method), slog.String("description", apiResp.Description))
		return nil, fmt.Errorf("API error: %s", apiResp.Description)
	}

	return &apiResp, nil
}

package main

import (
	"aienvoy/internal/pkg/config"
	"aienvoy/pkg/claudeweb"
	"aienvoy/pkg/cookiecloud"
	"log/slog"
)

func main() {
	cfg := config.GetConfig().CookieCloud
	cc := cookiecloud.New(cfg.Host, cfg.UUID, cfg.Pass)

	sessionKey, err := cc.GetCookie("claude.ai", "sessionKey")
	if err != nil {
		slog.Error("get cookie error", "err", err)
		return
	}

	claude := claudeweb.NewClaudeWeb(sessionKey.Value)
	cov, err := claude.CreateConversation("new conversation")
	if err != nil {
		slog.Error("create claude conversation error", "err", err)
		return
	}
	prompt := "what's the latest news"
	resp, err := claude.CreateChatMessage(cov.UUID, prompt)
	if err != nil {
		slog.Error("chat with claude error", "err", err)
		return
	}

	slog.Info("claude response message", "message", resp.Completion)
}
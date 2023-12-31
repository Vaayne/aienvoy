package handler

import (
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llm/bard"
	"github.com/Vaayne/aienvoy/pkg/llm/claude"
	"github.com/Vaayne/aienvoy/pkg/llm/claudeweb"
	"github.com/Vaayne/aienvoy/pkg/llm/openai"

	tb "gopkg.in/telebot.v3"
)

const (
	CommandBard          = "bard"
	CommandRead          = "read"
	CommandChatGPT35     = "gpt35"
	CommandChatGPT4      = "gpt4"
	CommandClaudeWeb     = "claude_web"
	CommandClaudeV2      = "claude_v2"
	CommandClaudeV1      = "claude_v1"
	CommandClaudeInstant = "claude_instant"
	CommandImagine       = "imagine"
)

func OnText(c tb.Context) error {
	text := strings.TrimSpace(c.Text())
	if text == "" {
		return c.Reply("empty message")
	}

	model := ""
	prompt := text
	if text[0] == '/' {
		texts := strings.Split(text, " ")
		model = texts[0][1:]
		if len(texts) > 1 {
			prompt = strings.Join(texts[1:], " ")
		} else {
			prompt = "hello"
		}

		switch model {
		case CommandBard:
			model = bard.ModelBard
		case CommandRead:
			return OnReadEase(c)
		case CommandChatGPT35:
			model = openai.ModelGPT3Dot5Turbo
		case CommandChatGPT4:
			model = openai.ModelGPT4
		case CommandClaudeWeb:
			model = claudeweb.ModelClaude2
		case CommandClaudeV2:
			model = claude.ModelClaudeV2
		case CommandClaudeV1:
			model = claude.ModelClaudeV1Dot3
		case CommandClaudeInstant:
			model = claude.ModelClaudeInstantV1Dot2
		case CommandImagine:
			return OnMidJourneyImagine(c)
		default:
			return c.Reply("Unsupported command!")
		}
	}

	// 1. create new conversation, no cache, model != ""
	// 2. create new conversation, use cache, model != ""
	// 3. use cache, model == ""
	llmCache, ok := getLLMConversationFromCache()
	if ok {
		if model == "" {
			model = llmCache.Model
		}
		return onLLMChat(c, llmCache.ConversationId, model, prompt)
	}
	if model != "" {
		return onLLMChat(c, "", model, prompt)
	}

	return c.Reply("Unsupported message")
}

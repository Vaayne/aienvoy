package awsbedrock

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/Vaayne/aienvoy/pkg/llm"
	llmconfig "github.com/Vaayne/aienvoy/pkg/llm/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime/types"
)

type Client struct {
	*bedrockruntime.Client
	config llmconfig.Config
}

func NewClient(cfg llmconfig.Config) (*Client, error) {
	if cfg.LLMType != llmconfig.LLMTypeAWSBedrock {
		return nil, fmt.Errorf("invalid config for bedrock, llmtype: %s", cfg.LLMType)
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	ab := cfg.AWSBedrock
	awsConfig, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion(ab.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			ab.AccessKey,
			ab.SecretKey,
			"",
		)))
	if err != nil {
		return nil, fmt.Errorf("get aws config error: %w", err)
	}
	return &Client{
		Client: bedrockruntime.NewFromConfig(awsConfig),
		config: cfg,
	}, nil
}

func (c *Client) ListModels() []string {
	if c.config.Models == nil || len(c.config.Models) == 0 {
		c.config.Models = c.config.AWSBedrock.ListModels()
	}
	return c.config.Models
}

func (c *Client) CreateChatCompletion(ctx context.Context, req llm.ChatCompletionRequest) (llm.ChatCompletionResponse, error) {
	slog.DebugContext(ctx, "chat start", "model", req.Model, "is_stream", false)
	bedrockRequest := &BedrockRequest{}
	bedrockRequest.FromChatCompletionRequest(req)

	output, err := c.InvokeModel(ctx, &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(req.Model),
		Body:        bedrockRequest.Marshal(),
		Accept:      aws.String("application/json"),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		slog.ErrorContext(ctx, "chat start", "model", req.Model, "is_stream", false, "err", err)
		return llm.ChatCompletionResponse{}, err
	}
	resp := &BedrockResponse{}
	resp.Unmarshal(output.Body)
	slog.DebugContext(ctx, "chat success", "model", req.Model, "is_stream", false)
	return resp.ToChatCompletionResponse(), nil
}

func (c *Client) CreateChatCompletionStream(ctx context.Context, req llm.ChatCompletionRequest, dataChan chan llm.ChatCompletionStreamResponse, errChan chan error) {
	slog.DebugContext(ctx, "chat start", "model", req.Model, "is_stream", true)
	bedrockRequest := &BedrockRequest{}
	bedrockRequest.FromChatCompletionRequest(req)

	output, err := c.InvokeModelWithResponseStream(ctx, &bedrockruntime.InvokeModelWithResponseStreamInput{
		ModelId:     aws.String(req.Model),
		Body:        bedrockRequest.Marshal(),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		errChan <- err
		return
	}

	sb := &strings.Builder{}

	for event := range output.GetStream().Events() {
		switch v := event.(type) {
		case *types.ResponseStreamMemberChunk:
			var resp BedrockResponse
			err := json.NewDecoder(bytes.NewReader(v.Value.Bytes)).Decode(&resp)
			if err != nil {
				slog.ErrorContext(ctx, "chat start", "model", req.Model, "is_stream", true, "err", err)
				errChan <- err
				return
			}
			sb.WriteString(resp.Completion)
			dataChan <- resp.ToChatCompletionStreamResponse()
		case *types.UnknownUnionMember:
			err = fmt.Errorf("unknown event type: %T", v)
			slog.ErrorContext(ctx, "chat start", "model", req.Model, "is_stream", true, "err", err)
			errChan <- err
			return
		default:
			err = fmt.Errorf("unknown event type: %T", v)
			slog.ErrorContext(ctx, "chat start", "model", req.Model, "is_stream", true, "err", err)
			errChan <- err
			return
		}
	}
	slog.DebugContext(ctx, "chat success", "model", req.Model, "is_stream", true)
	errChan <- io.EOF
}

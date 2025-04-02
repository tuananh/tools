package textsplitter

import (
	"fmt"
	"log/slog"

	"dario.cat/mergo"
	"github.com/obot-platform/tools/knowledge/pkg/datastore/defaults"
	"github.com/obot-platform/tools/knowledge/pkg/datastore/embeddings/load"
	dstypes "github.com/obot-platform/tools/knowledge/pkg/datastore/types"
	vs "github.com/obot-platform/tools/knowledge/pkg/vectorstore/types"
	"github.com/mitchellh/mapstructure"
	lcgosplitter "github.com/tmc/langchaingo/textsplitter"
)

type SplitterFunc func([]vs.Document) ([]vs.Document, error)

type TextSplitterOpts struct {
	ChunkSize    int    `json:"chunkSize" mapstructure:"chunkSize" usage:"Textsplitter Chunk Size" default:"2048" env:"KNOW_TEXTSPLITTER_CHUNK_SIZE" name:"textsplitter-chunk-size"`
	ChunkOverlap int    `json:"chunkOverlap" mapstructure:"chunkOverlap" usage:"Textsplitter Chunk Overlap" default:"256" env:"KNOW_TEXTSPLITTER_CHUNK_OVERLAP" name:"textsplitter-chunk-overlap"`
	ModelName    string `json:"modelName" mapstructure:"modelName" usage:"Textsplitter Model Name" default:"gpt-4o" env:"KNOW_TEXTSPLITTER_MODEL_NAME" name:"textsplitter-model-name"`
	EncodingName string `json:"encodingName" mapstructure:"encodingName" usage:"Textsplitter Encoding Name" default:"cl100k_base" env:"KNOW_TEXTSPLITTER_ENCODING_NAME" name:"textsplitter-encoding-name"`
}

// NewTextSplitterOpts returns the default options for a text splitter.
func NewTextSplitterOpts() TextSplitterOpts {
	return TextSplitterOpts{
		ChunkSize:    defaults.ChunkSizeTokens,
		ChunkOverlap: defaults.ChunkOverlapTokens,
		ModelName:    defaults.TokenModel,
		EncodingName: defaults.TokenEncoding,
	}
}

func (opts *TextSplitterOpts) Configure() error {
	return load.FillConfigEnv("KNOW_TEXTSPLITTER_", opts)
}

// NewLcgoTextSplitter returns a new langchain-go text splitter.
func NewLcgoTextSplitter(opts TextSplitterOpts) lcgosplitter.TokenSplitter {
	return lcgosplitter.NewTokenSplitter(
		lcgosplitter.WithChunkSize(opts.ChunkSize),
		lcgosplitter.WithChunkOverlap(opts.ChunkOverlap),
		lcgosplitter.WithModelName(opts.ModelName),
		lcgosplitter.WithEncodingName(opts.EncodingName),
	)
}

func NewLcgoMarkdownSplitter(opts TextSplitterOpts) *lcgosplitter.MarkdownTextSplitter {
	return lcgosplitter.NewMarkdownTextSplitter(
		lcgosplitter.WithChunkSize(opts.ChunkSize),
		lcgosplitter.WithChunkOverlap(opts.ChunkOverlap),
		lcgosplitter.WithModelName(opts.ModelName),
		lcgosplitter.WithEncodingName(opts.EncodingName),
		lcgosplitter.WithHeadingHierarchy(true),
		lcgosplitter.WithCodeBlocks(true),
	)
}

func GetTextSplitterConfig(name string) (any, error) {
	// TODO: expose splitter-specific config, not only our top-level options
	switch name {
	case "text", "markdown":
		return TextSplitterOpts{}, nil
	default:
		return nil, fmt.Errorf("unknown text splitter %q", name)
	}
}

func GetTextSplitter(name string, config any) (dstypes.TextSplitter, error) {
	switch name {
	case "text":
		cfg := NewTextSplitterOpts()
		if config != nil {
			var customCfg TextSplitterOpts
			if err := mapstructure.Decode(config, &customCfg); err != nil {
				return nil, fmt.Errorf("failed to decode text splitter configuration: %w", err)
			}
			slog.Debug("GetTextSplitter Text (before merge)", "config", customCfg)
			if err := mergo.Merge(&customCfg, cfg); err != nil {
				return nil, fmt.Errorf("failed to merge text splitter configuration: %w", err)
			}
			cfg = customCfg
		}
		slog.Debug("TextSplitter", "config", cfg)
		return FromLangchain(NewLcgoTextSplitter(cfg), "lcgo_text"), nil
	case "markdown":
		cfg := NewTextSplitterOpts()
		if config != nil {
			var customCfg TextSplitterOpts
			if err := mapstructure.Decode(config, &customCfg); err != nil {
				return nil, fmt.Errorf("failed to decode text splitter configuration: %w", err)
			}
			slog.Debug("GetTextSplitter Markdown (before merge)", "config", customCfg)
			if err := mergo.Merge(&customCfg, cfg); err != nil {
				return nil, fmt.Errorf("failed to merge text splitter configuration: %w", err)
			}
			cfg = customCfg
		}
		slog.Debug("MarkdownSplitter", "config", cfg)
		return FromLangchain(NewLcgoMarkdownSplitter(cfg), "lcgo_markdown"), nil
	default:
		return nil, fmt.Errorf("unknown text splitter %q", name)
	}
}

package types

import "encoding/json"

// MCP Protocol types
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// MCP Protocol Messages
type InitializeParams struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ClientInfo      Implementation     `json:"clientInfo"`
}

type ClientCapabilities struct {
	Roots    *RootsCapability    `json:"roots,omitempty"`
	Sampling *SamplingCapability `json:"sampling,omitempty"`
}

type RootsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type SamplingCapability struct{}

type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ServerInfo      Implementation     `json:"serverInfo"`
}

type ServerCapabilities struct {
	Tools     *ToolsCapability     `json:"tools,omitempty"`
	Resources *ResourcesCapability `json:"resources,omitempty"`
	Prompts   *PromptsCapability   `json:"prompts,omitempty"`
}

type ToolsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

type ResourcesCapability struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

type PromptsCapability struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// Tool types
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type ListToolsResult struct {
	Tools []Tool `json:"tools"`
}

type CallToolParams struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
}

type CallToolResult struct {
	Content []ContentItem `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ContentItem struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// Yokai search types
type YokaiSearchParams struct {
	Name     string `json:"name,omitempty"`
	Region   string `json:"region,omitempty"`
	Category string `json:"category,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

type YokaiBook struct {
	Title       string   `json:"title"`
	Author      string   `json:"author,omitempty"`
	Publisher   string   `json:"publisher,omitempty"`
	PublishDate string   `json:"publishDate,omitempty"`
	Description string   `json:"description,omitempty"`
	URL         string   `json:"url,omitempty"`
	ISBN        string   `json:"isbn,omitempty"`
	Subjects    []string `json:"subjects,omitempty"`
}

type YokaiSearchResult struct {
	Query   string      `json:"query"`
	Total   int         `json:"total"`
	Results []YokaiBook `json:"results"`
}

// YokaiProfile captures curated lore for a yokai entry.
type YokaiProfile struct {
	Name          string   `json:"name"`
	NativeName    string   `json:"nativeName,omitempty"`
	Region        string   `json:"region,omitempty"`
	Category      string   `json:"category,omitempty"`
	Summary       string   `json:"summary"`
	Legends       []string `json:"legends,omitempty"`
	Traits        []string `json:"traits,omitempty"`
	Motifs        []string `json:"motifs,omitempty"`
	FunFact       string   `json:"funFact,omitempty"`
	CreativeHooks []string `json:"creativeHooks,omitempty"`
}

// YokaiOfTheDayParams controls how the highlight tool selects a yokai.
type YokaiOfTheDayParams struct {
	Name     string `json:"name,omitempty"`
	Category string `json:"category,omitempty"`
	Region   string `json:"region,omitempty"`
	Seed     int64  `json:"seed,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// YokaiOfTheDayResult combines curated lore with recommended reading.
type YokaiOfTheDayResult struct {
	Profile          YokaiProfile `json:"profile"`
	Query            string       `json:"query"`
	TotalBooks       int          `json:"totalBooks"`
	RecommendedBooks []YokaiBook  `json:"recommendedBooks,omitempty"`
	StoryPrompt      string       `json:"storyPrompt,omitempty"`
	Notes            []string     `json:"notes,omitempty"`
}

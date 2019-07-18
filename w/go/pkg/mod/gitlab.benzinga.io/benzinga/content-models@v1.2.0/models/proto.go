package models

// EventType is the type of realtime event that is being processed.
type EventType string

// These are the possible ContentEvent types.
const (
	Created EventType = "Created"
	Updated EventType = "Updated"
	Removed EventType = "Removed"
)

// Range specifies a range of content items.
type Range struct {
	EarliestDate Time // DEPRECATED
	LatestDate   Time // DEPRECATED
	Before       Time
	Limit        int
	Skip         int
}

// Cursor specifies a cursor.
type Cursor struct {
	LastDate Time // DEPRECATED
	LastNID  int  // DEPRECATED
}

// Filter specifies a filter set for filtering content items.
type Filter struct {
	Keywords    []string
	Symbols     []string
	Channels    []int
	ContentType []string
	Sectors     []string
	Sentiments  []int `json:",omitempty"`
}

// ConnectRequest is the initial packet sent to the server.
type ConnectRequest struct {
	Version int
	Mode    string
}

// ConnectResponse is the reply to the intial packet, sent to the client.
type ConnectResponse struct {
	Status string
}

// QueryRequest is a request made in Historical mode to grab some data.
type QueryRequest struct {
	Search    string
	Filter    Filter
	Range     Range
	Cursor    Cursor // DEPRECATED
	RequestID int
}

// QueryResponse is the reply to a query request.
type QueryResponse struct {
	RequestID int
	Count     int
	Error     string
	Results   []Content
}

// AuthRequest is a request to perform a handshake
type AuthRequest struct {
	ExcludeBody bool
	RequestID   int
}

// AuthResponse is a response to an AuthRequest
type AuthResponse struct {
	RequestID int
}

// StoryBodyRequest is a request for a story's body by ID
type StoryBodyRequest struct {
	StoryID   string
	RequestID int
}

// StoryBodyResponse is the reply to a StoryBodyRequest
type StoryBodyResponse struct {
	Body      string
	Error     string `json:",omitempty"`
	RequestID int
}

// Event is a realtime content event.
type Event struct {
	ID      int64
	NodeID  int64
	Time    Time
	Content Content
	Event   EventType
}

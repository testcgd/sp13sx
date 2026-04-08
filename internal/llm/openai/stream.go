package openai

// Streaming transport is intentionally deferred.
// The backend emits a single completion event for now so the
// runtime and TUI can be wired without blocking future SSE support.

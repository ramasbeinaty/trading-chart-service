package tracing

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"google.golang.org/grpc/metadata"
)

// format: version-traceid-parentid-flags
// example: 00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01
type TraceContext struct {
	Version    string
	TraceID    string
	ParentID   string
	TraceFlags string
}

func GetTraceContext(ctx context.Context) (TraceContext, error) {
	var trace TraceContext

	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		// return trace if it's found and is properly formatted
		traceValues := md.Get("traceparent")

		if len(traceValues) > 0 {
			parts := strings.Split(traceValues[0], "-")

			if len(parts) == 4 {
				trace.Version = parts[0]
				trace.TraceID = parts[1]
				trace.ParentID = parts[2]
				trace.TraceFlags = parts[3]
				return trace, nil
			}
		}
	}

	// generate new trace
	return generateNewTraceContext()
}

func generateNewTraceContext() (TraceContext, error) {
	traceID, err := generateRandomHexString(16) // 128-bit trace ID
	if err != nil {
		return TraceContext{}, fmt.Errorf("failed to generate trace ID: %v", err)
	}

	parentID, err := generateRandomHexString(8) // 64-bit parent ID
	if err != nil {
		return TraceContext{}, fmt.Errorf("failed to generate parent ID: %v", err)
	}

	return TraceContext{
		Version:    "00",
		TraceID:    traceID,
		ParentID:   parentID,
		TraceFlags: "01",
	}, nil
}

func generateRandomHexString(size int) (string, error) {
	buff := make([]byte, size)
	if _, err := rand.Read(buff); err != nil {
		return "", err
	} else {
		return hex.EncodeToString(buff), nil
	}
}

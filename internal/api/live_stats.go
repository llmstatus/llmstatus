package api

import (
	"context"

	"github.com/llmstatus/llmstatus/internal/store/influx"
)

// LiveStatsReader is the subset of influx.LiveStatsReader used by the API.
type LiveStatsReader interface {
	AllProviderLiveStats(ctx context.Context) ([]influx.ProviderLiveStat, error)
	AllModelLiveStats(ctx context.Context) ([]influx.ModelLiveStat, error)
	AllModelSparklines(ctx context.Context) (map[string][]float64, error)
}

// compile-time check: influx.LiveStatsReader satisfies LiveStatsReader.
var _ LiveStatsReader = (influx.LiveStatsReader)(nil)

// WithLiveStatsReader wires an InfluxDB live stats reader into the Server.
func WithLiveStatsReader(r LiveStatsReader) func(*Server) {
	return func(s *Server) { s.liveStats = r }
}

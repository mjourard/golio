package riot

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mjourard/golio/api"
	"github.com/mjourard/golio/internal"
	"github.com/mjourard/golio/internal/mock"
)

func TestChampionClient_GetFreeRotation(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		want    *ChampionInfo
		doer    internal.Doer
		wantErr error
	}{
		{
			name: "get response",
			want: &ChampionInfo{},
			doer: mock.NewJSONMockDoer(ChampionInfo{}, 200),
		},
		{
			name: "unknown error status",
			wantErr: api.Error{
				Message:    "unknown error reason",
				StatusCode: 999,
			},
			doer: mock.NewStatusMockDoer(999),
		},
		{
			name:    "not found",
			wantErr: api.ErrNotFound,
			doer:    mock.NewStatusMockDoer(http.StatusNotFound),
		},
		{
			name: "rate limited",
			want: &ChampionInfo{},
			doer: rateLimitDoer(ChampionInfo{}),
		},
		{
			name: "unavailable once",
			want: &ChampionInfo{},
			doer: unavailableOnceDoer(ChampionInfo{}),
		},
		{
			name:    "unavailable twice",
			wantErr: api.ErrServiceUnavailable,
			doer:    mock.NewStatusMockDoer(http.StatusServiceUnavailable),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewClient(api.RegionEuropeWest, "API_KEY", tt.doer, logrus.StandardLogger())
			got, err := client.Champion.GetFreeRotation()
			require.Equal(t, err, tt.wantErr, fmt.Sprintf("want err %v, got %v", tt.wantErr, err))
			if tt.wantErr == nil {
				assert.Equal(t, got, tt.want)
			}
		})
	}
}

package riot

import (
	"fmt"
	"io"

	log "github.com/sirupsen/logrus"
)

type matchClient struct {
	c *Client
}

// Get returns a match specified by its ID
func (m *matchClient) Get(id int) (*Match, error) {
	logger := m.logger().WithField("method", "Get")
	var match *Match
	if err := m.c.getInto(fmt.Sprintf(endpointGetMatch, id), &match); err != nil {
		logger.Debug(err)
		return nil, err
	}
	return match, nil
}

// List returns a specified range of matches played on the account
func (m *matchClient) List(accountID string, filter MatchFilter) (*Matchlist, error) {
	logger := m.logger().WithField("method", "List")
	var matches *Matchlist
	queryParams := filter.GetQueryParams()
	if queryParams != "" {
		queryParams = "?" + queryParams
	}
	if err := m.c.getInto(
		fmt.Sprintf(endpointGetMatchesByAccount, accountID, queryParams),
		&matches,
	); err != nil {
		logger.Debug(err)
		return nil, err
	}
	return matches, nil
}

// MatchStreamValue value returned by ListStream, containing either a reference to a match or an error
type MatchStreamValue struct {
	*MatchReference
	Error error
}

// ListStream returns all matches played on this account as a stream, requesting new until there are no
// more new games
func (m *matchClient) ListStream(accountID string, filter MatchFilter) <-chan MatchStreamValue {
	logger := m.logger().WithField("method", "ListStream")
	cMatches := make(chan MatchStreamValue, 100)
	go func() {
		start := 0
		end := 100
		filter.BeginIndex = &start
		filter.EndIndex = &end
		for {
			matches, err := m.List(accountID, filter)
			if err != nil {
				logger.Debug(err)
				cMatches <- MatchStreamValue{Error: err}
				return
			}
			for _, match := range matches.Matches {
				cMatches <- MatchStreamValue{MatchReference: match}
			}
			if len(matches.Matches) < 100 {
				cMatches <- MatchStreamValue{Error: io.EOF}
				return
			}
			start += 100
			end += 100
		}
	}()
	return cMatches
}

// GetTimeline returns the timeline for the given match
// NOTE: timelines are not available for every match
func (m *matchClient) GetTimeline(matchID int) (*MatchTimeline, error) {
	logger := m.logger().WithField("method", "GetTimeline")
	var timeline MatchTimeline
	if err := m.c.getInto(fmt.Sprintf(endpointGetMatchTimeline, matchID), &timeline); err != nil {
		logger.Debug(err)
		return nil, err
	}
	return &timeline, nil
}

// ListIDsByTournamentCode returns all match ids for the given tournament
func (m *matchClient) ListIDsByTournamentCode(tournamentCode string) ([]int, error) {
	logger := m.logger().WithField("method", "ListIDsByTournamentCode")
	var ids []int
	if err := m.c.getInto(fmt.Sprintf(endpointGetMatchIDsByTournamentCode, tournamentCode), &ids); err != nil {
		logger.Debug(err)
		return nil, err
	}
	return ids, nil
}

// GetForTournament returns the match data for the given match in the given tournament
func (m *matchClient) GetForTournament(matchID int, tournamentCode string) (*Match, error) {
	logger := m.logger().WithField("method", "GetForTournament")
	var match Match
	if err := m.c.getInto(fmt.Sprintf(endpointGetMatchForTournament, matchID, tournamentCode), &match); err != nil {
		logger.Debug(err)
		return nil, err
	}
	return &match, nil
}

func (m *matchClient) logger() log.FieldLogger {
	return m.c.logger().WithField("category", "match")
}

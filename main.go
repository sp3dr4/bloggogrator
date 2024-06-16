package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
	"github.com/sp3dr4/bloggogrator/api"
	"github.com/sp3dr4/bloggogrator/internal/database"
	"github.com/sp3dr4/bloggogrator/internal/rss"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("CONN")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("connection open error: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("connection ping error: %v", err)
	}

	dbQueries := database.New(db)

	var pollActive bool = true
	pollActiveStr := os.Getenv("POLL_ENABLED")
	if pollActiveStr != "" {
		pollActive, err = strconv.ParseBool(pollActiveStr)
		if err != nil {
			log.Fatalf("invalid poll feature flag %v: %v", pollActiveStr, err)
		}
	}

	if pollActive {
		pollFrequencyStr := os.Getenv("POLL_FREQUENCY_SECONDS")
		pollFrequencySec, err := strconv.Atoi(pollFrequencyStr)
		if err != nil {
			log.Fatalf("invalid poll frequency seconds %v: %v", pollFrequencyStr, err)
		}

		pollAmountstr := os.Getenv("POLL_AMOUNT")
		pollAmount, err := strconv.ParseInt(pollAmountstr, 10, 32)
		if err != nil {
			log.Fatalf("invalid poll amount %v: %v", pollAmountstr, err)
		}

		feedsFetcher := func() ([]database.Feed, error) {
			return dbQueries.GetNextFeedsToFetch(context.Background(), int32(pollAmount))
		}

		feedMarker := func(id uuid.UUID, when time.Time) (database.Feed, error) {
			params := database.MarkFeedFetchedParams{ID: id, LastFetchedAt: sql.NullTime{Time: when, Valid: true}}
			return dbQueries.MarkFeedFetched(context.Background(), params)
		}

		postSaver := func(url, title, description string, publishedAt time.Time, feedId uuid.UUID) (*database.Post, error) {
			params := database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Url:         url,
				Title:       sql.NullString{String: title, Valid: title != ""},
				Description: sql.NullString{String: description, Valid: description != ""},
				PublishedAt: publishedAt,
				FeedID:      feedId,
			}
			p, err := dbQueries.CreatePost(context.Background(), params)
			if err != nil {
				pqErr, ok := err.(*pq.Error)
				if ok && pqErr.Code.Name() == "unique_violation" {
					return nil, nil
				}
				return nil, err
			}
			return &p, nil
		}

		go rss.Run(time.Duration(pollFrequencySec)*time.Second, feedsFetcher, feedMarker, postSaver)
	}

	api.Run(dbQueries)
}

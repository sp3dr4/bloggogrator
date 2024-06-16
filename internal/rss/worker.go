package rss

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sp3dr4/bloggogrator/internal/database"
)

const pubDateLayout = "Mon, 02 Jan 2006 15:04:05 -0700"

type GetNextFeeds func() ([]database.Feed, error)
type MarkFeed func(id uuid.UUID, when time.Time) (database.Feed, error)
type SavePost func(url, title, description string, publishedAt time.Time, feedId uuid.UUID) (*database.Post, error)

func Run(frequency time.Duration, getFeeds GetNextFeeds, mark MarkFeed, save SavePost) {
	ticker := time.NewTicker(frequency)
	for range ticker.C {
		log.Println("tick...")

		feeds, err := getFeeds()
		if err != nil {
			log.Printf("could not retrieve feeds: %v\n", err)
		}

		var wg sync.WaitGroup

		for _, feed := range feeds {
			wg.Add(1)

			go func() {
				defer wg.Done()

				feedContent, err := ReadRss(feed.Url)
				if err != nil {
					log.Printf("err reading rss %v: %v\n", feed.Name, err)
				} else {
					for _, item := range feedContent.Channel.Items {
						t, err := time.Parse(pubDateLayout, item.PubDate)
						if err != nil {
							log.Printf("date parse err: %v\n", err)
							continue
						}
						post, err := save(item.Link, item.Title, item.Description, t, feed.ID)
						if err != nil {
							log.Printf("save err: %v\n", err)
							continue
						}
						if post != nil {
							log.Printf("[%s] saved %v\n", feed.Name, post.Title)
						}
					}
				}

				_, err = mark(feed.ID, time.Now())
				if err != nil {
					log.Printf("err marking feed %v as fetched: %v\n", feed.Name, err)
				}

			}()
		}

		wg.Wait()
	}
}

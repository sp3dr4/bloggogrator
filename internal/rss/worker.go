package rss

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sp3dr4/bloggogrator/internal/database"
)

type GetNextFeeds func() ([]database.Feed, error)
type MarkFeed func(id uuid.UUID, when time.Time)

func Run(frequency time.Duration, getFeeds GetNextFeeds, mark MarkFeed) {
	ticker := time.NewTicker(frequency)
	for range ticker.C {
		log.Println("tick...")

		feeds, err := getFeeds()
		if err != nil {
			log.Printf("could not retrieve feeds: %v\n", err)
		}

		var wg sync.WaitGroup

		for _, fo := range feeds {
			wg.Add(1)

			go func() {
				defer wg.Done()
				feedContent, err := ReadRss(fo.Url)
				if err != nil {
					log.Printf("err reading rss %v: %v\n", fo.Name, err)
				} else {
					for _, item := range feedContent.Channel.Items {
						pubdate, _ := strings.CutSuffix(item.PubDate, " 00:00:00 +0000")
						log.Printf("[%s] %v -> %v\n", fo.Name, pubdate, item.Title)
					}
				}
				mark(fo.ID, time.Now())
			}()
		}

		wg.Wait()
	}
}

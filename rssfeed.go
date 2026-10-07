package main

import (
	"context"
	"encoding/xml"
	"html"
	"net/http"
	"io"
	"database/sql"
	"time"
	"github.com/google/uuid"
	"log"
	"strings"
	"example.com/gator/internal/database"
)

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")
	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	pser, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	end := RSSFeed{}
	err = xml.Unmarshal(pser, &end)
	if err != nil {
		return nil, err
	}
	end.Channel.Title = html.UnescapeString(end.Channel.Title)
	end.Channel.Description = html.UnescapeString(end.Channel.Description)
	for i := range end.Channel.Item {
		end.Channel.Item[i].Title = html.UnescapeString(end.Channel.Item[i].Title)
		end.Channel.Item[i].Description = html.UnescapeString(end.Channel.Item[i].Description)
	}
	return &end, nil
}
func scrapeFeeds(s *state)error{
	nextfeed,err:=s.db.GetNextFeedToFetch(context.Background())
	if err!=nil{
		return err
	}
	err =s.db.MarkFeedFetched(context.Background(),nextfeed.ID)
	if err!=nil{
		return err
	}
	feed,err:=fetchFeed(context.Background(),nextfeed.Url)
	if err!=nil{
		return err
	}
	for _,itwm:=range feed.Channel.Item{
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, itwm.PubDate); err == nil {
			publishedAt = sql.NullTime{Time: t, Valid: true}
		}
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			FeedID:      nextfeed.ID,
			Title:       itwm.Title,
			Description: sql.NullString{String: itwm.Description, Valid: true},
			Url:         itwm.Link,
			PublishedAt: publishedAt,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("Couldn't create post: %v", err)
			continue
		}
	}
	return nil
}
type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

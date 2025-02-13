package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const apiURL = "https://www.googleapis.com/youtube/v3/playlistItems"
const videoDetailsURL = "https://www.googleapis.com/youtube/v3/videos"

var apiKey string

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	apiKey = os.Getenv("YOUTUBE_API_KEY")
	if apiKey == "" {
		log.Fatal("YOUTUBE_API_KEY not set in .env file")
	}
}

func extractPlaylistID(url string) string {
	re := regexp.MustCompile(`[?&]list=([a-zA-Z0-9_-]+)`) // TODO: taken from someone else's random script. Check for safety & validate
	matches := re.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func fetchPlaylistItems(playlistID string) ([]string, error) {
	var videoIDs []string
	nextPageToken := ""

	for {
		url := fmt.Sprintf("%s?part=contentDetails&playlistId=%s&maxResults=50&pageToken=%s&key=%s",
			apiURL, playlistID, nextPageToken, apiKey)

		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var result struct {
			Items []struct {
				ContentDetails struct {
					VideoID string `json:"videoId"`
				} `json:"contentDetails"`
			} `json:"items"`
			NextPageToken string `json:"nextPageToken"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		for _, item := range result.Items {
			videoIDs = append(videoIDs, item.ContentDetails.VideoID)
		}

		if result.NextPageToken == "" {
			break
		}
		nextPageToken = result.NextPageToken
	}

	return videoIDs, nil
}

func fetchVideoDuration(videoID string) (time.Duration, error) {
	url := fmt.Sprintf("%s?part=contentDetails&id=%s&key=%s", videoDetailsURL, videoID, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			ContentDetails struct {
				Duration string `json:"duration"`
			} `json:"contentDetails"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if len(result.Items) == 0 {
		return 0, fmt.Errorf("no video details found for ID %s", videoID)
	}

	return parseISO8601Duration(result.Items[0].ContentDetails.Duration)
}

func parseISO8601Duration(isoDuration string) (time.Duration, error) {
	re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`) // TODO: check this
	matches := re.FindStringSubmatch(isoDuration)

	var hours, minutes, seconds int

	if matches[1] != "" {
		hours, _ = strconv.Atoi(matches[1])
	}
	if matches[2] != "" {
		minutes, _ = strconv.Atoi(matches[2])
	}
	if matches[3] != "" {
		seconds, _ = strconv.Atoi(matches[3])
	}

	return time.Duration(hours)*time.Hour + time.Duration(minutes)*time.Minute + time.Duration(seconds)*time.Second, nil
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%dh %dm %ds", hours, minutes, seconds)
}

func printDurationAtSpeed(totalDuration time.Duration, speed float64) {
	adjustedDuration := time.Duration(float64(totalDuration) / speed)
	fmt.Printf("At %.2fx: %s\n", speed, formatDuration(adjustedDuration))
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ytplaycount <playlist_url>")
		return
	}

	playlistURL := os.Args[1]
	playlistID := extractPlaylistID(playlistURL)
	if playlistID == "" {
		log.Fatal("Invalid YouTube playlist URL")
	}

	fmt.Printf("_________________________________________________________\n\n")
	fmt.Printf("Fetching: %s\n", playlistID)

	videoIDs, err := fetchPlaylistItems(playlistID)
	if err != nil {
		log.Fatalf("Error fetching playlist items: %v", err)
	}

	var totalDuration time.Duration
	for _, videoID := range videoIDs {
		videoDuration, err := fetchVideoDuration(videoID)
		if err != nil {
			log.Printf("Error fetching video duration for %s: %v", videoID, err)
			continue
		}
		totalDuration += videoDuration
	}

	fmt.Printf("\nTotal duration: %s\n\n", formatDuration(totalDuration))

	printDurationAtSpeed(totalDuration, 1.25)
	printDurationAtSpeed(totalDuration, 1.5)
	printDurationAtSpeed(totalDuration, 1.75)
	printDurationAtSpeed(totalDuration, 2)
	fmt.Printf("_________________________________________________________\n")
}

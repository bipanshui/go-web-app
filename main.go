package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

func homePage(w http.ResponseWriter, r *http.Request) {
	// Render the home html page from static folder
	http.ServeFile(w, r, "static/home.html")
}

func coursePage(w http.ResponseWriter, r *http.Request) {
	// Render the course html page
	http.ServeFile(w, r, "static/courses.html")
}

func aboutPage(w http.ResponseWriter, r *http.Request) {
	// Render the about html page
	http.ServeFile(w, r, "static/about.html")
}

func contactPage(w http.ResponseWriter, r *http.Request) {
	// Render the contact html page
	http.ServeFile(w, r, "static/contact.html")
}

func youtubeThumbnailHandler(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "missing url query parameter", http.StatusBadRequest)
		return
	}

	thumbnailURL, err := resolveYouTubeThumbnail(rawURL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"thumbnail_url": thumbnailURL,
	})
}

func resolveYouTubeThumbnail(rawURL string) (string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}

	host := strings.ToLower(parsedURL.Hostname())
	if strings.Contains(host, "youtu.be") {
		videoID := strings.Trim(parsedURL.Path, "/")
		if videoID != "" {
			return youtubeImageURL(videoID), nil
		}
	}

	if strings.Contains(host, "youtube.com") || strings.Contains(host, "youtube-nocookie.com") {
		if videoID := youtubeVideoIDFromURL(parsedURL); videoID != "" {
			return youtubeImageURL(videoID), nil
		}
	}

	if strings.Contains(host, "youtube.com") || strings.Contains(host, "youtu.be") || strings.Contains(host, "youtube-nocookie.com") {
		thumbnailURL, err := resolveYouTubeThumbnailFromOEmbed(rawURL)
		if err == nil && thumbnailURL != "" {
			return thumbnailURL, nil
		}

		thumbnailURL, err = resolveYouTubeThumbnailFromPage(rawURL)
		if err == nil && thumbnailURL != "" {
			return thumbnailURL, nil
		}
	}

	return defaultThumbnailDataURI(), nil
}

func youtubeVideoIDFromURL(parsedURL *url.URL) string {
	switch {
	case parsedURL.Path == "/watch":
		return parsedURL.Query().Get("v")
	case strings.HasPrefix(parsedURL.Path, "/shorts/"):
		return strings.TrimPrefix(parsedURL.Path, "/shorts/")
	case strings.HasPrefix(parsedURL.Path, "/embed/"):
		return strings.TrimPrefix(parsedURL.Path, "/embed/")
	case strings.HasPrefix(parsedURL.Path, "/v/"):
		return strings.TrimPrefix(parsedURL.Path, "/v/")
	default:
		return ""
	}
}

func youtubeImageURL(videoID string) string {
	return fmt.Sprintf("https://img.youtube.com/vi/%s/hqdefault.jpg", videoID)
}

func resolveYouTubeThumbnailFromOEmbed(rawURL string) (string, error) {
	oEmbedURL := fmt.Sprintf(
		"https://www.youtube.com/oembed?url=%s&format=json",
		url.QueryEscape(rawURL),
	)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(oEmbedURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("youtube oembed returned %s", resp.Status)
	}

	var payload struct {
		ThumbnailURL string `json:"thumbnail_url"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	return payload.ThumbnailURL, nil
}

func resolveYouTubeThumbnailFromPage(rawURL string) (string, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(rawURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("youtube page returned %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`(?i)<meta[^>]+property=["']og:image["'][^>]+content=["']([^"']+)["']`)
	match := re.FindStringSubmatch(string(body))
	if len(match) < 2 {
		return "", fmt.Errorf("youtube page did not include a thumbnail")
	}

	return match[1], nil
}

func defaultThumbnailDataURI() string {
	return "data:image/svg+xml;charset=utf-8," + url.PathEscape(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 640 360" role="img" aria-label="YouTube thumbnail placeholder">
<rect width="640" height="360" rx="28" fill="#f2f2f2"/>
<rect x="56" y="56" width="528" height="248" rx="22" fill="#ffffff" stroke="#d9d9d9"/>
<circle cx="320" cy="180" r="52" fill="#111827"/>
<path d="M304 156l48 24-48 24z" fill="#ffffff"/>
<text x="320" y="278" text-anchor="middle" font-family="Arial, sans-serif" font-size="26" fill="#4b5563">YouTube preview</text>
</svg>`)
}

func newMux() *http.ServeMux {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("static"))

	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("/api/youtube-thumbnail", youtubeThumbnailHandler)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/home", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})
	mux.HandleFunc("/home", homePage)
	mux.HandleFunc("/courses", coursePage)
	mux.HandleFunc("/about", aboutPage)
	mux.HandleFunc("/contact", contactPage)

	return mux
}

func main() {
	err := http.ListenAndServe("0.0.0.0:8080", newMux())
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/dhowden/tag"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jinzhu/configor"
)

// -------- CONFIG --------
type Config struct {
	Server struct {
		Addr string `yaml:"addr" default:":8080"`
	} `yaml:"server"`

	Music struct {
		Root string `yaml:"root" default:"./music"`
	} `yaml:"music"`
}

var cfg Config

// -------- MODELS --------
type Track struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Artist   string `json:"artist,omitempty"`
	Src      string `json:"src"`
	Cover    string `json:"cover,omitempty"`
	FilePath string `json:"-"`
}

// -------- GLOBAL CACHE --------
var (
	tracksMu sync.RWMutex
	tracks   []Track
)

// -------- UTILS --------
func isAudioFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	mimeType := mime.TypeByExtension(ext)
	return strings.HasPrefix(mimeType, "audio/")
}

func scanTracks(root string) ([]Track, error) {
	var result []Track

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !isAudioFile(path) {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		metadata, err := tag.ReadFrom(file)
		if err != nil {
			return nil
		}

		title := metadata.Title()
		if title == "" {
			title = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		}

		artist := metadata.Artist()

		// обложка
		var cover string
		picture := metadata.Picture()
		if picture != nil {
			cover = fmt.Sprintf("data:%s;base64,%s", picture.MIMEType, encodeBase64(picture.Data))
		}

		rel, _ := filepath.Rel(root, path)
		id := strings.ReplaceAll(rel, string(os.PathSeparator), "_")

		t := Track{
			ID:       id,
			Title:    title,
			Artist:   artist,
			Src:      fmt.Sprintf("/tracks/%s", rel),
			Cover:    cover,
			FilePath: path,
		}
		result = append(result, t)
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Title) < strings.ToLower(result[j].Title)
	})

	return result, nil
}

func encodeBase64(data []byte) string {
	return string(bytes.TrimSpace([]byte(base64.StdEncoding.EncodeToString(data))))
}

func refreshTracks() error {
	list, err := scanTracks(cfg.Music.Root)
	if err != nil {
		return err
	}

	tracksMu.Lock()
	defer tracksMu.Unlock()
	tracks = list
	return nil
}

func getPaginated(offset, limit int) []Track {
	tracksMu.RLock()
	defer tracksMu.RUnlock()
	if offset > len(tracks) {
		return []Track{}
	}
	end := offset + limit
	if end > len(tracks) {
		end = len(tracks)
	}
	return tracks[offset:end]
}

// -------- HANDLERS --------

// GET /api/tracks
func handleGetTracks(c *fiber.Ctx) error {
	offset := c.QueryInt("offset", 0)
	limit := c.QueryInt("limit", 10)

	data := getPaginated(offset, limit)
	return c.JSON(data)
}

// GET /api/tracks/refresh
func handleRefresh(c *fiber.Ctx) error {
	if err := refreshTracks(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(fiber.Map{"status": "ok", "count": len(tracks)})
}

// -------- MAIN --------

func main() {
	log.Println("app start")

	if err := configor.Load(&cfg, "config.yml"); err != nil {
		log.Fatalf("config load error: %v", err)
	}

	if _, err := os.Stat(cfg.Music.Root); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("music root does not exist: %s", cfg.Music.Root)
	}

	if err := refreshTracks(); err != nil {
		log.Printf("initial scan error: %v", err)
	}

	app := fiber.New()
	app.Use(logger.New())
	app.Static("/tracks", cfg.Music.Root)
	app.Static("/", "./res/dist")

	// API
	api := app.Group("/api")
	api.Get("/tracks", handleGetTracks)
	api.Get("/tracks/refresh", handleRefresh)

	// SPA fallback
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./res/dist/index.html")
	})

	log.Printf("Listening on %s, music root: %s", cfg.Server.Addr, cfg.Music.Root)
	log.Fatal(app.Listen(cfg.Server.Addr))
}

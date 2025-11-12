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
	"time"

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
	log.Printf("SCAN: open %s",root)
	var result []Track

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("WALK: %s err: %v", path, err)
			return err
		}
		if info.IsDir() {
			log.Printf("WALK: %s is dir", path)
			return nil
		}
		if !isAudioFile(path) {
			log.Printf("WALK: %s not audio file", path)
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			log.Printf("OS OPEN ERROR: open %s: %v",path, err)
			return nil
		}
		defer file.Close()

		log.Printf("META: read meta from %s", path)
		metadata, err := tag.ReadFrom(file)
		if err != nil {
			log.Printf("metadata read error for %s: %v", path, err)
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
		log.Printf("SCAN finish err: %s err: %v", root, err)
		return nil, err
	}
	log.Printf("SCAN finished")

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

// POST /api/tracks/add
func handleAddTrack(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "File not found in request",
		})
	}

	if fileHeader.Header.Get("Content-Type")[:5] != "audio" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Wrong file type",
		})
	}

	uploadDir := cfg.Music.Root
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Can't create dir",
		})
	}

	base := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s_%d%s", base, timestamp, filepath.Ext(fileHeader.Filename))
	savePath := filepath.Join(uploadDir, filename)

	// Сохраняем файл
	if err := c.SaveFile(fileHeader, savePath); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error on save file",
		})
	}
	log.Println("saved file:", savePath)

	return c.JSON(fiber.Map{
		"message":  "File uploaded",
		"filename": filename,
		"path":     savePath,
	})
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

	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024, // 20 MB
	})
	app.Use(logger.New())
	app.Static("/tracks", cfg.Music.Root)
	app.Static("/", "./res/dist")

	// API
	api := app.Group("/api")
	api.Get("/tracks", handleGetTracks)
	api.Get("/tracks/refresh", handleRefresh)
	api.Post("/tracks/add", handleAddTrack)

	// SPA fallback
	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendFile("./res/dist/index.html")
	})

	log.Printf("Listening on %s, music root: %s", cfg.Server.Addr, cfg.Music.Root)
	log.Fatal(app.Listen(cfg.Server.Addr))
}

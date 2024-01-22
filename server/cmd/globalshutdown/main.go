package main

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	app := fiber.New(fiber.Config{
		ProxyHeader: fiber.HeaderXForwardedFor,
	})

	app.Use(cors.New())

	db, err := sql.Open("sqlite3", "file::memory:?cache=shared")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Println(err)
		}
	}()

	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS pending (id TEXT PRIMARY KEY NOT NULL, ip TEXT NOT NULL, time TEXT NOT NULL)")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := stmt.Exec(); err != nil {
		log.Fatal(err)
	}

	go func() {
		for range time.Tick(30 * time.Minute) {
			stmt, err := db.Prepare("DELETE FROM pending WHERE time < ?")
			if err != nil {
				log.Println(err)
				continue
			}

			if _, err := stmt.Exec(time.Now().Add(-1 * time.Hour)); err != nil {
				log.Println(err)
				continue
			}
		}
	}()

	salt := uuid.New().String()
	app.Post("/shutdown", func(c *fiber.Ctx) error {
		var idString string
		err := json.Unmarshal(c.Body(), &idString)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		id, err := uuid.Parse(idString)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		h := sha256.New()
		h.Write([]byte(c.IP() + salt))
		ipHash := string(h.Sum(nil))

		var count int
		if err := db.QueryRow("SELECT COUNT(*) FROM pending WHERE ip = ?", ipHash).Scan(&count); err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if count >= 5 {
			return c.SendStatus(fiber.StatusTooManyRequests)
		}

		stmt, err := db.Prepare("INSERT OR REPLACE INTO pending VALUES (?, ?, ?)")
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if _, err := stmt.Exec(id, ipHash, time.Now()); err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.SendStatus(fiber.StatusOK)
	})

	app.Post("/pending", func(c *fiber.Ctx) error {
		var idString string
		err := json.Unmarshal(c.Body(), &idString)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		idParsed, err := uuid.Parse(idString)
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusBadRequest)
		}

		var id string
		row := db.QueryRow("DELETE FROM pending WHERE id = ? RETURNING 1", idParsed.String())
		if err := row.Scan(&id); err != nil {
			if err == sql.ErrNoRows {
				return c.JSON(false)
			}

			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		return c.JSON(true)
	})

	app.Listen(":3000")
}

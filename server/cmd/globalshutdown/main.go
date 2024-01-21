package main

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	app := fiber.New()
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

	stmt, err := db.Prepare("CREATE TABLE IF NOT EXISTS pending (id TEXT PRIMARY KEY NOT NULL)")
	if err != nil {
		log.Fatal(err)
	}

	if _, err := stmt.Exec(); err != nil {
		log.Fatal(err)
	}

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

		stmt, err := db.Prepare("INSERT OR IGNORE INTO pending VALUES (?)")
		if err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		if _, err := stmt.Exec(id); err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		var count int
		row := db.QueryRow("SELECT COUNT(*) FROM pending")
		if err := row.Scan(&count); err != nil {
			log.Println(err)
			return c.SendStatus(fiber.StatusInternalServerError)
		}

		maxCount := 1000
		if count > maxCount {
			stmt, err := db.Prepare("DELETE FROM pending WHERE rowid <= ?")
			if err != nil {
				log.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}

			if _, err := stmt.Exec(count - maxCount); err != nil {
				log.Println(err)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
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

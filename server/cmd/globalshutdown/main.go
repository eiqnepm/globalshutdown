package main

import (
	"encoding/json"
	"log"
	"slices"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/google/uuid"
)

var (
	pendingMachines  []uuid.UUID
	mPendingMachines sync.Mutex
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	app := fiber.New()
	app.Use(cors.New())

	app.Post("/shutdown", func(c *fiber.Ctx) error {
		var idString string
		err := json.Unmarshal(c.Body(), &idString)
		if err != nil {
			return err
		}

		id, err := uuid.Parse(idString)
		if err != nil {
			return err
		}

		mPendingMachines.Lock()
		defer mPendingMachines.Unlock()
		if !slices.Contains(pendingMachines, id) {
			pendingMachines = append(pendingMachines, id)
		}

		return c.SendStatus(fiber.StatusOK)
	})

	app.Post("/pending", func(c *fiber.Ctx) error {
		var idString string
		err := json.Unmarshal(c.Body(), &idString)
		if err != nil {
			return err
		}

		id, err := uuid.Parse(idString)
		if err != nil {
			return c.SendStatus(fiber.StatusBadRequest)
		}

		if !slices.Contains(pendingMachines, id) {
			return c.JSON(false)
		}

		mPendingMachines.Lock()
		defer mPendingMachines.Unlock()
		var newPendingMachines []uuid.UUID
		for _, pending := range pendingMachines {
			if pending == id {
				continue
			}

			newPendingMachines = append(newPendingMachines, pending)
		}

		pendingMachines = newPendingMachines
		return c.JSON(true)
	})

	app.Listen(":3000")
}

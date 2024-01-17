package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/getlantern/systray"
	"github.com/getlantern/systray/example/icon"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.design/x/clipboard"
)

var (
	pendingMachines  []uuid.UUID
	mPendingMachines sync.Mutex
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)

	fScheme := flag.String("scheme", "https", "Specifies the server scheme for the client")
	fHost := flag.String("host", "globalshutdown.eiqnepm.duckdns.org", "Specifies the server host for the client")
	fPort := flag.Int("port", 3000, "Specifies the server port for the client")
	flag.Parse()

	systray.Run(func() {
		systray.SetIcon(icon.Data)
		systray.SetTooltip("GlobalShutdown")

		if _, err := os.Stat("id.txt"); errors.Is(err, os.ErrNotExist) {
			f, err := os.Create("id.txt")
			if err != nil {
				log.Println(err)
				systray.Quit()
			}

			_, err = f.WriteString(uuid.New().String())
			if err != nil {
				log.Println(err)
				systray.Quit()
			}
		}

		f, err := os.ReadFile("id.txt")
		if err != nil {
			log.Println(err)
		}

		id, err := uuid.Parse(string(f))
		if err != nil {
			log.Println(err)
			systray.Quit()
		}

		err = clipboard.Init()
		if err != nil {
			log.Println(err)
			systray.Quit()
		}

		mCopy := systray.AddMenuItem("Copy ID", "")
		go func() {
			for {
				<-mCopy.ClickedCh
				clipboard.Write(clipboard.FmtText, []byte(id.String()))
			}
		}()

		mQuit := systray.AddMenuItem("Quit", "")
		go func() {
			<-mQuit.ClickedCh
			systray.Quit()
		}()

		go func() {
			for range time.Tick(5 * time.Minute) {
				u := url.URL{
					Scheme: *fScheme,
					Host:   net.JoinHostPort(*fHost, strconv.Itoa(*fPort)),
					Path:   "/pending",
				}

				agent := fiber.Post(u.String())
				agent.JSON(id.String())
				statusCode, body, errs := agent.Bytes()
				if len(errs) > 0 {
					log.Println(errs)
					continue
				}

				if statusCode != fiber.StatusOK {
					log.Println(statusCode)
					continue
				}

				var pending bool
				err = json.Unmarshal(body, &pending)
				if err != nil {
					log.Println(err)
					continue
				}

				if pending != true {
					continue
				}

				if err := exec.Command(`C:\Windows\System32\shutdown.exe`, "/s", "/f", "/t", "0").Run(); err != nil {
					log.Println(err)
				}
			}
		}()

	}, func() {})
}

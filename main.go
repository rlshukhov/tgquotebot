package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/thanhpk/randstr"
	tele "gopkg.in/telebot.v3"
)

func main() {
	templateHtml, err := os.ReadFile("./quote.html")
	if err != nil {
		log.Fatal(err)
	}

	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tele.OnText, func(c tele.Context) error {
		user := c.Message().OriginalSender
		if user == nil {
			return c.Send("Перешли мне сообщение и я сделаю из него стикер.")
		}

		author := user.FirstName
		if user.LastName != "" {
			author = author + " " + user.LastName
		}

		photos, err := b.ProfilePhotosOf(user)
		if err != nil {
			return err
		}
		if len(photos) == 0 {
			return c.Send("Нет аватарки, я так не умею(")
		}

		avatarPhoto := photos[0]
		avatarFile, err := b.FileByID(avatarPhoto.FileID)
		if err != nil {
			return err
		}

		token := randstr.String(16)

		err = b.Download(&avatarFile, "/tmp/quote/"+token+"_avatar.jpeg")
		if err != nil {
			return err
		}

		templateCopy := string(templateHtml)
		templateCopy = strings.Replace(templateCopy, "{{author}}", author, -1)
		templateCopy = strings.Replace(templateCopy, "{{message}}", c.Text(), -1)
		templateCopy = strings.Replace(templateCopy, "{{avatar_file_path}}", token+"_avatar.jpeg", -1)

		htmlFile := "/tmp/quote/" + token + ".html"
		pngFile := "/tmp/quote/" + token + ".png"
		webpFile := "/tmp/quote/" + token + ".webp"

		err = os.WriteFile(htmlFile, []byte(templateCopy), 0644)
		if err != nil {
			return err
		}

		err = shellexec("wkhtmltoimage --format png --transparent --width 512 --quality 100 --zoom 1.5 --enable-local-file-access " + htmlFile + " " + pngFile)
		if err != nil {
			return err
		}

		err = shellexec("cwebp " + pngFile + " -o " + webpFile)
		if err != nil {
			return err
		}

		sticker := &tele.Sticker{
			File: tele.FromDisk(webpFile),
		}

		return c.Send(sticker)
	})

	b.Start()
}

func shellexec(command string) error {
	cmd := exec.Command("bash", "-c", command)
	return cmd.Run()
}

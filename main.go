package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/tebeka/selenium"
	"github.com/thanhpk/randstr"
	tele "gopkg.in/telebot.v3"
)

var (
	b             *tele.Bot
	templateHtml  string
	seleniumUrl   string
	seleniumMutex sync.Mutex
	service       selenium.WebDriver
	session       string
)

const (
	storagePath = "/tmp/quote/"
)

func main() {
	go runStaticFilesServer()

	seleniumUrl = os.Getenv("SELENIUM") + "/wd/hub"

	time.Sleep(10 * time.Second) // wait until selenium start

	var err error
	service, err = selenium.NewRemote(selenium.Capabilities{"browserName": "chrome"}, seleniumUrl)
	if err != nil {
		fmt.Printf("Failed to connect to the Selenium WebDriver: %v\n", err)
		log.Fatalln(err)
	}
	session = service.SessionID()
	defer service.Quit()

	t, err := os.ReadFile("./quote.html")
	if err != nil {
		log.Fatal(err)
	}
	templateHtml = string(t)

	pref := tele.Settings{
		Token:  os.Getenv("TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err = tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
		return
	}

	b.Handle(tele.OnText, handleMessage)
	b.Start()
}

func handleMessage(c tele.Context) error {
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

	token := randstr.String(16)
	avatarFileName := "avatar-placeholder.png"
	avatarFilePath := ""

	if len(photos) > 0 {
		avatarPhoto := photos[0]
		avatarFile, err := b.FileByID(avatarPhoto.FileID)
		if err != nil {
			return err
		}

		avatarFileName = token + "_avatar.jpeg"
		avatarFilePath = storagePath + avatarFileName
		err = b.Download(&avatarFile, avatarFilePath)
		if err != nil {
			return err
		}
	}

	htmlFileName := token + ".html"
	htmlFilePath := storagePath + htmlFileName
	pngFile := storagePath + token + ".png"
	webpFile := storagePath + token + ".webp"

	defer func() {
		os.Remove(htmlFilePath)
		os.Remove(pngFile)
		os.Remove(webpFile)
		if avatarFilePath != "" {
			os.Remove(avatarFilePath)
		}
	}()

	templateCopy := templateHtml
	templateCopy = strings.Replace(templateCopy, "{{author}}", html.EscapeString(author), -1)
	templateCopy = strings.Replace(templateCopy, "{{message}}", html.EscapeString(c.Text()), -1)
	templateCopy = strings.Replace(templateCopy, "{{avatar_file_path}}", avatarFileName, -1)

	err = os.WriteFile(htmlFilePath, []byte(templateCopy), 0644)
	if err != nil {
		return err
	}

	err = htmlToPng(htmlFileName, pngFile)
	if err != nil {
		return err
	}

	_, err = shellexec("cwebp " + pngFile + " -o " + webpFile)
	if err != nil {
		return err
	}

	sticker := &tele.Sticker{
		File: tele.FromDisk(webpFile),
	}

	return c.Send(sticker)
}

func htmlToPng(htmlFileName string, pngFilePath string) error {
	time.Sleep(100 * time.Millisecond)

	seleniumMutex.Lock()
	defer seleniumMutex.Unlock()

	ip, err := getMyIP()
	if err != nil {
		fmt.Printf("Failed to get ip: %v\n", err)
		return err
	}

	fileUrl := "http://" + ip + ":80/" + htmlFileName
	if err := service.Get(fileUrl); err != nil {
		fmt.Printf("Failed to load page: %v\n", err)
		return err
	}
	time.Sleep(200 * time.Millisecond)

	params := map[string]interface{}{
		"color": map[string]interface{}{"r": 0, "g": 0, "b": 0, "a": 0},
	}
	if err := sendCommand(session, seleniumUrl, "Emulation.setDefaultBackgroundColorOverride", params); err != nil {
		fmt.Printf("Failed to set default background color override: %v\n", err)
		return err
	}

	script := `return document.querySelector('.card').getBoundingClientRect();`
	size, err := service.ExecuteScript(script, nil)
	if err != nil {
		fmt.Printf("Failed to get element size: %v\n", err)
		return err
	}

	deviceMetrics := map[string]interface{}{
		"width":             512,
		"height":            int(size.(map[string]any)["height"].(float64)) + 32,
		"deviceScaleFactor": 1,
		"mobile":            false,
		"scale":             1,
	}
	if err := sendCommand(session, seleniumUrl, "Emulation.setDeviceMetricsOverride", deviceMetrics); err != nil {
		fmt.Printf("Failed to set device metrics override: %v\n", err)
		return err
	}

	screenshot, err := service.Screenshot()
	if err != nil {
		fmt.Printf("Failed to take screenshot: %v\n", err)
		return err
	}

	if err := os.WriteFile(pngFilePath, screenshot, 0644); err != nil {
		fmt.Printf("Failed to save screenshot: %v\n", err)
		return err
	}

	return nil
}

type commandParams struct {
	Cmd    string                 `json:"cmd"`
	Params map[string]interface{} `json:"params"`
}

func sendCommand(sessionID, url, command string, params map[string]interface{}) error {
	cmd := commandParams{
		Cmd:    command,
		Params: params,
	}
	cmdJSON, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed to marshal command: %v", err)
	}

	requestURL := fmt.Sprintf("%s/session/%s/chromium/send_command_and_get_result", url, sessionID)
	resp, err := http.Post(requestURL, "application/json", bytes.NewBuffer(cmdJSON))
	if err != nil {
		return fmt.Errorf("failed to send command: %v", err)
	}
	err = resp.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to close response body: %v", err)
	}

	return nil
}

func shellexec(command string) (string, error) {
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func getMyIP() (string, error) {
	o, err := shellexec("nslookup app | awk '/^Address: / { print $2 }' | tail -n 1")
	if err != nil {
		return "", err
	}
	return strings.Trim(o, " \n\r\t\v\x00"), nil
}

func runStaticFilesServer() {
	dir := "/tmp/quote"

	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	err := http.ListenAndServe("0.0.0.0:80", nil)
	if err != nil {
		log.Fatal(err)
	}
}

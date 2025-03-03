package browserQuery

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var COUNT_TABS = 0

type Client struct {
	Conn http.ResponseWriter
	Ch   chan string
}

var (
	clientsMu sync.RWMutex
	clients   = make(map[*Client]bool)
)

func HandleSSE(c echo.Context) error {
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")

	flusher, ok := c.Response().Writer.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming unsupported")
	}

	client := &Client{
		Conn: c.Response(),
		Ch:   make(chan string),
	}

	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, client)
		clientsMu.Unlock()
		close(client.Ch)
		fmt.Println("[LOG] <HandleSSE> Client disconnected")
	}()

	go func() {
		for msg := range client.Ch {
			fmt.Fprintf(client.Conn, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}()

	// Keep the connection open.  No need to read from the client.
	<-c.Request().Context().Done() // Block until client disconnects
	fmt.Println("[LOG] <HandleSSE> Client disconnected")
	return nil
}

// every five minutes, check if there are clients gt than max_Clients
func TrimClients() {
	panic("test")
}

func main() {
	server := echo.New()

	// CORS middleware with custom configuration
	server.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{"*"}, // Or specific origins
		AllowMethods:     []string{echo.GET, echo.HEAD, echo.PUT, echo.PATCH, echo.POST, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
	}))

	//server.GET("/sse", HandleSSE)
	server.POST("/upload", UploadTabs)
	server.GET("/upload/count", UploadCount)

	server.Start(":53891")
}

// -- //
type TabInfo struct {
	Active          bool        `json:"active"`
	Attention       bool        `json:"attention"`
	Audible         bool        `json:"audible"`
	AutoDiscardable bool        `json:"autoDiscardable"`
	CookieStoreId   string      `json:"cookieStoreId"`
	Discarded       bool        `json:"discarded"`
	Height          int         `json:"height"`
	Hidden          bool        `json:"hidden"`
	Highlighted     bool        `json:"highlighted"`
	ID              int         `json:"id"`
	Incognito       bool        `json:"incognito"`
	Index           int         `json:"index"`
	IsArticle       bool        `json:"isArticle"`
	IsInReaderMode  interface{} `json:"isInReaderMode"` // Or a specific type if known
	LastAccessed    int64       `json:"lastAccessed"`
	MutedInfo       struct {
		Muted bool `json:"muted"`
	} `json:"mutedInfo"`
	OpenerTabId  int  `json:"openerTabId"`
	Pinned       bool `json:"pinned"`
	SharingState struct {
		Camera     bool        `json:"camera"`
		Microphone bool        `json:"microphone"`
		Screen     interface{} `json:"screen"` // Or a specific type if known
	} `json:"sharingState"`
	Status         string `json:"status"`
	SuccessorTabId int    `json:"successorTabId"`
	Width          int    `json:"width"`
	WindowId       int    `json:"windowId"`
}

type uploadTabsBody struct {
	Body string `json:"body"`
}

func UploadTabs(c echo.Context) error {
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	var tabs []TabInfo
	a := c.Request().Body
	err := json.NewDecoder(a).Decode(&tabs)
	if err != nil {
		fmt.Println("[ERROR]", err.Error())
		return c.String(400, "NO")
	}
	COUNT_TABS = len(tabs)
	fmt.Println("[SUCCESS]", len(tabs))
	return c.String(200, "OK")
}

func UploadCount(c echo.Context) error {
	cnt := c.QueryParam("count")
	if cnt == "" {
		return c.String(400, "invalid query parameter `count`")
	}

	SendToRainmeter(cnt)
	fmt.Println("[SUCCESS]", cnt)
	return c.String(200, "OK")
}

func SendToRainmeter(cnt string) {
	cmd := exec.Command("Rainmeter.exe", "!SetVariable", "TabsCount", cnt, "Tabs")
	err := cmd.Run()
	if err != nil {
		log.Fatalf("%s", err.Error())
	}
	//Rainmeter.exe !SetVariable TabsCount "4" "illustro\Clock - Copy"
}

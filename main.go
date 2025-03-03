package browserQuery

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/eissar/nest/sse"
	"github.com/labstack/echo/v4"
)

//	type Client struct {
//		Conn http.ResponseWriter
//		Ch   chan string
//	}
type Client sse.Client

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

type TabInfoCallback func(c echo.Context, t []TabInfo)

// handler for response from client
func UploadTabs(c echo.Context) error {
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")

	var tabs []TabInfo
	a := c.Request().Body
	err := json.NewDecoder(a).Decode(&tabs)
	if err != nil {
		fmt.Println("[ERROR]", err.Error())
		return c.String(400, "NO")
	}
	fmt.Println("[SUCCESS]", len(tabs))
	return c.String(200, "OK")
}

func UploadCount(c echo.Context) error {
	cnt := c.QueryParam("count")
	if cnt == "" {
		return c.String(400, "invalid query parameter `count`")
	}

	fmt.Println("[SUCCESS]", cnt)
	return c.String(200, "OK")
}

// creates closure to handle tabs response from client
func UploadTabsHandler(callback TabInfoCallback) echo.HandlerFunc {
	// handler for response from client
	//
	// TabInfoCallback func(c echo.Context, t TabInfo)
	//
	return func(c echo.Context) error {
		var tabs []TabInfo
		a := c.Request().Body
		err := json.NewDecoder(a).Decode(&tabs)
		if err != nil {
			fmt.Println("[ERROR]", err.Error())
			return c.String(400, "NO")
		}
		fmt.Println("[SUCCESS]", len(tabs))

		callback(c, tabs)
		return c.String(200, "OK")
	}

}

func RegisterRootRoutes(server *echo.Echo) {
	server.GET("/eagleApp/sse", HandleSSE)
	server.POST("/api/uploadTabs", UploadTabs)
}

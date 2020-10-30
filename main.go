package main

/**
启动时接收服务端websocket连接信息, 连接到服务端
*/
import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8080", "http service address")

func retryConnect(err error) {
	log.Println("error occurred:", err)
	for i := 3; i > 0; i-- {
		log.Printf("retry to connect to server in %s second after", strconv.Itoa(i))
		time.Sleep(time.Duration(1) * time.Second) // 当连接不上服务端时每隔3秒钟重新连接一次服务端
	}

	main()
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/agent/select_action"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		retryConnect(err)
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				retryConnect(nil)
				log.Println("read:", err)
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			log.Println("done")
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(""))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
		time.Sleep(time.Duration(1) * time.Second) // 每隔1秒钟消费一次动作
	}
}

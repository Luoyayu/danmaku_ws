package danmaku_ws

import (
	ctx "context"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"time"

	"github.com/fatih/color"
	ws "github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var FilterRegexp *regexp.Regexp

func init() {
	FilterRegexp, _ = regexp.Compile(`(.*)【(.*)】|(.*)【(.*)`)
}

func filterDanmaku(danmaku string) []byte {
	return FilterRegexp.Find([]byte(danmaku))
}

type Client struct {
	Id   int64
	Conn *ws.Conn
	Ack  chan bool

	Debug        bool
	Sync         bool
	Filter       bool
	FilterRegexp string
}

// 0	4	Packet Length [length of WebSocket Frame]
// 4	2	Header Length [1000 => 0010]
// 6	2	Protocol Version [0002]
// 8	4	Operation
// 12	4	Sequence Id [0000 0001]
// 16	-	Body

func (c *Client) Init(context ctx.Context) error {
	if c.Filter {
		FilterRegexp, _ = regexp.Compile(c.FilterRegexp)
		log.Println("[INFO] Apply Danmaku Filter")
	}

	if c.Sync {
		log.Println("[INFO] Sync Print Danmaku")
	} else {
		log.Println("[INFO] Async Print Danmaku")
	}

	if room, err := RoomInit(c.Id); err == nil {
		if room.Data.LiveStatus == 0 {
			fmt.Println("🚧 Not on the air 🚧")
		}
		c.Id = room.Data.RoomId
		c.Logf("[INFO] Ready to estimate conn to danmaku server %q\n", DanmakuHost)

		c.Ack = make(chan bool)

		if c.Conn, _, err = (&ws.Dialer{}).Dial(DanmakuHost, nil); err != nil {
			return err
		}

		go c.sendAuth()
		go c.recvMsg(context)

		if <-c.Ack == true {
			fmt.Printf("Conn has benn estimated to danmaku server in room %v\n", c.Id)
			go c.sendHeartBeat(context)
		}

		select {
		case <-context.Done():
			return nil
		}
	} else {
		return errors.Wrap(err, fmt.Sprintf("[ERR] Init room for %v\n", c.Id))
	}
}

func (c *Client) sendAuth() {
	authParams := map[string]interface{}{
		"clientver": "1.6.3",
		"platform":  "web",
		"protover":  1,
		"roomid":    c.Id,
		"uid":       int(rand.Float64()*200000000000000.0 + 100000000000000.0),
		"type":      2,
	}

	body, _ := json.Marshal(authParams)
	handshake := fmt.Sprintf("%08x00100001%08x00000001", len(string(body))+16, OpsAuth)

	buf := make([]byte, len(handshake)>>1)
	hex.Decode(buf, []byte(handshake))
	c.Log("[INFO] Send auth package to danmaku server.")
	c.Conn.WriteMessage(ws.BinaryMessage, append(buf, body...))
}

func (c *Client) sendHeartBeat(context ctx.Context) {
	for {
		select {
		case <-context.Done():
			c.Log("[INFO] Stop send heart beat package.")
			return
		default:
			buf := make([]byte, 16)
			hex.Decode(buf, []byte("0000001f001000010000000200000001"))
			c.Conn.WriteMessage(ws.BinaryMessage, buf)
			time.Sleep(30 * time.Second)
		}
	}
}

func (c *Client) splitBufferAsync(buffer []byte) {
	for i, packSize := uint32(0), uint32(0); i < uint32(len(buffer)); i += packSize {
		packSize = binary.BigEndian.Uint32(buffer[i : i+4])
		go c.handleMsg(buffer[i : i+packSize])
	}
	return
}

func (c *Client) splitBufferSync(buffer []byte) (bufferPacks [][]byte) {
	//fmt.Println(hex.Dump(buffer))
	for i, packSize := uint32(0), uint32(0); i < uint32(len(buffer)); i += packSize {
		packSize = binary.BigEndian.Uint32(buffer[i : i+4])
		//log.Println(hex.Dump(buffer[i:i+4]), " => packSize = ", packSize)
		bufferPacks = append(bufferPacks, buffer[i:i+packSize])
	}
	return
}

func (c *Client) recvMsg(context ctx.Context) {
	for {
		select {
		case <-context.Done():
			c.Log("[INFO] Stop receive danmaku.")
			return
		default:
			_, msg, _ := c.Conn.ReadMessage()
			c.handleMsg(msg)
		}
	}
}

func (c *Client) handleMsg(msg []byte) {
	op, body := msg[11], msg[16:]
	switch op {
	case OpsNotify:
		c.Log("[INFO] Receive OpsNotify")
		c.handleOpsNotify(body)
	case OpsAuthACK:
		c.Log("[INFO] Receive OpsAuthACK")
		c.Ack <- true
	case OpsHeartBeatACK:
		c.Log("[INFO] Receive OpsHeartBeatACK")
		popularity := binary.BigEndian.Uint32(body)
		c.Logf("[INFO] popularity: %v\n", popularity)
	}
}

func (c *Client) handleOpsNotify(msgContent []byte) {
	var jMap map[string]interface{}
	if err := json.Unmarshal(msgContent, &jMap); err != nil {
		c.Log("[ERR] Can't Parse Message to Json, try do Zlib UnCompress")
		if c.Sync {
			bufferPacks := c.splitBufferSync(zlibDecompress(msgContent))
			c.Logf("[INFO] Cut to %d packs from message\n", len(bufferPacks))
			for i := range bufferPacks {
				c.Logf("[INFO] [%2d/%2d]Sub Message", i+1, len(bufferPacks))
				c.handleMsg(bufferPacks[i])
			}
		} else {
			c.splitBufferAsync(zlibDecompress(msgContent))
		}
	}
	cmd := jMap["cmd"]

	c.Logf("[INFO] OpsNotify#Command: %v\n", cmd)
	c.Logf("[INFO] OpsNotify#Message-Content: %#v\n", jMap)

	switch {
	case cmd == `DANMU_MSG`:
		c.Log("[INFO] OpsNotify: DANMU_MSG")

		var (
			info       = jMap["info"].([]interface{})
			danmaku    = info[1].(string)
			posterName = info[2].([]interface{})[1].(string)
		) // HACK ME!

		// always highlight filter
		if filterDanmaku(danmaku) != nil {
			color.New(color.FgBlue).Printf("‍🌐 %s（%s）\n", danmaku, posterName)
		} else if !c.Filter {
			fmt.Printf("%s（%s）\n", danmaku, posterName)
		}
	case cmd == `SUPER_CHAT_MESSAGE_JPN` || cmd == `SUPER_CHAT_MESSAGE`:
		var (
			data       = jMap["data"].(map[string]interface{})
			danmaku    = data["message"].(string)
			posterName = data["user_info"].(map[string]interface{})["uname"].(string)
		) // HACK ME!

		if !c.Filter {
			color.New(color.FgRed).Printf("%s（%s）\n", danmaku, posterName)
		}
	}
}

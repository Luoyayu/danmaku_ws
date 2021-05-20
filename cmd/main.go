package main

import (
	ctx "context"
	"danmaku_ws"
	"flag"
	"log"
	"os/exec"
	"strings"
	"time"
)

var (
	rid          = flag.Int64("rid", -1, "room id")
	async        = flag.Bool("async", false, "async but faster")
	debug        = flag.Bool("debug", false, "debug for websockets conn")
	filter       = flag.Bool("filter", false, "danmaku filter switch ")
	filterRegexp = flag.String("regexp", `(.*)【(.*)】|(.*)【(.*)`, "danmaku filter regexp")
	play         = flag.Bool("play", false, "play bilibili live")
	player       = flag.String("player", "/usr/local/bin/mpv", "player used to play bilibili live")
	playerArgs   = flag.String("playerArgs", "", "args for player")
	delaySecond  = flag.Float64("delay", 0.5, "delay second")
)

func playLive(danmakuServerCancel ctx.CancelFunc) {
	room, err := danmaku_ws.RoomInit(*rid)
	if err != nil {
		log.Fatal(err)
	}

	if room.Data.LiveStatus != 1 {
		return
	}

	qn := "4"
	if strings.Contains(*playerArgs, "no-video") || strings.Contains(*playerArgs, "vo=null") {
		qn = "1"
	}

	playUrl, err := danmaku_ws.GetPlayUrl(room.Data.RoomId, qn)
	if err != nil {
		log.Fatal(err)
	}

	args := []string{" "}
	args = append(args, strings.Split(*playerArgs, " ")...)
	args = append(args, "--hwdec=yes", playUrl.Data.DUrl[0].Url)

	if err := (&exec.Cmd{
		Path: *player,
		Args: args}).Run(); err != nil {
		log.Fatal("player:", err)
	}
	log.Println("exit player")
	danmakuServerCancel()
}

func main() {
	flag.Parse()
	if *rid == -1 {
		log.Fatal("-rid: room id is needed!")
	}

	context, cancel := ctx.WithCancel(ctx.Background())

	if *play {
		go playLive(cancel)
	}

	time.Sleep(time.Duration(*delaySecond * float64(time.Second)))

	go func() {
		if err := (&danmaku_ws.Client{
			Id:           *rid,
			Debug:        *debug,
			Filter:       *filter,
			FilterRegexp: *filterRegexp,
			Sync:         !*async}).Init(context); err != nil {
			log.Fatal(err)
		}
	}()

	select {
	case <-context.Done():
		return
	}
}

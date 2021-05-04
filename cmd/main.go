package main

import (
	ctx "context"
	"danmaku_ws"
	"flag"
	"log"
)

func main() {
	var (
		rid          = flag.Int64("rid", -1, "room id")
		async        = flag.Bool("async", false, "async but faster")
		debug        = flag.Bool("debug", false, "debug for websockets conn")
		filter       = flag.Bool("filter", false, "danmaku filter switch ")
		filterRegexp = flag.String("regexp", `(.*)【(.*)】|(.*)【(.*)`, "danmaku filter regexp")
	)

	flag.Parse()
	if *rid == -1 {
		log.Fatal("-rid: room id is needed!")
	}

	c := &danmaku_ws.Client{Id: *rid, Debug: *debug, Filter: *filter, FilterRegexp: *filterRegexp, Sync: !*async}
	context, cancel := ctx.WithCancel(ctx.Background())

	if err := c.Init(context); err != nil {
		log.Fatal(err)
	}
	cancel() // no sense
}

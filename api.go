package danmaku_ws

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
)

type Room struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
	Data    *struct {
		RoomId     int64 `json:"room_id"`
		ShortId    int64 `json:"short_id"`
		Uid        int64 `json:"uid"`
		LiveStatus int   `json:"live_status"`
		LiveTime   int64 `json:"live_time"`
	} `json:"data"`
}

func RoomInit(roomID interface{}) (room *Room, err error) {
	var ret interface{}
	if ret, err = GetDefault("https://api.live.bilibili.com/room/v1/Room/room_init",
		map[string]interface{}{"id": roomID}, &Room{}); err == nil {
		room = ret.(*Room)
	}
	return room, errors.Wrap(err, "room_init")
}

type PlayUrl struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Message string `json:"message"`
	Data    *struct {
		DUrl []struct {
			Url string
		} `json:"durl"`
	} `json:"data"`
}

func GetPlayUrl(longRoomID interface{}, quality string) (playUrl *PlayUrl, err error) {
	var ret interface{}
	if ret, err = GetDefault("https://api.live.bilibili.com/room/v1/Room/playUrl",
		map[string]interface{}{"cid": longRoomID, "quality": quality, "platform": "h5"}, &PlayUrl{}); err == nil {
		playUrl = ret.(*PlayUrl)
	}
	return playUrl, errors.Wrap(err, "playUrl")
}

func GetDefault(url_ string, params map[string]interface{}, in interface{}) (out interface{}, err error) {
	out = in
	l := url.Values{}
	for k, v := range params {
		l.Add(k, fmt.Sprint(v))
	}

	var resp *http.Response
	req, _ := http.NewRequest("GET", url_+"?"+l.Encode(), nil)
	if resp, err = (&http.Client{}).Do(req); err == nil {
		defer resp.Body.Close()
		if err = json.NewDecoder(resp.Body).Decode(&in); err != nil {
			return
		}
	}
	return
}

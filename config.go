package danmaku_ws

const DanmakuHost = "wss://broadcastlv.chat.bilibili.com/sub"

const (
	OpsHeartBeat    = 2 // 0000 0002
	OpsHeartBeatACK = 3 // 0000 0003

	OpsAuth    = 7 // 0000 0007
	OpsAuthACK = 8 // 0000 0008

	OpsNotify = 5 // 0000 0005
)

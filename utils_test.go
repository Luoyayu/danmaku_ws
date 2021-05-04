package danmaku_ws

import (
	"reflect"
	"testing"
)

func Test_extractTranslator(t *testing.T) {
	type args struct {
		danmaku string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			args: args{
				danmaku: "普通弹幕",
			},
			want: nil,
		},
		{
			args: args{
				danmaku: "【中文方括号】",
			},
			want: []byte("【中文方括号】"),
		},
		{
			args: args{
				danmaku: "人名【中文方括号】",
			},
			want: []byte("人名【中文方括号】"),
		},
		{
			args: args{
				danmaku: "“中文引号”",
			},
			want: nil,
		},
		{
			args: args{
				danmaku: "【一半的中文方括号",
			},
			want: []byte("【一半的中文方括号"),
		},
		{
			args: args{
				danmaku: "【】",
			},
			want: []byte("【】"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterDanmaku(tt.args.danmaku); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterDanmaku() = %v, want %v", got, tt.want)
			}
		})
	}
}

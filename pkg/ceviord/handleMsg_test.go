package ceviord

import "testing"

func TestReplaceMsg(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "URL-http",
			args: args{msg: "http://google.com"},
			want: "ゆーあーるえる。",
		},
		{
			name: "URL-https",
			args: args{msg: "https://google.com"},
			want: "ゆーあーるえる。",
		},
		{name: `ほげ\nふが`, args: args{msg: "ほげ\nふが"}, want: `ほげ ふが`},
		{name: `ほげ\r\nふが`, args: args{msg: "ほげ\r\nふが"}, want: "ほげ\r ふが"},
		{name: `半角チルダ`, args: args{msg: "あ~"}, want: "あー"},
		{name: `全角チルダ`, args: args{msg: "あ〜"}, want: "あー"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ReplaceMsg(tt.args.msg); got != tt.want {
				t.Errorf("ReplaceMsg() = %v, want %v", got, tt.want)
			}
		})
	}
}

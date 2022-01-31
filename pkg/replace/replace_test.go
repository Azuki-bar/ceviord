package replace

import (
	"testing"
)

func TestRecords_Replace(t *testing.T) {
	type args struct {
		msg string
	}
	tests := []struct {
		name string
		ds   dicts
		args args
		want string
	}{
		{name: "a->b and b->a", ds: dicts{
			Dict{Word: "a", Yomi: "b"}, Dict{Word: "b", Yomi: "a"}},
			args: args{msg: "ab"}, want: "ba",
		},
		{
			name: "a->b and b->a", ds: dicts{
				Dict{Word: "a", Yomi: "b"}, Dict{Word: "b", Yomi: "a"},
			},
			args: args{msg: "aaaaabbbbb"}, want: "bbbbbaaaaa",
		},
		{
			name: "a -> b b->c ...", ds: dicts{
				Dict{Word: "a", Yomi: "b"},
				Dict{Word: "b", Yomi: "c"},
				Dict{Word: "c", Yomi: "d"},
				Dict{Word: "d", Yomi: "e"},
				Dict{Word: "e", Yomi: "f"},
				Dict{Word: "f", Yomi: "g"},
				Dict{Word: "g", Yomi: "h"},
				Dict{Word: "h", Yomi: "i"},
				Dict{Word: "i", Yomi: "j"},
				Dict{Word: "j", Yomi: "k"},
				Dict{Word: "k", Yomi: "l"},
				Dict{Word: "l", Yomi: "m"},
				Dict{Word: "m", Yomi: "n"},
				Dict{Word: "n", Yomi: "o"},
				Dict{Word: "o", Yomi: "p"},
				Dict{Word: "p", Yomi: "q"},
				Dict{Word: "q", Yomi: "r"},
				Dict{Word: "r", Yomi: "s"},
				Dict{Word: "s", Yomi: "t"},
				Dict{Word: "t", Yomi: "u"},
				Dict{Word: "u", Yomi: "v"},
				Dict{Word: "v", Yomi: "w"},
				Dict{Word: "w", Yomi: "x"},
				Dict{Word: "x", Yomi: "y"},
				Dict{Word: "y", Yomi: "z"},
				Dict{Word: "z", Yomi: "a"},
			},
			args: args{msg: "a"}, want: "b",
		},
		{
			name: "replace for long", ds: dicts{
				Dict{Word: "a", Yomi: "aa"}, Dict{Word: "b", Yomi: "bb"},
			},
			args: args{msg: "ab"}, want: "aabb",
		},
		{
			name: "recursive change", ds: dicts{
				Dict{Word: "ab", Yomi: "ba"}, Dict{Word: "b", Yomi: "c"},
			},
			args: args{msg: "ababab"}, want: "bababa",
		},
		{
			name: "japanese char replace", ds: dicts{
				Dict{Word: "初音ミク", Yomi: "鏡音リン"}, Dict{Word: "b", Yomi: "c"},
			},
			args: args{msg: `こんにちは初音ミクだよ。`}, want: "こんにちは鏡音リンだよ。",
		},
		{
			name: "no replace", ds: dicts{
				Dict{Word: "初音ミク", Yomi: "鏡音リン"},
			},
			args: args{msg: `こんにちは。`}, want: `こんにちは。`,
		},
		{
			name: "no records", ds: dicts{},
			args: args{msg: `こんにちは。`}, want: `こんにちは。`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ds.replace(tt.args.msg); got != tt.want {
				t.Errorf("replace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestApplySysDict(t *testing.T) {
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
			args: args{msg: "http://example.com"},
			want: "ゆーあーるえる。",
		},
		{
			name: "URL-https",
			args: args{msg: "https://example.com"},
			want: "ゆーあーるえる。",
		},
		{
			name: "URL-ftps",
			args: args{msg: "ftps://example.com"},
			want: "ftps://example.com",
		},
		{
			name: "name and https url",
			args: args{msg: `recommend contents -> https://example.com`},
			want: `recommend contents -> ゆーあーるえる。`,
		},
		{
			name: "https url and content",
			args: args{msg: `recommend contents
https://example.com

`},
			want: `recommend contents ゆーあーるえる。  `,
		},
		{
			name: "https url and content",
			args: args{msg: `https://example.com <- recommend contents!`},
			want: `ゆーあーるえる。`,
		},
		{
			name: "https url and content",
			args: args{msg: `https://example.com
something contents.

`},
			want: `ゆーあーるえる。 something contents.  `,
		},
		{name: `ほげ\nふが`, args: args{msg: "ほげ\nふが"}, want: `ほげ ふが`},
		{name: `ほげ\r\nふが`, args: args{msg: "ほげ\r\nふが"}, want: "ほげ\r ふが"},
		{name: `半角チルダ`, args: args{msg: "あ~"}, want: "あー"},
		{name: `全角チルダ`, args: args{msg: "あ〜"}, want: "あー"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ApplySysDict(tt.args.msg); got != tt.want {
				t.Errorf("ApplySysDict() = %v, want %v", got, tt.want)
			}
		})
	}
}

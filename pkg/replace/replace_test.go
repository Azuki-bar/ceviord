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
		ds   Dicts
		args args
		want string
	}{
		{name: "a->b and b->a", ds: Dicts{
			Dict{UserDictInput: UserDictInput{Word: "a", Yomi: "b"}},
			Dict{UserDictInput: UserDictInput{Word: "b", Yomi: "a"}},
		},
			args: args{msg: "ab"}, want: "ba",
		},
		{
			name: "a->b and b->a", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "a", Yomi: "b"}},
				Dict{UserDictInput: UserDictInput{Word: "b", Yomi: "a"}},
			},
			args: args{msg: "aaaaabbbbb"}, want: "bbbbbaaaaa",
		},
		{
			name: "a -> b b->c ...", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "a", Yomi: "b"}},
				Dict{UserDictInput: UserDictInput{Word: "b", Yomi: "c"}},
				Dict{UserDictInput: UserDictInput{Word: "c", Yomi: "d"}},
			},
			args: args{msg: "a"}, want: "b",
		},
		{
			name: "replace for long", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "a", Yomi: "aa"}},
				Dict{UserDictInput: UserDictInput{Word: "aa", Yomi: "aaa"}},
				Dict{UserDictInput: UserDictInput{Word: "aaa", Yomi: "aaaa"}},
			},
			args: args{msg: "ab"}, want: "aab",
		},
		{
			name: "replace for long", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "a", Yomi: "aa"}},
				Dict{UserDictInput: UserDictInput{Word: "b", Yomi: "bb"}},
			},
			args: args{msg: "ab"}, want: "aabb",
		},
		{
			name: "recursive change", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "ab", Yomi: "ba"}},
				Dict{UserDictInput: UserDictInput{Word: "b", Yomi: "c"}},
			},
			args: args{msg: "ababab"}, want: "bababa",
		},
		{
			name: "japanese char replace", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "初音ミク", Yomi: "鏡音リン"}},
				Dict{UserDictInput: UserDictInput{Word: "b", Yomi: "c"}},
			},
			args: args{msg: `こんにちは初音ミクだよ。`}, want: "こんにちは鏡音リンだよ。",
		},
		{
			name: "no replace", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "初音ミク", Yomi: "鏡音リン"}},
			},
			args: args{msg: `こんにちは。`}, want: `こんにちは。`,
		},
		{
			name: "no records", ds: Dicts{},
			args: args{msg: `こんにちは。`}, want: `こんにちは。`,
		},
		{
			name: "GITHUB", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "github", Yomi: "ぎっとはぶ"}},
			},
			args: args{msg: `GITHUB`}, want: `ぎっとはぶ`,
		},
		{
			name: "Github", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "GITHUB", Yomi: "ぎっとはぶ"}},
			},
			args: args{msg: `GitHub`}, want: `ぎっとはぶ`,
		},
		{
			name: "Github", ds: Dicts{
				Dict{UserDictInput: UserDictInput{Word: "GITHUB", Yomi: "ぎっとはぶ"}},
			},
			args: args{msg: `GitHubABC`}, want: `ぎっとはぶABC`,
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
			want: "URL ",
		},
		{
			name: "URL-https",
			args: args{msg: "https://example.com"},
			want: "URL ",
		},
		{
			name: "URL-ftps",
			args: args{msg: "ftps://example.com"},
			want: "ftps://example.com",
		},
		{
			name: "name and https url",
			args: args{msg: `recommend contents -> https://example.com`},
			want: `recommend contents -> URL `,
		},
		{
			name: "https url and content",
			args: args{msg: `recommend contents
https://example.com

`},
			want: `recommend contents URL   `,
		},
		{
			name: "https url and content",
			args: args{msg: `https://example.com <- recommend contents!`},
			want: `URL `,
		},
		{
			name: "https url and content",
			args: args{msg: `https://example.com
something contents.

`},
			want: `URL  something contents.  `,
		},
		{name: `ほげ\nふが`, args: args{msg: "ほげ\nふが"}, want: `ほげ ふが`},
		{name: `ほげ\r\nふが`, args: args{msg: "ほげ\r\nふが"}, want: "ほげ\r ふが"},
		{name: `半角チルダ`, args: args{msg: "あ~"}, want: "あー"},
		{name: `全角チルダ`, args: args{msg: "あ〜"}, want: "あー"},
		{name: `custom emoji`, args: args{msg: `<:thinking_vim:1234>`}, want: ""},
		{name: `custom anime emoji`, args: args{msg: `<a:thinking_vim:1234>`}, want: ""},
		{name: `打ち消しされない`, args: args{msg: `~~`}, want: "ーー"},
		{name: `イタリック`, args: args{msg: `~a~`}, want: "ーaー"},
		{name: `打ち消し`, args: args{msg: `~~a~~`}, want: " ピーー"},
		{name: `複数文字を打ち消す`, args: args{msg: `~~aa~~`}, want: " ピーー"},
		{name: `途中からイタリック`, args: args{msg: `ああ~いい~`}, want: "ああーいいー"},
		{name: `打ち消しできていない`, args: args{msg: `ああ~~いい~`}, want: "ああーーいいー"},
		{name: `途中から打ち消し`, args: args{msg: `ああ~~いい~~`}, want: "ああ ピーー"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ApplySysDict(tt.args.msg); got != tt.want {
				t.Errorf("ApplySysDict() = %v, want %v", got, tt.want)
			}
		})
	}
}

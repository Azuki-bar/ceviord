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
				Dict{UserDictInput: UserDictInput{Word: "d", Yomi: "e"}},
				Dict{UserDictInput: UserDictInput{Word: "e", Yomi: "f"}},
				Dict{UserDictInput: UserDictInput{Word: "f", Yomi: "g"}},
				Dict{UserDictInput: UserDictInput{Word: "g", Yomi: "h"}},
				Dict{UserDictInput: UserDictInput{Word: "h", Yomi: "i"}},
				Dict{UserDictInput: UserDictInput{Word: "i", Yomi: "j"}},
				Dict{UserDictInput: UserDictInput{Word: "j", Yomi: "k"}},
				Dict{UserDictInput: UserDictInput{Word: "k", Yomi: "l"}},
				Dict{UserDictInput: UserDictInput{Word: "l", Yomi: "m"}},
				Dict{UserDictInput: UserDictInput{Word: "m", Yomi: "n"}},
				Dict{UserDictInput: UserDictInput{Word: "n", Yomi: "o"}},
				Dict{UserDictInput: UserDictInput{Word: "o", Yomi: "p"}},
				Dict{UserDictInput: UserDictInput{Word: "p", Yomi: "q"}},
				Dict{UserDictInput: UserDictInput{Word: "q", Yomi: "r"}},
				Dict{UserDictInput: UserDictInput{Word: "r", Yomi: "s"}},
				Dict{UserDictInput: UserDictInput{Word: "s", Yomi: "t"}},
				Dict{UserDictInput: UserDictInput{Word: "t", Yomi: "u"}},
				Dict{UserDictInput: UserDictInput{Word: "u", Yomi: "v"}},
				Dict{UserDictInput: UserDictInput{Word: "v", Yomi: "w"}},
				Dict{UserDictInput: UserDictInput{Word: "w", Yomi: "x"}},
				Dict{UserDictInput: UserDictInput{Word: "x", Yomi: "y"}},
				Dict{UserDictInput: UserDictInput{Word: "y", Yomi: "z"}},
				Dict{UserDictInput: UserDictInput{Word: "z", Yomi: "a"}},
			},
			args: args{msg: "a"}, want: "b",
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

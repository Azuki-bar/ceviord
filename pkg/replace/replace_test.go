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
		rs   Records
		args args
		want string
	}{
		{name: "a->b and b->a", rs: Records{
			Record{Before: "a", After: "b"}, Record{Before: "b", After: "a"},
		},
			args: args{msg: "ab"}, want: "ba",
		},
		{name: "a->b and b->a", rs: Records{
			Record{Before: "a", After: "b"}, Record{Before: "b", After: "a"},
		},
			args: args{msg: "aaaaabbbbb"}, want: "bbbbbaaaaa",
		},
		{name: "a -> b b->c ...", rs: Records{
			Record{Before: "a", After: "b"},
			Record{Before: "b", After: "c"},
			Record{Before: "c", After: "d"},
			Record{Before: "d", After: "e"},
			Record{Before: "e", After: "f"},
			Record{Before: "f", After: "g"},
			Record{Before: "g", After: "h"},
			Record{Before: "h", After: "i"},
			Record{Before: "i", After: "j"},
			Record{Before: "j", After: "k"},
			Record{Before: "k", After: "l"},
			Record{Before: "l", After: "m"},
			Record{Before: "m", After: "n"},
			Record{Before: "n", After: "o"},
			Record{Before: "o", After: "p"},
			Record{Before: "p", After: "q"},
			Record{Before: "q", After: "r"},
			Record{Before: "r", After: "s"},
			Record{Before: "s", After: "t"},
			Record{Before: "t", After: "u"},
			Record{Before: "u", After: "v"},
			Record{Before: "v", After: "w"},
			Record{Before: "w", After: "x"},
			Record{Before: "x", After: "y"},
			Record{Before: "y", After: "z"},
			Record{Before: "z", After: "a"},
		},
			args: args{msg: "a"}, want: "b",
		},
		{name: "replace for long", rs: Records{
			Record{Before: "a", After: "aa"}, Record{Before: "b", After: "bb"},
		},
			args: args{msg: "ab"}, want: "aabb",
		},
		{name: "recursive change", rs: Records{
			Record{Before: "ab", After: "ba"}, Record{Before: "b", After: "c"},
		},
			args: args{msg: "ababab"}, want: "bababa",
		},
		{name: "japanese char replace", rs: Records{
			Record{Before: "初音ミク", After: "鏡音リン"}, Record{Before: "b", After: "c"},
		},
			args: args{msg: `こんにちは初音ミクだよ。`}, want: "こんにちは鏡音リンだよ。",
		},
		{name: "no replace", rs: Records{
			Record{Before: "初音ミク", After: "鏡音リン"},
		},
			args: args{msg: `こんにちは。`}, want: `こんにちは。`,
		},
		{name: "no records", rs: Records{},
			args: args{msg: `こんにちは。`}, want: `こんにちは。`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.rs.Replace(tt.args.msg); got != tt.want {
				t.Errorf("Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}

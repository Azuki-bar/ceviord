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
			Dict{Before: "a", After: "b"}, Dict{Before: "b", After: "a"}},
			args: args{msg: "ab"}, want: "ba",
		},
		{
			name: "a->b and b->a", ds: dicts{
				Dict{Before: "a", After: "b"}, Dict{Before: "b", After: "a"},
			},
			args: args{msg: "aaaaabbbbb"}, want: "bbbbbaaaaa",
		},
		{
			name: "a -> b b->c ...", ds: dicts{
				Dict{Before: "a", After: "b"},
				Dict{Before: "b", After: "c"},
				Dict{Before: "c", After: "d"},
				Dict{Before: "d", After: "e"},
				Dict{Before: "e", After: "f"},
				Dict{Before: "f", After: "g"},
				Dict{Before: "g", After: "h"},
				Dict{Before: "h", After: "i"},
				Dict{Before: "i", After: "j"},
				Dict{Before: "j", After: "k"},
				Dict{Before: "k", After: "l"},
				Dict{Before: "l", After: "m"},
				Dict{Before: "m", After: "n"},
				Dict{Before: "n", After: "o"},
				Dict{Before: "o", After: "p"},
				Dict{Before: "p", After: "q"},
				Dict{Before: "q", After: "r"},
				Dict{Before: "r", After: "s"},
				Dict{Before: "s", After: "t"},
				Dict{Before: "t", After: "u"},
				Dict{Before: "u", After: "v"},
				Dict{Before: "v", After: "w"},
				Dict{Before: "w", After: "x"},
				Dict{Before: "x", After: "y"},
				Dict{Before: "y", After: "z"},
				Dict{Before: "z", After: "a"},
			},
			args: args{msg: "a"}, want: "b",
		},
		{
			name: "replace for long", ds: dicts{
				Dict{Before: "a", After: "aa"}, Dict{Before: "b", After: "bb"},
			},
			args: args{msg: "ab"}, want: "aabb",
		},
		{
			name: "recursive change", ds: dicts{
				Dict{Before: "ab", After: "ba"}, Dict{Before: "b", After: "c"},
			},
			args: args{msg: "ababab"}, want: "bababa",
		},
		{
			name: "japanese char replace", ds: dicts{
				Dict{Before: "初音ミク", After: "鏡音リン"}, Dict{Before: "b", After: "c"},
			},
			args: args{msg: `こんにちは初音ミクだよ。`}, want: "こんにちは鏡音リンだよ。",
		},
		{
			name: "no replace", ds: dicts{
				Dict{Before: "初音ミク", After: "鏡音リン"},
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

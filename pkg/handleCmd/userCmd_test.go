package handleCmd

import (
	"reflect"
	"strings"
	"testing"
)

func Test_bye_parse(t *testing.T) {
	type args struct {
		in0 []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: "succeed pattern",
			args: args{
				in0: []string{"bye"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &byeOld{}
			if err := b.parse(tt.args.in0); (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_change_parse(t *testing.T) {
	type fields struct {
		changeTo string
	}
	type args struct {
		cmds []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "change to castA",
			fields:  fields{changeTo: "castA"},
			args:    args{cmds: []string{"change", "castA"}},
			wantErr: false,
		},
		{
			name:    "change to ``",
			fields:  fields{changeTo: ""},
			args:    args{cmds: []string{"change", ""}},
			wantErr: false,
		},
		{
			name:    "change to ` `",
			fields:  fields{changeTo: " "},
			args:    args{cmds: []string{"change", " "}},
			wantErr: false,
		},
		{
			name:    "change to castA with verbose args",
			fields:  fields{changeTo: "castA"},
			args:    args{cmds: []string{"change", "castA", "castB"}},
			wantErr: false,
		},
		{
			name:    "change cast is not provided",
			args:    args{cmds: []string{"change"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &changeOld{}
			if err := c.parse(tt.args.cmds); (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(c, &changeOld{changeTo: tt.fields.changeTo}) {
				t.Errorf("parse failed")
			}
		})
	}
}

func Test_dictAdd_parse(t *testing.T) {
	type fields struct {
		word string
		yomi string
	}
	type args struct {
		cmd []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "change eng -> kana",
			fields:  fields{word: "hoge", yomi: "ほげ"},
			args:    args{cmd: []string{"add", "hoge", "ほげ"}},
			wantErr: false,
		},
		{
			name:    "change eng -> eng",
			fields:  fields{word: "hoge", yomi: "fuga"},
			args:    args{cmd: []string{"add", "hoge", "fuga"}},
			wantErr: false,
		},
		{
			name:    "change emoji -> eng",
			fields:  fields{word: "🤔", yomi: "thinking"},
			args:    args{cmd: []string{"add", "🤔", "thinking"}},
			wantErr: false,
		},
		{
			name:    "not trim long string less 300",
			fields:  fields{word: strings.Repeat("a", 300), yomi: strings.Repeat("b", 300)},
			args:    args{cmd: []string{"add", strings.Repeat("a", 300), strings.Repeat("b", 300)}},
			wantErr: false,
		},
		{
			name:    "trim long string over 300",
			fields:  fields{word: strings.Repeat("a", 300), yomi: strings.Repeat("b", 300)},
			args:    args{cmd: []string{"add", strings.Repeat("a", 301), strings.Repeat("b", 301)}},
			wantErr: false,
		},
		{
			name:    "add `hoge` to empty",
			fields:  fields{word: "hoge", yomi: ""},
			args:    args{cmd: []string{"add", "hoge", ""}},
			wantErr: false,
		},
		{
			name:    "add `hoge` to ` `",
			fields:  fields{word: "hoge", yomi: " "},
			args:    args{cmd: []string{"add", "hoge", " "}},
			wantErr: false,
		},
		{
			name:    "add with verbose args",
			fields:  fields{word: "hoge", yomi: "fuga"},
			args:    args{cmd: []string{"add", "hoge", "fuga", "piyo"}},
			wantErr: false,
		},
		{
			name:    "not provide replace after words",
			args:    args{cmd: []string{"add", "hoge"}},
			wantErr: true,
		},
		{
			name:    "not provide replace words",
			args:    args{cmd: []string{"add"}},
			wantErr: true,
		},
		{
			name:    "not provide oparation cmd",
			args:    args{cmd: []string{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dictAdd{}
			if err := d.parse(tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(d, &dictAdd{word: tt.fields.word, yomi: tt.fields.yomi}) {
				t.Errorf("parse failed")
			}
		})
	}
}

func Test_dictDel_parse(t *testing.T) {
	type fields struct {
		ids []uint
	}
	type args struct {
		cmd []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "1 delete id provided",
			fields:  fields{ids: []uint{1}},
			args:    args{cmd: []string{"delete", "1"}},
			wantErr: false,
		},
		{
			name:    "2 delete id provided",
			fields:  fields{ids: []uint{1, 2}},
			args:    args{cmd: []string{"delete", "1", "2"}},
			wantErr: false,
		},
		{
			name:    "provide not uint id",
			fields:  fields{ids: []uint{}},
			args:    args{cmd: []string{"delete", "-1"}},
			wantErr: true,
		},

		{
			name:    "provide only NaN id",
			fields:  fields{ids: []uint{}},
			args:    args{cmd: []string{"delete", "hoge"}},
			wantErr: true,
		},
		{
			name:    "provide NaN id last",
			fields:  fields{ids: []uint{}},
			args:    args{cmd: []string{"delete", "1", "hoge"}},
			wantErr: true,
		},
		{
			name:    "provide NaN id and correct id",
			fields:  fields{ids: []uint{}},
			args:    args{cmd: []string{"delete", "hoge", "1"}},
			wantErr: true,
		},
		{
			name:    "delete arg not satisfied",
			fields:  fields{ids: nil},
			args:    args{cmd: []string{"delete"}},
			wantErr: true,
		},
		{
			name:    "cmd not satisfied",
			fields:  fields{ids: nil},
			args:    args{cmd: []string{}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dictDel{}
			if err := d.parse(tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(d, &dictDel{ids: tt.fields.ids}) {
				t.Errorf("parse failed")
			}
		})
	}
}

func Test_dictList_parse(t *testing.T) {
	type fields struct {
		isLatest bool
		from     uint
		to       uint
		limit    uint
	}
	type args struct {
		cmd []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "`dict list` shows latest 10 records",
			fields:  fields{isLatest: true, limit: 10},
			args:    args{cmd: []string{"list"}},
			wantErr: false,
		},
		{
			name:    "`dict list LIMITS` shows latest LIMITS records",
			fields:  fields{isLatest: true, limit: 100},
			args:    args{cmd: []string{"list", "100"}},
			wantErr: false,
		},
		{
			name:    "`dict list all` shows 2^32 -1 records",
			fields:  fields{isLatest: true, limit: 1<<32 - 1},
			args:    args{cmd: []string{"list", "all"}},
			wantErr: false,
		},
		{
			name:    "`dict list hoge` cannot parse id",
			fields:  fields{isLatest: true},
			args:    args{cmd: []string{"list", "hoge"}},
			wantErr: true,
		},
		{
			name:    "`dict list from 1 to 10`",
			fields:  fields{from: 1, to: 10},
			args:    args{cmd: []string{"list", "1", "10"}},
			wantErr: false,
		},
		{
			name:    "`dict list -1 10` negative number is not supported",
			fields:  fields{},
			args:    args{cmd: []string{"list", "-1", "10"}},
			wantErr: true,
		},
		{
			name:    "`dict list all 1` not support",
			fields:  fields{},
			args:    args{cmd: []string{"list", "all", "1"}},
			wantErr: true,
		},
		{
			name:    "`dict list 1 all` not support",
			fields:  fields{},
			args:    args{cmd: []string{"list", "1", "all"}},
			wantErr: true,
		},
		{
			name:    "`dict list` with verbose opt",
			fields:  fields{isLatest: false, from: 1, to: 10},
			args:    args{cmd: []string{"list", "1", "10", "100"}},
			wantErr: false,
		},
		{
			name:    "`dict list` with verbose opt are ignored",
			fields:  fields{isLatest: false, from: 1, to: 10},
			args:    args{cmd: []string{"list", "1", "10", "hoge"}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dictList{}
			if err := d.parse(tt.args.cmd); (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(d, &dictList{
				isLatest: tt.fields.isLatest,
				from:     tt.fields.from,
				to:       tt.fields.to,
				limit:    tt.fields.limit},
			) {
				t.Errorf("parse failed wants %+v; got %+v",
					&dictList{isLatest: tt.fields.isLatest, from: tt.fields.from, to: tt.fields.to, limit: tt.fields.limit}, d)
			}
		})
	}
}

func Test_dict_parse(t *testing.T) {
	type fields struct {
		sub userMainCmd
	}
	type args struct {
		cmds []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "dict add",
			fields:  fields{sub: &dictAdd{word: "word", yomi: "yomi"}},
			args:    args{cmds: []string{"dict", "add", "word", "yomi"}},
			wantErr: false,
		},
		{
			name:    "dict add failed",
			fields:  fields{sub: nil},
			args:    args{cmds: []string{"dict", "add", "word"}},
			wantErr: true,
		},
		{
			name:    "dict del",
			fields:  fields{sub: &dictDel{ids: []uint{1}}},
			args:    args{cmds: []string{"dict", "del", "1"}},
			wantErr: false,
		},
		{
			name:    "dict delete",
			fields:  fields{sub: &dictDel{ids: []uint{1}}},
			args:    args{cmds: []string{"dict", "delete", "1"}},
			wantErr: false,
		},
		{
			name:    "dict rm",
			fields:  fields{sub: &dictDel{ids: []uint{1}}},
			args:    args{cmds: []string{"dict", "rm", "1"}},
			wantErr: false,
		},
		{
			name:    "dict del id not provided",
			fields:  fields{sub: nil},
			args:    args{cmds: []string{"dict", "del"}},
			wantErr: true,
		},
		{
			name:    "dict list",
			fields:  fields{sub: &dictList{isLatest: true, from: 0, to: 0, limit: 10}},
			args:    args{cmds: []string{"dict", "list"}},
			wantErr: false,
		},
		{
			name:    "dict ls",
			fields:  fields{sub: &dictList{isLatest: true, from: 0, to: 0, limit: 10}},
			args:    args{cmds: []string{"dict", "ls"}},
			wantErr: false,
		},
		{
			name:    "dict show",
			fields:  fields{sub: &dictList{isLatest: true, from: 0, to: 0, limit: 10}},
			args:    args{cmds: []string{"dict", "show"}},
			wantErr: false,
		},
		{
			name:    "not supported cmd",
			fields:  fields{sub: nil},
			args:    args{cmds: []string{"dict", "hoge"}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &dictOld{}
			if err := d.parse(tt.args.cmds); (err != nil) != tt.wantErr {
				t.Errorf("parse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(d, &dictOld{sub: tt.fields.sub}) {
				t.Errorf("parse failed; want %+v, but %+v", &dictOld{sub: tt.fields.sub}, d)
			}
		})
	}
}

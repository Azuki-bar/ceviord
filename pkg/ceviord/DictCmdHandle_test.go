package ceviord

import (
	"ceviord/pkg/replace"
	"testing"
)

type ReplacerMock struct{}

func (ReplacerMock) Add(dict *replace.UserDictInput) error    { return nil }
func (ReplacerMock) Delete(dictId uint) (replace.Dict, error) { return replace.Dict{}, nil }

func Test_handleDictCmd(t *testing.T) {
	ceviord.dictController = ReplacerMock{}
	type args struct {
		content  string
		authorId string
		guildId  string
		dictCmd  string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "dict cmd not called",
			args:    args{content: prefix + "", dictCmd: "dict"},
			wantErr: true,
		},
		{
			name:    "dict sub cmd not called",
			args:    args{content: prefix + "dic", dictCmd: "dict"},
			wantErr: true,
		},
		{
			name:    "dict sub cmd not called",
			args:    args{content: prefix + "dict", dictCmd: "dict"},
			wantErr: true,
		},
		{
			name:    "not defined dict sub cmd",
			args:    args{content: prefix + "dict a", dictCmd: "dict"},
			wantErr: true,
		},
		{
			name:    "use defined dict sub cmd but replace string not shown",
			args:    args{content: prefix + "dict add", dictCmd: "dict"},
			wantErr: true,
		},
		{
			name:    "use defined dict sub cmd but replace string not shown",
			args:    args{content: prefix + "dict add old", dictCmd: "dict"},
			wantErr: true,
		},
		{
			name:    "use defined dict sub cmd but replace string not shown",
			args:    args{content: prefix + "dict add old new", dictCmd: "dict"},
			wantErr: false,
		},
		{
			name:    "use defined dict sub cmd but replace string not shown",
			args:    args{content: prefix + "dict add old new1 new2", dictCmd: "dict"},
			wantErr: false,
		},
		{
			name:    "use defined dict sub cmd but replace string not shown",
			args:    args{content: prefix + "dict del 1", dictCmd: "dict"},
			wantErr: false,
		},
		{
			name:    "use defined dict sub cmd but replace string not shown",
			args:    args{content: prefix + "dict del 1 s", dictCmd: "dict"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := handleDictCmd(tt.args.content, tt.args.authorId, tt.args.guildId, tt.args.dictCmd); (err != nil) != tt.wantErr {
				t.Errorf("handleDictCmd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

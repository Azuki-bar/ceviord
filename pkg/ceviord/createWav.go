package ceviord

import (
	"github.com/gotti/cevigo/pkg/cevioai"
	"log"
)

type cevioWav struct {
	talker    cevioai.ITalker2V40
	isSucceed chan error
}

func NewTalker() *cevioWav {
	c := cevioWav{isSucceed: make(chan error, 0)}
	c.talker = cevioai.NewITalker2V40(cevioai.CevioAiApiName)
	_, err := c.talker.SetCast("さとうさらら")
	if err != nil {
		log.Fatalln(err)
	}
	return &c
}

func (c *cevioWav) OutputWaveToFile(talkWard string, path string) (err error) {
	_, err = c.talker.OutputWaveToFile(talkWard, path)
	return err
}

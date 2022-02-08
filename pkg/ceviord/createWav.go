package ceviord

import (
	"github.com/gotti/cevigo/pkg/cevioai"
	"log"
)

type CevioWav interface {
	OutputWaveToFile(talkWord, path string) (err error)
	ApplyEmotions(param *Parameter) (err error)
}

type cevioWav struct {
	talker    cevioai.ITalker2V40
	isSucceed chan error
}

func NewTalker() *cevioWav {
	c := cevioWav{isSucceed: make(chan error, 0)}
	c.talker = cevioai.NewITalker2V40(cevioai.CevioAiApiName)
	_, err := c.talker.SetCast("さとうささら")
	if err != nil {
		log.Fatalln(err)
	}
	return &c
}

func (c *cevioWav) OutputWaveToFile(talkWard string, path string) (err error) {
	_, err = c.talker.OutputWaveToFile(talkWard, path)
	return err
}

func (c *cevioWav) ApplyEmotions(param *Parameter) error {
	c.talker.SetVolume(param.Volume)
	c.talker.SetSpeed(param.Speed)
	c.talker.SetTone(param.Tone)
	c.talker.SetToneScale(param.Tonescale)
	c.talker.SetAlpha(param.Alpha)
	comp, err := c.talker.GetComponents()
	if err != nil {
		return err
	}
	for n, v := range param.Emotions {
		com, err := comp.ByName(n)
		if err != nil {
			return err
		}
		com.SetValue(v)
	}
	return nil
}

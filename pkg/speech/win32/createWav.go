package win32

import (
	"github.com/azuki-bar/ceviord/pkg/ceviord"
	"github.com/gotti/cevigo/pkg/cevioai"
)

type cevioWav struct {
	talker cevioai.ITalker2V40
}

func NewTalker(para *ceviord.Parameter) *cevioWav {
	c := cevioWav{talker: cevioai.NewITalker2V40(cevioai.CevioAiApiName)}
	c.ApplyEmotions(para)
	return &c
}

func (c *cevioWav) OutputWaveToFile(talkWard string, path string) (err error) {
	_, err = c.talker.OutputWaveToFile(talkWard, path)
	return err
}

func (c *cevioWav) ApplyEmotions(param *ceviord.Parameter) error {
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

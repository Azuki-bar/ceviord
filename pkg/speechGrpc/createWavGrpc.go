package speechGrpc

import (
	"ceviord/pkg/ceviord"
	pb "ceviord/spec"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"os"
	"time"
)

type cevioWavGrpc struct {
	ttsClient pb.TtsClient
	grpcConn  *grpc.ClientConn
	param     *ceviord.Parameter
}

// NewTalker returns wav create connection and connection close function.
func NewTalker(connTar string, param *ceviord.Parameter) (*cevioWavGrpc, func() error) {
	conn, err := grpc.Dial(connTar, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	client := pb.NewTtsClient(conn)
	c := &cevioWavGrpc{
		ttsClient: client,
		param:     param,
	}
	return c, c.grpcConn.Close
}
func (c *cevioWavGrpc) OutputWaveToFile(talkWord, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := c.ttsClient.CreateWav(ctx, &pb.CevioTtsRequest{
		Text:     talkWord,
		Cast:     c.param.Cast,
		Volume:   uint32(c.param.Volume),
		Speed:    uint32(c.param.Speed),
		Pitch:    uint32(c.param.Tone),
		Alpha:    uint32(c.param.Alpha),
		Intro:    uint32(c.param.Tonescale),
		Emotions: typeCast(c.param.Emotions),
	})
	if err != nil {
		return fmt.Errorf("grpc execute failed `%w`", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("temp file create failed `%w`", err)
	}
	_, err = f.Write(res.Audio)
	return err
}
func (c *cevioWavGrpc) ApplyEmotions(param *ceviord.Parameter) error {
	c.param = param
	return nil
}

func typeCast(m map[string]int) map[string]uint32 {
	n := make(map[string]uint32)
	for k, v := range m {
		n[k] = uint32(v)
	}
	return n
}

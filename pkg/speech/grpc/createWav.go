package grpc

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/azuki-bar/ceviord/pkg/ceviord"
	pb "github.com/azuki-bar/ceviord/spec"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type cevioWavGrpc struct {
	ttsClient pb.TtsClient
	grpcConn  *grpc.ClientConn
	param     ceviord.Parameter
	token     string
	logger    *zap.Logger
}

// NewTalker returns wav create connection and connection close function.
func NewTalker(logger *zap.Logger, connConf *ceviord.Conn, param *ceviord.Parameter) (*cevioWavGrpc, func() error) {
	conn, err := grpc.Dial(connConf.Cevio.EndPoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalln(err)
	}
	client := pb.NewTtsClient(conn)
	c := &cevioWavGrpc{
		ttsClient: client,
		param:     *param,
		token:     connConf.Cevio.Token,
		logger:    logger,
	}
	return c, c.grpcConn.Close
}
func (c *cevioWavGrpc) OutputWaveToFile(talkWord, path string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req := &pb.CevioTtsRequest{
		Text:     talkWord,
		Cast:     c.param.Cast,
		Volume:   uint32(c.param.Volume),
		Speed:    uint32(c.param.Speed),
		Pitch:    uint32(c.param.Tone),
		Alpha:    uint32(c.param.Alpha),
		Intro:    uint32(c.param.Tonescale),
		Emotions: typeCast(c.param.Emotions),
		Token:    c.token,
	}
	res, err := c.ttsClient.CreateWav(ctx, req)
	if err != nil {
		c.logger.Warn("grpc execute failed", zap.Error(err))
		return fmt.Errorf("grpc execute failed `%w`", err)
	}
	f, err := os.Create(path)
	if err != nil {
		c.logger.Warn("temp file create failed", zap.Error(err))
		return fmt.Errorf("temp file create failed `%w`", err)
	}
	_, err = f.Write(res.Audio)
	return err
}
func (c *cevioWavGrpc) ApplyEmotions(param *ceviord.Parameter) error {
	c.param = *param
	return nil
}

func typeCast(m map[string]int) map[string]uint32 {
	n := make(map[string]uint32)
	for k, v := range m {
		n[k] = uint32(v)
	}
	return n
}

package grpc

import (
	"context"
	"io"
	"math/rand"
	"postParser/internal/adapters/driven/grpc/pb"
	"postParser/internal/logger"
	"postParser/internal/repo"
	"time"

	errs "github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	Stream pb.PostParser_MultiPartClient
	req    *pb.FileUploadReq
	err    error
)

type Transmitter interface {
	Transmit(repo.AppDistributorUnit)
}

type TransmitAdapter struct{}

func NewTransmitter() Transmitter {
	return &TransmitAdapter{}
}
func (t *TransmitAdapter) Transmit(adu repo.AppDistributorUnit) {
	//connectString := os.Getenv("HOST") + ":" + os.Getenv("PORT")
	connectString := "localhost:3100"
	conn, err := grpc.Dial(connectString, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.L.Error(errs.Wrap(err, "rpc.Transmit.grpc.dial"))
	}

	client := pb.NewPostParserClient(conn)

	err = doTransmit(client, conn, adu)
	if err != nil {
		logger.L.Error(errs.Wrap(err, "rpc.Transmit.grpc.doTransmit"))
	}
}

func doTransmit(c pb.PostParserClient, conn *grpc.ClientConn, adu repo.AppDistributorUnit) error {

	if Stream == nil {
		logger.L.Info("no stream yet, creating one ")
		req = &pb.FileUploadReq{
			Info: &pb.FileUploadReq_FileInfo{
				FileInfo: &pb.FileInfo{
					FileName: adu.H.FormName,
					FileType: "zxc",
					StreamID: uint32(rand.Intn(50)),
				},
			},
		}

		Stream, err = c.MultiPart(context.Background())
		if err != nil {
			return errs.Wrap(err, "Creating stream")
		}
		logger.L.Infof("in rpc.doTransmit stream created %v:%#v", Stream, Stream)

		err = Stream.Send(req)
		if err != nil {
			return err
		}
	}

	if adu.B.B == nil {
		logger.L.Info("in rpc.doTransmit closing and getting results from stream ", Stream)
		res, err := Stream.CloseAndRecv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}

		logger.L.Infof("in rpc.doTransmit uploaded file %q with size of %d bytes\n", res.FileName, res.FileSize)

		logger.L.Info("in rpc.doTransmit closing conn")
		conn.Close()
		Stream = nil
		return nil
	}
	time.Sleep(time.Millisecond * 100)
	req = &pb.FileUploadReq{
		Info: &pb.FileUploadReq_FileData{
			FileData: adu.B.B,
		},
	}
	logger.L.Infof("in rpc.doTransmit transmitting to existed stream %#v unit %v\n ", Stream, adu.H)
	err = Stream.Send(req)
	if err != nil {
		return errs.Wrap(err, "Stream.Send")
	}
	return nil
}

package app_test

import (
	"context"
	"net"
	"plumbing"
	v2 "plumbing/v2"
	app "rlp/app"
	"time"

	"google.golang.org/grpc"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Start", func() {
	It("receive messages via egress client", func() {
		// create fake doppler
		router := newMockDopplerServer()

		lis, err := net.Listen("tcp", "localhost:0")
		defer lis.Close()
		Expect(err).ToNot(HaveOccurred())

		grpcServer := grpc.NewServer()
		plumbing.RegisterDopplerServer(grpcServer, router)
		go grpcServer.Serve(lis)

		// point our app at the fake doppler
		rlp := app.Start(
			app.WithIngressAddrs([]string{lis.Addr().String()}),
		)
		go rlp.Start()

		// create public client consuming log API
		conn, err := grpc.Dial(rlp.EgressAddr.String(), grpc.WithInsecure())
		Expect(err).ToNot(HaveOccurred())

		defer conn.Close()
		egressRequest := &v2.EgressRequest{}
		egressClient := v2.NewEgressClient(conn)
		egressStream, err := egressClient.Receiver(context.Background(), egressRequest)

		var subscriber plumbing.Doppler_SubscribeServer
		Eventually(router.SubscribeInput.Stream, 5).Should(Receive(&subscriber))

		go func() {
			payload := buildLogMessage()
			response := &plumbing.Response{
				Payload: payload,
			}

			for {
				err := subscriber.Send(response)
				if err != nil {
					return
				}
			}
		}()

		envelope, err := egressStream.Recv()
		Expect(err).ToNot(HaveOccurred())
		Expect(envelope.Timestamp).To(Equal(int64(99)))
	})
})

func buildLogMessage() []byte {
	e := &events.Envelope{
		Origin:    proto.String("foo"),
		EventType: events.Envelope_LogMessage.Enum(),
		LogMessage: &events.LogMessage{
			Message:     []byte("foo"),
			MessageType: events.LogMessage_OUT.Enum(),
			Timestamp:   proto.Int64(time.Now().UnixNano()),
			AppId:       proto.String("test-app"),
		},
	}
	b, _ := proto.Marshal(e)
	return b
}

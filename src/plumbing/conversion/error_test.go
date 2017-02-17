package conversion_test

import (
	"plumbing/conversion"
	v2 "plumbing/v2"

	"github.com/cloudfoundry/sonde-go/events"
	"github.com/gogo/protobuf/proto"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
)

var _ = Describe("HTTP", func() {
	Context("given a v1 envelope", func() {
		It("converts to a v2 envelope", func() {
			v1Envelope := &events.Envelope{
				EventType:  events.Envelope_Error.Enum(),
				Origin:     proto.String("fake-origin"),
				Deployment: proto.String("some-deployment"),
				Job:        proto.String("some-job"),
				Index:      proto.String("some-index"),
				Ip:         proto.String("some-ip"),
				Error: &events.Error{
					Source:  proto.String("test-source"),
					Code:    proto.Int32(12345),
					Message: proto.String("test-message"),
				},
			}

			expectedV2Envelope := &v2.Envelope{
				Tags: map[string]*v2.Value{
					"__v1_type":  {&v2.Value_Text{"Error"}},
					"source":     {&v2.Value_Text{"test-source"}},
					"code":       {&v2.Value_Integer{12345}},
					"origin":     {&v2.Value_Text{"fake-origin"}},
					"deployment": {&v2.Value_Text{"some-deployment"}},
					"job":        {&v2.Value_Text{"some-job"}},
					"index":      {&v2.Value_Text{"some-index"}},
					"ip":         {&v2.Value_Text{"some-ip"}},
				},
				Message: &v2.Envelope_Log{
					Log: &v2.Log{
						Payload: []byte("test-message"),
						Type:    v2.Log_OUT,
					},
				},
			}

			converted := conversion.ToV2(v1Envelope)

			_, err := proto.Marshal(converted)
			Expect(err).ToNot(HaveOccurred())

			Expect(*converted).To(MatchFields(IgnoreExtras, Fields{
				"Tags":    Equal(expectedV2Envelope.Tags),
				"Message": Equal(expectedV2Envelope.Message),
			}))
		})
	})
})

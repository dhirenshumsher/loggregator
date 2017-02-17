package ingress

import v2 "plumbing/v2"

type NullConverter struct{}

func (NullConverter) Convert(data []byte) (envelope *v2.Envelope, err error) {
	return nil, nil
}

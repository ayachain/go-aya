package ipfsapi

import (
	"context"
	"errors"
	"testing"

	coreiface "github.com/ipfs/interface-go-ipfs-core"
)

var apiNotImplemented = errors.New("api not implemented")

func (tp *IpfsAPI) makeAPI(ctx context.Context) (coreiface.CoreAPI, error) {
	api, err := tp.MakeAPISwarm(ctx, false, 1)
	if err != nil {
		return nil, err
	}

	return api[0], nil
}

type Provider interface {
	// Make creates n nodes. fullIdentity set to false can be ignored
	MakeAPISwarm(ctx context.Context, fullIdentity bool, n int) ([]coreiface.CoreAPI, error)
}

func (tp *IpfsAPI) MakeAPISwarm(ctx context.Context, fullIdentity bool, n int) ([]coreiface.CoreAPI, error) {

	if tp.apis != nil {
		tp.apis <- 1
		go func() {
			<-ctx.Done()
			tp.apis <- -1
		}()
	}

	return tp.Provider.MakeAPISwarm(ctx, fullIdentity, n)
}

type IpfsAPI struct {

	Provider
	apis chan int

}

func MakeIPFSAPI( p Provider, ctx context.Context ) (coreiface.CoreAPI, error) {

	running := 1
	apis := make(chan int)
	zeroRunning := make(chan struct{})

	go func() {

		for i := range apis {
			running += i
			if running < 1 {
				close(zeroRunning)
				return
			}
		}

	}()

	tp := &IpfsAPI{Provider: p, apis: apis}

	return tp.makeAPI(ctx)
}

func (tp *IpfsAPI) hasApi(t *testing.T, tf func(coreiface.CoreAPI) error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	api, err := tp.makeAPI(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if err := tf(api); err != nil {
		t.Fatal(api)
	}
}

module github.com/ayachain/go-aya

require (
	github.com/Kubuxu/go-os-helper v0.0.1
	github.com/ayachain/go-aya-alvm v0.0.0-00010101000000-000000000000
	github.com/ayachain/go-aya-alvm-json v0.0.0-00010101000000-000000000000
	github.com/dgrijalva/jwt-go v3.2.0+incompatible // indirect
	github.com/ethereum/go-ethereum v1.8.27
	github.com/hashicorp/go-multierror v1.0.0
	github.com/ipfs/go-ipfs v0.0.0-00010101000000-000000000000
	github.com/ipfs/go-ipfs-api v0.0.1
	github.com/ipfs/go-ipfs-cmds v0.0.8
	github.com/ipfs/go-ipfs-config v0.0.4
	github.com/ipfs/go-ipfs-files v0.0.3
	github.com/ipfs/go-ipfs-util v0.0.1
	github.com/ipfs/go-log v0.0.1
	github.com/ipfs/go-merkledag v0.0.6
	github.com/ipfs/go-metrics-prometheus v0.0.2
	github.com/ipfs/go-mfs v0.0.11
	github.com/ipfs/go-unixfs v0.0.8
	github.com/ipfs/interface-go-ipfs-core v0.0.8
	github.com/jbenet/goprocess v0.1.3
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.2.8
	github.com/libp2p/go-libp2p-loggables v0.1.0
	github.com/libp2p/go-libp2p-peer v0.2.0
	github.com/multiformats/go-multiaddr v0.0.4
	github.com/multiformats/go-multiaddr-dns v0.0.2
	github.com/multiformats/go-multiaddr-net v0.0.1
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3
	github.com/valyala/fasttemplate v1.0.1 // indirect
	github.com/whyrusleeping/yamux v1.2.0 // indirect
)

replace (
	github.com/ayachain/go-aya-alvm => ../go-aya-alvm
	github.com/ayachain/go-aya-alvm-json => ../go-aya-alvm-json
	github.com/ipfs/go-ipfs => ../../ipfs/go-ipfs
)

go 1.12

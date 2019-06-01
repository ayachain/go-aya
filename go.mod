module github.com/ayachain/go-aya

go 1.12

require (
	github.com/Kubuxu/go-os-helper v0.0.1
	github.com/ayachain/go-aya-alvm v0.0.0-00010101000000-000000000000
	github.com/ayachain/go-aya-alvm-adb v0.0.0
	github.com/hashicorp/go-multierror v1.0.0
	github.com/ipfs/go-cid v0.0.2
	github.com/ipfs/go-ipfs v0.4.21
	github.com/ipfs/go-ipfs-cmds v0.0.8
	github.com/ipfs/go-ipfs-config v0.0.3
	github.com/ipfs/go-ipfs-files v0.0.3
	github.com/ipfs/go-ipfs-util v0.0.1
	github.com/ipfs/go-log v0.0.1
	github.com/ipfs/go-merkledag v0.0.3
	github.com/ipfs/go-metrics-prometheus v0.0.2
	github.com/ipfs/interface-go-ipfs-core v0.0.8
	github.com/jbenet/goprocess v0.1.3
	github.com/libp2p/go-libp2p-loggables v0.0.1
	github.com/multiformats/go-multiaddr v0.0.4
	github.com/multiformats/go-multiaddr-dns v0.0.2
	github.com/multiformats/go-multiaddr-net v0.0.1
	github.com/prometheus/client_golang v0.9.3
)

replace (
	github.com/ayachain/go-aya-alvm => ../go-aya-alvm
	github.com/ayachain/go-aya-alvm-adb => ../go-aya-alvm-adb
)

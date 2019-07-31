module github.com/ayachain/go-aya

go 1.12

require (
	github.com/Kubuxu/go-os-helper v0.0.1
	github.com/allegro/bigcache v1.2.0 // indirect
	github.com/aristanetworks/goarista v0.0.0-20190531155855-fef20d617fa7 // indirect
	github.com/ayachain/go-aya-alvm v0.0.0-00010101000000-000000000000
	github.com/ayachain/go-aya-alvm-adb v0.0.0-00010101000000-000000000000
	github.com/deckarep/golang-set v1.7.1 // indirect
	github.com/ethereum/go-ethereum v1.8.27
	github.com/hashicorp/go-multierror v1.0.0
	github.com/ipfs/go-block-format v0.0.2
	github.com/ipfs/go-cid v0.0.2
	github.com/ipfs/go-datastore v0.0.5
	github.com/ipfs/go-ipfs v0.4.21
	github.com/ipfs/go-ipfs-cmds v0.0.8
	github.com/ipfs/go-ipfs-config v0.0.3
	github.com/ipfs/go-ipfs-files v0.0.3
	github.com/ipfs/go-ipfs-util v0.0.1
	github.com/ipfs/go-log v0.0.1
	github.com/ipfs/go-merkledag v0.0.3
	github.com/ipfs/go-metrics-prometheus v0.0.2
	github.com/ipfs/go-mfs v0.0.7
	github.com/ipfs/go-unixfs v0.0.6
	github.com/ipfs/interface-go-ipfs-core v0.0.8
	github.com/jbenet/goprocess v0.1.3
	github.com/libp2p/go-libp2p v0.0.28
	github.com/libp2p/go-libp2p-loggables v0.0.1
	github.com/libp2p/go-libp2p-peer v0.1.1
	github.com/libp2p/go-libp2p-peerstore v0.0.6
	github.com/libp2p/go-libp2p-pubsub v0.0.3
	github.com/multiformats/go-multiaddr v0.0.4
	github.com/multiformats/go-multiaddr-dns v0.0.2
	github.com/multiformats/go-multiaddr-net v0.0.1
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v0.9.3
	github.com/prometheus/common v0.4.0
	github.com/rjeczalik/notify v0.9.2 // indirect
	github.com/syndtr/goleveldb v1.0.0
	github.com/whyrusleeping/go-logging v0.0.0-20170515211332-0457bb6b88fc
	go4.org v0.0.0-20190313082347-94abd6928b1d
)

replace (
	github.com/ayachain/go-aya-alvm => ../go-aya-alvm
	github.com/ayachain/go-aya-alvm-adb => ../go-aya-alvm-adb
)

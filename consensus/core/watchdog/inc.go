//
// When the received data arrives, there are several situations.
//
// 1. This node often sends incorrect data
// 2. This node is no longer trusted by me.
// 3. I trust this node very much.
// 4. I unconditionally trust the data given to me by this node.
// 5. more...
//
//
// Therefore, I call the type of watchdog, meaning to prevent malicious nodes,
// let the watchdog scare it away. In order to reduce the intensity of the work
// of notaries, the dog will help you filter some unnecessary or untrustworthy
// messages, or shield the data sent by some nodes. In order to effectively
// utilize resources, we will not remove malicious nodes from the whole network.
// When sending wrong data, it is not the intention of the node itself, such as
// uncertain network delay or transmission error.
//
//
// Because according to the original mechanism of "Aya", every node needs to verify
// its correctness before transmitting data. If the wrong data is sent, although
// consensus will not be reached, it will still cause network congestion because
// of a large number of wrong requests.
//
//
// The black-and-white list will remain on the local nodes, but will not be
// synchronized to the whole network. If possible, we may submit these data to
// the chain for storage as a node credit system.
//
//
// The class that implements the watchdog interface should have its own evaluation
// mechanism for outsiders, such as correct message scoring and wrong message
// scoring according to the severity of the consequences.
//
//
// In Aya, a node has a unique label for transmission, and things need to be signed.
// An address can correspond to multiple nodes, while a node intelligently corresponds
// to an address. Therefore, it is necessary to consider that when a node belonging
// to an address sends an error or malicious message, it should also correspond to
// all controlled nodes under the corresponding address. Processing, relatively if
// the result is correct, should increase the trust of all the nodes to which they
// belong
//
// We believe that the watchdog's black-and-white list data is not allowed to be added
// directly, because malicious attackers may hold a large number of nodes to manually
// add trust between Byzantine nodes, which is unanimous in consensus, so let the
// watchdog learn automatically during operation to protect notaries.
//
// If you use temporary distrust as the result of processing, you should make small trust
// adjustments to temporary distrust nodes in the watchdog at intervals to give it a
// chance to run the nodes correctly.
//
// by oblivioned
//
package watchdog

import pubsub "github.com/libp2p/go-libp2p-pubsub"

type FinalResult int

const (
	FinalResult_Trust_Evey	FinalResult = 128
	FinalResult_Success		FinalResult = 1
	FinalResult_ELV1		FinalResult = -1
	FinalResult_ELV2		FinalResult = -2
	FinalResult_ELV3		FinalResult = -3
	FinalResult_ELV4		FinalResult = -4
	FinalResult_Distrust	FinalResult = -50
	FinalResult_Never		FinalResult = -128
)

type FinalResultDefer func(FinalResult)

type MsgFromDogs struct {
	pubsub.Message
	ResultDefer FinalResultDefer
}

type WatchDog interface {

	// Whether the door can be opened for the designated node, if the familiar
	// person (white list) opens the door directly, the stranger (black list) closes
	// the door.
	TaskMessage( msg pubsub.Message ) *MsgFromDogs

	// We provide a reference logic. When the score is higher than 0, we can pass it.
	// When the score is lower than 0, we should give up processing the message and
	// store the message in a watchdog's cache until a trusted node broadcasts the
	// same data as it, and then add points.
	CreditScoring( peeridOrAddress string ) int8

	// When the node closes, the cache data is written to prevent loss. If the write
	// error occurs, some data will be discarded appropriately, which will not affect
	// the next node operation.
	Close()

}
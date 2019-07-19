package logs

import "github.com/whyrusleeping/go-logging"

const (
	AModules_TxPool					= "ATxPool"
	AModules_CVFS					= "ACvfs"
	AModules_CVFS_Indexes			= "IndexesServices"

	AVDBServices_Electorals			= "ElectoralServices"
)


func ConfigLogs() {

	logging.SetLevel(logging.DEBUG, AModules_TxPool)
	logging.SetLevel(logging.DEBUG, AModules_CVFS)
	logging.SetLevel(logging.DEBUG, AModules_CVFS_Indexes)


}
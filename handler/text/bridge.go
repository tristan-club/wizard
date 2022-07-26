package text

const (
	InSufficientBalance = "😭 * Insufficient Balance * \nPlease Deposit Some %s To Your %s Chain Address: \n\n _↓↓↓Click To Copy↓↓↓_\n\n `%s` \n\n_↑↑↑Click To Copy↑↑↑_"
	EnterBridgeAmount   = "💰 *My METIS On BNB*\n\n*METIS*: %s\n\n_Please enter the amount  you would like to transfer to Andromeda_\n\n ✳️  Should between %s \\- %s  "
	BridgeConfirmOrder  = "📝 *Confirm Order*\n*From*: BNB\n*To*: Andromeda\n*Amount*: %s METIS\n*Estimated Gas*: %s BNB\n*Fee*: %s METIS\n\n✳️ Enter your pin code to continue"
)

const (
	BridgeSubmitted         = "📝 *Processing*\n*Stage 1/4*: Transferring To Bridge"
	BridgeFromPending       = "📝 *Processing*\n*Stage 2/4*: Waiting Confirm\n*TXN URL*: [click to view](%s)"
	BridgeFromSuccess       = "📝 *Processing*\n*Stage 3/4*: Transferring to Andromeda"
	BridgeToPending         = "📝 *Processing*\n*Stage 4/4*: Waiting Confirm\n*TXN URL*: [click to view](%s)"
	BridgeTransactionFailed = "⛔️ *Transaction Failed*\n*Error Message*: [click to view](%s) click to view"
	BridgeToSuccess         = "✨ *Success*\n*You Get*: %s METIS\n*Your METIS Total*: %s\n*TXN URL*: [click to view](%s)"
	BridgeOtherFailed       = "⛔️ *Transaction Failed*\n*Error Message*: %s"
)

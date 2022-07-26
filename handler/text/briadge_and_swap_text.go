package text

const (
	SwapOrder = "📝 *Confirm Order*\n" +
		"*From*:  %s %s\n" +
		"*To*: METIS\n" +
		"*Vender*: PancakeSwap\n" +
		"*Estimated Gas*: %s\n" +
		"*Slippage Point*: 3%%\n" +
		"*You Will Receive*: %s \\~ %s\n\n" +
		"✳️ Enter your pin code to continue"

	SwapEnterAmount = "💰 *My Balance*\n\n" +
		"*%s*: %s\n\n" +
		"✳️ Please input the amount  \n" +
		"    \\(%s \\- %s\\)  below👇  "

	SwapApproveStart      = "📝 *Processing*\n*Stage 1/4*: Approve"
	SwapApproveProcessing = "📝 *Processing*\n*Stage 2/4* : Waiting Confirm\n*TXN URL*:  [click to view](%s) "
	SwapApproveToSwap     = "📝 *Processing*\n*Stage 3/4*: Approved \\- Swapping"
	SwapProcessing1       = "📝 *Processing*\n*Stage 4/4*: Waiting Confirm\n*TXN URL*: [click to view](%s)"
	SwapProcessing2       = "📝 *Processing*\n*TXN URL*: [click to view](%s)"
	SwapSuccess           = "✨ *Success*\n*You Get*: %s METIS\n*Your METIS Total*: %s\n*TXN URL*: [click to view](%s)"
	SwapFailed            = "❌ *Filed*\n*Transaction Failed reason*%s\n*TXN URL*: [click to view](%s)"
	ContinueBridge        = "Bridge METIS to Andromeda"
	CheckTimeOut1         = "📝 *Still Processing*\n*Stage 4/4*: Waiting Confirm\n*Message*: It has been 30 minutes, since the txn was sent to the BNB chain. It is uncommon and we will check this txn. Also you could open this link below for more details.\n*TXN URL*: [click to view](%s)"
	CheckTimeout2         = "📝 *Still Processing*\n*Message*: It has been 30 minutes, since the txn was sent to the BNB chain. It is uncommon and we will check this txn. Also you could open this link below for more details.\n*TXN URL*: [click to view](%s)"

	BridgeOrder = "✳️ Create order success,here is the order information:\n\n" +
		"```\nBridging:\n" +
		"Transfer Assert: [BNB]METIS => [Andromeda]METIS\n" +
		"FEE: [Andromeda]METIS = %s\n" +
		"Estimated GAS: [BNB]BNB=%s\n" +
		"Estimated Result: [Andromeda]METIS = %s ~ %s\n" +
		"```\n" +
		"Please make sure you have enough %s and BNB to finish these steps." +
		"Enter your pin code to continue\n"
)

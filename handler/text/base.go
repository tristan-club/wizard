package text

var CustomStartMenu = ""
var StartInviteText = ""

const (
	SelectChain              = "✳️ Select chain "
	SelectAsset              = "✳️ Select asset to send"
	SelectEnvelopeRewardType = "✳️ *Select red envelope award type*"
	SelectEnvelopeType       = "✳️ *Select red envelope type*"
	SelectAssetType          = "✳️ *Select the type of asset to add*"
	EnterAmount              = "✳️ Enter an amount."
	EnterAmountWithRange     = "✳️ Enter an amount between %s and %s."
	EnterQuantity            = "✳️ Enter quantity"
	EnterQuantityWithRange   = "✳️ Enter an quantity between %d and %d."
	EnterReceiverAddress     = "✳️ Enter receiver wallet address"
	EnterTokenAddress        = "✳️ Enter token address you want to add"
	EnterMetamaskAddress     = "✳️ Enter your metamask address"
	EnterPinCode             = "✳️ Enter your pin code"
	EnterOldPinCode          = "✳️ Enter your old pin code"
	EnterNewPinCode          = "✳️ Enter your new pin code, at least 6 characters."
	EnterEnvelopeQuantity    = "✳️ Enter red envelope quantity, min %d, max %d."
	EnterTokenName           = "✳️ Enter token name.\n 1-28 English letters (eg Tether USD), numbers, characters, spaces and hyphens are accepted."
	EnterTokenSymbol         = "✳️ Enter token symbol.\n1-10 characters (eg USDT). Spaces cannot be included, but English letters, numeric characters, etc. can be included."
	EnterInitialSupply       = "✳️ Enter token initial supply.\n The initial amount is the number of tokens to be generated, the minimum amount is %d, and the maximum is %d"
	EnterBool                = "✳️ Select"
	EnterMintable            = "✳️ Choose whether your tokens can be mint later "
	EnterEnvelopeOption      = "✳️ Choose whether your tokens can be mint later "
	EnterBridgeAsset         = "✳️ *Choose an asset on BNB to swap for METIS*"
	EnterPrivateWithDelete   = "✳️ Enter your private\n\n*Note*:\nTo keep your private key safe\nWe recommend that you delete the private key message after the operation is complete"
	EnterPinCodeToDelete     = "✳️ Enter the pincode to clear the account\nNote that this operation will erase all your data and cannot be recovered"

	ChosenChain  = "You have chosen the %s."
	ChosenAsset  = "Your have chosen asset %s."
	ChosenCommon = "You have chosen %s"
)

const (
	Introduce = "\U0001F973🙌 Welcome! Tristan MetaWallet is the first-ever web 3.0 crypto wallet for Telegram users and communities\nWith MetaWallet you can:\n🔗 Create your 1st crypto wallet with your telegram account    \n💸 Receive or transfer of your crypto asset and NFTs\n🚀 Issue tokens\n💵 Send or receive Red Envelopes with your social contacts\n🎁 Airdrop tokens to your community \n🎮 Launch 3rd party dApps and Games by one click, welcome to join @tristanmetawallet for more discussion."

	CreateAccountSuccess   = "Congratulations\\! Your Meta wallet has been created \\. \n\nThe wallet address is\\: `%s`\\.\nPin Code is `%s`\\.\nYour Wallet Pin Code is the only way to access your crypto asset in MetaWallet and CAN NOT be recovered if lost\\."
	GetAccountSuccess      = "Your MetaWallet address is\\: `%s` \\."
	RecommendChangePinCode = "Would you like to change the Pincode\\?\nPlease use `/change_pin_code` command\\."

	CheckDm             = "We have forwarded you the details. Please lookout for DM from \"MetaWalletBot\" "
	UserNoInit          = "%s \\, you have not created an account yet\\, please forward to private chat with bot to initialize your account\\."
	UserNoInitInPrivate = "%s \\, you have not created an account yet, use \"/start\" to initialize your account\\."
	BalanceSuccess      = "Your balance is"

	MessageDisappearSoon = "*NOTE*\\: This message will be clear shortly\nPlease save your pincode in time or use the `/change_pin_code` command to change your pincode"

	TransactionProcessing           = "*Your transaction is processing*\n*TXN URL*: [click to view](%s)"
	TransactionProcessingNoMarkDown = "Your transaction is processing, you can view it on %s"
	EnvelopePreparing               = "The red envelope account is created and the recharge operation is in progress\n*TXN URL: [click to view](%s)*"
	TransferSuccess                 = "*Your transfer succeed*\n*To*: %s\n*Asset*: %s   *Value*: %s\n*TXN URL*: [click to view](%s)"
	TransactionSuccess              = "*Your transaction succeed, [click to view](%s)*"
	TransferSuccessNoMarkdown       = "Your transfer to %s asset %s value %s succeed, you can view it on %s"

	WaitForResult                     = "Your transaction is under processing..."
	CreateEnvelopeSuccess             = "*Your red envelope No\\#%s is created, [click to view](%s)*"
	ShareEnvelopeSuccess              = "User %s create a red envelope NO\\#%s by %s total value %s\\!"
	DCShareEnvelopeSuccess            = "User %s create a red envelope NO\\#%s by %s total value %s\\!\nUse the `/open_red_envelope` command to open the red envelope"
	IssueTokenSuccess                 = "*Your token issued successfully*\nThe contract address is `%s`\n[click to view](%s)"
	AirdropSuccess                    = "*The airdrop operation succeeded*\n*TXN URL*: [click to view](%s)"
	AirdropSuccessInGroup             = "User %s successfully initiated the %s %s token airdrop operation\\nEveryone in the following list got %s %s :\n%s\n*TXN URL*: [click to view](%s)"
	EnvelopeDetail                    = "\U0001F9E7Red Envelope Created by %s\n\n💰Remaining Amount\\: %s %s\n\n🎁Claimed: %d\\/%d"
	OpenEnvelope                      = "OPEN"
	Bridge                            = "BRIDGE"
	ChangePinCode                     = "CHANGE PIN CODE"
	SubmitMetamask                    = "SUBMIT METAMASK"
	ContinueToBridge                  = "continue to bridge"
	OpenEnvelopeTransactionProcessing = "User %s open red envelope NO\\#%s is processing\\, get %s %s"
	OpenEnvelopeSuccess               = "User %s open red envelope NO\\#%s succeed \\! get %s %s \\!\n*TXN URL: [click to view](%s)*"
	EnvelopeCreateFailed              = "Red envelope creation failed."
	ChangePinCodeSuccess              = "Pin code has been updated."

	BusinessError = "%s used %s failed (%s)"
	ServerError   = "Ops, something went wrong"

	SwitchPrivate      = "✳️ %s\\, Please forward to private chat with bot for detail"
	NoAssetToOperation = "Ops, you don't have a token to proceed to the next step."
	ClickStart         = "✳️ Please use /start command to initialize your wallet."

	NeedGroup = "This command only works in group chat."

	ProcessTimeOut = "Process timeout."

	ButtonForwardPrivateChat = "FORWARD TO CONTINUE"
	ButtonForwardCreate      = "FORWARD TO CREATE"
	ButtonJoin               = "JOIN"
	ButtonHelp               = "HELP"

	DepositAsset = "Please Deposit Some %s To Your %s Chain Address"

	OperationSuccess    = "Operation Success!"
	OperationProcessing = "Operation Processing..."

	GetPrivateSuccess              = "Your private key is:\n`%s`\nPlease keep it safe\nThis message will be deleted shortly"
	GetPrivateSuccessNeedDeleteMsg = "Your private key is:\n`%s`\nPlease keep it safe\nIn order to protect your private key security\nPlease delete this message in time"

	NoMetamaskAddress  = "Please enter your MetaMask address"
	HasMetamaskAddress = "Your account has been associate with MetaMask address `%s`\\.\n" +
		"Entering a new address will replace the original one,\n" +
		"which may cause the reward for the original address to be invalid\\."
	BindMetamaskAddressSuccess = "Operation Success\\!\nThe MetaMask address currently associate with your account is `%s`"

	OpenWebApp = "Please open web app"

	ChangeYourPin = "Please change your Pincode for your asset security!"

	InitKeyboardButtonSuccess = "Init keyboard button success!"
)

const (
	ForbiddenError = "Operation failed\\. %s\\, bot can't send messages to you\\. Please forward to the bot and click `Start` first\\. "
)

const (
	StartBotDefaultText = "Hi %s\\, %s invites you to join the group\nAn incentive event is being held there\\. You are welcome to participate together\\."
	StartBotSimpleText  = "An incentive event is being held there\\. You are welcome to participate together\\."
)

const (
	KBAccount       = "💎 My Account"
	KBProfile       = "🏅 Social Profile"
	KBActivity      = "🔥 Trending Activity"
	KBLastActivity  = "🕘 Last Activity"
	KBBalance       = "💰 Balance"
	KBHelp          = "❓️ Guide & Help"
	KBChangePinCode = "ℹ️ Change Pincode"
	KBTransfer      = "♻ Transfer"
)

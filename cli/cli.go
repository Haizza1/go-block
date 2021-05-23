package cli

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/Haizza1/go-block/blockchain"
	"github.com/Haizza1/go-block/wallet"
)

type CommandLine struct{}

// print usage will print the cli usage
func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("	getbalance -address <ADDRESS> - get the balance for the given address")
	fmt.Println(" 	createBlockchain -address <ADDRESS> create a blockchain with the given address")
	fmt.Println(" 	printchain - Prints the blocks in the Blockchain")
	fmt.Println(" 	send -from <FROM> -to <TO> -amount <AMOUNT> - Send Send amount of coins")
	fmt.Println("	createWallet - creates a new Wallet")
	fmt.Println("	listaddresses - list the address in our wallet file")
	fmt.Println("	reindex - Rebuilds The unspent transactions outputs set")
}

// validateArgs will check if the given args
func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit()
	}
}

func (cli *CommandLine) validateAddress(address string) {
	if !wallet.ValidateAddress(address) {
		fmt.Printf("Adress %s is invalid\n", address)
		runtime.Goexit()
	}
}

func (cli *CommandLine) reindexUTXO() {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	UTXSet := blockchain.UTXOSet{BlockChain: chain}
	UTXSet.Reindex()

	count := UTXSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

// listAddresses will print all the addresses in the wallet.data file
func (cli *CommandLine) listAddresses() {
	wallets, _ := wallet.CreateWallets()
	addresses := wallets.GetAllAddress()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

// createWallet will create new wallet and saved to the wallet file
func (cli *CommandLine) createWallet() {
	wallets, _ := wallet.CreateWallets()
	address := wallets.AddWallet()
	wallets.SaveFile()
	fmt.Printf("New address is: %s\n", address)
}

// printChain will print all the blocks in the blockchain
func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockChain("")
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Previos Hash: %x\n", block.PrevHash)
		fmt.Printf("Block Hash: %x\n", block.Hash)

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}

		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}
}

// create blockchain will create a new blockchain instance with the given address
func (cli *CommandLine) createBLockChain(address string) {
	cli.validateAddress(address)
	chain := blockchain.InitBLockChain(address)
	chain.Database.Close()
	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address string) {
	cli.validateAddress(address)
	chain := blockchain.ContinueBlockChain(address)
	UTXIOSet := blockchain.UTXOSet{BlockChain: chain}
	defer chain.Database.Close()

	balance := 0
	pubKeyHash := wallet.Base58Encode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	unspentTxs := UTXIOSet.FindUTXO(pubKeyHash)

	for _, out := range unspentTxs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %d\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount int) {
	cli.validateAddress(from)
	cli.validateAddress(to)

	chain := blockchain.ContinueBlockChain(from)
	UTXIOSet := &blockchain.UTXOSet{BlockChain: chain}
	defer chain.Database.Close()

	tx := blockchain.NewTransaction(from, to, amount, UTXIOSet)
	block := chain.AddBlock([]*blockchain.Transaction{tx})
	UTXIOSet.Update(block)
	fmt.Println("Success!")
}

// Run will start the comman line app and validate the args
func (cli *CommandLine) Run() {
	cli.validateArgs()

	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBLockchainCmd := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	listAddressesCmd := flag.NewFlagSet("listaddresses", flag.ExitOnError)
	createWalletCmd := flag.NewFlagSet("createWallet", flag.ExitOnError)
	reindexCmd := flag.NewFlagSet("reindex", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBLockChainAddress := createBLockchainCmd.String("address", "", "The address to send genesis block reward to")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		blockchain.CheckError(err)

	case "send":
		err := sendCmd.Parse(os.Args[2:])
		blockchain.CheckError(err)

	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.CheckError(err)

	case "createBlockchain":
		err := createBLockchainCmd.Parse(os.Args[2:])
		blockchain.CheckError(err)

	case "createWallet":
		err := createWalletCmd.Parse(os.Args[2:])
		blockchain.CheckError(err)

	case "listaddresses":
		err := listAddressesCmd.Parse(os.Args[2:])
		blockchain.CheckError(err)

	case "reindex":
		err := reindexCmd.Parse(os.Args[2:])
		blockchain.CheckError(err)

	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if reindexCmd.Parsed() {
		cli.reindexUTXO()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			cli.printUsage()
			runtime.Goexit()
		}

		cli.getBalance(*getBalanceAddress)
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			cli.printUsage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if createWalletCmd.Parsed() {
		cli.createWallet()
	}

	if listAddressesCmd.Parsed() {
		cli.listAddresses()
	}

	if createBLockchainCmd.Parsed() {
		if *createBLockChainAddress == "" {
			cli.printUsage()
			runtime.Goexit()
		}

		cli.createBLockChain(*createBLockChainAddress)
	}
}

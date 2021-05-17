package cli

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/TariqueNasrullah/iotchain/analysis"
	"github.com/TariqueNasrullah/iotchain/blockchain"
	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
)

// CommandLine structure
type CommandLine struct{}
type transData []string

func (trans *transData) String() string {
	return fmt.Sprint(*trans)
}

func (trans *transData) Set(value string) error {
	if len(*trans) > 0 {
		return errors.New("interval flag already set")
	}
	for _, dt := range strings.Split(value, ",") {
		*trans = append(*trans, dt)
	}
	return nil
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" node -addr ADDRESS -connect ADDRESS - RUN as node")
	fmt.Println(" address -f ADDRESS - Get addresses from a node")
	fmt.Println(" cleanup - Cleansup database")
	fmt.Println(" populate - Populates DB with test data")
	fmt.Println(" keygen - Generate Key")
	fmt.Println(" print - Print Chain")
	fmt.Println(" client - Client options")
}

func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(0)
	}
}

// Run runs commandline
func (cli *CommandLine) Run() {
	opts := badger.DefaultOptions(blockchain.DBPATH)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		panic(err)
	}
	badgerRepo := blockchain.NewBadgerRepository(db)

	cli.validateArgs()

	runNodeCmd := flag.NewFlagSet("node", flag.ExitOnError)
	nodeAddress := runNodeCmd.String("addr", "", "Node address")
	remoteNodeAddress := runNodeCmd.String("connect", "", "Address of node to with to connecect to")

	addressListCmd := flag.NewFlagSet("address", flag.ExitOnError)
	addressListCmdNodeAddress := addressListCmd.String("f", "", "Node address from which addresses are required")

	clearDbCmd := flag.NewFlagSet("cleanup", flag.ExitOnError)
	populateCmd := flag.NewFlagSet("populate", flag.ExitOnError)

	keyGenCmd := flag.NewFlagSet("keygen", flag.ExitOnError)
	keyGenCmdUsername := keyGenCmd.String("u", "", "Username")
	keyGenCmdPassword := keyGenCmd.String("p", "", "Password")
	keyGenCmdServerAddr := keyGenCmd.String("f", "", "Server address")

	printchainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	printchainCmdToken := printchainCmd.String("token", "", "Toekn")

	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	clientCmdSync := clientCmd.Bool("sync", false, "Sync Local Chain from a Miner")
	clientCmdMinerAddr := clientCmd.String("f", "", "Miner Address")
	clientCmdToken := clientCmd.String("token", "", "Token")
	clientCmdBlock := clientCmd.Bool("b", false, "Generate block")
	var transactions transData
	clientCmd.Var(&transactions, "t", "Comma seperated list of transactions")
	clientCmdBlockCount := clientCmd.Int("count", 1, "Number of blocks")

	testCmd := flag.NewFlagSet("test", flag.ExitOnError)
	testCmdAddr := testCmd.String("f", "", "address")

	analyzeCmd := flag.NewFlagSet("analyze", flag.ExitOnError)
	analyzeCmdBlockCount := analyzeCmd.Int("count", 1, "Number of blocks to generate (positive integer)")
	analyzeCmdBlockSise := analyzeCmd.Int("size", 1, "Block size in Kb (positive Integer)")
	analyzeCmdServerAddr := analyzeCmd.String("f", "", "Server Address")

	switch os.Args[1] {
	case "node":
		err := runNodeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "address":
		err := addressListCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "cleanup":
		err := clearDbCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "populate":
		err := populateCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "keygen":
		err := keyGenCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "print":
		err := printchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "client":
		err := clientCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "test":
		err := testCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "analyze":
		err := analyzeCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(0)
	}

	if runNodeCmd.Parsed() {
		blockchain.NodeAddress = *nodeAddress
		network := blockchain.Network{}

		chain, err := blockchain.InitBlockChain(blockchain.DBPATH, badgerRepo)
		if err != nil {
			logrus.Fatalf("Can't Initialize blockchain database %v\n", err)
		}
		blockchain.Chain = chain
		defer chain.Database.Close()

		go network.Serve(*nodeAddress)

		if *remoteNodeAddress == "" {
			logrus.Infoln("Server starting as stand alone")
		} else {
			network.SendAddress(*remoteNodeAddress)
			network.DiscoverAndConnect()
			bestHeightNode := network.FindBestHeightNode()

			if bestHeightNode == "" {
				logrus.Warn("No Node found with best height")
			} else {
				err := blockchain.ClearDB()
				if err != nil {
					logrus.Warnf("%v\n", err)
				} else {
					logrus.Info("Database Cleaned Up")
				}
				network.GetFullChain(bestHeightNode)
			}
		}
		time.Sleep(time.Duration(time.Second * 5))
		blockchain.PrintConnectedNodes()

		logrus.Infof("NODE BOOTED SUCCESSFULLY! Ready for mining!")
		for {
		}
	}
	if addressListCmd.Parsed() {
		if *addressListCmdNodeAddress == "" {
			addressListCmd.Usage()
			os.Exit(1)
		}

		network := blockchain.Network{}
		network.GetAddress(*addressListCmdNodeAddress)
	}
	if clearDbCmd.Parsed() {
		err := blockchain.ClearDB()
		if err != nil {
			logrus.Fatalf("%v\n", err)
		}
		logrus.Info("Database Cleaned Up")
	}
	if populateCmd.Parsed() {
		chain, err := blockchain.InitBlockChain(blockchain.DBPATH, badgerRepo)
		if err != nil {
			logrus.Fatalf("Can't Initialize blockchain database %v\n", err)
		}
		blockchain.Chain = chain
		defer chain.Database.Close()

		err = populateDb()
		if err != nil {
			logrus.Errorf("%v\n", err)
		}
		logrus.Info("Db populated")
	}

	if keyGenCmd.Parsed() {
		key, err := blockchain.GenerateKey(blockchain.KEYPATH)
		if err != nil {
			logrus.Fatalf("%v\n", err)
		}

		if *keyGenCmdUsername == "" && *keyGenCmdPassword == "" && *keyGenCmdServerAddr == "" {
			err := key.SaveFile(blockchain.KEYPATH)
			if err != nil {
				logrus.Fatalf("%v\n", err)
			}
			logrus.Infof("Key Generation Successfull")
			return
		}
		if *keyGenCmdUsername == "" || *keyGenCmdPassword == "" || *keyGenCmdServerAddr == "" {
			keyGenCmd.Usage()
			os.Exit(1)
		}

		network := blockchain.Network{}
		token, err := network.GetToken(*keyGenCmdUsername, *keyGenCmdPassword, *keyGenCmdServerAddr)
		if err != nil {
			logrus.Fatalf("%v\n", err)
		}

		key.Token = token
		err = key.SaveFile(blockchain.KEYPATH)
		if err != nil {
			logrus.Fatalf("%v\n", err)
		}
		fmt.Printf("%s\n", key)
		logrus.Infof("Key Generation Successfull")
	}
	if printchainCmd.Parsed() {
		var token []byte
		if *printchainCmdToken == "" {
			key, err := blockchain.LoadKey(blockchain.KEYPATH)
			if err != nil {
				logrus.Fatal(err)
			}
			token = key.Token
		} else {
			tkn, err := hex.DecodeString(*printchainCmdToken)
			if err != nil {
				logrus.Fatalf("%v\n", err)
			}
			token = tkn
		}

		chain, err := blockchain.InitBlockChain(blockchain.DBPATH, badgerRepo)
		if err != nil {
			logrus.Fatalf("Can't Initialize blockchain database %v\n", err)
		}
		blockchain.Chain = chain
		defer chain.Database.Close()

		network := blockchain.Network{}
		err = network.PrintChain(token)
		if err != nil {
			if err == badger.ErrKeyNotFound {
				log.Fatal("Token is invalid")
			}
			logrus.Fatalf("%v\n", err)
		}
	}
	if clientCmd.Parsed() {
		if *clientCmdMinerAddr == "" {
			clientCmd.Usage()
			logrus.Fatal("Miner address didn't mentioned")
		}

		var token []byte

		if *clientCmdToken == "" {
			key, err := blockchain.LoadKey(blockchain.KEYPATH)
			if err != nil {
				logrus.Fatal("Unable to load keys...")
			}
			token = key.Token
		} else {
			tkn, err := hex.DecodeString(*clientCmdToken)
			if err != nil {
				logrus.Fatal(err)
			}
			token = tkn
		}
		if *clientCmdBlockCount <= 0 {
			logrus.Fatal("Block count must be a positive number")
		}

		if *clientCmdSync == true {
			chain, err := blockchain.InitBlockChain(blockchain.DBPATH, badgerRepo)
			if err != nil {
				logrus.Fatal(err)
			}
			defer chain.Database.Close()
			blockchain.Chain = chain

			network := blockchain.Network{}
			err = network.DiscoverAndDownload(*clientCmdMinerAddr, token)
			if err != nil {
				logrus.Fatal(err)
			}
		} else if *clientCmdBlock == true {
			if len(transactions) == 0 {
				clientCmd.Usage()
				logrus.Fatal("No transaction data provided")
			}
			chain, err := blockchain.InitBlockChain(blockchain.DBPATH, badgerRepo)
			if err != nil {
				logrus.Fatal(err)
			}
			defer chain.Database.Close()
			blockchain.Chain = chain

			network := blockchain.Network{}
			for i := 1; i <= *clientCmdBlockCount; i++ {
				logrus.Infof("Generting Block: %v\n", i)
				err = network.CreateBlock(*clientCmdMinerAddr, token, transactions)
				if err != nil {
					logrus.Fatal(err)
				}
				logrus.Info("Block Mined Successfully")
			}
		}
	}
	if testCmd.Parsed() {
		network := blockchain.Network{}
		network.Test(*testCmdAddr)
	}
	if analyzeCmd.Parsed() {
		if *analyzeCmdBlockCount < 1 || *analyzeCmdBlockSise < 1 || len(*analyzeCmdServerAddr) == 0 {
			analyzeCmd.Usage()
			os.Exit(1)
		}
		analysis.BlockSize = *analyzeCmdBlockSise * 1000

		key, err := blockchain.LoadKey(blockchain.KEYPATH)
		if err != nil {
			logrus.Fatal(err)
		}
		token := key.Token

		chain, err := blockchain.InitBlockChain(blockchain.DBPATH, badgerRepo)
		if err != nil {
			logrus.Fatal(err)
		}
		defer chain.Database.Close()
		blockchain.Chain = chain
		network := blockchain.Network{}

		// generate fixed size transaction
		randomByte := make([]byte, analysis.BlockSize)
		rand.Read(randomByte)
		var transactions []string
		transactions = append(transactions, hex.EncodeToString(randomByte))
		// end

		for i := 1; i <= *analyzeCmdBlockCount; i++ {
			logrus.Infof("Generting Block: %v\n", i)
			err = network.CreateBlock(*analyzeCmdServerAddr, token, transactions)
			if err != nil {
				logrus.Fatal(err)
			}
			logrus.Info("Block Mined Successfully")
		}
	}
}

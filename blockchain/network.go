package blockchain

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	// KnownNodes holds known nodes address
	KnownNodes = make(map[string]struct{})
	// ConnectedNodes holds the address of connected nodes
	ConnectedNodes = make(map[string]*grpc.ClientConn)
	// Protocol defination
	Protocol = "tcp"

	// NodeAddress is the node address
	NodeAddress = ""
)

const (
	port = "8000"
)

// Network structure
type Network struct {
	Version        int
	Protocol       string
	NodeAddress    string
	KnownNodes     []string
	ConnectedNodes []string
}

// Serve serves
func (network *Network) Serve(addr string) {
	lis, err := net.Listen(Protocol, addr)
	if err != nil {
		logrus.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	RegisterMinerServer(s, &Server{})

	logrus.Info("Server started : ", addr)
	if err := s.Serve(lis); err != nil {
		logrus.Fatalf("failed to serve: %v", err)
	}
}

// Connect establish a connection to a grpc server and returns the connection and error
func (network *Network) Connect(srvAddr string) (*grpc.ClientConn, error) {
	conn, err := grpc.DialContext(context.Background(), srvAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(time.Second*10)))

	return conn, err
}

// SendAddress sends addr to a server
func (network *Network) SendAddress(srvAddr string) {
	conn, err := network.Connect(srvAddr)
	if err != nil {
		return
	}

	clinet := NewMinerClient(conn)

	response, err := clinet.SendAddress(context.Background(), &SendAddressRequest{Addr: NodeAddress})
	if err != nil {
		return
	}
	if response.StatusCode == 200 {
		ConnectedNodes[conn.Target()] = conn
		logrus.Infof("Connected to %v\n", conn.Target())
	}
	return
}

// GetAddress gets addresses from micro services
func (network *Network) GetAddress(srvAddr string) []string {
	var addrList []string

	client := NewMinerClient(ConnectedNodes[srvAddr])

	stream, err := client.GetAddress(context.Background(), &GetAddressRequest{})
	if err != nil {
		log.Printf("Err: %v\n", err)
		return addrList
	}

	for {
		addr, err := stream.Recv()
		if err == io.EOF {
			break
		}
		addrList = append(addrList, addr.Address)
	}
	return addrList
}

// DiscoverAndConnect connects to all nodes
func (network *Network) DiscoverAndConnect() {
	queue := []string{}
	for key := range ConnectedNodes {
		queue = append(queue, key)
	}

	for len(queue) != 0 {
		addr := queue[0]
		queue = queue[1:]
		addrList := network.GetAddress(addr)

		for _, newAddresses := range addrList {
			if newAddresses != NodeAddress && !contains(ConnectedNodes, newAddresses) {
				network.SendAddress(newAddresses)
				queue = append(queue, newAddresses)
			}
		}
	}
}

func (network *Network) discoverNodes(srvAddr string) {
	if !contains(ConnectedNodes, srvAddr) {
		conn, err := network.Connect(srvAddr)
		if err == nil {
			ConnectedNodes[conn.Target()] = conn
		}
	}
	queue := []string{}
	for srvAddr := range ConnectedNodes {
		queue = append(queue, srvAddr)
	}

	for len(queue) != 0 {
		addr := queue[0]
		queue = queue[1:]

		addrList := network.GetAddress(addr)

		for _, newAddresses := range addrList {
			if !contains(ConnectedNodes, newAddresses) {
				queue = append(queue, newAddresses)
				conn, err := network.Connect(newAddresses)
				if err != nil {
					continue
				}
				ConnectedNodes[conn.Target()] = conn
			}
		}
	}
}

// DiscoverAndDownload discovres the network and download best chain to local db
func (network *Network) DiscoverAndDownload(srvAddr string, token []byte) error {
	network.discoverNodes(srvAddr)
	fmt.Println(" --- Discovered nodes")
	for key := range ConnectedNodes {
		fmt.Println(key)
	}

	myHeight, err := Chain.Height(token)
	if err != nil {
		return err
	}
	bestHeightNode := network.FindBestHeightNodeByToken(token, myHeight)

	if bestHeightNode == "" {
		logrus.Info("No best node found to sync chain")
		return nil
	}
	logrus.Info("Best height node : ", bestHeightNode)

	err = network.GetChain(bestHeightNode, token)
	if err != nil {
		logrus.Infof("Chain Synchronization error: %v", err)
	}
	logrus.Info("Chain Synchronization Successfull")
	return nil
}

// CreateBlock creates block and send to a miner
func (network *Network) CreateBlock(srvAddr string, token []byte, transData []string) error {
	network.discoverNodes(srvAddr)
	if len(ConnectedNodes) == 0 {
		return errors.New("Unable to discover at lest one miner node")
	}
	discoveredNodeListString := []string{}
	for addr := range ConnectedNodes {
		discoveredNodeListString = append(discoveredNodeListString, addr)
	}

	var trans []*Transaction
	for _, data := range transData {
		trans = append(trans, &Transaction{Data: []byte(data)})
	}
	lastHash, err := Chain.LastHash(token)
	if err != nil {
		return err
	}
	key, err := LoadKey(KEYPATH)
	if err != nil {
		return err
	}

	block := Block{
		Transactions: trans,
		Token:        token,
		PrevHash:     lastHash,
		PublicKey:    key.PublicKey,
	}
	for {
		logrus.Infoln("Sigining Block")
		err = block.Sign(key.PrivateKey)
		if err != nil {
			return err
		}
		if ok := block.VerifySignature(); ok {
			logrus.Warnln("Sigining verification failed, retrying...")
			break
		}
		logrus.Infoln("Signing Success..")
	}

	serializedBlock, err := block.Serialize()
	if err != nil {
		return err
	}

	try := 1
	for {
		if try == 5 {
			return errors.New("Unable to mine block")
		}
		selectedAddr := discoveredNodeListString[rand.Intn(len(discoveredNodeListString))]
		logrus.Infof("Choosen miner address: %v\n", selectedAddr)
		err := network.Mine(selectedAddr, serializedBlock)
		try++
		if err != nil {
			logrus.Errorf("Unable to mine this node: %v\n", err.Error())
			logrus.Info("Retrying....")
			continue
		}
		break
	}

	return nil
}

// Mine send mine request to a miner
func (network *Network) Mine(srvAddr string, block []byte) error {
	client := NewMinerClient(ConnectedNodes[srvAddr])

	response, err := client.Mine(context.Background(), &MineRequest{Block: block})
	if err != nil {
		return err
	}
	deserilizedBlock, err := Deserialize(response.Block)
	if err != nil {
		return err
	}

	pow := NewProof(deserilizedBlock)
	validat := pow.Validate()
	if !validat {
		logrus.Error("Returned block pow invalid")
	} else {
		logrus.Info("returned block is valid")
	}

	err = Chain.AddBlock(deserilizedBlock)
	if err != nil {
		return err
	}
	fmt.Println("-- Mined Block")
	fmt.Printf("%s\n", deserilizedBlock)
	return nil
}

// GetFullHeight gets full height from a node
func GetFullHeight(srvAddr string, myHeight int64) (int64, error) {

	client := NewMinerClient(ConnectedNodes[srvAddr])

	response, err := client.FullHeight(context.Background(), &FullHeightRequest{Height: myHeight})
	if err != nil {
		return int64(0), err
	}
	return response.Height, nil
}

// FindBestHeightNode finds a node which has best height
func (network *Network) FindBestHeightNode() string {
	var addr string

	myHeight := Chain.FullHeight()
	max := myHeight

	for srvAddr := range ConnectedNodes {
		height, err := GetFullHeight(srvAddr, myHeight)
		if err != nil {
			logrus.Warnf("Error: %v", err)
		} else {
			if height > max {
				max = height
				addr = srvAddr
			}
		}
	}
	return addr
}

// FindBestHeightNodeByToken finds best height node by token
func (network *Network) FindBestHeightNodeByToken(token []byte, myHeight int64) string {
	var addr string
	max := myHeight

	for srvAddr := range ConnectedNodes {
		height, err := Getheight(srvAddr, token)
		if err != nil {
			logrus.Warnf("Error: %v", err)
		} else {
			if height > max {
				max = height
				addr = srvAddr
			}
		}
	}
	return addr
}

// Getheight get heights of a chain
func Getheight(srvAddr string, token []byte) (int64, error) {

	client := NewMinerClient(ConnectedNodes[srvAddr])
	resp, err := client.Height(context.Background(), &HeightRequest{Token: token})
	if err != nil {
		return 0, err
	}
	return resp.Height, nil
}

// GetFullChain downloads full blockchain from srvAddr node
func (network *Network) GetFullChain(srvAddr string) error {
	client := NewMinerClient(ConnectedNodes[srvAddr])
	stream, err := client.GetFullChain(context.Background(), &GetFullChainRequest{})
	if err != nil {
		return err
	}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}
		err = Chain.Database.Update(func(txn *badger.Txn) error {
			txn.Set(response.Key, response.Value)
			logrus.Info("Db Updating")
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// GetChain gets chain from server/miner
func (network *Network) GetChain(srvAddr string, token []byte) error {

	client := NewMinerClient(ConnectedNodes[srvAddr])
	stream, err := client.GetChain(context.Background(), &GetChainRequest{Token: token})
	if err != nil {
		return err
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		block, err := Deserialize(resp.Block)
		if err != nil {
			return err
		}
		if block.IsGenesis() {
			err := Chain.AddGenesis(block)
			if err != nil {
				return err
			}
		} else {
			err := Chain.AddBlock(block)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// PropagateBlock propagates a block accross the network
func (network *Network) PropagateBlock(block []byte, srvAddr string) {
	client := NewMinerClient(ConnectedNodes[srvAddr])
	_, err := client.PropagateBlock(context.Background(), &PropagateBlockRequest{Block: block})
	if err != nil {
		logrus.Warnf("%v\n", err)
	}
}

// GetToken gets token from a miner
func (network *Network) GetToken(username, password, srvAddr string) ([]byte, error) {
	conn, err := grpc.DialContext(context.Background(), srvAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(time.Second*10)))
	if err != nil {
		logrus.Warnf("%v\n", err)
		return []byte{}, err
	}
	defer conn.Close()

	client := NewMinerClient(conn)
	resp, err := client.Token(context.Background(), &TokenRequest{Username: username, Password: password})
	if err != nil {
		return []byte{}, err
	}
	return resp.Token, nil
}

// Printchain prints the chain of an address
func (network *Network) Printchain(token []byte) error {
	address, err := Address(token)
	if err != nil {
		return err
	}
	var lastHash []byte
	err = Chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(address)
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			lastHash = val
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	itr := Iterator{CurrentHash: lastHash, Database: Chain.Database}

	for {
		block := itr.Next()
		if block == nil {
			break
		}

		fmt.Printf("%s\n", block)

		if len(block.PrevHash) == 0 {
			break
		}
	}
	return nil
}

// Ping pings a server/miner
func (network *Network) Ping(srvAddr string) bool {
	conn, err := grpc.DialContext(context.Background(), srvAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(time.Second*10)))
	if err != nil {
		return false
	}
	defer conn.Close()

	client := NewMinerClient(conn)
	_, err = client.Ping(context.Background(), &PingRequest{})
	if err != nil {
		return false
	}
	return true
}

func contains(data map[string]*grpc.ClientConn, val string) bool {
	if _, ok := data[val]; !ok {
		return false
	}
	return true
}

// Test tests
func (network *Network) Test(srvAddr string) {
	conn, err := grpc.DialContext(context.Background(), srvAddr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(time.Duration(time.Second*10)))
	if err != nil {
		logrus.Panic(err)
		return
	}
	defer conn.Close()

	client := NewMinerClient(conn)

	trans := Transaction{Data: []byte("ok")}
	block := Block{
		PrevHash:     []byte("prevhash"),
		Transactions: []*Transaction{&trans},
		Token:        []byte("token"),
		PublicKey:    []byte("pubkey"),
	}
	_, err = block.Serialize()
	if err != nil {
		logrus.Fatal(err)
	}

	resp, err := client.Test(context.Background(), &TestRequest{})
	if err != nil {
		logrus.Fatal(err)
	}
	rblock, err := Deserialize(resp.Block)
	if err != nil {
		logrus.Panic(err)
	}
	pow := NewProof(rblock)
	validate := pow.Validate()

	if !validate {
		logrus.Error("received block is not valid")
	} else {
		logrus.Info("received block is valid")
	}
	fmt.Printf("%s\n", rblock)
}

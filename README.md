# Private-IoT-blockchain [![Build Status](https://travis-ci.org/TariqueNasrullah/Private-IoT-blockchain.svg?branch=master)](https://travis-ci.org/github/TariqueNasrullah/Private-IoT-blockchain) ![Go](https://github.com/TariqueNasrullah/Private-IoT-blockchain/workflows/Go/badge.svg?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/TariqueNasrullah/Private-IoT-blockchain)](https://goreportcard.com/report/github.com/TariqueNasrullah/Private-IoT-blockchain)

## Spining Up Miner Node
Stand Alone

    go run main.go node _node_addr:port

Example

    root@797c26b08d71:/go/src/app# go run main.go node -addr 172.17.0.2:8000
    INFO[0000] Server starting as stand alone               
    INFO[0000] Server started : 172.17.0.2:8000             
      --Connected Nodes
    INFO[0005] NODE BOOTED SUCCESSFULLY! Ready for mining!
        
Connect Miner node with existing network

    go run main.go node _node_addr:port -connect _known_miner_addr:8000
    
Example

    root@819389d44dbc:/go/src/app# go run main.go node -addr 172.17.0.3:8000 -connect 172.17.0.2:8000
    INFO[0000] Server started : 172.17.0.3:8000             
    INFO[0000] Connected to 172.17.0.2:8000                 
    WARN[0000] No Node found with best height               
      --Connected Nodes
        172.17.0.2:8000
    INFO[0005] NODE BOOTED SUCCESSFULLY! Ready for mining!
    
## Client
### -Registration
    
    go run main.go keygen -f _miner_addr:port -u username -p password
    
### -Sync client local chain from miners

    go run main.go client -sync -f 172.17.0.3:8000
    
### -Generate Block

    go run main.go client -b -t comma_seperated_transaction_list -f _miner_addr:port
    
Example
    
    root@006f30561cb8:/go/src/app# go run main.go client -b -t 100,200,300 -f 172.17.0.3:8000
    INFO[0000] Choosen miner address: 172.17.0.2:8000       
    INFO[0000] returned block is valid                      
    -- Mined Block

    ----Block: 
     PrevHash  : 000651B7DA0CE05411B6F49E1B9CA8624074C5282B12343BCF19922CD0B991F5
     Hash      : 00010909003287AD52140B30946FCC89913809AD95509651B9F37FD6DBA11629
     Nounce    : 901
     Signature : A8704460E4DB3F12171600100E4C8A9089EEC5168A483DDB0DEDEC326AA81CB678065C54A21931C78ED6366BEB3FCB41AFE001FC16F5F022559F3256DF9D7A6E
     Token     : 28DCF8AC5DB136B362A926FDB3F59E8CBCE7CAA4B9423616BD5566A657C44435E2343E2FB7BB17917D7A72FBC8022B18B909E1D46410A6DF254F92C644295337
     PublicKey : 39DF79DE5A375AA3CA5C21700A5159E2E9D947D4545FC7FDC9987F317BD275E6E1AB995FB8415E3269649AE211BA2DDAE820900D119CCB4FAB68206EE2EA4C4E
       ├──Transaction  : 100
       ├──Transaction  : 200
       ├──Transaction  : 300
    INFO[0000] Block Mined Successfully

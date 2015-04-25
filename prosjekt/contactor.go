package main

import(
	"network"
	"time"
	"fmt"
	"strconv"
)

var liftsOnline = make(map[string]network.ConnectionUDP)
var disconnElevChan = make(chan network.ConnectionUDP)

func writemap() {
	for {
		fmt.Println("Size av liftsOnline:",len(liftsOnline))
		fmt.Println("liftsOnline:",liftsOnline)
		fmt.Println("Master har ID:",selectMaster(liftsOnline))
		time.Sleep(time.Second*10)
	}
}

func selectMaster(lifts map[string]network.ConnectionUDP)string {
	master := "-1"
	for key := range lifts{
		m, _ := strconv.Atoi(master)
		k, _ := strconv.Atoi(key)
		if k > m{
			master = key
		}
	}
	return master
}

//func masterAlive()bool{
//}
/* ---HUSKELISTE
	ElevInfo må pakkes i ElevatorRun
	* ConnTimer tar kun inn network.ConnectionUDP, og vet ikke ID'en
	* -> fikset at den sletter nå
*/ /*
func Slave(){

	updateOutChan := make(chan network.Message)
	updateInChan := make(chan network.Message)
	extOrderChan := make(chan ButtonSignal)
	newOrderChan := make(chan ButtonSignal)
	fromMasterChan := make(chan ButtonSignal)
		
	//go ELEVATORRUN
	go networkHandler()
	
	for{
		select{
			case update := <- updateOutChan
				updateMsg := network.Message{Content: network.Info, Addr: "broadcast", ElevInfo: update}
				network.MessageChan <- updateMsg
			case extOrdButtonSignal := <- extOrderChan
				extOrdMsg := network.Message{Content: network.NewOrder, Addr: <addresse til master>, Floor: extOrdButtonSignal.Floor, Button: extOrdButtonSignal.Button}
				network.MessageChan <- extOrdMsg //Få pakket forunftig beskjed her først da, med recieverAddr til MASTER
			case order := <- newOrderChan
				fromMasterChan <- order
		}	
	}
}

*/

func main(){
	var kanal = make(chan int)
	
	go networkHandler()
	
	<-kanal
}

func networkHandler(){
	network.Init()
	go writemap()
	for{
		select{
		case msg := <-network.RecieveChan:
			messageHandler(network.ParseMessage(msg))
		case conn := <- disconnElevChan:
			deleteLift(conn.Addr)
		}
			
			
	}
}

//  msg.Addr byttet med id i liftsonline(key)
func messageHandler(msg network.Message) {
	id := network.FindID(msg.Addr)
	switch msg.Content{
		case network.Alive:
			if conn, alive := liftsOnline[id]; alive{
				conn.Timer.Reset(network.ResetConnTime)
			} else{
				newConn := network.ConnectionUDP{msg.Addr, time.NewTimer(network.ResetConnTime)}
				
				liftsOnline[id] = newConn
				go connTimer(newConn)
			}
		case network.NewOrder:
			//kanskje skrive ut noe?
			// KANAL skrive orderen til canalen
			
			//cost :=  69 //costfunksjonen inn her
			
			//costMsg := network.Message{Content: network.Cost, Floor: msg.Floor, Button: msg.Button, Cost: cost}
			//network.MessageChan <- costMsg
		case network.CompletedOrder:
			//Slette order
		case network.Info:
			// LAGRER INFO OM HEISEN M/ TILHØRENDE ID, hvor?
		}		
}

func connTimer(conn network.ConnectionUDP){
	for{
	<-conn.Timer.C
	disconnElevChan <- conn
	}
}

//endret slik at den sletter vha idkey, ikke addr key
func deleteLift(addr string){
	id := network.FindID(msg.Addr)
	delete(liftsOnline, id)
}

package main
/*
updateINOUTChan blir nå av tupen ElevatorInfo. Det tror jeg er greit,
men da kanskje legge inn id i ElevatorInfo også for lettere arbeit.?
*/
import(
	"network"
	"time"
	"fmt"
	"strconv"
	"net"
	."control"
	."driver"
	"filehandler"
)

var liftsOnline = make(map[string]network.ConnectionUDP)
var LiftsOnlineInfo = make(map[string]network.ElevatorInfo)
var disconnElevChan = make(chan network.ConnectionUDP)
var myID string = findMyID()

var uncompleteOrders Matrix

func writemap(checkMasterChan chan string) {
	for {
		fmt.Printf("%v lifts online, master id: %v\n",len(liftsOnline),selectMaster(liftsOnline, checkMasterChan))
		fmt.Println("liftsOnline:",liftsOnline)
		time.Sleep(time.Second*10)
	}
}

func selectMaster(lifts map[string]network.ConnectionUDP, checkMasterChan chan string)string {
	master := myID
	for key := range lifts{
		m, _ := strconv.Atoi(master)
		k, _ := strconv.Atoi(key)
		if k > m{
			master = key
		}
	}
	checkMasterChan<-master
	return master
}


//}
/* ---HUSKELISTE
	ElevInfo må pakkes i ElevatorRun

*/ /* */
func main(){
	hengekanal := make(chan int)

	updateOutChan := make(chan network.ElevatorInfo)
	updateInChan := make(chan network.ElevatorInfo)//Trenger ikke egen case, messageHandler tar ansvar
	checkMasterChan := make(chan string)
	//checkMasterChan<-myID	
	newOrderChan := make(chan ButtonSignal,20)
	completeOrderChan := make(chan ButtonSignal)
	extOrderChan := make(chan ButtonSignal)
	fromMasterChan := make(chan ButtonSignal)//Kanalen som går inn til elev og gir oppdrag

	updateLightsSlaveChan := make(chan int)
	completeOrderBroadcastChan := make(chan ButtonSignal)	
	backupChan := make(chan /*map[string]*/Matrix)//network.ElevatorInfo)
	//elevInfoChan := make(chan network.ElevatorInfo)
	
	// Holder orden på master/slave-rollen	
	//terminateChan := make(chan bool)
	//terminatedChan := make(chan bool)

	go backupHandler(backupChan)
	
	go networkHandler(updateInChan, checkMasterChan, newOrderChan, completeOrderChan, completeOrderBroadcastChan, updateLightsSlaveChan)
	go InitElevator(updateOutChan,  checkMasterChan, completeOrderChan, extOrderChan, fromMasterChan)
	go Master(updateOutChan, checkMasterChan, newOrderChan, completeOrderChan, extOrderChan, fromMasterChan, completeOrderBroadcastChan, updateLightsSlaveChan)

	<-hengekanal
}

//func checkMaster(checkMasterChan chan string){


func Slave(updateOutChan chan network.ElevatorInfo, checkMasterChan chan string, newOrderChan chan ButtonSignal, completeOrderChan chan ButtonSignal, extOrderChan chan ButtonSignal, fromMasterChan chan ButtonSignal, completeOrderBroadcastChan chan ButtonSignal, updateLightsSlaveChan chan int){		
	var masterAddr string 
	var master string
	fmt.Println("JEG ER NÅ EN SLAVE")
	for{
		select{
			case master = <- checkMasterChan:
				if master == myID{
					//terminateChan<-true
					//<-terminatedChan
					go Master(updateOutChan, checkMasterChan, newOrderChan, completeOrderChan, extOrderChan, fromMasterChan, completeOrderBroadcastChan, updateLightsSlaveChan)
					return
				} //PROBLEM UNDER- i init fasen, vil ikke masterADdr eksistere
				masterAddr = liftsOnline[master].Addr
			case update := <- updateOutChan:
				updateMsg := network.Message{Content: network.Info, Addr: "broadcast", ElevInfo: update}
				network.MessageChan <- updateMsg
			case extOrdButtonSignal := <- extOrderChan:
				fmt.Println("Slave mottatt fra extOrderChan")
				extOrdMsg := network.Message{Content: network.NewOrder, Addr: masterAddr, Floor: extOrdButtonSignal.Floor, Button: extOrdButtonSignal.Button}
				network.MessageChan <- extOrdMsg 
			// PROBLEM - lagre extOrd lokalt i tilfelle intet svar, kjøre da for å være sikker??
				
	
			case orderButtonSignal := <- newOrderChan:
				fromMasterChan <- orderButtonSignal
			case completeOrder := <- completeOrderChan:
				complOrdMsg := network.Message{Content: network.CompletedOrder, Addr: masterAddr, Floor: completeOrder.Floor, Button: completeOrder.Button}
				network.MessageChan <- complOrdMsg
			case <-updateLightsSlaveChan:
				for i := 0; i < network.NumFloors; i++{
					for j := 0; j < network.NumButtons-1; j++{
						if LiftsOnlineInfo[master].Matrix[i][j] {
							Elev_set_button_lamp(ButtonSignal{Light: 1})
						} else {
							Elev_set_button_lamp(ButtonSignal{Light: 0})						
						}
					}	
				}
		}	
	}
}

func Master(updateOutChan chan network.ElevatorInfo, checkMasterChan chan string, newOrderChan chan ButtonSignal, completeOrderChan chan ButtonSignal, extOrderChan chan ButtonSignal, fromMasterChan chan ButtonSignal, completeOrderBroadcastChan chan ButtonSignal, updateLightsSlaveChan chan int){
	fmt.Println("JEG ER NÅ MASTER")
	//masterAddr := liftsOnline[myID].Addr
	//masterA := liftsOnline[myID]
	fmt.Println("----------MASTERADDRESSE: ", myID)
	fmt.Println(liftsOnline)
	
	for{
		select{
			case master := <- checkMasterChan:
				if master != myID{
					//terminateChan<-true
					//<-terminatedChan
					go Slave(updateOutChan, checkMasterChan, newOrderChan, completeOrderChan, extOrderChan, fromMasterChan, completeOrderBroadcastChan, updateLightsSlaveChan)
					return
				}
			case update := <- updateOutChan:
				updateMsg := network.Message{Content: network.Info, Addr: "broadcast", ElevInfo: update}
				network.MessageChan <- updateMsg
			case orderButtonSignal := <- newOrderChan: //Finner beste heis til å ta jobben og sender til den
				fmt.Println("newOrderChan leverer videre")
				//returnerer heis best egnet for jobbet
				//heisID := myID//costfunction(orderButtonSignal) //order inneholder opp/ned+etasje
		// --- LAGRE ordren i uncompleteOrders og slette igjen når completed er mottatt??
				if uncompleteOrders[orderButtonSignal.Floor][orderButtonSignal.Button] == true{
					fmt.Println("BLOKKERT ORDERE")
					continue
				} 
				uncompleteOrders[orderButtonSignal.Floor][orderButtonSignal.Button] = true
				Elev_set_button_lamp(orderButtonSignal)
				
				heisID := myID//costfunction(orderButtonSignal) //order inneholder opp/ned+etasje
		// --- LAGRE ordren i uncompleteOrders og slette igjen når completed er mottatt??
				cost := 100
				newCost := 101				
				for key := range LiftsOnlineInfo {
					heisInfo := LiftsOnlineInfo[key]
					/*if (heisInfo.Dir > 0 && heisInfo.Floor > orderButtonSignal.Floor) {
						newCost = control.ComplexCost(heisInfo.Floor, orderButtonSignal.Floor	
					} else if (heisInfo.Dir < 0 && heisInfo.Floor < orderButtonSignal.Floor) {
						newCost = control.ComplexCost(heisInfo.Dir, heisInfo.Floor, heisInfo.Matrix, orderButtonSignal.Floor
					} else {
						newCost = control.SimpleCost(heisInfo.Floor, orderButtonSignal.Floor)
					}
					if newCost < cost {
						cost = newCost
						bestID = key
					}
				}*/
					
					if heisInfo.Dir == 0 {
						newCost = SimpleCost(heisInfo.Floor, orderButtonSignal.Floor)			
					} else if (heisInfo.Dir > 0 && ((heisInfo.Floor < orderButtonSignal.Floor) && orderButtonSignal.Button == BUTTON_CALL_UP)){
						newCost = SimpleCost(heisInfo.Floor, orderButtonSignal.Floor)
					} else if (heisInfo.Dir < 0 && ((heisInfo.Floor > orderButtonSignal.Floor) && orderButtonSignal.Button == BUTTON_CALL_DOWN)){
						newCost = SimpleCost(heisInfo.Floor, orderButtonSignal.Floor)
					} else {
						newCost = ComplexCost(heisInfo.Dir, heisInfo.Floor, heisInfo.Matrix, orderButtonSignal.Floor)
					}
					if newCost < cost {
						cost = newCost
						heisID = key
					}
				}

				if heisID == myID{
					fromMasterChan<-orderButtonSignal
				}else{
					addrOrderReciever := liftsOnline[heisID].Addr
					newOrderMsg := network.Message{Content: network.NewOrder, Addr: addrOrderReciever, Floor: orderButtonSignal.Floor, Button: orderButtonSignal.Button}
					network.MessageChan <- newOrderMsg
				}
				updateMsg := network.Message{Content: network.Light, Addr: "broadcast", ElevInfo: network.ElevatorInfo{Matrix: uncompleteOrders}}
				network.MessageChan <- updateMsg
			case extOrdButtonSignal := <- extOrderChan: //Type ButtonSignal
				//---------- LEGGE TIL I UNCOMPLETE ORDERS, HVIS IKKE FÅTT SLETTET VHA completeOrderChan innen gitt tid (si 10 sek), ta ordren og legg inn i interne ordre matrisen

				
					
				


				fmt.Println("Master mottatt fra extOrderChan")
				//extOrdMsg := network.Message{Content: network.NewOrder, Addr: masterAddr, Floor: extOrdButtonSignal.Floor, Button: extOrdButtonSignal.Button}
				//fmt.Println("MASTERADDRESSE: ", masterAddr)
				//network.MessageChan <- extOrdMsg //Få pakket forunftig beskjed her først da, med recieverAddr til MASTER
				newOrderChan <- extOrdButtonSignal
			
			case completeOrderLocal := <- completeOrderChan: //Type ButtonSignal
				// slette fra uncompleteOrders (PS; SYNKE uncompleteOrders til alle for backup?) (no orders lost)
				//_ = completeOrderLocal
				uncompleteOrders[completeOrderLocal.Floor][completeOrderLocal.Button] = false
				Elev_set_button_lamp(completeOrderLocal)
			case completeOrderBroadcast := <-completeOrderBroadcastChan:
				// Her sjekke om den orderen finnes i min lokale kø fra extOrderChan, evt slette
				uncompleteOrders[completeOrderBroadcast.Floor][completeOrderBroadcast.Button] = false
				Elev_set_button_lamp(completeOrderBroadcast)
			case <-updateLightsSlaveChan:
				continue
		}
	}
}

func findMyID() string/*int*/ {
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                IDstr := ipnet.IP.String()
                IDstr = IDstr[12:]
                //ID, _ := strconv.Atoi(IDstr)
                return IDstr//ID
            }
        }
			//fmt.Println(err)
//BØR FIKSES SALIG-.-.-.-.-.-.-.-.-.-.-.-.--.--.-.-	
	} 
	return "69"
}

func backupHandler(backupChan chan Matrix/*network.ElevatorInfo.Matrix*/ ){
	//filehandler.Init()
	loadBackupTime := time.Second*10
	saveBackupTime := time.Second
	
	loadTimer := time.NewTimer(loadBackupTime)
	saveTimer := time.NewTimer(saveBackupTime)
	for{
		select{
		case <-saveTimer.C: //tar 
			filehandler.SaveBackup(LiftsOnlineInfo)
			//restart timer!!
			saveTimer.Reset(saveBackupTime)
		case conn := <- disconnElevChan: //LEGGER inn ordre fra død heis
			id := network.FindID(conn.Addr)
			m := matrixCompareOr(LiftsOnlineInfo[id].Matrix, LiftsOnlineInfo[myID].Matrix)
			backupChan<-m //Sender OR'et matrise til elevator.
			//________ BURDE KANSKJE FORDELES PÅ NYTT IGJEN???? OG INTERNE ORDRE BURDE IGNORERES
			delete(LiftsOnlineInfo, id)
			//sletting av  liftsOnline blir gjort i networkHandler
		case <-loadTimer.C: //SJEKKER EGEN MATRISE MOT BACKUPEN
			backupMap := filehandler.LoadBackup()
			var backupMatrix /*(network.ElevatorInfo).*/Matrix = backupMap[myID].Matrix
			if backupMatrix != LiftsOnlineInfo[myID].Matrix{
				backupMatrix = matrixCompareOr(backupMatrix, LiftsOnlineInfo[myID].Matrix)
				backupChan<-backupMatrix
			}
			loadTimer.Reset(loadBackupTime)
		}
	}
}




func matrixCompareOr(m1 Matrix, m2 Matrix) (Matrix){
	var m3 Matrix
	for i := 0; i < network.NumFloors; i++ {
		for j := 0; j < network.NumButtons; j++{
			if m1[i][j] == true || m2[i][j] == true{
				m3[i][j] = true
			}
		}
	}
	return m3
}

func networkHandler(updateInChan chan network.ElevatorInfo, checkMasterChan chan string, newOrderChan chan ButtonSignal, completeOrderChan chan ButtonSignal, completeOrderBroadcastChan chan ButtonSignal, updateLightsSlaveChan chan int){
	network.Init()
	go writemap(checkMasterChan)
	for{
		select{
		case msg := <-network.RecieveChan:
			messageHandler(network.ParseMessage(msg), updateInChan, checkMasterChan, newOrderChan, completeOrderChan, completeOrderBroadcastChan, updateLightsSlaveChan)
		case conn := <- disconnElevChan:
			deleteLift(conn.Addr, checkMasterChan)
		}
			
			
	}
}

//  msg.Addr byttet med id i liftsonline(key)
func messageHandler(msg network.Message, updateInChan chan network.ElevatorInfo, checkMasterChan chan string, newOrderChan chan ButtonSignal, completeOrderChan chan ButtonSignal, completeOrderBroadcastChan chan ButtonSignal, updateLightsSlaveChan chan int) {
	id := network.FindID(msg.Addr)
	switch msg.Content{
		case network.Alive:
			if conn, alive := liftsOnline[id]; alive{
				conn.Timer.Reset(network.ResetConnTime)
			} else{
				newConn := network.ConnectionUDP{msg.Addr, time.NewTimer(network.ResetConnTime)}
				
				liftsOnline[id] = newConn
				go connTimer(newConn)
				selectMaster(liftsOnline, checkMasterChan) //CHECKMASTER
			}
		case network.NewOrder:
			//kanskje skrive ut noe?
			// KANAL skrive orderen til canalen
			/*tempButtonSignal := ButtonSignal{Floor: msg.Floor, Button: msg.Button} 
			newOrderChan<-tempButtonSignal*/
			newOrderChan<-ButtonSignal{Floor: msg.Floor, Button: msg.Button, Light: 1} 
			//cost :=  69 //costfunksjonen inn her
			
			//costMsg := network.Message{Content: network.Cost, Floor: msg.Floor, Button: msg.Button, Cost: cost}
			//network.MessageChan <- costMsg
		case network.CompletedOrder:
			//Slette order - denne casen er kun for master??
			completeOrderBroadcastChan <- ButtonSignal{Floor: msg.Floor, Button: msg.Button}
		case network.Info:
			// LAGRER INFO OM HEISEN M/ TILHØRENDE ID, hvor?
			//elevInfoChan<-msg.ElevInfo	//HVA med å bare utføre arbeitet her
			LiftsOnlineInfo[id] = msg.ElevInfo
			updateLightsSlaveChan<-1
		case network.Light:
			updateLightsSlaveChan<-1
			//fmt.Println(LiftsOnlineInfo)
// bruke updateInChan	--- endret til elevInfoChan??		
		}		
}
/*
func elevInfoSaver(elevInfoChan chan network.ElevatorInfo){
		
}*/


func connTimer(conn network.ConnectionUDP){
	for{
	<-conn.Timer.C
	disconnElevChan <- conn
	}
}

//endret slik at den sletter vha idkey, ikke addr key
func deleteLift(addr string, checkMasterChan chan string){
	id := network.FindID(addr)
	delete(liftsOnline, id)
	selectMaster(liftsOnline, checkMasterChan) //CHECKMASTER
}

package main

import (
	"fmt"
	"net"
	"bufio"
	"os"
	"runtime"
	"strings"
	"time"
)

const selfIP = "129.241.187.144"
const targIP = "129.241.187.136"
const broadcastIP = "129.241.187.255"

const writePort = "20006"
const recievePort = "30000"
const fixSizePort = "34933"
const readPort = "33546"


func broadcastLocalIP(){

	for {
		bcConn := setupSenderUDP()
		melding := findLocalIP()

		fmt.Println("ip funnet:" + melding)

		byteMelding := []byte(melding)
		bcConn.Write(byteMelding)

		fmt.Println("melding sendt")
		time.Sleep(100*time.Millisecond)
		}
	return
}

func listenForIP(){
	str := make([]byte, 1024)
	listenConn := SetupListenUDP()
	
	fmt.Println("satt opp lyttesocket")

	for {
		fmt.Println("starten av lytteløkke")
		time.Sleep(200*time.Millisecond)
		_, senderAddr, err := listenConn.ReadFromUDP(str[:])

		fmt.Println("mottatt melding")	

		if err != nil {
			fmt.Println("ReadFromUDP error")
		}
		fmt.Println("message: ")
		fmt.Println(string(str))
		fmt.Println("from addr: ")
		fmt.Println(senderAddr)
	}
	return
}


func connect(targetIP string, targetPort string) *net.TCPConn {
	TCPAddr, _ := net.ResolveTCPAddr("tcp", targetIP + ":" + targetPort)	
	conn, err := net.DialTCP("tcp", nil, TCPAddr)
	//defer conn.Close()
    if err!= nil {
    	fmt.Fprintln(os.Stderr, "Tried to connect to: " + targetIP + ":" + targetPort)
    	fmt.Fprintln(os.Stderr, " Connection error: " + err.Error())
       	return conn
    }
	return conn
}

func acceptTCPConn() *net.TCPConn {
	listener = net.ListenTCP("tcp", 0)
	incConn = net.AcceptTCP(listener)
	return incConn
}

/*
func disconnect(connection *net.TCPConn) {
	connection.Close()
	return
}
*/

func write(connection *net.TCPConn, msg string, reader *bufio.Reader){
	for {
    	fmt.Print("Enter text: ")
    	msg, _ = reader.ReadString('\n')

    	connection.Write([]byte(msg[:len(msg)-1] + "\000"))
		
		time.Sleep(1*time.Millisecond)
		
		if strings.ToLower(msg) == "disconnect\n" {break}	
	}
	return
}

func read(reader *bufio.Reader){
	for {
		str, err := reader.ReadString('\000')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Delim error: " + err.Error())
			return
		}
		fmt.Println("Melding: " + str)
		if strings.ToLower(str) == "disconnect\000" {break}
	}
	return
}


func findLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	for _, address := range addrs {
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return "0.0.0.0"
}

func setupSenderUDP() *net.UDPConn {
	addr, _ := net.ResolveUDPAddr("udp", broadcastIP + ":" + writePort)
	sock, _ := net.DialUDP("udp", nil, addr)
	/*
	melding := "sup"
	byteMelding := []byte(melding)
	sock.Write(byteMelding)
	*/
	return sock
}

func SetupListenUDP() *net.UDPConn {
	addr, _ := net.ResolveUDPAddr("udp", ":" + writePort)
	sock, _ := net.ListenUDP("udp", addr)
	
	/*
	_, senderAddr, _ := sock.ReadFromUDP(str[:])
	fmt.Println("message: ")
	fmt.Println(string(str))
	fmt.Println("from addr: ")
	fmt.Println(senderAddr)
	*/
	
	/*
	if err != nil {
		fmt.println("Error: " + err.Error())
		return
	}
	*/
	return sock
}


func main() {
	fmt.Println("starter main")
	ch := make(chan string)
	
	runtime.GOMAXPROCS(runtime.NumCPU())
	

	go listenForIP()
	fmt.Println("startet listen")
	
	broadcastLocalIP()

	fmt.Println("ferdig med broadcast")

	/*
	Conn := connect(targIP, readPort)
	
	var message string = "Initialized"
	sendMsgReader := bufio.NewReader(os.Stdin)
	receiveMsgReader := bufio.NewReader(Conn)
	
	go read(receiveMsgReader)
	time.Sleep(1*time.Millisecond)
	go write(Conn, message, sendMsgReader)
	*/
	
	
	<- ch

	//Conn.Close()
	
}



/*
func readFixSizeTCP(reader *bufio.Reader){
	str := make([]byte, 1024)
	n, err := reader.Read(str)    // returnerer lengden på beskjeden (n)
	if n != 1024 || err != nil {
		fmt.Fprintln(os.Stderr, "Message size: " + strconv.Itoa(n))
		fmt.Fprintln(os.Stderr, "Fixed size error: " + err.Error())
		return
	}
	fmt.Println("Melding: " + string(str))
	return
}
*/
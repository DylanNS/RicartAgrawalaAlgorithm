package main
import (
"fmt"
"net"
"os"
"strconv"
"time"
"bufio"
"strings"
)
//Variáveis globais interessantes para o processo
var err string
var myPort string //porta do meu servidor
var nServers int //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores
 //dos outros processos
var ServConn *net.UDPConn //conexão do meu servidor (onde recebo
 //mensagens dos outros processos)

var id int //numero identificador do processo
var mylogicalClock int
var estouNaCS bool
var estouEsperando bool
var received_all_replies bool
var sharedResource *net.UDPConn
var queued_request []int
var lc_requisicao int
var replies_received []int


func CheckError(err1 error){
	if err1 != nil {
		fmt.Println("Erro: ", err1)
		os.Exit(0)
	}
}

func PrintError(err error) {
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}

func Am_I_priority(pj_id int, lc_pj int) bool {

	//my logical clock at request is lower than another's
	if lc_requisicao < lc_pj {
		//I am priority =)
		return true
	} else if lc_pj > lc_requisicao {
		//my logical clock at request isn't lower than another's
		//I'm not priority TT
		return false
	} else {
		//my logical clock at request is equal to another's
		if id < pj_id {
			//but my id is lower
			//So i am priority =)
			return true
		} else {
			//my id isn't lower
			//I'm not priority TT
			return false
		}
	}
}

func queue_request_from_pj(pj_id int){
	//queue request from pj without replying
	queued_request = append(queued_request,pj_id)
}

func reply2pj(pj_id int){
	//reply immediately to pj
	//build reply menssage
	str_lc:= strconv.Itoa(mylogicalClock)
	//transformar id em string 
	str_id := strconv.Itoa(id)
	// concatenar todas
	mymsg :=  str_id + "," + str_lc + ",reply"
	buf := []byte(mymsg)
	//Enviar mensagem para o pj
	index := pj_id - 1
	//reply to pj
     _,err := CliConn[index].Write(buf)
     if err != nil {
        fmt.Println(mymsg, err)
	}
}


func MaxInt(x1, x2 int) int {


	if x1 > x2 {
		return x1
	}
	return x2
}

func procurar_in_list (pj_id int) bool{

	for _,i := range replies_received {
		if i == pj_id{
			return true
		}
	}
	return false
}

func doServerJob() {
//Ler (uma vez somente) da conexão UDP a mensagem
//Escreve na tela a msg recebida


	 buf := make([]byte, 1024)
	 for {

		 n, _, err := ServConn.ReadFromUDP(buf)
		 //aux = "id,logical_clock"
		 aux := string (buf[0:n])
		 //stream_msg = ["id" , "logical_clock"]
		 stream_msg := strings.Split(aux,",")
		 str_pj_id := stream_msg[0]
		 str_lc_pj := stream_msg[1]
		 pj_id,err := strconv.Atoi(str_pj_id)
		 lc_pj,err := strconv.Atoi(str_lc_pj)
		 fmt.Println("Recebi mensagem de ",str_pj_id," com relogio de ",str_lc_pj, " do tipo ",stream_msg[2])
		 //If menssage is request
		 if stream_msg[2] == "request" {
			 //I am in CS or (I want CS and preference is mine) then queue pj
			if estouNaCS || (estouEsperando && Am_I_priority(pj_id,lc_pj)){
				//queue request from pj withou replying
				fmt.Printf("\nEnfilerei %d com relogio %d, pois estouNaCS = %t, estouEsperando = %t, meuid = %d, meu relogio = %d\n",pj_id,lc_pj,estouNaCS,estouEsperando,id,lc_requisicao)
				queue_request_from_pj(pj_id)

			} else {
				//reply immediately to pj
				fmt.Printf("\nEnviando reply <id,clock> = < %d , %d > para %d \n",id,mylogicalClock,pj_id)
				reply2pj(pj_id)
			}
		} else if stream_msg[2] == "reply" {
			//If menssage is reply
			//procurar pj na lista de replies recebidas
			if (!procurar_in_list(pj_id)){
				//se nao tiver coloca na lista de replies recebidas
				replies_received = append(replies_received,pj_id)
			}
			//Se recebeu todas replies
			if (len(replies_received) >= nServers){
				//ativar flag de recebido todas replies
				fmt.Println("Recebi todas replies: ",len(replies_received))
				received_all_replies = true
			}
		} else { fmt.Println("Mensagem desconhecida ",aux)  }
		//update logical clock = max(my,other) + 1
		mylogicalClock = MaxInt(mylogicalClock,lc_pj) + 1
		 fmt.Println("Atualizei meu relogio para ",mylogicalClock)
		 if err != nil {
		 	fmt.Println("Error: ",err)
		 } 
 	}
}

func initConnections() {
	id, _ = strconv.Atoi(os.Args[1])
	myPort = os.Args[ id + 1]
	nServers = len(os.Args) - 2
	/*Esse 2 tira o nome (no caso Process) e tira o id.*/

	connections := make([]*net.UDPConn, nServers, nServers)
	
	for i:=0; i<nServers; i++ {

		port := os.Args[i+2]
		
			ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1" + string (port) )
			PrintError(err)
 
    		LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
 		   	PrintError(err)
 
    		connections[i], err = net.DialUDP("udp",LocalAddr, ServerAddr)
			PrintError(err)
		
	}
	ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1:10001")
		PrintError(err)
    LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
 		PrintError(err) 
	sharedResource , err = net.DialUDP("udp",LocalAddr, ServerAddr)
		PrintError(err)
	
	CliConn = connections

	 /* Lets prepare a address at any address at port 10001*/   
	 ServerAddr,err = net.ResolveUDPAddr("udp", myPort)
	 CheckError(err)
	
	 /* Now listen at selected port */
	 ServConn, err = net.ListenUDP("udp", ServerAddr)
	 CheckError(err)
	 //init process's logical clock with 0
	 mylogicalClock = 0
}

func readInput(ch chan string) {
	// Non-blocking async routine to listen for terminal input
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

	
func Usar_CS(lc_requisicao int, text_simples string){
	//Enviar mensagem e dormir
	//Construir mensagem
	//tranformar logic_clock em string
	estouNaCS = true
	str_lc_requisicao := strconv.Itoa(lc_requisicao)
	//transformar id em string 
	str_id := strconv.Itoa(id)
	// concatenar todas
	mymsg :=  str_id + "," + str_lc_requisicao + "," + text_simples
	//transformar mensagem em bytes
	buf := []byte(mymsg)
	//Enviar mensagem para o Shared Resource
     _,err := sharedResource.Write(buf)
     if err != nil {
        fmt.Println(mymsg, err)
	}
	//Dormir um pouco
	time.Sleep(time.Second*3)
}

func Solicitando_acesso_CS(lclock int){
	//Build menssage
	//Converter relogio logico para string
	str_logical_clock := strconv.Itoa(lclock)
	//transformar id em string 
	str_id := strconv.Itoa(id)
	//E concatenar todas
	mymsg :=  str_id + "," + str_logical_clock + ",request"
	//transformar mensagem em bytes
	buf := []byte(mymsg)
	//Multicast request to all n-1 processes
	//Enviar mensagem para n-1 processes
	for _, conn2process := range CliConn {

     	_,err := conn2process.Write(buf)
     	if err != nil {
        	fmt.Println(mymsg, err)
		}
	}
}

func reply_any_queued_request(){

	//Build menssage
	//Converter relogio logico para string
	str_logical_clock := strconv.Itoa(mylogicalClock)
	//transformar id em string 
	str_id := strconv.Itoa(id)
	//E concatenar todas
	mymsg :=  str_id + "," + str_logical_clock + ",reply"
	//transformar mensagem em bytes
	buf := []byte(mymsg)
	//Reply to all queued processes
	//Enviar mensagem para processos na fila
	for _,id:= range queued_request {

		index := id -1
		//reply queued request
		_,err := CliConn[index].Write(buf)
		if err != nil {
		   fmt.Println(mymsg, err)
	   }
   }
}

func Liberar_CS(){
	//to exit the CS
	//released, reset the flags.
	estouNaCS = false
	estouEsperando = false
	received_all_replies = false
	//reply to any queued request
	reply_any_queued_request()
	//clear reply received list
	replies_received = nil
}

func Ricart_Agrawala(lc_requisicao int, text_simples string){

	estouEsperando = true
	Solicitando_acesso_CS(lc_requisicao)
	//Wait until received N-1 replies
	fmt.Println("Esperando todas replies")
	for !received_all_replies {}
	fmt.Println("Entrei na CS!")
	Usar_CS(lc_requisicao,text_simples)
	fmt.Println("Sai da CS!")
	Liberar_CS()
	fmt.Println("Liberei CS!")
}

func main(){
	initConnections()
	estouNaCS = false
	estouEsperando = false
	//O fechamento de conexões devem ficar aqui, assim só fecha
	//conexão quando a main morrer
	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}
	//Todo Process fará a mesma coisa: ouvir msg e mandar infinitos
	//i’s para os outros processos
	ch := make(chan string)
	go readInput(ch)
	for {
			
		//Server
		go doServerJob()
		// When there is a request (from stdin). Do it!
		select {
			case x, valid := <- ch :
				if valid {
                        compare,_ := strconv.Atoi(x)
						if ( compare != id && x == "x"){
							//Ver se esta na CS ou esperando
							if (estouNaCS || estouEsperando){
								fmt.Println("x ignorado!")
							} else {
								fmt.Printf("Solicitando acesso com ID = %d e Logical Clock = %d\n",id , mylogicalClock)
								text_simples := "CS sugou"
								lc_requisicao = mylogicalClock
								go Ricart_Agrawala(lc_requisicao,text_simples)
							}
							
						} else{

							mylogicalClock++
							fmt.Printf("Atualizado logicalClock para %d \n",mylogicalClock)
						}
				} else {
						 
					fmt.Println("Channel closed!")
						
				}
				
			default:
			
				// Do nothing in the non-blocking approach.
			
				time.Sleep(time.Second * 1)
		}
			
		// Wait a while
		time.Sleep(time.Second * 1)
	}
}
		
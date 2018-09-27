package main
import (
"fmt"
"net"
"strings"
"os"
)

//Variáveis globais interessantes para o processo
var err string
var ServConn *net.UDPConn //conexão do meu servidor (onde recebo
 //mensagens dos outros processos)

 func CheckError(err1 error){
	if err1 != nil {
		fmt.Println("Erro: ", err1)
		os.Exit(0)
	}
}
	
func main () {
	Address, err := net.ResolveUDPAddr("udp", ":10001")
	CheckError(err)
	ServConn, err := net.ListenUDP("udp", Address)
	CheckError(err)
	defer ServConn.Close()
	buf := make([]byte, 1024)
	for {


			n, _, err := ServConn.ReadFromUDP(buf)
			//aux = "id,int_logic_clock,CS sugou"
			aux := string (buf[0:n])
			//stream_msg = ["id","int_logic_clock","CS sugou"]
			stream_msg := strings.Split(aux,",")
			//Imprime na tela o ID, relogio logico e mensagem
			fmt.Printf("\nProcess ID: %s\nLogic Clock: %s\nMenssage Text: %s",stream_msg[0],stream_msg[1],stream_msg[2])
			if err != nil {
				fmt.Println("Error: ",err)
			} 

		
		
	}
}



		
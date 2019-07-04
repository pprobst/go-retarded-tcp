/*
------
Server
------
*/

package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "os/exec"
    "strings"
    "utils"
)

const BUFFER_SIZE = 1024
const PORT = "7005"
const IP = "127.0.0.1" // localhost
const N = 3

func main() {
    fmt.Println("Ouvindo...")
    server, err := net.Listen("tcp", IP+":"+PORT)
    utils.CheckError(err)
    defer server.Close()

    for {
        conn, err := server.Accept()
        utils.CheckError(err)
        go handleConn(conn)
    }
}

func handleConn(conn net.Conn) {
    file, err := os.OpenFile(genFilename()+".jpg", 
    os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    utils.CheckError(err)
    defer file.Close()

    var id_frame int = 0;
    var bad_frames int = 0;
    var lost_frames int = 0;
    var nack bool = false

    recData := make([]string, BUFFER_SIZE) // dados a serem escritos no arquivo

    scanner := bufio.NewScanner(conn)

    fmt.Println("---START---")
    for scanner.Scan() {
        frame_rec := scanner.Text() 
        if string(frame_rec) == "EOF" { 
            break 
        }

        fmt.Print("checa antes\n")
        ack, data := checkFrame(string(frame_rec), id_frame)
        fmt.Print("checa depois\n")
        fmt.Print(id_frame)

        if (ack == 0) { // deu tudo ok
            if id_frame % N == 0 {
                fmt.Print(" - Envia ACK")
                fmt.Fprintf(conn, "ACK:" + frame_rec[0:8] + "\n")
            } else { fmt.Fprintf(conn, "\n") }
        } else if (ack == 1) { // chegou o frame errado -> algum foi perdido
            lost_frames += 1
            nack = true
            fmt.Print(" - Frame perdido")
            if id_frame % N == 0 {
                fmt.Print(" - Erro no frame de ACK; envia ACK com ressalva") 
                fmt.Fprintf(conn, "ACK:" + frame_rec[0:8] + " (RESSALVA)" + "\n")
            } else { fmt.Fprintf(conn, "NACK:" + frame_rec[0:8] +"\n") }
        } else { // chegou frame com os dados corrompidos
            bad_frames += 1 
            nack = true
            fmt.Print(" - Frame malvado")
            if id_frame % N == 0 {
                fmt.Print(" - Erro no frame de ACK; envia ACK com ressalva") 
                fmt.Fprintf(conn, "ACK:" + frame_rec[0:8] + " (RESSALVA)" + "\n")
            } else { fmt.Fprintf(conn, "NACK:" + frame_rec[0:8] +"\n") } 
        }

        data = remByteStuffing(data)
        recData[id_frame] = string(data)

        if nack && id_frame % N == 0 {
            id_frame -= N; // "go back n"
            nack = false
        }

        id_frame += 1 
        fmt.Println()
    }

    fmt.Println("> Frames malvados: ", bad_frames)
    fmt.Println("> Frames perdidos: ", lost_frames)
    fmt.Println("---END---")
    writeToFile(file, recData)
}

// Escreve os dados recebidos no arquivo.
func writeToFile(f *os.File, data []string) {
    for _, s := range(data) {
        f.WriteString(s) 
    }
}

// Retorna true se as strings s1 == s2;
// false caso s1 != s2.
func eqStr(s1, s2 string) bool {
    switch strings.Compare(s1, s2) {
    case 0:
        return true
    default:
        return false
    }
}

// Remove os escapes dos dados recebidos,
// retornando a string de dados original.
func remByteStuffing(s string) (res string) {
    s = strings.ReplaceAll(s, "ESCESC", "ESC")
    res = strings.ReplaceAll(s, "ESC/n", "\n")
    return
}

// Retorna true se os o ID do "frame" do servidor bater com o
// ID do frame enviado pelo cliente.
func checkID(server_id_frame int, client_id_frame string) bool {
    var this_id_frame string = fmt.Sprintf("%08b", server_id_frame)
    return eqStr(this_id_frame, client_id_frame)
}

// Retorna true se o destino do cliente bater com o IP+PORT do servidor.
func checkDest(client_dest string) bool {
    client_dest = client_dest[32:]
    server_ip := utils.Pack32BinaryIP4(IP)
    server_port := utils.Pack16BinaryPort(PORT)
    server_dest := server_ip + server_port
    return eqStr(server_dest, client_dest)
}

// Retorna true se os hashes MD5 dos dados baterem.
func checkChecksum(server_checksum, client_checksum string) bool {
    if utils.CheckMD5Str(server_checksum, client_checksum) {
        return true 
    }
    return false
}

// Faz todas as verificações necessárias para o servidor aceitar 
// o frame do cliente.
func checkFrame(frame string, serv_id_frame int) (ack int, rec_data string) {
    ack = 0
    var frame_id string = frame[0:8]
    var dest string = frame[8:88] // orig + dest + port = 80 "bits"
    var checksum string = frame[len(frame)-32:]
    var data string = frame[88:len(frame)-32]
    rec_data = data

    var server_checksum string = fmt.Sprintf("%x", utils.MD5(data))

    if !checkID(serv_id_frame, frame_id) {
        ack = 1 // código para frame perdido 
        return ack, rec_data
    }

    if !(checkChecksum(server_checksum, checksum) && checkDest(dest)) {
        ack = 2 // código para frame malvado
    }

    return ack, rec_data 
}

// Gera um filename a partir do programa uuidgen.
func genFilename() string {
    out, err := exec.Command("uuidgen").Output()
    utils.CheckError(err)
    return string(out)
}

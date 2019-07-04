/*
------
Client
------
*/

package main

import (
    "bufio"
    "fmt"
    "io"
    "net"
    "os"
    "strings"
    "strconv"
    "utils"
    "time"
)

const BUFFER_SIZE = 1024
const N_SERVER = 3
var g_ip string
var g_port string
var g_this_ip string

func main() {
    // Pega o ip e o port
    if len(os.Args) != 3 {
        fmt.Println("Exemplo de uso: go run client.go 127.0.0.1 7005")
        return
    }

    g_ip = os.Args[1]
    g_port = os.Args[2]
    g_this_ip = "8.8.8.8"

    conn, err := net.Dial("tcp", g_ip+":"+g_port)
    utils.CheckError(err)
    fmt.Println("Discando " + g_ip + ":" + g_port + "...")

    // Leitura do arquivo
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Entre com 'enviar <nome_do_arquivo>' para enviar um arquivo ao servidor:\n")
    inputUser, _ := reader.ReadString('\n')
    arrayCommands := strings.Split(inputUser, " ")

    if arrayCommands[0] == "enviar" {
        sendFile(arrayCommands[1], conn)
    } else {
        fmt.Println("Comando malvado!")
    }
}

// Envia um dado ao servidor.
func sendFile(filename string, conn net.Conn) {
    defer conn.Close()

    file, err := os.Open(strings.TrimSpace(filename))
    utils.CheckError(err)

    defer file.Close()

    r := bufio.NewReader(file)
    buf := make([]byte, BUFFER_SIZE)
    var id_frame int = 0
    var bckpFrames []string
    var shouldResend = false
    var idx int64 = 0

    // lê o arquivo até terminar e vai enviando os dados
    fmt.Println("---START---")
    for {
        n, err := io.ReadFull(r, buf[:cap(buf)])
        buf = buf[:n]
        if err != nil  {
            if err == io.EOF {
                fmt.Println("---END---")
                fmt.Fprintf(conn, "EOF" + "\n")
                break
            }
            if err != io.ErrUnexpectedEOF {
                fmt.Println(err)
                break
            }
        }
        res := byteStuffing(string(buf))
        frame := framing(res, id_frame)

        bckpFrames = append(bckpFrames, frame)

        frame = geraErroArtificial(frame, res, id_frame)

        // envia o frame para o servidor
        conn.Write([]byte(frame))
        fmt.Println("frame", id_frame, "-->", len(string(frame)), " bytes enviados")
        id_frame += 1
        time.Sleep(10 * time.Millisecond)

        // resposta do servidor
        // pode ser "ACK\n" ou "NACK:<id binário>\n" ou simplesmente vazio
        reply, _ := bufio.NewReader(conn).ReadString('\n')
        fmt.Print("     ", reply)

        if reply != "\n" {
            if reply[0:4] == "NACK" {
                shouldResend = true
            } else if reply[0:3] == "ACK" {
                binID := reply[4:12]
                idx, _ = strconv.ParseInt(binID, 2, 64); 
                fmt.Println("     IDX:", idx)
                if shouldResend {
                    resendFrames(bckpFrames, idx-(N_SERVER-1), conn)  
                } 
                shouldResend = false
            }
        }
    }
}

// Gera um frame errado/"perdido" para o servidor pedir o reenvio.
func geraErroArtificial(frame, res string, id_frame int) string {
    if id_frame == 100 || id_frame == 50 || id_frame == 20 || id_frame == 102 {
        frame = strings.ReplaceAll(frame, "a", "Z")
    } else if id_frame == 10 || id_frame == 79 || id_frame == 41 {
        frame = framing(res, id_frame+1) 
    }
    return frame
}

// Reenvia os frames ao servidor a partir do índice idx.
func resendFrames(frames []string, idx int64, conn net.Conn) {
    for i := idx; i < int64(len(frames)); i++ {
        fmt.Println("Reenviando frame ", i, "-->", len(frames[i]), " bytes enviados")
        conn.Write([]byte(frames[i])) 
    }
}

// Realiza o bytestuffing nos dados, substituindo escapes.
func byteStuffing(s string) (res string) {
    s = strings.ReplaceAll(s, "ESC", "ESCESC")
    res = strings.ReplaceAll(s, "\n", "ESC/n")
    return
}

// Realiza a formatação do endereço de destino;
// ip concatenado com o port em binário.
func formatDest(ip, port string) (s string) {
    this_ip := utils.Pack32BinaryIP4(g_this_ip)
    ip = utils.Pack32BinaryIP4(ip)
    port = utils.Pack16BinaryPort(port)
    s = this_ip + ip + port
    return
}

// --------------------- FORMATAÇÃO ---------------------
//
// id_frame | origem destino port | dados | checksum | \n
//    8            32+32+16          ...       32       
//   bin             bin             bin       hex
//
// * Por que id_frame e destino estão em binário?
//   Para manter uma quantidade fixa de caracteres na string.
// * E quanto ao checksum, por que hexadecimal?
//   Hashes MD5, neste caso, terão sempre 32 caracteres hexadecimais (128 bits).
func framing(data string, id_frame int) (frame string) {
    frame = fmt.Sprintf("%08b", id_frame)
    frame += utils.FormatOrigDest(g_ip, g_this_ip, g_port)
    var checksum string = fmt.Sprintf("%x", utils.MD5(data))
    frame += data + checksum + "\n"
    return
}

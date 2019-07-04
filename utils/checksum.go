/*
* checksum.go
* -----------
* Responsável por gerar hashes MD5 e verificar pela
* igualdade entre dois hashes ("checksum").
* Cada hash gerado pelo MD5 possui 128 bits (16 bytes).
*
* Implementado a partir de:
* https://en.wikipedia.org/wiki/MD5#Pseudocode
*
*/

package utils

import (
    "bytes"
    "encoding/binary"
    "fmt"
    //"io/ioutil"
    "math"
    "os"
    "strings"
)

/*
func main() {
    //data, err := ioutil.ReadFile("img.jpg")
    //checkError(err)
    //fmt.Println(string(data))

    //fmt.Println(checksum(string(data)))
    s1 := md5("The quick brown fox jumps over the lazy dog")
    fmt.Printf("Hashcode: %x\n", s1)
    s2 := md5("the quick brown fox jumps over the lazy dog")
    fmt.Printf("Hashcode: %x\n", s2)
    fmt.Println(checkmd5(s1, s2))
}
*/

var shift = [...]uint{7, 12, 17, 22, 5, 9, 14, 20, 4, 11, 16, 23, 6, 10, 15, 21}
var table [64]uint32

func init() {
    for i := range table {
        table[i] = uint32((1 << 32) * math.Abs(math.Sin(float64(i+1))))
    }
}

func MD5(s string) (r [16]byte) {
    padded := bytes.NewBuffer([]byte(s))
    padded.WriteByte(0x80) // 128
    for padded.Len()%64 != 56 {
        padded.WriteByte(0)
    }
    messageLenBits := uint64(len(s)) * 8
    binary.Write(padded, binary.LittleEndian, messageLenBits)

    var a0, b0, c0, d0 uint32 = 0x67452301, 0xEFCDAB89, 0x98BADCFE, 0x10325476
    var buffer [16]uint32

    // lê a cada 64 bytes
    for binary.Read(padded, binary.LittleEndian, buffer[:]) == nil {
        a1, b1, c1, d1 := a0, b0, c0, d0
        for j := 0; j < 64; j++ {
            var f uint32
            bufferIndex := j
            round := j >> 4
            switch round {
            case 0:
                f = (b1 & c1) | (^b1 & d1)
            case 1:
                f = (b1 & d1) | (c1 & ^d1)
                bufferIndex = (bufferIndex*5 + 1) & 0x0F
            case 2:
                f = b1 ^ c1 ^ d1
                bufferIndex = (bufferIndex*3 + 5) & 0x0F
            case 3:
                f = c1 ^ (b1 | ^d1)
                bufferIndex = (bufferIndex * 7) & 0x0F
            }
            sa := shift[(round<<2)|(j&3)]
            a1 += f + buffer[bufferIndex] + table[j]
            a1, d1, c1, b1 = d1, c1, b1, a1<<sa|a1>>(32-sa)+b1
        }
        a0, b0, c0, d0 = a0+a1, b0+b1, c0+c1, d0+d1
    }

    binary.Write(bytes.NewBuffer(r[:0]), binary.LittleEndian, []uint32{a0, b0, c0, d0})
    return
}

func CheckMD5(r1, r2 [16]byte) bool {
    /*
    var equal bool = true
    if fmt.Sprintf("%x", r1) != fmt.Sprintf("%x", r2) {
        equal = false
    }
    return equal
    */
    return CheckMD5Str(fmt.Sprintf("%x", r1), fmt.Sprintf("%x", r2))
}

func CheckMD5Str(r1, r2 string) bool {
    switch strings.Compare(r1, r2) {
    case 0:
        return true
    default:
        return false
    } 
}

func checkError(err error) {
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
        os.Exit(1)
    }
}

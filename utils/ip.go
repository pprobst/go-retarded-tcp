package utils

import (
    "bytes"
    "encoding/binary"
    "fmt"
    "strconv"
    "math/big"
    "net"
)

func IP4toInt(IPv4Address net.IP) int64 {
    IPv4Int := big.NewInt(0)
    IPv4Int.SetBytes(IPv4Address.To4())
    return IPv4Int.Int64()
}

func Pack32BinaryIP4(ip4Address string) string {
    ipv4Decimal := IP4toInt(net.ParseIP(ip4Address))

    buf := new(bytes.Buffer)
    err := binary.Write(buf, binary.BigEndian, uint32(ipv4Decimal))

    if err != nil {
        fmt.Println("Unable to write to buffer:", err)
    }

    // Hex
    result := fmt.Sprintf("%x", buf.Bytes())

    ui, _ := strconv.ParseUint(result, 16, 64) 

    var resBin string = fmt.Sprintf("%032b", ui)
    return resBin
    //return result
    //return buf.Bytes()
}

func Pack16BinaryPort(port string) string {
    intPort, _ := strconv.Atoi(port)
    port = fmt.Sprintf("%016b", intPort)
    return port
}

func FormatOrigDest(ip, ip2, port string) (s string) {
    this_ip := Pack32BinaryIP4(ip2)
    ip = Pack32BinaryIP4(ip)
    port = Pack16BinaryPort(port)
    s = this_ip + ip + port
    return
}

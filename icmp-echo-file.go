package main

import (
    "bufio"
    "flag"
    "fmt"
    "io"
    "net"
    "os"
    "time"
)

func sendPacket(conn net.PacketConn, destination *net.IPAddr, payload []byte, seq int) {
    packet := make([]byte, len(payload)+8)
    packet[0] = 8
    packet[1] = 0
    packet[4] = uint8(os.Getpid() >> 8)
    packet[5] = uint8(os.Getpid() & 0xff)
    packet[6] = uint8(seq >> 8)
    packet[7] = uint8(seq & 0xff)
    copy(packet[8:], payload)

    cklen := len(packet)
    cksum := uint32(0)
    for i := 0; i < cklen-1; i += 2 {
        cksum += uint32(packet[i+1])<<8 | uint32(packet[i])
    }
    if cklen&1 == 1 {
        cksum += uint32(packet[cklen-1])
    }
    cksum = (cksum >> 16) + (cksum & 0xffff)
    cksum = cksum + (cksum >> 16)

    packet[2] ^= uint8(^cksum & 0xff)
    packet[3] ^= uint8(^cksum >> 8)

    _, err := conn.WriteTo(packet, destination)
    if err != nil { panic(err) }
}

func findLocalAddress(def string) (string) {
    if def != "" { return def }

    localAddr := ""
    if interfaces, err := net.Interfaces(); err == nil {
        for _, iface := range interfaces {
            if iface.Flags&net.FlagLoopback == 0 && iface.Flags&net.FlagPointToPoint == 0 {
                if addrs, err := iface.Addrs(); err == nil && len(addrs) > 0 {
                    for _, addr := range addrs {
                        if ip, _, err := net.ParseCIDR(addr.String()); err == nil && ip.To4() != nil {
                            if localAddr != "" { panic("unable to determine which interface to use; specify source address with -l") }
                            localAddr = ip.String()
                        }
                    }
                }
            }
        }
    }
    if localAddr == "" { panic("unable to find an interface to use; specify source address with -l") }

    return localAddr
}

func main() {
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [-i <num>] [-s <num>] [-l <address>] <destination host> <file>\n", os.Args[0])
        flag.PrintDefaults()
    }

    defer func() {
        if r := recover(); r != nil {
            fmt.Printf("Error: %s\n", r)
        }
    }()

    interval := flag.Int("i", 500, "Interval between pings in milliseconds")
    laddr := flag.String("l", "", "Source IP address")
    psize := flag.Int("s", 128, "Size of ICMP packet data section in bytes")

    flag.Parse()
    if flag.NArg() != 2 { panic("invalid number of arguments") }

    destination := flag.Arg(0)
    filename := flag.Arg(1)

    raddr, err := net.ResolveIPAddr("ip4", destination)
    if err != nil { panic(err) }

    file, err := os.Open(filename)
    if err != nil { panic(err) }
    defer file.Close()
    reader := bufio.NewReader(file)

    conn, err := net.ListenPacket("ip4:icmp", findLocalAddress(*laddr))
    if err != nil { panic(err) }
    defer conn.Close()

    seq := 0
    buf := make([]byte, *psize)
    limiter := time.Tick(time.Millisecond * time.Duration(*interval))
    for {
        n, err := reader.Read(buf)
        if err != nil && err != io.EOF { panic(err) }
        if n == 0 { break }

        sendPacket(conn, raddr, buf, seq)
        seq++
        <-limiter
    }
}

package main

import (
	"bufio"
	"io"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/armon/go-socks5"
)

type myc struct {
	server_addr      string
	server_port      string
	local_ip         string
	local_port       string
	remote_port      string
	token            string
	socks5_port      string
	user_socks5_port string
}

var data = &myc{
	server_addr: "192.168.31.72",
	server_port: "7000",
	local_ip:    "127.0.0.1",
	local_port:  "22",
	remote_port: "6000",
	token:       "12345678",
	socks5_port: "5000",
}

func start() {

	ip_port := data.server_addr + ":" + data.server_port
	conn, err := net.Dial("tcp", ip_port)
	if err != nil {
		println("dial error")
		println("please check your server")
		time.Sleep(time.Second * 5)
		return
	}

	send_content := []byte(
		"token = " + data.token +
			"\nport = " + data.remote_port +
			"\nsocks5_port = " + data.socks5_port +
			"\nuser_socks5_port = " + data.user_socks5_port +
			"\n")
	_, err = conn.Write(send_content)
	if err != nil {
		println("send port error")
	}
	// you can add check way here
	// reader := bufio.NewReader(conn)
	// b, err := reader.ReadBytes(byte('\n'))
	// if err != nil {
	// 	println("read error")
	// 	start()
	// 	return
	// }
	// print("recv : ", string(b))
	// if string(b) != "ok\n" {
	// 	start()
	// 	return
	// }
	defer conn.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go sock5_foward()
	go port_foward()
	wg.Wait()

}
func port_foward() {

	for {
		remote, err := net.Dial("tcp", data.server_addr+":"+data.server_port)
		if err != nil {
			// println("wating for user")
			time.Sleep(time.Second * 1)
			continue
		}
		local, err := net.Dial("tcp", data.local_ip+":"+data.local_port)
		if err != nil {
			continue
		}
		go forward(remote, local)
	}
}
func sock5_foward() {
	var server *socks5.Server
	var err error
	server, err = socks5.New(&socks5.Config{})
	if err != nil {
		panic(err)
	}
	for {
		conn, err := net.Dial("tcp", data.server_addr+":"+data.socks5_port)
		if err != nil {
			// println("wating for user socks5")
			time.Sleep(time.Second * 1)
			continue
		}
		defer conn.Close()
		println("socks5 working")
		_ = conn.SetDeadline(time.Time{})
		// 使用该 socks5 库提供的 ServeConn 方法
		err = server.ServeConn(conn)
		if err != nil {
			println(err)
		}
	}

}

func forward(srcConn, dstConn net.Conn) {
	var wg sync.WaitGroup
	wg.Add(2)

	println("start foward")
	go func() {
		defer wg.Done()
		defer srcConn.Close()
		defer dstConn.Close()
		_, err := io.Copy(srcConn, dstConn)
		if err != nil {
			return
		}

	}()
	go func() {
		defer wg.Done()
		defer dstConn.Close()
		defer srcConn.Close()
		_, err := io.Copy(dstConn, srcConn)
		if err != nil {
			return
		}
	}()
	wg.Wait()
}

func main() {
	ini()
	start()
}
func ini() {
	ini, err := os.Open("./myc.ini")
	if err != nil {
		panic("open error")
	}
	defer ini.Close()
	scanner := bufio.NewScanner(ini)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case "server_addr":
			data.server_addr = value
		case "server_port":
			data.server_port = value
		case "local_ip":
			data.local_ip = value
		case "remote_port":
			data.remote_port = value
		case "local_port":
			data.local_port = value
		case "token":
			data.token = value
		case "socks5_port":
			data.socks5_port = value
		case "user_socks5_port":
			data.user_socks5_port = value
		}
	}
	if err := scanner.Err(); err != nil {
		panic("scan error")
	}
	println("server_addr=", data.server_addr)
	println("server_port=", data.server_port)

	println("remote_port=", data.remote_port)
	println("local_ip   =", data.local_ip)
	println("local_port =", data.local_port)

	println("socks5_port=", data.socks5_port)
	println("user_socks5_port=", data.user_socks5_port)
}

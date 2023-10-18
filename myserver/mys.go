package main

import (
	"bufio"
	"errors"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

type file_ini struct {
	server_addr string
	bind_port   string
	user_port   string
	token       string
	bind1_port  string
	user1_port  string
}

var data = &file_ini{
	bind_port: "7000",
	token:     "12345678",
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	// 遍历所有网络接口，找到一个有效的IPv4或IPv6地址
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil // 返回IPv4地址
			}
			return ipnet.IP.String(), errors.New("无法获取本机服务器IPv4地址") // 返回IPv6地址
		}
	}
	return "", errors.New("无法获取本机服务器IP地址")
}
func start() {
	server_addr, err := getLocalIP()
	if err != nil {
		panic("get ip error")
	}
	data.server_addr = server_addr

	listener_bind, err := net.Listen("tcp", data.server_addr+":"+data.bind_port)
	if err != nil {
		println("listen error", err)
		return
	}

	conn, err := listener_bind.Accept()
	if err != nil {
		println("accept error", err)
	}
	defer conn.Close()
	for {
		if recv_port(conn) {
			break
		}
	}
	// conn.Write([]byte("ok\n"))
	println("check over")
	listener_bind.Close()
	var wg sync.WaitGroup
	wg.Add(2)
	go port_foward(data.user_port, data.bind_port)
	go port_foward(data.user1_port, data.bind1_port)
	wg.Wait()
}
func port_foward(user_port, bind_port string) {
	println("listening port:" + user_port)
	listener_user, err := net.Listen("tcp", data.server_addr+":"+user_port)
	if err != nil {
		panic("user port error,please check port " + user_port)
	}
	defer listener_user.Close()
	for {
		conn_u, err := listener_user.Accept()
		if err != nil {
			continue
		}
		listener_bind, err := net.Listen("tcp", data.server_addr+":"+bind_port)
		if err != nil {
			panic("bind port error,please check port " + bind_port)
		}
		conn, err := listener_bind.Accept()
		if err != nil {
			continue
		}
		go forward(conn, conn_u)
		listener_bind.Close()
	}

}
func port_foward_old(listener2, listener net.Listener) {
	defer listener.Close()
	defer listener2.Close()
	for {
		conn_u, err := listener2.Accept()
		if err != nil {
			continue
		}
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go forward(conn, conn_u)
	}

}
func forward(srcConn, dstConn net.Conn) {
	println("start forward")
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer srcConn.Close()
		defer dstConn.Close()
		_, err := io.Copy(srcConn, dstConn)
		if err != nil {
			return
		}
		println("discon")
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
func recv_port(conn net.Conn) bool {
	reader := bufio.NewReader(conn)

	t := check_string(reader, "token")
	if len(t) != 0 && data.token == t {
		port := check_string(reader, "port")
		if port != "" {
			data.user_port = port
			socks5_port := check_string(reader, "socks5_port")
			if socks5_port != "" {
				data.bind1_port = socks5_port
				user_port := check_string(reader, "user_socks5_port")
				if user_port != "" {
					data.user1_port = user_port
					return true
				}
			}
		}
	}
	return false
}
func check_string(reader *bufio.Reader, check string) string {
	b, _ := reader.ReadBytes(byte('\n'))
	println("do check:", string(b))
	parts := strings.Split(string(b), "=")
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == check {
		return value
	}
	return ""
}
func main() {

	ini()
	for {
		start()
		println("rechecking .....")
	}
}

func ini() {
	ini, err := os.Open("./mys.ini")
	if err != nil {
		panic("open error")
	}
	defer ini.Close()
	scanner := bufio.NewScanner(ini)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if len(parts) != 2 {
			continue
		}
		switch key {
		case "bind_port":
			data.bind_port = value
		case "token":
			data.token = value
		}
	}
	if err := scanner.Err(); err != nil {
		panic("scan error")
	}
	println("bind_port =", data.bind_port)
	println("token = ", data.token)
}

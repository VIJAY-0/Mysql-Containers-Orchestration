package LoadBalancer

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
)

type Config struct {
	SlaveDSNs []string
	Port      string
}

type LoadBalancer struct {
	SchedulingAlgorithm string
	SlaveDSNs           *[]string
	MasterDSN           *string
	Slaves              []*sql.DB
	mu                  sync.Mutex
	index               int
}

func NewLoadBalancer(dsn []string) (*LoadBalancer, error) {

	var slaves []*sql.DB

	for _, d := range dsn {
		db, err := sql.Open("mysql", d)
		if err != nil {
			return nil, fmt.Errorf("unable to connect to slave :%v", err)
		}
		fmt.Printf("connected to slave :%s\n", d)
		slaves = append(slaves, db)
	}

	return &LoadBalancer{Slaves: slaves}, nil

}

func (lb *LoadBalancer) getNextSlave() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	slave := (*lb.SlaveDSNs)[lb.index]
	lb.index = (lb.index + 1) % len(*lb.SlaveDSNs)
	return slave
}

func isWriteQuery(query string) bool {
	query = strings.ToUpper(query)

	return strings.HasPrefix(query, "INSERT") ||
		strings.HasPrefix(query, "UPDATE") ||
		strings.HasPrefix(query, "DELETE") ||
		strings.HasPrefix(query, "CREATE") ||
		strings.HasPrefix(query, "DROP") ||
		strings.HasPrefix(query, "REPLACE") ||
		strings.HasPrefix(query, "ALTER")
}

func (lb *LoadBalancer) HandleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	var data []byte

	// buffer := make([]byte, 1024)
	// for {
	// 	n, err := clientConn.Read(buffer)
	// 	if err != nil {
	// 		if err == io.EOF {
	// 			break
	// 		}
	// 		log.Printf("error reading form client :%v", err)
	// 		return
	// 	}
	// 	data = append(data, buffer[:n]...)

	// 	if n < len(buffer) {
	// 		break
	// 	}
	// }

	query := string(data)
	// query := "Create"
	var server string

	if isWriteQuery(query) {
		log.Printf("Forwading to Master: %v\n", query)
		server = *lb.MasterDSN
	} else {
		log.Printf("Forwading to Slave: %v\n", query)
		server = lb.getNextSlave()
	}

	slaveConn, err := net.Dial("tcp", server)
	if err != nil {
		log.Printf("Failed to connect to slave: %v\n", err)
		return
	}
	defer slaveConn.Close()

	log.Printf("connected to : %v\n", server)

	// _, err = slaveConn.Write(data)

	if err != nil {
		log.Printf("Failed to forward query to server:%v", err)
		return
	}

	log.Println("Sucesfully forwarded query to server ")
	log.Println("Communication started")

	go io.Copy(slaveConn, clientConn)
	io.Copy(clientConn, slaveConn)

}

// conn = mysql.connector.connect(
//     host="localhost",
//     user="root",
//     password="rootpassword",
//     port=3305
// )

// cursor.execute("SHOW VARIABLES LIKE 'server_id';")
// for x in cursor:
//     print(x)

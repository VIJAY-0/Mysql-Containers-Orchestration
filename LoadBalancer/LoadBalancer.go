package LoadBalancer

import (
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type Config struct {
	SlaveDSNs []string
	Port      string
}

type LoadBalancer struct {
	SchedulingAlgorithm string
	SlaveDSNs           *[]string
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

func (lb *LoadBalancer) HandleConnection(clientConn net.Conn) {
	defer clientConn.Close()

	slave := lb.getNextSlave()

	slaveConn, err := net.Dial("tcp", slave)
	if err != nil {
		log.Printf("Failed to connect to slave: %v\n", err)
		return
	}
	defer slaveConn.Close()

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

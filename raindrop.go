package main

import (
	log "github.com/sirupsen/logrus"
	"flag"
	"sync"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"gopkg.in/go-playground/validator.v9"
	"time"
)

const (
	nodeBits uint8 = 5  // 2^5=32
	typeBits uint8 = 6  // 2^6=64
	stepBits uint8 = 11 // 2^11=2048

	nodeMax uint8 = -1 ^ (-1 << nodeBits)
	typeMax uint8 = -1 ^ (-1 << typeBits)

	stepMask int64 = -1 ^ (-1 << stepBits)

	typeShift = stepBits
	nodeShift = typeShift + stepBits
	timeShift = nodeShift + nodeBits
)

var (
	nodeIndex int64

	// 2018-02-22 16:44:54.639553 +0800 CST m=+0.000783951
	Epoch int64 = 1519289094639

	Validate   *validator.Validate
	ServerNode *Node
)

type ID int64

func setupLog() error {
	// TODO Extract the variable to configuration file (config.yaml).
	level := "debug"
	parseLevel, e := log.ParseLevel(level)
	if e != nil {
		log.Fatalf("parse level string failed. error: %v", e)
	}
	log.SetLevel(parseLevel)

	formatter := &log.TextFormatter{
		// Never try to modify the timestamp below
		// Easy way to remember: 1 2 3 4 5 6
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	return nil
}

type Node struct {
	mutex sync.Mutex
	node  int64
	typ   int64
	step  int64
	time  int64
}

func init() {
	setupLog()

	Validate = validator.New()

	// init node index from command argument
	nodeIdx := flag.Int("node", 1, "node index")
	flag.Parse()

	log.Infof("Raindrop nodeIndex=%d", *nodeIdx)
	if uint8(*nodeIdx) > nodeMax || *nodeIdx < 1 {
		log.Fatal("Input node index value illegal. node: [1-31]")
	}
	nodeIndex = int64(*nodeIdx)
	ServerNode = newNode(nodeIndex)
}

func main() {
	r := gin.Default()
	r.GET("/ticking", doTicking)
	r.Run()
}

type TickType struct {
	Typ int64 `form:"t" binding:"required" validate:"min=1,max=64"`
}

func newNode(node int64) *Node {
	return &Node{node: node, typ: 0, time: 0, step: 0}
}

func doTicking(c *gin.Context) {
	var tt TickType
	if err := c.BindQuery(&tt); err == nil {
		e := Validate.Struct(tt)
		if e != nil {
			log.Error(e)
			c.JSON(http.StatusBadRequest,
				gin.H{"status": "ERROR",
					"message": err.Error(),
				})
			return
		}
		id, err := ServerNode.ticking(tt.Typ)
		if err != nil {
			log.Error(err)
			c.JSON(http.StatusInternalServerError,
				gin.H{"status": "ERROR",
					"message": err.Error(),
				})
		}
		log.Infof("node=%v", *ServerNode)

		c.JSON(http.StatusOK,
			gin.H{"status": "OK",
				"id": id,
				"type": tt.Typ})
		return
	} else {
		log.Error(err)
		c.JSON(http.StatusBadRequest,
			gin.H{"status": "ERROR",
				"message": err.Error(),
			})
		return
	}
}

func (n *Node) ticking(typ int64) (ID, error) {
	if uint8(typ) > typeMax || typ < 1 {
		log.Errorf("type=%d invalid.", typ)
		return -1, errors.New("The parameter: type, invalid.")
	}

	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.typ = typ
	now := time.Now().UnixNano() / 1000000

	if n.time == now {
		n.step = (n.step + 1) & stepMask
		if n.step == 0 {
			for now <= n.time {
				now = time.Now().UnixNano() / 1000000
			}
		}
	} else {
		n.step = 0
	}

	n.time = now

	id := ID((now-Epoch)<<timeShift |
		(n.node << nodeShift) |
		(n.typ << typeShift) |
		(n.step),
	)
	return id, nil
}

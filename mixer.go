package mixer

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/bluele/gcache"
	"github.com/ilooky/logger"
	"time"
)

type Config struct {
	Handler IHandler
}

type IHandler interface {
	Send(data []byte) error

	Close()
}

type Message struct {
	Id    rune
	packs []*pack
}

type pack struct {
	id    rune
	index string
	flag  rune
	data  []byte
}

func (pck *pack) Data() []byte {
	var data bytes.Buffer
	data.WriteRune(pck.id)
	data.WriteString(pck.index)
	data.Write(pck.data)
	data.WriteRune(pck.flag)
	return data.Bytes()
}

const (
	PckHead = 4
	PckSize = 10
)

type mixer struct {
	Config
	Queue    chan *pack
	msgQueue chan *Message
	cache    gcache.Cache
	maxId    rune
}

func NewMixer(cfg Config) *mixer {
	return &mixer{
		Config:   cfg,
		Queue:    make(chan *pack),
		msgQueue: make(chan *Message),
		cache:    gcache.New(26).LRU().Expiration(time.Minute * 3).Build(),
		maxId:    'A'}
}

func (mixer *mixer) Merge(data []byte) (result []byte, err error) {
	fmt.Println(string(data))
	pck := data[:4]
	fmt.Println(string(pck))
	return nil, nil
}

//min:1
func (mixer *mixer) getId() rune {
	if mixer.maxId == 'Z' {
		return 'A'
	}
	defer func() {
		mixer.maxId = mixer.maxId + 1
	}()
	return mixer.maxId
}

//all data
//flag:   1位：0 | 1
//Id: 	  1位：26
//index:  2位：702
//data: 64 - 1 - 1 - 2 = 60
// 702*60/1024=41.1328125
func (mixer *mixer) Submit(data []byte) error {
	var msg Message
	msg.Id = mixer.getId()
	index := ""
	for len(data)+PckHead > PckSize {
		index = GetIndex(index)
		if pack, err := mixer.toPackage(msg.Id, index, data[:PckSize-PckHead], '0'); err != nil {
			return err
		} else {
			data = data[PckSize-PckHead:]
			msg.packs = append(msg.packs, pack)
		}
	}
	pack, err := mixer.toPackage(msg.Id, GetIndex(index), data, '1')
	if err != nil {
		return err
	}
	msg.packs = append(msg.packs, pack)
	err = mixer.saveSendCache(msg.Id, msg)
	if err != nil {
		return err
	}
	mixer.msgQueue <- &msg
	return nil
}

func (mixer *mixer) toPackage(id rune, index string, data []byte, flag rune) (*pack, error) {
	var pck pack
	pck.id = id
	pck.index = index
	pck.flag = flag
	pck.data = data
	return &pck, nil
}

func (mixer *mixer) send(data []byte) {
	go func() {
		err := mixer.Handler.Send(data)
		if err != nil {
			fmt.Println("发送数据失败，err:", err)
		}
	}()
}

func (mixer *mixer) saveSendCache(key rune, value interface{}) error {
	return mixer.cache.Set("send_"+string(key), value)
}

const recv = "recv_"

func (mixer *mixer) SaveRecv(id, index, flag string, data []byte) error {
	msg, err := mixer.cache.Get(recv + id)
	var m Message
	noExist := errors.Is(err, gcache.KeyNotFoundError)
	if err != nil && !noExist {
		return err
	}
	if err == nil {
		m = msg.(Message)
	}
	msgId := []rune(id)[0]
	if noExist {
		m = Message{
			Id: msgId,
		}
	}
	pck, err := mixer.toPackage(msgId, index, data, []rune(flag)[0])
	if err != nil {
		return err
	}
	m.packs = append(m.packs, pck)
	return mixer.cache.Set(recv+id, m)
}

func (mixer *mixer) GetRecvById(id string, lastIndex string) (string, error) {
	msg, err := mixer.cache.Get(recv + id)
	if err != nil {
		return "", err
	}
	m := msg.(Message)
	lack := m.lack(lastIndex)
	if len(lack) == 0 {
		var data bytes.Buffer
		for i := range m.packs {
			data.Write(m.packs[i].data)
		}
		return string(data.Bytes()), nil
	} else {
		return "", errors.New("the data is incomplete")
	}
}

//缺少的包
func (m *Message) lack(lastIndex string) []string {
	size := len(m.packs)
	last := ToInt(lastIndex)
	if size == last {
		return []string{}
	}
	var recvIds = make(map[int]string)
	for _, pack := range m.packs {
		recvIds[ToInt(pack.index)] = pack.index
	}
	var lackIds []string
	for i := 1; i < last+1; i++ {
		v, ok := recvIds[i]
		if !ok {
			lackIds = append(lackIds, v)
		}
	}
	return lackIds
}

func (mixer *mixer) Run() {
	go func() {
		for {
			select {
			case pck := <-mixer.Queue:
				go mixer.send(pck.Data())
			default:
			}
		}
	}()
	for {
		select {
		case msg := <-mixer.msgQueue:
			for i := range msg.packs {
				mixer.Queue <- msg.packs[i]
			}
		default:
		}
	}
}

func (mixer *mixer) Close() {
	mixer.cache.Purge()
	mixer.Handler.Close()
	logger.Info("释放资源")
}

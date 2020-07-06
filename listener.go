package proc

import (
	"fmt"
	"github.com/aberic/gnomon"
	"github.com/aberic/gnomon/log"
	"io/ioutil"
	"time"
)

var (
	proc         *Proc
	remote, host string
	scheduled    *time.Timer // 超时检查对象
	delay        time.Duration
	stop         chan struct{} // 释放当前角色chan
)

func init() {
	proc = &Proc{}
	timeDistance := gnomon.EnvGetInt64D(timeDistanceEnv, 1500)
	delay = time.Millisecond * time.Duration(timeDistance)
	remote = gnomon.EnvGet(listenAddr)
	host = gnomon.EnvGet(hostname)
	fmt.Println("remote: ", remote)
	scheduled = time.NewTimer(delay)
	stop = make(chan struct{}, 1)
}

// ListenStart 开启监听发送
func ListenStart() {
	log.Debug("listener start", log.Server("proc"), log.Field("remote", remote))
	if gnomon.StringIsNotEmpty(remote) {
		go send()
	}
}

func send() {
	scheduled.Reset(time.Millisecond * time.Duration(5))
	for {
		select {
		case <-scheduled.C:
			if err := proc.run(); nil == err {
				log.Debug("send", log.Server("proc"), log.Field("proc", proc))
				if _, err := gnomon.HTTPPostJSON(remote, proc); nil != err {
					log.Error("send", log.Err(err))
				}
			} else {
				log.Error("send", log.Err(err))
			}
			scheduled.Reset(delay)
		case <-stop:
			return
		}
	}
}

// Proc 监听发送完整对象
type Proc struct {
	Hostname string
	CPUGroup *CPUGroup
	MemInfo  *MemInfo
	LoadAvg  *LoadAvg
	//Swaps    *Swaps
	Version *Version
	Stat    *Stat
	//CGroup   *CGroup
	UsageCPU  float64
	Mounts    *Mounts
	Disk      *Disk
	DiskStats *DiskStats
	SockStat  *SockStat
}

func (p *Proc) run() error {
	if err := obtainCPUGroup().Info(); nil == err {
		p.CPUGroup = obtainCPUGroup()
	}
	if err := obtainMemInfo().Info(); nil == err {
		p.MemInfo = obtainMemInfo()
	}
	if err := obtainLoadAvg().Info(); nil == err {
		p.LoadAvg = obtainLoadAvg()
	}
	//swaps := &Swaps{}
	//if err := swaps.Info(); nil == err {
	//	p.Swaps = swaps
	//}
	if err := obtainVersion().Info(); nil == err {
		p.Version = obtainVersion()
	}
	if err := obtainStat().Info(); nil == err {
		p.Stat = obtainStat()
	}
	//cGroup := &CGroup{}
	//if err := cGroup.Info(); nil == err {
	//	p.CGroup = cGroup
	//}
	if usage, err := UsageCPU(); nil == err {
		p.UsageCPU = usage
	}
	if err := obtainMounts().Info(); nil == err {
		p.Mounts = obtainMounts()
	}
	if err := obtainDisk().Info(); nil != err {
		p.Disk = obtainDisk()
	}
	if err := obtainDiskStats().Info(); nil != err {
		p.DiskStats = obtainDiskStats()
	}
	if err := obtainSockStat().Info(); nil != err {
		p.SockStat = obtainSockStat()
	}
	bs, err := ioutil.ReadFile(host)
	if nil != err {
		return err
	}
	p.Hostname = gnomon.StringTrim(string(bs))
	return nil
}

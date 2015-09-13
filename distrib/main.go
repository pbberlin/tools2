// Package distrib enables distributed processing of
// any slice of structs implementing the Worker
// interface.
package distrib

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync/atomic"
	"time"
)

var lnp = log.New(os.Stdout, "", 0) // logger no prefix
var lpf = lnp.Printf                // shortcut

type Worker interface {
	Work()
}

// Packet is a struct, that is passed around between stages.
// Packet.Work() is evoked during stage2.
type Packet struct {
	TaskID   int
	WorkerID int
	Worker   // Anonymous interface; any struct that has a Work() method
}

func (w Packet) String() string {
	return fmt.Sprintf(" packet#%-2v  Wkr%v", w.TaskID, w.WorkerID)
}

type Options struct {
	NumWorkers       int           //
	Want             int32         // Maximum results before pipeline is torn down
	TimeOutDur       time.Duration // Max duration of Work() before timeout + abandonment.
	CollectRemainder bool          // Upon exit: Wait for remaining packets to reach stage3, or flush them where they are.
	TailingSleep     bool          // Upon exit: Wait a short while, checking the return of all goroutines.
}

var DefaultOptions = Options{
	NumWorkers:       6,
	Want:             int32(10),
	TimeOutDur:       time.Millisecond * 15,
	CollectRemainder: true,
	TailingSleep:     false,
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Distrib contains the simplest pipeline with three stages.
//
// All send and receive operations are clad in
// select statements with a <-fin branch,
// ensuring clear exit of all goroutines.
//
// Distrib processes packages of type WLoad.
// The actual work is always contained in Packet.Work().
//
// Distrib has two exit modes.
// CollectRemainder or inversely "FlushAndAbandon".
//
// FlushAndAbandon was supposed to set inn and out to nil,
// blocking every further communication.
// It then closes fin, causing all goroutines to exit.
// Setting inn and out to nil alarmed the race detector.
// Instead we signal stage1 to stop with "lcnt = -10"
//
// CollectRemainder equally prevents further feeding of the pipeline.
// It then waits until all sent packets
// have been received at stage3,
// relying on the synchronized sent counter.
//
// Distrib provides a timeout for Packet.Work().
// Timeout for workWrap requires spawning of a goroutine
// and an individual receiver channel for each packet.
// Upon timeout the channel receiving the work is
// closed.
// The timed out workers will hang around in blocked mode
// until they are flushed with close(fin) at exit.
//
// Stage3 of Distrib is a *synchroneous* loop,
// delivering us from waitgroup.Wait...
//
// Signalling stage1 to stage3 that packets are exhausted
// happens by setting Options.Want to zero.
//
// Signalling stage3 to stage1 that loading should be
// stopped, happens by setting lcnt to -10.
// Both need to be synchronized.
//
// Todo/Caveat: CollectRemainder==false && len(jobs) << Want
// leads to premature flushing
//
func Distrib(jobs []Worker, opt Options) []*Packet {

	inn := make(chan *Packet) // stage1 => stage2
	out := make(chan *Packet) // stage2 => stage3

	fin := make(chan struct{}) // flush all

	lcnt := int32(0) // load counter; always incrementing; except: to signal stop loading from downstream
	sent := int32(0) // sent packages - might get decremented for timed out work
	recv := int32(0) // received packages - converges against sent packages - unless in flushing mode

	ticker := time.NewTicker(10 * time.Millisecond)
	tick := ticker.C // channel from ticker

	var returns = make([]*Packet, 0, int(opt.Want)+5)

	packets := make([]*Packet, 0, len(jobs))
	for i, job := range jobs {
		pack := &Packet{}
		pack.Worker = job
		pack.TaskID = i
		packets = append(packets, pack)
	}

	//
	// stage 1
	go func() {
		for {

			idx := int(atomic.LoadInt32(&lcnt))

			if idx > len(packets)-1 {
				lpf("=== input packets exhausted ===")
				atomic.StoreInt32(&opt.Want, int32(0))
				return
			}
			if idx < 0 { // signal from downstream
				lpf("=== loading stage 1 terminated ===")
				return
			}
			select {
			case inn <- packets[idx]:
				atomic.AddInt32(&lcnt, 1)
				atomic.AddInt32(&sent, 1) // sent++, sent can be decremented later on, therefore distinct var "lcnt"
			case <-fin:
				return
			}
		}
	}()

	//
	// stage 2
	for i := 0; i < opt.NumWorkers; i++ {
		go func(wkrID int) {
			for {
				timeout := time.After(opt.TimeOutDur)
			MarkX:
				select {
				case packet := <-inn:

					// Wrap work into a go routine.
					// It puts the result into chan res.
					res := make(chan *Packet)
					workWrap := func(pck *Packet) { // packet given as argument to prevent race cond race_1_b
						pck.WorkerID = wkrID // signature
						pck.Work()
						select {
						case res <- pck:
						case <-fin:
						}
					}
					go workWrap(packet)

					//
					// Now put workWrap() in competition with timeout.
					select {
					case completed := <-res:
						select {
						case out <- completed:
						case <-fin:
							return
						}
					case <-timeout:
						atomic.AddInt32(&sent, -1)
						lpf("TOUT  snt%2v  %v", atomic.LoadInt32(&sent), *packet) // race_1_b
						// => stage 3 has to check recv >= sent in separate select-tick branch
						//    because no WLoad packet is sent-received upon this timeout.
						break MarkX // skipping this packet
					}

				case <-fin:
					return
				}
			}
		}(i)
	}

	//
	// stage 3
	// synchroneous
	func() {
		for {
			select {

			case <-tick:
				// Exit after collecting remainder
				// Sent might be decremented in timed out works
				if opt.CollectRemainder &&
					recv >= atomic.LoadInt32(&opt.Want) &&
					recv >= atomic.LoadInt32(&sent) &&
					true {
					return
				}

			case packet := <-out:

				recv++

				lpf("rcv%-2v of %-2v snt%-2v  %v", recv, atomic.LoadInt32(&opt.Want), atomic.LoadInt32(&sent), packet)
				// lpf("  %v", stringspb.IndentedDump(packet.Workload))

				returns = append(returns, packet)

				if recv >= atomic.LoadInt32(&opt.Want) {

					if recv == atomic.LoadInt32(&opt.Want) {
						lpf("=== enough ===")
					}

					// inn = nil  // race detector objected to this line
					atomic.StoreInt32(&lcnt, -10) // signalling stop loading to stage 1

					// Exit immediately
					if !opt.CollectRemainder {
						lpf("=== flush all stages and abandon remaining results ===")
						// out = nil // race detector objected to this line
						close(fin)
						if opt.TailingSleep {
							time.Sleep(60 * time.Millisecond) // few messages from exiting goroutines might occur
						}
						return
					}

				}

			case <-fin:
				return
			}
		}
	}()

	lpf("=== cleanup ===")

	ticker.Stop() // cleaning up the ticker really necessary?

	if opt.CollectRemainder {
		close(fin) // flush remaining packets whereever they are
		if opt.TailingSleep {
			time.Sleep(60 * time.Millisecond) // 1 or 2 messages from timed out workWrap might occur
		}
	}

	return returns

}

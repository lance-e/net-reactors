package netreactors

import (
	"log"

	"golang.org/x/sys/unix"
)

type (
	// ChannelList []*Channel
	pollFdList []unix.PollFd
	channelMap map[int32]*Channel
)
type Poller struct {
	ownerLoop_ *EventLoop //event loop owner
	pollfds_   pollFdList //cathe of pollfd
	channels_  channelMap //map of fd and Channel
}

// *************************
// public:
// *************************

func NewPoller(loop *EventLoop) *Poller {
	return &Poller{
		ownerLoop_: loop,
		pollfds_:   make(pollFdList, 0),
		channels_:  make(channelMap),
	}
}

func (p *Poller) Poll(timeoutMs int, activeChannels *[]*Channel) {
	n, err := unix.Poll(p.pollfds_, timeoutMs)
	if err != nil || n < 0 {
		log.Panicf("Poller.Poll failed ,n:%d , err:%s \n", n, err.Error())
	}
	if n > 0 {
		log.Printf("%d events happended\n", n)
		p.fillActiveChannels(n, activeChannels)
	} else if n == 0 {
		log.Printf("nothing happended\n")
	}
}

func (p *Poller) UpdateChannel(channel *Channel) {
	p.AssertInLoopGoroutine()
	log.Printf("UpdateChannel: fd=%d , events=%d\n", channel.fd_, channel.events_)
	if channel.index_ < 0 {
		// a new one , add to p.pollfds_
		if _, ok := p.channels_[channel.fd_]; ok {
			log.Panicln("Poller.UpdateChannel: channel's fd already exist")
		}
		var pollfd = unix.PollFd{
			Fd:      channel.fd_,
			Events:  channel.events_,
			Revents: 0,
		}
		p.pollfds_ = append(p.pollfds_, pollfd)
		idx := len(p.pollfds_) - 1
		channel.SetIndex(idx)
		p.channels_[pollfd.Fd] = channel
	} else {
		//update existing one
		if _, ok := p.channels_[channel.fd_]; !ok {
			log.Panicln("Poller.UpdateChannel: channel's fd doesn't exist")
		}
		if p.channels_[channel.fd_] != channel {
			log.Panicln("Poller.UpdateChannel: the channel corresponding to this fd isn't some one")
		}
		idx := channel.Index()
		if idx < 0 || idx >= len(p.pollfds_) {
			log.Panicln("Poller.UpdateChannel: the index of channel is invalid")
		}
		pfd := &p.pollfds_[idx]
		if pfd.Fd != channel.Fd() && pfd.Fd != -channel.Fd()-1 {
			log.Panicln("Panicln.UpdateChannel: the fd is invalid")
		}
		pfd.Events = channel.events_
		pfd.Revents = channel.revents_
		if channel.IsNoneEvent() {
			//no event , ignore this pollfd
			pfd.Fd = -channel.Fd() - 1
		}
	}
}

func (p *Poller) RemoveChannel(channel *Channel) {
	p.AssertInLoopGoroutine()
	log.Printf("fd = %d\n", channel.Fd())
	if _, ok := p.channels_[channel.Fd()]; !ok {
		log.Panicf("Poller.RemoveChannel:channel not found\n")
	}
	if p.channels_[channel.Fd()] != channel {
		log.Panicf("Poller.RemoveChannel:channel isn't the target channel\n ")
	}
	if !channel.IsNoneEvent() {
		log.Panicf("Poller.RemoveChannel:channel isn't none event\n")
	}
	idx := channel.Index()
	if idx < 0 || idx >= len(p.pollfds_) {
		log.Panicf("Poller.RemoveChannel:channel's index is wrong\n")
	}
	pfd := p.pollfds_[idx]
	if pfd.Fd != -channel.Fd()-1 || pfd.Events != channel.Events() {
		log.Panicf("Poller.RemoveChannel:channel's data is wrong\n")
	}
	delete(p.channels_, channel.Fd())
	if idx == len(p.pollfds_)-1 {
		p.pollfds_ = p.pollfds_[:len(p.pollfds_)-1]
	} else {
		channelAtEnd := p.pollfds_[len(p.pollfds_)-1].Fd
		target := p.pollfds_[idx]
		end := p.pollfds_[len(p.pollfds_)-1]
		//swap the tail and target channel
		p.pollfds_[idx] = end
		p.pollfds_[len(p.pollfds_)-1] = target

		if channelAtEnd < 0 {
			channelAtEnd = -channelAtEnd - 1
		}
		p.channels_[channelAtEnd].SetIndex(idx)

		p.pollfds_ = p.pollfds_[:len(p.pollfds_)-1]
	}

}

func (p *Poller) AssertInLoopGoroutine() {
	p.ownerLoop_.AssertInLoopGoroutine()
}

// *************************
// private:
// *************************

func (p *Poller) fillActiveChannels(numEvents int, activeChannels *[]*Channel) {
	for i := 0; i < len(p.pollfds_) && numEvents > 0; i++ {
		pfd := p.pollfds_[i]
		if pfd.Revents > 0 {
			numEvents--
			v, ok := p.channels_[pfd.Fd]
			if !ok {
				panic("fillActiveChannels: Can't find the target event fd")
			} else {
				if v.fd_ != pfd.Fd {
					panic("fillActiveChannels: channel's fd doesn't equal to pollfd's fd")
				}
				v.SetRevents(pfd.Revents)
				// pfd.Revents = 0
				*activeChannels = append(*activeChannels, v)
			}

		}
	}
}

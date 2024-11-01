package netreactors

import (
	"fmt"
	"log"
	"syscall"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	PrependSize    = 8
	BufferInitSize = 1024
)

//Buffer:
/* +-------------------+------------------+------------------+ */
/* | prependable bytes |  readable bytes  |  writable bytes  | */
/* |                   |     (content)    |                  | */
/* +-------------------+------------------+------------------+ */
/* |                   |                  |                  | */
/* 0      <=      readerindex   <=   writerindex    <=    size */

type Buffer struct {
	buffer_      []byte
	readerIndex_ int
	writerIndex_ int
}

// *************************
// public:
// *************************

func NewBuffer() *Buffer {
	return &Buffer{
		buffer_:      make([]byte, BufferInitSize+PrependSize),
		readerIndex_: PrependSize,
		writerIndex_: PrependSize,
	}
}

func (b *Buffer) ReadFd(fd int, saveErrno *error) int {
	tempBuffer := make([]byte, 65536)
	vec := make([]unix.Iovec, 2)
	writeable := b.WriteableBytes()
	vec[0].Base = b.bytePointer(b.writerIndex_)
	vec[0].Len = uint64(writeable)
	vec[1].Base = &tempBuffer[0]
	vec[1].Len = uint64(len(tempBuffer))
	iovcnt := 2
	if writeable > len(tempBuffer) {
		iovcnt = 1
	}
	n, _, errno := syscall.Syscall(syscall.SYS_READV, uintptr(fd), uintptr(unsafe.Pointer(&vec[0])), uintptr(iovcnt))
	if errno != 0 || n < 0 {
		*saveErrno = errno
	} else if n < uintptr(writeable) {
		b.writerIndex_ += int(n)
	} else {
		fmt.Printf("Buffer.ReadFd:will use tempBuffer,and have use %d bytes\n", len(tempBuffer[:int(n)-writeable]))
		b.buffer_ = append(b.buffer_, tempBuffer[:int(n)-writeable]...)
		b.writerIndex_ += int(n)
	}
	/* if int(n) == writeable+len(tempBuffer) { */
	/* //read again */
	/* } */
	return int(n)
}

func (b *Buffer) Peek() *byte {
	return b.bytePointer(b.readerIndex_)
}

func (b *Buffer) ReadableBytes() int {
	return b.writerIndex_ - b.readerIndex_
}
func (b *Buffer) WriteableBytes() int {
	return len(b.buffer_) - b.writerIndex_
}
func (b *Buffer) PrependBytes() int {
	return b.readerIndex_
}

func (b *Buffer) Retrieve(size int) {
	if size > b.ReadableBytes() {
		log.Panicf("Buffer.Retrieve: size is more than the readable bytes size\n")
	}
	if size < b.ReadableBytes() {
		b.readerIndex_ += size
	} else {
		b.RetrieveAll()
	}
}

func (b *Buffer) RetrieveAll() {
	b.readerIndex_ = PrependSize
	b.writerIndex_ = PrependSize
}

func (b *Buffer) RetrieveAllString() []byte {
	return b.RetrieveAsString(b.ReadableBytes())
}
func (b *Buffer) RetrieveAsString(size int) []byte {
	if size > b.ReadableBytes() {
		log.Panicf("Buffer.RetrieveAsString: size is more than readable bytes size\n")
	}
	ans := b.buffer_[b.readerIndex_ : b.readerIndex_+size]
	b.Retrieve(size)
	return ans
}

// *************************
// private:
// *************************

func (b *Buffer) begin() *byte {
	return &b.buffer_[0]
}

func (b *Buffer) bytePointer(idx int) *byte {
	return &b.buffer_[idx]
}

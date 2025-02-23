package netreactors

import (
	"log"
	"syscall"
	"unsafe"
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

var (
	tempBuffer = make([]byte, 65536)
)

func (b *Buffer) ReadFd(fd int, saveErrno *error) int {
	vec := make([]syscall.Iovec, 2)
	writeable := b.WriteableBytes()
	vec[0].Base = b.bytePointer(b.writerIndex_)
	vec[0].Len = uint64(writeable)
	vec[1].Base = &tempBuffer[0]
	vec[1].Len = uint64(len(tempBuffer))

	//readv
	n, _, errno := syscall.Syscall(syscall.SYS_READV, uintptr(fd), uintptr(unsafe.Pointer(&vec[0])), 2)
	size := int(int64(n))
	if errno != 0 || size < 0 {
		*saveErrno = errno
	} else if size <= writeable {
		b.writerIndex_ += size
	} else {
		Dlog.Printf("Buffer.ReadFd:will use tempBuffer,and have use %d bytes\n", len(tempBuffer[:size-writeable]))
		b.writerIndex_ = len(b.buffer_)
		b.Append(tempBuffer[:size-writeable])
	}
	/* if int(n) == writeable+len(tempBuffer) { */
	/* //read again */
	/* } */
	return size
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

func (b *Buffer) Append(data []byte) {
	// allacte more capacity
	if b.WriteableBytes() < len(data) {
		//(here is already include the use Prepend() or without read anything's case)
		if b.WriteableBytes()+b.readerIndex_-PrependSize >= len(data) {
			// move readable data into front ,don't need to allocate new capacity

			readable := b.ReadableBytes()
			copy(b.buffer_[PrependSize:], b.buffer_[b.readerIndex_:b.writerIndex_])
			b.readerIndex_ = PrependSize
			b.writerIndex_ = readable + PrependSize
		} else {
			// resize
			//FIX: we can move readable data into front first if have free capacity in front
			readable := b.ReadableBytes()
			newbuffer := make([]byte, b.writerIndex_+len(data)+8192) //reserve 8192 byte
			copy(newbuffer[PrependSize:], b.buffer_[b.readerIndex_:b.writerIndex_])
			b.readerIndex_ = PrependSize
			b.writerIndex_ = readable + PrependSize
			b.buffer_ = newbuffer
		}
	}

	//begin append data into buffer_
	copy(b.buffer_[b.writerIndex_:], data)
	b.writerIndex_ += len(data)

}

func (b *Buffer) Prepend(data []byte) {
	if b.PrependBytes() < len(data) {
		log.Panicf("Buffer.Prepend: size is more than the prependable bytes size\n")
	}
	b.readerIndex_ -= len(data)
	copy(b.buffer_[b.readerIndex_:b.readerIndex_+len(data)], data[:])
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

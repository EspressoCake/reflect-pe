package lib

import (
	"syscall"
	. "unsafe"
)

type WinAPI interface {
	Memcopy(start, end, size uintptr)
	VirtualAlloc(size uint) (Pointer, error)
	CstrVal(ptr Pointer) (out []byte)
	LoadLibrary(ptrName uintptr) (Pointer, error)
	GetProcAddress(libraryAddress, ptrName Pointer) (uintptr, error)
	Incr64(src Pointer, val uint64)
	Incr32(src Pointer, val uint32)
	Incr16(src Pointer, val uint16)
	NtFlushInstructionCache(ptr uintptr) error
	CreateThread(ptr Pointer) (uintptr, error)
	WaitForSingleObject(handle uintptr) error
	CloseHandle(handle uintptr)
}

type Win struct {
}

func (w *Win) VirtualAlloc(size uint) (Pointer, error) {
	ret, _, err := virtualAlloc.Call(
		uintptr(0),
		uintptr(size),
		uintptr(0x00001000|0x00002000), // MEM_COMMIT | MEM_RESERVE
		uintptr(0x40))                  // PAGE_EXECUTE_READWRITE

	if err != syscall.Errno(0) {
		return nil, err
	}
	return Pointer(ret), nil
}

func (w *Win) Memcopy(start, end, size uintptr) {
	for i := uintptr(0); i < size; i += Sizeof(uint(0)) {
		*(*uint)(Pointer(end + i)) = *(*uint)(Pointer(start + i))
	}
}

func (w *Win) Incr64(src Pointer, val uint64) {
	*(*uint64)(src) += val
}
func (w *Win) Incr32(src Pointer, val uint32) {
	*(*uint32)(src) += val
}
func (w *Win) Incr16(src Pointer, val uint16) {
	*(*uint16)(src) += val
}

func (w *Win) CstrVal(ptr Pointer) (out []byte) {
	var byteVal byte
	out = make([]byte, 0)
	for i := 0; ; i++ {
		byteVal = *(*byte)(Pointer(ptr))
		if byteVal == 0x00 {
			break
		}
		out = append(out, byteVal)
		ptr = ptrOffset(ptr, 1)
	}
	return out
}

func (w *Win) LoadLibrary(ptrName uintptr) (Pointer, error) {
	ret, _, err := loadLibrary.Call(ptrName)

	if err != syscall.Errno(0) {
		return nil, err
	}
	return Pointer(ret), nil
}

func (w *Win) GetProcAddress(libraryAddress, ptrName Pointer) (uintptr, error) {
	ret, _, err := getProcAddress.Call(
		ptrValue(libraryAddress),
		ptrValue(ptrName))

	if err != syscall.Errno(0) {
		return 0, err
	}
	return ret, nil
}

func (w *Win) NtFlushInstructionCache(ptr uintptr) error {
	_, _, err := ntFlushInstructionCache.Call(
		ptr,
		uintptr(0),
		uintptr(0))

	if err != syscall.Errno(0) {
		return err
	}
	return nil
}

func (w *Win) CreateThread(ptr Pointer) (uintptr, error) {
	ret, _, err := createThread.Call(
		uintptr(0),
		uintptr(0),
		ptrValue(ptr),
		uintptr(0),
		uintptr(0),
		uintptr(0))
	if err != syscall.Errno(0) {
		return 0, err
	}
	return ret, nil
}

func (w *Win) WaitForSingleObject(handle uintptr) error {
	_, _, err := waitForSingleObject.Call(
		handle,
		syscall.INFINITE)
	if err != syscall.Errno(0) {
		return err
	}
	return nil
}

func (w *Win) CloseHandle(handle uintptr) {
	syscall.CloseHandle(syscall.Handle(handle))
}

type ImageImportDescriptor struct {
	OriginalFirstThunk uint32
	TimeDateStamp      uint32
	ForwarderChain     uint32
	Name               uint32
	FirstThunk         uint32
}

type DebugDirectory struct {
	Characteristics  uint32
	TimeDateStamp    uint32
	MajorVersion     uint16
	MinorVersion     uint16
	Type             uint32
	SizeOfData       uint32
	AddressOfRawData uint32
	PointerToRawData uint32
}

type Pogo struct {
	Signature uint32
	Entries   uint32
}

type PogoEntry struct {
	Start_rva uint32
	Size      uint32
	Name      uint8
}
type ImageExportDescriptor struct {
	Characteristics       uint32
	TimeDateStamp         uint32
	MajorVersion          uint16
	MinorVersion          uint16
	Name                  uint32
	Base                  uint32
	NumberOfFunctions     uint32
	NumberOfNames         uint32
	AddressOfFunctions    uint32
	AddressOfName         uint32
	AddressOfNameOrdinals uint32
}

const IMAGE_REL_BASED_HIGH = 0x1
const IMAGE_REL_BASED_LOW = 0x2
const IMAGE_REL_BASED_HIGHLOW = 0x3
const IMAGE_REL_BASED_DIR64 = 0xa
const POGO_TYPE = 0xd

type ImageBaseRelocation struct {
	VirtualAddress uint32
	SizeOfBlock    uint32
}

type ImageReloc struct {
	OffsetType uint16
}

func (c *ImageReloc) GetOffset() uint16 {
	return c.OffsetType & 0x0fff
}

func (c *ImageReloc) GetType() uint16 {
	return (c.OffsetType & 0xf000) >> 12
}

type OriginalImageThunkData struct {
	Ordinal uint
}
type ImageThunkData struct {
	AddressOfData uintptr
}
type ImageImportByName struct {
	Hint uint16
	Name byte
}

var (
	kernel32                = syscall.MustLoadDLL("kernel32.dll")
	ntdll                   = syscall.MustLoadDLL("ntdll.dll")
	virtualAlloc            = kernel32.MustFindProc("VirtualAlloc")
	virtualProtect          = kernel32.MustFindProc("VirtualProtect")
	loadLibrary             = kernel32.MustFindProc("LoadLibraryA")
	getProcAddress          = kernel32.MustFindProc("GetProcAddress")
	createThread            = kernel32.MustFindProc("CreateThread")
	waitForSingleObject     = kernel32.MustFindProc("WaitForSingleObject")
	rtlCopyMemory           = ntdll.MustFindProc("RtlCopyMemory")
	ntFlushInstructionCache = ntdll.MustFindProc("NtFlushInstructionCache")
)
package dsound

import (
	"strconv"
	"syscall"
)

const (
	DS_OK                    = DSRESULT(0x00000000)
	DSERR_ACCESSDENIED       = DSRESULT(0x80070005)
	DSERR_ALLOCATED          = DSRESULT(0x8878000A)
	DSERR_ALREADYINITIALIZED = DSRESULT(0x88780082)
	DSERR_BADFORMAT          = DSRESULT(0x88780064)
	DSERR_BUFFERLOST         = DSRESULT(0x88780096)
	DSERR_CONTROLUNAVAIL     = DSRESULT(0x8878001E)
	DSERR_GENERIC            = DSRESULT(0x80004005)
	DSERR_INVALIDCALL        = DSRESULT(0x88780032)
	DSERR_INVALIDPARAM       = DSRESULT(0x80070057)
	DSERR_NOAGGREGATION      = DSRESULT(0x80040110)
	DSERR_NODRIVER           = DSRESULT(0x88780078)
	DSERR_NOINTERFACE        = DSRESULT(0x80000004)
	DSERR_OTHERAPPHASPRIO    = DSRESULT(0x887800A0)
	DSERR_OUTOFMEMORY        = DSRESULT(0x8007000E)
	DSERR_PRIOLEVELNEEDED    = DSRESULT(0x88780046)
	DSERR_UNINITIALIZED      = DSRESULT(0x887800AA)
	DSERR_UNSUPPORTED        = DSRESULT(0x80004001)
)

type DSRESULT uintptr

func (r DSRESULT) Error() string {
	switch r {
	case DS_OK:
		return "DS_OK: The operation completed successfully."
	case DSERR_ACCESSDENIED:
		return "DSERR_ACCESSDENIED: Access is denied."
	case DSERR_ALLOCATED:
		return "DSERR_ALLOCATED: The call failed because resources (such as a priority level) were already being used by another caller."
	case DSERR_ALREADYINITIALIZED:
		return "DSERR_ALREADYINITIALIZED: This object is already initialized"
	case DSERR_BADFORMAT:
		return "DSERR_BADFORMAT: The specified WAVE format is not supported."
	case DSERR_BUFFERLOST:
		return "DSERR_BUFFERLOST: The buffer memory has been lost, and must be restored."
	case DSERR_CONTROLUNAVAIL:
		return "DSERR_CONTROLUNAVAIL: The control (vol,pan,etc.) requested by the caller is not available."
	case DSERR_GENERIC:
		return "DSERR_GENERIC: An undetermined error occured inside the DirectSound subsystem."
	case DSERR_INVALIDCALL:
		return "DSERR_INVALIDCALL: This call is not valid for the current state of this object."
	case DSERR_INVALIDPARAM:
		return "DSERR_INVALIDPARAM: An invalid parameter was passed to the returning function."
	case DSERR_NOAGGREGATION:
		return "DSERR_NOAGGREGATION: This object does not support aggregation."
	case DSERR_NODRIVER:
		return "DSERR_NODRIVER: No sound driver is available for use."
	case DSERR_NOINTERFACE:
		return "DSERR_NOINTERFACE: The requested COM interface is not available."
	case DSERR_OTHERAPPHASPRIO:
		return "DSERR_OTHERAPPHASPRIO: Another app has a higher priority level, preventing this call from succeeding."
	case DSERR_OUTOFMEMORY:
		return "DSERR_OUTOFMEMORY: Not enough free memory is available to complete the operation."
	case DSERR_PRIOLEVELNEEDED:
		return "DSERR_PRIOLEVELNEEDED: The caller does not have the priority level required for the function to succeed."
	case DSERR_UNINITIALIZED:
		return "DSERR_UNINITIALIZED: This object has not been initialized"
	case DSERR_UNSUPPORTED:
		return "DSERR_UNSUPPORTED: The function called is not supported at this time."
	}
	return "0x" + strconv.FormatUint(uint64(r), 16) + ": An unknown error occurred."
}

func dsResult(r1, r2 uintptr, err syscall.Errno) error {
	if r1 != 0 {
		return DSRESULT(r1)
	}
	return nil
}

func dllDSResult(r1, r2 uintptr, lastErr error) error {
	if lastErr.(syscall.Errno) != 0 {
		return lastErr
	}
	if r1 != 0 {
		return DSRESULT(r1)
	}
	return nil
}

type comProc uintptr

func (p comProc) Call(a ...uintptr) (r1, r2 uintptr, lastErr syscall.Errno) {
	switch len(a) {
	case 0:
		return syscall.Syscall(uintptr(p), uintptr(len(a)), 0, 0, 0)
	case 1:
		return syscall.Syscall(uintptr(p), uintptr(len(a)), a[0], 0, 0)
	case 2:
		return syscall.Syscall(uintptr(p), uintptr(len(a)), a[0], a[1], 0)
	case 3:
		return syscall.Syscall(uintptr(p), uintptr(len(a)), a[0], a[1], a[2])
	case 4:
		return syscall.Syscall6(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], 0, 0)
	case 5:
		return syscall.Syscall6(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], 0)
	case 6:
		return syscall.Syscall6(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5])
	case 7:
		return syscall.Syscall9(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5], a[6], 0, 0)
	case 8:
		return syscall.Syscall9(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], 0)
	case 9:
		return syscall.Syscall9(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8])
	case 10:
		return syscall.Syscall12(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], 0, 0)
	case 11:
		return syscall.Syscall12(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], 0)
	case 12:
		return syscall.Syscall12(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11])
	case 13:
		return syscall.Syscall15(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11], a[12], 0, 0)
	case 14:
		return syscall.Syscall15(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11], a[12], a[13], 0)
	case 15:
		return syscall.Syscall15(uintptr(p), uintptr(len(a)), a[0], a[1], a[2], a[3], a[4], a[5], a[6], a[7], a[8], a[9], a[10], a[11], a[12], a[13], a[14])
	}
	panic("too many arguments")
}
